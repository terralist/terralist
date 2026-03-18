package saml

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Config implements auth.Configurator interface and
// handles the configuration parameters for SAML authentication.
type Config struct {
	// IdPMetadataURL is the URL where the IdP metadata can be fetched from.
	// Either IdPMetadataURL, IdPMetadataFile, or both IdPEntityID and IdPSSOURL must be provided.
	IdPMetadataURL string

	// IdPMetadataFile is the local file path to the IdP metadata XML file.
	// Either IdPMetadataURL, IdPMetadataFile, or both IdPEntityID and IdPSSOURL must be provided.
	IdPMetadataFile string

	// IdPEntityID is the Identity Provider entity ID.
	// Can be used instead of IdPMetadataURL/IdPMetadataFile if IdPSSOURL is also provided.
	IdPEntityID string

	// IdPSSOURL is the Identity Provider Single Sign-On URL.
	// Can be used instead of IdPMetadataURL/IdPMetadataFile if IdPEntityID is also provided.
	IdPSSOURL string

	// IdPSSOCertificate is the Identity Provider SSO certificate (PEM format).
	// Required if certificate cannot be extracted from IdP metadata.
	IdPSSOCertificate string

	// NameAttribute is the SAML attribute name that contains the user's name.
	// Defaults to common SAML attribute names if not specified.
	NameAttribute string

	// EmailAttribute is the SAML attribute name that contains the user's email.
	// Defaults to common SAML attribute names if not specified.
	EmailAttribute string

	// GroupsAttribute is the SAML attribute name that contains the user's groups.
	// This is optional and used for RBAC group mapping.
	GroupsAttribute string

	// CertFile is the path to the certificate file (PEM format) used for signing SAML requests.
	// This is optional but recommended for production deployments.
	CertFile string

	// KeyFile is the path to the private key file (PEM format) used for signing SAML requests.
	// This is optional but recommended for production deployments.
	KeyFile string

	// PrivateKeySecret is the passphrase for the private key if it is encrypted.
	// This is optional and only required if the private key file is encrypted.
	PrivateKeySecret string

	// TerralistSchemeHostAndPort is the base URL of the Terralist instance.
	// Used to construct the ACS (Assertion Consumer Service) redirect URL.
	TerralistSchemeHostAndPort string

	// HTTPClientTimeout is the timeout for HTTP requests to fetch IdP metadata.
	// Default: 30 seconds.
	HTTPClientTimeout time.Duration

	// AssertionClockSkew is the allowed time difference between SP and IdP clocks.
	// SAML 2.0 spec recommends allowing clock skew (typically 5 minutes).
	// Default: 5 minutes.
	AssertionClockSkew time.Duration

	// RequestIDExpiration is how long SAML request IDs are kept to prevent replay attacks.
	// Default: 1 hour.
	RequestIDExpiration time.Duration

	// RequestIDCleanupInterval is how often expired request IDs are cleaned up.
	// Default: 15 minutes.
	RequestIDCleanupInterval time.Duration

	// MetadataRefreshInterval is the default interval for refreshing IdP metadata.
	// Default: 24 hours.
	MetadataRefreshInterval time.Duration

	// MetadataRefreshCheckInterval is how often to check if metadata needs refresh.
	// Default: 1 hour.
	MetadataRefreshCheckInterval time.Duration

	// MaxAssertionAge is the maximum age of SAML assertions from IssueInstant.
	// Default: 1 hour.
	MaxAssertionAge time.Duration

	// AllowIdPInitiated specifies whether to allow IdP-initiated SSO.
	// Security best practice is to disable this. Default: false.
	AllowIdPInitiated bool

	// DisableRequestIDValidation disables SAML request ID validation.
	// This can be useful in Kubernetes environments where requests may be
	// routed to different pods. Default: false.
	DisableRequestIDValidation bool
}

func (c *Config) SetDefaults() {
	if c.NameAttribute == "" {
		c.NameAttribute = "displayName"
	}

	if c.EmailAttribute == "" {
		c.EmailAttribute = "email"
	}

	// Set SAML timing defaults with SAML spec-compliant values
	if c.HTTPClientTimeout == 0 {
		c.HTTPClientTimeout = 30 * time.Second
	}

	if c.AssertionClockSkew == 0 {
		c.AssertionClockSkew = 5 * time.Minute
	}

	if c.RequestIDExpiration == 0 {
		c.RequestIDExpiration = 1 * time.Hour
	}

	if c.RequestIDCleanupInterval == 0 {
		c.RequestIDCleanupInterval = 15 * time.Minute
	}

	if c.MetadataRefreshInterval == 0 {
		c.MetadataRefreshInterval = 24 * time.Hour
	}

	if c.MetadataRefreshCheckInterval == 0 {
		c.MetadataRefreshCheckInterval = 1 * time.Hour
	}

	if c.MaxAssertionAge == 0 {
		c.MaxAssertionAge = 1 * time.Hour
	}

	// AllowIdPInitiated defaults to false for security
	// (no change needed since bool defaults to false)
}

func (c *Config) Validate() error {
	// Validate that we have either metadata source OR direct IdP config
	hasMetadataSource := c.IdPMetadataURL != "" || c.IdPMetadataFile != ""
	hasDirectIdPConfig := c.IdPEntityID != "" && c.IdPSSOURL != ""

	if !hasMetadataSource && !hasDirectIdPConfig {
		return fmt.Errorf("missing required IdP configuration: either provide IdPMetadataURL/IdPMetadataFile, or both IdPEntityID and IdPSSOURL")
	}

	if hasMetadataSource && hasDirectIdPConfig {
		return fmt.Errorf("cannot specify both metadata source (IdPMetadataURL/IdPMetadataFile) and direct IdP config (IdPEntityID/IdPSSOURL) at the same time")
	}

	if c.IdPMetadataURL != "" && c.IdPMetadataFile != "" {
		return fmt.Errorf("both IdPMetadataURL and IdPMetadataFile cannot be set at the same time")
	}

	if c.IdPEntityID != "" && c.IdPSSOURL == "" {
		return fmt.Errorf("IdPEntityID requires IdPSSOURL to be set")
	}

	if c.IdPSSOURL != "" && c.IdPEntityID == "" {
		return fmt.Errorf("IdPSSOURL requires IdPEntityID to be set")
	}

	if c.TerralistSchemeHostAndPort == "" {
		return fmt.Errorf("missing required Terralist scheme host and port")
	}

	// Validate that Terralist URL uses HTTPS scheme (required for SAML security)
	terralistURL, err := url.Parse(c.TerralistSchemeHostAndPort)
	if err != nil {
		return fmt.Errorf("invalid Terralist URL: %w", err)
	}

	// Check if URL scheme is HTTPS
	if strings.ToLower(terralistURL.Scheme) != "https" {
		return fmt.Errorf("SAML requires HTTPS transport - Terralist URL must use https:// scheme, got: %s", terralistURL.Scheme)
	}

	if c.CertFile != "" && c.KeyFile == "" {
		return fmt.Errorf("cert file specified but key file is missing")
	}

	if c.KeyFile != "" && c.CertFile == "" {
		return fmt.Errorf("key file specified but cert file is missing")
	}

	// Validate timing configurations
	if c.HTTPClientTimeout <= 0 {
		return fmt.Errorf("HTTPClientTimeout must be positive")
	}

	if c.AssertionClockSkew < 0 {
		return fmt.Errorf("AssertionClockSkew cannot be negative")
	}

	if c.RequestIDExpiration <= 0 {
		return fmt.Errorf("RequestIDExpiration must be positive")
	}

	if c.RequestIDCleanupInterval <= 0 {
		return fmt.Errorf("RequestIDCleanupInterval must be positive")
	}

	if c.MetadataRefreshInterval <= 0 {
		return fmt.Errorf("MetadataRefreshInterval must be positive")
	}

	if c.MetadataRefreshCheckInterval <= 0 {
		return fmt.Errorf("MetadataRefreshCheckInterval must be positive")
	}

	if c.MaxAssertionAge <= 0 {
		return fmt.Errorf("MaxAssertionAge must be positive")
	}

	// Validate that cleanup interval is reasonable compared to expiration
	if c.RequestIDCleanupInterval > c.RequestIDExpiration {
		return fmt.Errorf("RequestIDCleanupInterval (%v) cannot be longer than RequestIDExpiration (%v)",
			c.RequestIDCleanupInterval, c.RequestIDExpiration)
	}

	return nil
}
