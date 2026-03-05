package saml

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"terralist/pkg/auth"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/rs/zerolog/log"
)

// Provider is the concrete implementation of auth.Provider for SAML authentication.
type Provider struct {
	IdPMetadataURL             string
	IdPMetadataFile            string
	IdPEntityID                string
	IdPSSOURL                  string
	IdPSSOCertificate          string
	SPEntityID                 string
	ACSUrl                     string
	MetadataURL                string
	NameAttribute              string
	EmailAttribute             string
	GroupsAttribute            string
	CertFile                   string
	KeyFile                    string
	PrivateKeySecret           string
	TerralistSchemeHostAndPort string

	// Configurable SAML timing and behavior constants
	HTTPClientTimeout            time.Duration
	AssertionClockSkew           time.Duration
	RequestIDExpiration          time.Duration
	RequestIDCleanupInterval     time.Duration
	MetadataRefreshInterval      time.Duration
	MetadataRefreshCheckInterval time.Duration
	MaxAssertionAge              time.Duration
	AllowIdPInitiated            bool
	DisableRequestIDValidation   bool

	// Internal state: SAML components
	idpMetadata     *saml.EntityDescriptor
	spCertificate   *x509.Certificate
	spPrivateKey    *rsa.PrivateKey
	serviceProvider *saml.ServiceProvider
	requestTracker  *requestTracker

	// Metadata refresh state
	metadataLastRefresh time.Time
	metadataMutex       sync.RWMutex
	stopMetadataRefresh chan struct{}
}

const (
	// relayStateMaxSize is the maximum size for RelayState in bytes
	// SAML 2.0 specification recommends max 80 bytes for RelayState, but
	// many implementations including Google Workspace support larger values.
	// We use 512 bytes to accommodate OAuth payloads while preventing abuse.
	relayStateMaxSize = 512
)

// newHTTPClient creates an HTTP client with configurable timeout.
func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}

// requestTracker tracks SAML AuthnRequest IDs to prevent replay attacks.
// Request IDs are stored with their creation timestamp and automatically
// cleaned up after expiration.
type requestTracker struct {
	requestIDs  map[string]time.Time
	mutex       sync.RWMutex
	stopCleanup chan struct{}
}

// newRequestTracker creates a new request tracker with configurable cleanup interval.
func newRequestTracker(cleanupInterval time.Duration) *requestTracker {
	rt := &requestTracker{
		requestIDs:  make(map[string]time.Time),
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup goroutine with configurable interval
	go rt.cleanupExpired(cleanupInterval)

	return rt
}

// Track adds a request ID to the tracker with the current timestamp.
func (rt *requestTracker) Track(requestID string) {
	if requestID == "" {
		return
	}

	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	rt.requestIDs[requestID] = time.Now()
}

// ValidateAndConsume checks if a request ID exists and hasn't expired,
// then removes it to prevent reuse (replay attack prevention).
// Returns true if the request ID is valid, false otherwise.
func (rt *requestTracker) ValidateAndConsume(requestID string, expiration time.Duration) bool {
	if requestID == "" {
		return false
	}

	rt.mutex.Lock()
	defer rt.mutex.Unlock()

	createdAt, exists := rt.requestIDs[requestID]
	if !exists {
		return false
	}

	// Check if request ID has expired
	if time.Since(createdAt) > expiration {
		delete(rt.requestIDs, requestID)
		return false
	}

	// Consume the request ID (remove it) to prevent replay
	delete(rt.requestIDs, requestID)
	return true
}

// cleanupExpired periodically removes expired request IDs.
func (rt *requestTracker) cleanupExpired(cleanupInterval time.Duration) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rt.mutex.Lock()
			now := time.Now()
			for id, createdAt := range rt.requestIDs {
				if now.Sub(createdAt) > cleanupInterval*4 { // Cleanup items older than 4x cleanup interval
					delete(rt.requestIDs, id)
				}
			}
			rt.mutex.Unlock()
		case <-rt.stopCleanup:
			return
		}
	}
}

// Stop stops the cleanup goroutine. Should be called when the provider is being destroyed.
func (rt *requestTracker) Stop() {
	close(rt.stopCleanup)
}

func (p *Provider) Name() string {
	return "SAML"
}

// GetAuthorizeUrl initiates the SAML SSO flow by creating a SAML AuthnRequest
// and redirecting to the IdP's SSO endpoint.
// The state parameter is used to maintain the OAuth flow state for Terraform compatibility.
func (p *Provider) GetAuthorizeUrl(state string) string {
	// Ensure IdP metadata is loaded
	metadata := p.getIdPMetadata()
	if metadata == nil {
		if err := p.loadIdPMetadata(); err != nil {
			log.Error().
				AnErr("error", SanitizeErrorForLogging(err)).
				Str("source", "metadata_load").
				Msg("SAML authentication request failed: failed to load IdP metadata")
			// Return error URL with sanitized message
			return fmt.Sprintf("%s?error=metadata_load_failed&error_description=%s", p.ACSUrl, url.QueryEscape(SanitizeErrorForURL(err)))
		}
		metadata = p.getIdPMetadata()
	}

	// Ensure service provider is initialized
	if p.serviceProvider == nil {
		if err := p.initializeServiceProvider(); err != nil {
			log.Error().
				AnErr("error", SanitizeErrorForLogging(err)).
				Str("source", "sp_init").
				Msg("SAML authentication request failed: failed to initialize service provider")
			return fmt.Sprintf("%s?error=sp_init_failed&error_description=%s", p.ACSUrl, url.QueryEscape(SanitizeErrorForURL(err)))
		}
	}

	// Get SSO URL from IdP metadata
	idpSSOURL := p.getSSOURL()
	if idpSSOURL == "" {
		log.Error().
			Str("source", "sso_url").
			Str("entity_id", metadata.EntityID).
			Msg("SAML authentication request failed: no SSO URL found in IdP metadata")
		return fmt.Sprintf("%s?error=no_sso_url&error_description=%s", p.ACSUrl, url.QueryEscape("no SSO URL found in IdP metadata"))
	}

	// Create SAML AuthnRequest using the library
	authnRequest, err := p.serviceProvider.MakeAuthenticationRequest(
		idpSSOURL,
		saml.HTTPRedirectBinding,
		saml.HTTPPostBinding,
	)
	if err != nil {
		log.Error().
			AnErr("error", SanitizeErrorForLogging(err)).
			Str("source", "authn_request").
			Str("idp_sso_url", idpSSOURL).
			Msg("SAML authentication request failed: failed to create AuthnRequest")
		return fmt.Sprintf("%s?error=authn_request_failed&error_description=%s", p.ACSUrl, url.QueryEscape(SanitizeErrorForURL(err)))
	}

	// Track the request ID to prevent replay attacks
	if authnRequest.ID != "" {
		p.requestTracker.Track(authnRequest.ID)
	}

	// Validate RelayState before passing to SAML AuthnRequest (CSRF protection)
	if err := validateRelayState(state); err != nil {
		log.Error().
			AnErr("error", SanitizeErrorForLogging(err)).
			Str("source", "relaystate_validation").
			Str("request_id", authnRequest.ID).
			Int("relaystate_size", len(state)).
			Msg("SAML authentication request failed: invalid RelayState")
		return fmt.Sprintf("%s?error=invalid_relaystate&error_description=%s", p.ACSUrl, url.QueryEscape(SanitizeErrorForURL(err)))
	}

	// Get the redirect URL with SAMLRequest and RelayState (state)
	redirectURL, err := authnRequest.Redirect(state, p.serviceProvider)
	if err != nil {
		log.Error().
			AnErr("error", SanitizeErrorForLogging(err)).
			Str("source", "redirect").
			Str("request_id", authnRequest.ID).
			Msg("SAML authentication request failed: failed to create redirect URL")
		return fmt.Sprintf("%s?error=redirect_failed&error_description=%s", p.ACSUrl, url.QueryEscape(SanitizeErrorForURL(err)))
	}

	// Log successful authentication request initiation
	log.Info().
		Str("request_id", authnRequest.ID).
		Str("idp_entity_id", metadata.EntityID).
		Str("idp_sso_url", idpSSOURL).
		Str("sp_entity_id", p.SPEntityID).
		Time("issue_instant", authnRequest.IssueInstant).
		Msg("SAML authentication request initiated")

	return redirectURL.String()
}

// GetUserDetails parses the SAMLResponse (provided as base64 encoded XML in the code parameter)
// and extracts user attributes to populate the user struct.
func (p *Provider) GetUserDetails(samlResponse string, user *auth.User) error {
	// Decode base64 SAMLResponse
	samlResponseXML, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		log.Error().
			AnErr("error", SanitizeErrorForLogging(err)).
			Str("source", "base64_decode").
			Msg("SAML authentication failed: failed to decode SAML response")
		sanitizedMsg, _ := SanitizeError(err)
		return fmt.Errorf("%s", sanitizedMsg)
	}

	// Ensure service provider is initialized for response validation
	if p.serviceProvider == nil {
		if err := p.initializeServiceProvider(); err != nil {
			log.Error().
				AnErr("error", SanitizeErrorForLogging(err)).
				Str("source", "sp_init").
				Msg("SAML authentication failed: failed to initialize service provider")
			sanitizedMsg, _ := SanitizeError(err)
			return fmt.Errorf("%s", sanitizedMsg)
		}
	}

	// Parse ACS URL for validation
	acsURL, err := url.Parse(p.ACSUrl)
	if err != nil {
		log.Error().
			AnErr("error", SanitizeErrorForLogging(err)).
			Str("source", "acs_url_parse").
			Str("acs_url", p.ACSUrl).
			Msg("SAML authentication failed: invalid ACS URL")
		sanitizedMsg, _ := SanitizeError(err)
		return fmt.Errorf("%s", sanitizedMsg)
	}

	// First, parse the response to extract the request ID (InResponseTo)
	// We need to do a preliminary parse to get the InResponseTo value
	var response saml.Response
	if err := xml.Unmarshal(samlResponseXML, &response); err != nil {
		log.Error().
			AnErr("error", SanitizeErrorForLogging(err)).
			Str("source", "xml_parse").
			Msg("SAML authentication failed: failed to parse SAML response XML")
		sanitizedMsg, _ := SanitizeError(err)
		return fmt.Errorf("%s", sanitizedMsg)
	}

	// Extract request ID from the response (InResponseTo field is directly on Response)
	requestID := response.InResponseTo

	// Validate request ID exists and hasn't been used (replay attack prevention)
	// Note: InResponseTo is optional in SAML 2.0 but should always be present for SP-initiated flows.
	// If missing, we skip validation (less secure but allows compatibility with some IdPs).
	// In production, consider requiring InResponseTo for all responses.
	if !p.DisableRequestIDValidation && requestID != "" {
		if !p.requestTracker.ValidateAndConsume(requestID, p.RequestIDExpiration) {
			log.Warn().
				Str("request_id", requestID).
				Str("source", "request_id_validation").
				Msg("SAML authentication failed: invalid or replayed request ID")
			return fmt.Errorf("authentication failed: invalid or expired response")
		}
	} else if !p.DisableRequestIDValidation && requestID == "" {
		log.Warn().
			Str("source", "request_id_missing").
			Msg("SAML authentication warning: InResponseTo missing from SAML response")
	} else if p.DisableRequestIDValidation {
		log.Debug().
			Str("request_id", requestID).
			Msg("SAML request ID validation disabled")
	}

	// Build list of valid request IDs for library validation
	// If we have a request ID, pass it; otherwise pass empty slice
	possibleRequestIDs := []string{}
	if requestID != "" {
		possibleRequestIDs = []string{requestID}
	}

	// Parse and validate SAML Response using the library
	assertion, err := p.serviceProvider.ParseXMLResponse(samlResponseXML, possibleRequestIDs, *acsURL)
	if err != nil {
		log.Error().
			AnErr("error", SanitizeErrorForLogging(err)).
			Str("request_id", requestID).
			Str("source", "response_validation").
			Str("issuer", response.Issuer.Value).
			Msg("SAML authentication failed: failed to parse/validate SAML response")
		sanitizedMsg, _ := SanitizeError(err)
		return fmt.Errorf("%s", sanitizedMsg)
	}

	// Validate assertion expiration (NotBefore/NotOnOrAfter/IssueInstant) with clock skew tolerance
	if err := p.validateAssertionExpiration(assertion); err != nil {
		log.Error().
			AnErr("error", SanitizeErrorForLogging(err)).
			Str("request_id", requestID).
			Str("assertion_id", assertion.ID).
			Str("source", "assertion_expiration").
			Str("issuer", assertion.Issuer.Value).
			Msg("SAML authentication failed: assertion expiration validation failed")
		sanitizedMsg, _ := SanitizeError(err)
		return fmt.Errorf("%s", sanitizedMsg)
	}

	// Extract attributes from the assertion
	attributes := make(map[string][]string)
	if len(assertion.AttributeStatements) > 0 {
		attrStatement := assertion.AttributeStatements[0]
		for _, attr := range attrStatement.Attributes {
			// Get the attribute name (handle both Name and FriendlyName)
			attrName := attr.Name
			if attrName == "" {
				attrName = attr.FriendlyName
			}

			// Extract attribute values
			var values []string
			for _, val := range attr.Values {
				values = append(values, val.Value)
			}
			attributes[attrName] = values
		}
	}

	// Debug log all extracted attributes
	log.Debug().
		Interface("saml_attributes", attributes).
		Int("attribute_count", len(attributes)).
		Msg("Extracted SAML attributes from assertion")

	// Map attributes to user struct
	if err := p.mapAttributesToUser(attributes, user); err != nil {
		log.Error().
			AnErr("error", SanitizeErrorForLogging(err)).
			Str("request_id", requestID).
			Str("assertion_id", assertion.ID).
			Str("source", "attribute_mapping").
			Str("subject", assertion.Subject.NameID.Value).
			Str("issuer", assertion.Issuer.Value).
			Msg("SAML authentication failed: failed to map attributes")
		// Attribute errors are safe to expose (e.g., "name attribute not found")
		return err
	}

	// Debug log the final user groups
	log.Debug().
		Strs("user_groups", user.Groups).
		Int("groups_count", len(user.Groups)).
		Msg("Final user groups after SAML authentication")

	// Log successful authentication
	var notBefore, notOnOrAfter time.Time
	if assertion.Conditions != nil {
		notBefore = assertion.Conditions.NotBefore
		notOnOrAfter = assertion.Conditions.NotOnOrAfter
	}

	log.Info().
		Str("request_id", requestID).
		Str("assertion_id", assertion.ID).
		Str("subject", assertion.Subject.NameID.Value).
		Str("issuer", assertion.Issuer.Value).
		Str("user_email", user.Email).
		Str("user_name", user.Name).
		Time("issue_instant", assertion.IssueInstant).
		Time("not_before", notBefore).
		Time("not_on_or_after", notOnOrAfter).
		Msg("SAML authentication successful")

	return nil
}

// validateAssertionExpiration validates that the SAML assertion is within its valid time window.
// It checks NotBefore and NotOnOrAfter conditions with clock skew tolerance.
// This provides explicit validation even though the library may also perform checks.
func (p *Provider) validateAssertionExpiration(assertion *saml.Assertion) error {
	if assertion == nil {
		return fmt.Errorf("assertion is nil")
	}

	now := time.Now()

	// Check Conditions if present
	if assertion.Conditions != nil {
		conditions := assertion.Conditions

		// Validate NotBefore: assertion should not be used before this time
		// With clock skew tolerance, we allow assertions that are slightly in the future
		if !conditions.NotBefore.IsZero() {
			notBeforeWithSkew := conditions.NotBefore.Add(-p.AssertionClockSkew)
			if now.Before(notBeforeWithSkew) {
				return fmt.Errorf("assertion is not yet valid: NotBefore=%v, current time=%v (with %v clock skew tolerance)",
					conditions.NotBefore, now, p.AssertionClockSkew)
			}
		}

		// Validate NotOnOrAfter: assertion should not be used on or after this time
		// With clock skew tolerance, we allow assertions that are slightly expired
		if !conditions.NotOnOrAfter.IsZero() {
			notOnOrAfterWithSkew := conditions.NotOnOrAfter.Add(p.AssertionClockSkew)
			if now.After(notOnOrAfterWithSkew) {
				return fmt.Errorf("assertion has expired: NotOnOrAfter=%v, current time=%v (with %v clock skew tolerance)",
					conditions.NotOnOrAfter, now, p.AssertionClockSkew)
			}
		}
	}

	// Also validate IssueInstant to ensure assertion is not too old
	// This is a secondary check - assertions older than MaxAssertionAge are suspicious
	if !assertion.IssueInstant.IsZero() {
		age := now.Sub(assertion.IssueInstant)
		if age > p.MaxAssertionAge {
			return fmt.Errorf("assertion is too old: IssueInstant=%v, age=%v (max allowed=%v)",
				assertion.IssueInstant, age, p.MaxAssertionAge)
		}
		// Also check if assertion is from the future (beyond clock skew)
		if age < -p.AssertionClockSkew {
			return fmt.Errorf("assertion is from the future: IssueInstant=%v, current time=%v (with %v clock skew tolerance)",
				assertion.IssueInstant, now, p.AssertionClockSkew)
		}
	}

	return nil
}

// extractIdPCertificatesFromMetadata extracts IdP SSO certificates from metadata.
// Returns all signing certificates found in the metadata.
func (p *Provider) extractIdPCertificatesFromMetadata() ([]*x509.Certificate, error) {
	metadata := p.getIdPMetadata()
	if metadata == nil {
		return nil, fmt.Errorf("metadata not loaded")
	}

	if len(metadata.IDPSSODescriptors) == 0 {
		return nil, nil
	}

	idpSSODescriptor := metadata.IDPSSODescriptors[0]
	var certificates []*x509.Certificate

	for _, keyDesc := range idpSSODescriptor.KeyDescriptors {
		if keyDesc.Use == "signing" || keyDesc.Use == "" {
			for _, x509Cert := range keyDesc.KeyInfo.X509Data.X509Certificates {
				// Decode base64 certificate
				certData, err := base64.StdEncoding.DecodeString(x509Cert.Data)
				if err != nil {
					continue
				}

				// Parse certificate
				cert, err := x509.ParseCertificate(certData)
				if err != nil {
					continue
				}

				certificates = append(certificates, cert)
			}
		}
	}

	return certificates, nil
}

// loadIdPMetadata loads and parses the IdP metadata from URL, file, or constructs it from direct config.
func (p *Provider) loadIdPMetadata() error {
	if p.IdPMetadataURL != "" {
		// Fetch from URL using the library
		metadataURL, err := url.Parse(p.IdPMetadataURL)
		if err != nil {
			log.Error().
				AnErr("error", err).
				Str("metadata_url", p.IdPMetadataURL).
				Str("source", "metadata_url_parse").
				Msg("SAML metadata load failed: invalid IdP metadata URL")
			return fmt.Errorf("invalid IdP metadata URL: %w", err)
		}

		log.Info().
			Str("metadata_url", p.IdPMetadataURL).
			Msg("Fetching IdP metadata from URL")

		httpClient := newHTTPClient(p.HTTPClientTimeout)
		metadata, err := samlsp.FetchMetadata(context.Background(), httpClient, *metadataURL)
		if err != nil {
			log.Error().
				AnErr("error", err).
				Str("metadata_url", p.IdPMetadataURL).
				Str("source", "metadata_fetch").
				Msg("SAML metadata load failed: failed to fetch IdP metadata from URL")
			return fmt.Errorf("failed to fetch IdP metadata from URL: %w", err)
		}

		p.metadataMutex.Lock()
		p.idpMetadata = metadata
		p.metadataLastRefresh = time.Now()
		p.metadataMutex.Unlock()

		log.Info().
			Str("entity_id", metadata.EntityID).
			Str("metadata_url", p.IdPMetadataURL).
			Time("valid_until", metadata.ValidUntil).
			Dur("cache_duration", metadata.CacheDuration).
			Msg("IdP metadata loaded successfully from URL")

		// Start metadata refresh goroutine for URL-based metadata
		p.startMetadataRefresh()
	} else if p.IdPMetadataFile != "" {
		// Read from file
		log.Info().
			Str("metadata_file", p.IdPMetadataFile).
			Msg("Loading IdP metadata from file")

		file, err := os.Open(p.IdPMetadataFile)
		if err != nil {
			log.Error().
				AnErr("error", err).
				Str("metadata_file", p.IdPMetadataFile).
				Str("source", "file_open").
				Msg("SAML metadata load failed: failed to open IdP metadata file")
			return fmt.Errorf("failed to open IdP metadata file: %w", err)
		}
		defer file.Close()

		// Parse metadata XML using xml.Decoder
		metadata := &saml.EntityDescriptor{}
		if err := xml.NewDecoder(file).Decode(metadata); err != nil {
			log.Error().
				AnErr("error", err).
				Str("metadata_file", p.IdPMetadataFile).
				Str("source", "xml_parse").
				Msg("SAML metadata load failed: failed to parse IdP metadata")
			return fmt.Errorf("failed to parse IdP metadata: %w", err)
		}

		p.metadataMutex.Lock()
		p.idpMetadata = metadata
		p.metadataLastRefresh = time.Now()
		p.metadataMutex.Unlock()

		log.Info().
			Str("entity_id", metadata.EntityID).
			Str("metadata_file", p.IdPMetadataFile).
			Time("valid_until", metadata.ValidUntil).
			Dur("cache_duration", metadata.CacheDuration).
			Msg("IdP metadata loaded successfully from file")
	} else if p.IdPEntityID != "" && p.IdPSSOURL != "" {
		// Construct metadata from direct IdP configuration
		idpSSODescriptor := saml.IDPSSODescriptor{
			SingleSignOnServices: []saml.Endpoint{
				{
					Binding:  saml.HTTPRedirectBinding,
					Location: p.IdPSSOURL,
				},
				{
					Binding:  saml.HTTPPostBinding,
					Location: p.IdPSSOURL,
				},
			},
		}

		// Add certificate if provided
		if p.IdPSSOCertificate != "" {
			// Parse PEM certificate
			block, _ := pem.Decode([]byte(p.IdPSSOCertificate))
			if block == nil {
				return fmt.Errorf("failed to parse IdP SSO certificate PEM")
			}

			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse IdP SSO certificate: %w", err)
			}

			// Encode certificate as base64 for metadata
			certDER := cert.Raw
			certBase64 := base64.StdEncoding.EncodeToString(certDER)

			// Add KeyDescriptor with certificate
			idpSSODescriptor.KeyDescriptors = []saml.KeyDescriptor{
				{
					Use: "signing",
					KeyInfo: saml.KeyInfo{
						X509Data: saml.X509Data{
							X509Certificates: []saml.X509Certificate{
								{
									Data: certBase64,
								},
							},
						},
					},
				},
			}
		}

		p.metadataMutex.Lock()
		p.idpMetadata = &saml.EntityDescriptor{
			EntityID:          p.IdPEntityID,
			IDPSSODescriptors: []saml.IDPSSODescriptor{idpSSODescriptor},
		}
		p.metadataLastRefresh = time.Now()
		p.metadataMutex.Unlock()

		log.Info().
			Str("entity_id", p.IdPEntityID).
			Str("sso_url", p.IdPSSOURL).
			Str("source", "direct_config").
			Msg("IdP metadata constructed from direct configuration")
	} else {
		return fmt.Errorf("no IdP metadata source provided")
	}

	metadata := p.getIdPMetadata()
	if metadata == nil {
		return fmt.Errorf("failed to load IdP metadata")
	}

	// Validate that certificate is present (either from metadata or provided)
	// Only check if we loaded from metadata (not direct config, as direct config handles it)
	if p.IdPMetadataURL != "" || p.IdPMetadataFile != "" {
		certificates, err := p.extractIdPCertificatesFromMetadata()
		if err == nil && len(certificates) == 0 {
			// No certificate found in metadata, check if provided via config
			if p.IdPSSOCertificate == "" {
				return fmt.Errorf("IdP SSO certificate is required but was not found in metadata and TERRALIST_SAML_IDP_SSO_CERTIFICATE was not provided")
			}
		}
	} else if p.IdPEntityID != "" && p.IdPSSOURL != "" {
		// Using direct config - certificate is required if not already added
		if p.IdPSSOCertificate == "" {
			return fmt.Errorf("IdP SSO certificate is required when using direct IdP configuration (IdPEntityID/IdPSSOURL) - provide TERRALIST_SAML_IDP_SSO_CERTIFICATE")
		}
	}

	return nil
}

// startMetadataRefresh starts a background goroutine to periodically refresh IdP metadata.
// Only starts if metadata is loaded from URL (not file or direct config).
func (p *Provider) startMetadataRefresh() {
	// Only refresh metadata loaded from URL
	if p.IdPMetadataURL == "" {
		return
	}

	// Initialize stop channel if not already initialized
	if p.stopMetadataRefresh == nil {
		p.stopMetadataRefresh = make(chan struct{})
	}

	go p.metadataRefreshLoop()
}

// metadataRefreshLoop periodically checks and refreshes IdP metadata.
func (p *Provider) metadataRefreshLoop() {
	ticker := time.NewTicker(p.MetadataRefreshCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if metadata needs refresh
			if p.shouldRefreshMetadata() {
				if err := p.refreshMetadata(); err != nil {
					// Log error but continue - don't crash on refresh failure
					// Metadata will be refreshed on next check
					log.Warn().
						AnErr("error", err).
						Msg("Failed to refresh IdP metadata, will retry on next check")
				} else {
					log.Info().
						Msg("Successfully refreshed IdP metadata")
				}
			}
		case <-p.stopMetadataRefresh:
			return
		}
	}
}

// shouldRefreshMetadata checks if metadata should be refreshed based on:
// - Time since last refresh
// - Metadata ValidUntil attribute
// - Metadata CacheDuration attribute.
func (p *Provider) shouldRefreshMetadata() bool {
	p.metadataMutex.RLock()
	defer p.metadataMutex.RUnlock()

	if p.idpMetadata == nil {
		return false
	}

	now := time.Now()

	// Check if metadata has expired (ValidUntil)
	if !p.idpMetadata.ValidUntil.IsZero() {
		// Refresh if metadata expires within the next check interval
		if p.idpMetadata.ValidUntil.Before(now.Add(p.MetadataRefreshCheckInterval)) {
			return true
		}
	}

	// Check CacheDuration if specified
	if p.idpMetadata.CacheDuration > 0 {
		refreshTime := p.metadataLastRefresh.Add(p.idpMetadata.CacheDuration)
		if refreshTime.IsZero() || time.Since(p.metadataLastRefresh) >= p.MetadataRefreshCheckInterval {
			// Check if metadata is expired or we need to check for updates
			return true
		}
		if now.After(refreshTime) {
			return true
		}
	}

	// Default: refresh every metadataRefreshInterval
	if time.Since(p.metadataLastRefresh) >= p.MetadataRefreshInterval {
		return true
	}

	return false
}

// refreshMetadata fetches and updates IdP metadata from the configured URL.
func (p *Provider) refreshMetadata() error {
	if p.IdPMetadataURL == "" {
		return fmt.Errorf("cannot refresh metadata: no metadata URL configured")
	}

	metadataURL, err := url.Parse(p.IdPMetadataURL)
	if err != nil {
		return fmt.Errorf("invalid IdP metadata URL: %w", err)
	}

	// Fetch new metadata
	httpClient := newHTTPClient(p.HTTPClientTimeout)
	newMetadata, err := samlsp.FetchMetadata(context.Background(), httpClient, *metadataURL)
	if err != nil {
		return fmt.Errorf("failed to fetch IdP metadata from URL: %w", err)
	}

	// Validate new metadata has required components
	if len(newMetadata.IDPSSODescriptors) == 0 {
		return fmt.Errorf("refreshed metadata has no IDPSSODescriptors")
	}

	// Validate metadata hasn't expired
	now := time.Now()
	if !newMetadata.ValidUntil.IsZero() && newMetadata.ValidUntil.Before(now) {
		return fmt.Errorf("refreshed metadata has already expired: ValidUntil=%v", newMetadata.ValidUntil)
	}

	// Update metadata atomically
	p.metadataMutex.Lock()
	oldMetadata := p.idpMetadata
	p.idpMetadata = newMetadata
	p.metadataLastRefresh = time.Now()
	p.metadataMutex.Unlock()

	// Reinitialize service provider with new metadata
	// This ensures certificates and SSO URLs are updated
	if p.serviceProvider != nil {
		if err := p.initializeServiceProvider(); err != nil {
			// Rollback on failure
			p.metadataMutex.Lock()
			p.idpMetadata = oldMetadata
			p.metadataMutex.Unlock()
			return fmt.Errorf("failed to reinitialize service provider with new metadata: %w", err)
		}
	}

	log.Info().
		Str("entity_id", newMetadata.EntityID).
		Time("valid_until", newMetadata.ValidUntil).
		Msg("IdP metadata refreshed successfully")

	return nil
}

// getIdPMetadata safely returns a copy of the IdP metadata.
// Callers should not modify the returned metadata.
func (p *Provider) getIdPMetadata() *saml.EntityDescriptor {
	p.metadataMutex.RLock()
	defer p.metadataMutex.RUnlock()
	return p.idpMetadata
}

// initializeServiceProvider initializes the SAML Service Provider using the library.
func (p *Provider) initializeServiceProvider() error {
	// Load SP certificate and key if provided
	if p.CertFile != "" && p.KeyFile != "" {
		// Load certificate
		certData, err := os.ReadFile(p.CertFile)
		if err != nil {
			return fmt.Errorf("failed to read certificate file: %w", err)
		}

		block, _ := pem.Decode(certData)
		if block == nil {
			return fmt.Errorf("failed to parse certificate PEM")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %w", err)
		}
		p.spCertificate = cert

		// Load and decrypt private key
		keyData, err := os.ReadFile(p.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to read private key file: %w", err)
		}

		block, _ = pem.Decode(keyData)
		if block == nil {
			return fmt.Errorf("failed to parse private key PEM")
		}

		// Handle encrypted PEM blocks (legacy support)
		// Note: x509.DecryptPEMBlock is deprecated but still needed for legacy compatibility
		var keyDER []byte
		if p.PrivateKeySecret != "" {
			// If passphrase provided, try to decrypt (handles both encrypted and unencrypted)
			//nolint:staticcheck // SA1019: x509.DecryptPEMBlock is deprecated but required for legacy PEM encryption compatibility
			decrypted, err := x509.DecryptPEMBlock(block, []byte(p.PrivateKeySecret))
			if err != nil {
				return fmt.Errorf("failed to decrypt private key with provided passphrase: %w", err)
			}
			keyDER = decrypted
		} else {
			// No passphrase - assume unencrypted
			keyDER = block.Bytes
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(keyDER)
		if err != nil {
			// Try PKCS8 format if PKCS1 fails
			key, err := x509.ParsePKCS8PrivateKey(keyDER)
			if err != nil {
				return fmt.Errorf("failed to parse private key: %w", err)
			}
			var ok bool
			privateKey, ok = key.(*rsa.PrivateKey)
			if !ok {
				return fmt.Errorf("private key is not an RSA key")
			}
		}
		p.spPrivateKey = privateKey
	} else {
		// Generate a self-signed certificate if not provided
		// This is for development/testing - production should use proper certificates
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate private key: %w", err)
		}

		template := x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject: pkix.Name{
				CommonName: p.SPEntityID,
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().Add(365 * 24 * time.Hour),
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			return fmt.Errorf("failed to create certificate: %w", err)
		}

		cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %w", err)
		}

		p.spCertificate = cert
		p.spPrivateKey = privateKey
	}

	// Parse ACS URL
	acsURL, err := url.Parse(p.ACSUrl)
	if err != nil {
		return fmt.Errorf("invalid ACS URL: %w", err)
	}

	// Get metadata URL
	metadataURL, err := url.Parse(p.MetadataURL)
	if err != nil {
		return fmt.Errorf("invalid metadata URL: %w", err)
	}

	// Create service provider
	sp := &saml.ServiceProvider{
		EntityID:          p.SPEntityID,
		Key:               p.spPrivateKey,
		Certificate:       p.spCertificate,
		IDPMetadata:       p.getIdPMetadata(),
		AcsURL:            *acsURL,
		MetadataURL:       *metadataURL,
		AllowIDPInitiated: p.AllowIdPInitiated,
	}

	p.serviceProvider = sp

	return nil
}

// ValidateRelayState validates the RelayState parameter according to SAML 2.0 specifications.
// This provides CSRF protection by ensuring RelayState is properly formatted and within size limits.
// Returns an error if validation fails.
func ValidateRelayState(relayState string) error {
	if relayState == "" {
		return fmt.Errorf("relayState cannot be empty")
	}

	// Validate size: SAML 2.0 spec recommends max 80 bytes
	if len(relayState) > relayStateMaxSize {
		return fmt.Errorf("relayState exceeds maximum size: %d bytes (max %d bytes)", len(relayState), relayStateMaxSize)
	}

	// Validate format: should be valid base64 (for OAuth payload encoding)
	// We don't decode here to avoid unnecessary processing, but we check basic format
	// The actual decryption/parsing happens later in the OAuth flow
	if len(relayState) == 0 {
		return fmt.Errorf("relayState is empty")
	}

	// Basic format validation: base64 strings should only contain valid characters
	// This is a lightweight check - full validation happens during OAuth payload parsing
	base64Chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	for _, char := range relayState {
		if !strings.ContainsRune(base64Chars, char) {
			return fmt.Errorf("relayState contains invalid characters (expected base64)")
		}
	}

	return nil
}

// validateRelayState is an internal wrapper for ValidateRelayState.
func validateRelayState(relayState string) error {
	return ValidateRelayState(relayState)
}

// getSSOURL returns the SSO URL from the IdP metadata.
func (p *Provider) getSSOURL() string {
	metadata := p.getIdPMetadata()

	if metadata == nil {
		return ""
	}

	if len(p.idpMetadata.IDPSSODescriptors) > 0 {
		idpSSODescriptor := p.idpMetadata.IDPSSODescriptors[0]
		for _, ssoService := range idpSSODescriptor.SingleSignOnServices {
			if ssoService.Binding == saml.HTTPRedirectBinding {
				return ssoService.Location
			}
		}
		// Fallback to first SSO service if redirect binding not found
		if len(idpSSODescriptor.SingleSignOnServices) > 0 {
			return idpSSODescriptor.SingleSignOnServices[0].Location
		}
	}

	return ""
}

// GetSPMetadata returns the SP metadata XML as bytes.
// This can be used to serve the metadata endpoint.
func (p *Provider) GetSPMetadata() ([]byte, error) {
	// Ensure service provider is initialized
	if p.serviceProvider == nil {
		if err := p.initializeServiceProvider(); err != nil {
			return nil, fmt.Errorf("failed to initialize service provider: %w", err)
		}
	}

	// Generate metadata EntityDescriptor using the SAML library
	metadata := p.serviceProvider.Metadata()

	// Marshal to XML
	metadataXML, err := xml.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata to XML: %w", err)
	}

	// Add XML declaration
	xmlBytes := []byte(xml.Header)
	xmlBytes = append(xmlBytes, metadataXML...)

	return xmlBytes, nil
}

// resolveNameAttribute resolves the name attribute, supporting templating.
// Templates use {{attributeName}} syntax to reference SAML attributes.
// If no template is detected, falls back to direct attribute lookup.
func (p *Provider) resolveNameAttribute(attributes map[string][]string) (string, error) {
	nameTemplate := p.NameAttribute

	// Check if this looks like a template (contains {{}})
	if strings.Contains(nameTemplate, "{{") && strings.Contains(nameTemplate, "}}") {
		// Parse and execute template
		tmpl, err := template.New("name").Parse(nameTemplate)
		if err != nil {
			return "", fmt.Errorf("invalid name attribute template: %w", err)
		}

		// Create template data - Go templates expect a struct/map, not individual variables
		// We'll create a map where attribute names are accessible as .AttributeName
		templateData := make(map[string]string)
		for key, values := range attributes {
			if len(values) > 0 {
				templateData[key] = values[0] // Use first value for templates
			}
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, templateData); err != nil {
			return "", fmt.Errorf("failed to execute name attribute template: %w", err)
		}

		result := strings.TrimSpace(buf.String())
		// Check if template resulted in "<no value>" (missing key) or empty
		if result != "" && result != "<no value>" {
			return result, nil
		}
		// If template resulted in empty string, fall through to fallback logic
	} else {
		// Direct attribute lookup
		if nameAttr, ok := attributes[nameTemplate]; ok && len(nameAttr) > 0 {
			return nameAttr[0], nil
		}
	}

	// Fallback: Try common SAML attribute names
	for _, attrName := range []string{"displayName", "name", "givenName", "cn", "uid"} {
		if nameAttr, ok := attributes[attrName]; ok && len(nameAttr) > 0 {
			return nameAttr[0], nil
		}
	}

	return "", fmt.Errorf("name attribute not found in SAML response")
}

// mapAttributesToUser maps SAML attributes to the user struct.
func (p *Provider) mapAttributesToUser(attributes map[string][]string, user *auth.User) error {
	// Extract name - support templating
	userName, err := p.resolveNameAttribute(attributes)
	if err != nil {
		return fmt.Errorf("failed to resolve name attribute: %w", err)
	}

	user.Name = userName

	// Extract email
	if emailAttr, ok := attributes[p.EmailAttribute]; ok && len(emailAttr) > 0 {
		user.Email = emailAttr[0]
	} else {
		// Try common SAML attribute names
		for _, attrName := range []string{"email", "mail", "emailAddress"} {
			if emailAttr, ok := attributes[attrName]; ok && len(emailAttr) > 0 {
				user.Email = emailAttr[0]
				break
			}
		}
	}

	if user.Email == "" {
		return fmt.Errorf("email attribute not found in SAML response")
	}

	// Extract groups (optional)
	if p.GroupsAttribute != "" {
		if groupsAttr, ok := attributes[p.GroupsAttribute]; ok {
			user.Groups = groupsAttr
		}
	}

	return nil
}
