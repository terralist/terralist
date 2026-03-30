package saml

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/crewjam/saml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestCertificate creates a test X.509 certificate for testing.
func createTestCertificate(t *testing.T) *x509.Certificate {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test Org"},
			Country:       []string{"US"},
			Province:      []string{"CA"},
			Locality:      []string{"Test City"},
			StreetAddress: []string{"Test Street"},
			PostalCode:    []string{"12345"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert
}

// createTestMetadata creates a test SAML EntityDescriptor for testing.
func createTestMetadata(t *testing.T) *saml.EntityDescriptor {
	cert := createTestCertificate(t)
	certDER := cert.Raw
	certBase64 := base64.StdEncoding.EncodeToString(certDER)

	return &saml.EntityDescriptor{
		EntityID: "https://test-idp.example.com",
		IDPSSODescriptors: []saml.IDPSSODescriptor{
			{
				SSODescriptor: saml.SSODescriptor{
					RoleDescriptor: saml.RoleDescriptor{
						KeyDescriptors: []saml.KeyDescriptor{
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
						},
					},
				},
				SingleSignOnServices: []saml.Endpoint{
					{
						Binding:  saml.HTTPRedirectBinding,
						Location: "https://test-idp.example.com/sso",
					},
				},
			},
		},
	}
}

// TestRequestTracker_Track tests that request IDs are tracked correctly.
func TestRequestTracker_Track(t *testing.T) {
	rt := newRequestTracker(15 * time.Minute)
	defer rt.Stop()

	requestID := "test-request-id-123"

	// Track a request ID
	rt.Track(requestID)

	// Verify it was tracked
	rt.mutex.RLock()
	_, exists := rt.requestIDs[requestID]
	rt.mutex.RUnlock()

	assert.True(t, exists, "Request ID should be tracked")
}

// TestRequestTracker_TrackEmptyID tests that empty request IDs are ignored.
func TestRequestTracker_TrackEmptyID(t *testing.T) {
	rt := newRequestTracker(15 * time.Minute)
	defer rt.Stop()

	// Track empty request ID
	rt.Track("")

	// Verify no entries were created
	rt.mutex.RLock()
	count := len(rt.requestIDs)
	rt.mutex.RUnlock()

	assert.Equal(t, 0, count, "Empty request ID should not be tracked")
}

// TestRequestTracker_ValidateAndConsume tests that request IDs can be validated and consumed.
func TestRequestTracker_ValidateAndConsume(t *testing.T) {
	rt := newRequestTracker(15 * time.Minute)
	defer rt.Stop()

	requestID := "test-request-id-456"

	// Track a request ID
	rt.Track(requestID)

	// Validate and consume it
	valid := rt.ValidateAndConsume(requestID, time.Hour)
	assert.True(t, valid, "Request ID should be valid")

	// Verify it was consumed (removed)
	rt.mutex.RLock()
	_, exists := rt.requestIDs[requestID]
	rt.mutex.RUnlock()

	assert.False(t, exists, "Request ID should be consumed (removed)")
}

// TestRequestTracker_ValidateAndConsume_ReplayAttack tests replay attack prevention.
func TestRequestTracker_ValidateAndConsume_ReplayAttack(t *testing.T) {
	rt := newRequestTracker(15 * time.Minute)
	defer rt.Stop()

	requestID := "test-request-id-789"

	// Track a request ID
	rt.Track(requestID)

	// First validation should succeed
	valid1 := rt.ValidateAndConsume(requestID, time.Hour)
	assert.True(t, valid1, "First validation should succeed")

	// Second validation should fail (replay attack)
	valid2 := rt.ValidateAndConsume(requestID, time.Hour)
	assert.False(t, valid2, "Replay attack should be prevented")
}

// TestRequestTracker_ValidateAndConsume_UnknownID tests that unknown request IDs are rejected.
func TestRequestTracker_ValidateAndConsume_UnknownID(t *testing.T) {
	rt := newRequestTracker(15 * time.Minute)
	defer rt.Stop()

	requestID := "unknown-request-id"

	// Try to validate an unknown request ID
	valid := rt.ValidateAndConsume(requestID, time.Hour)
	assert.False(t, valid, "Unknown request ID should be rejected")
}

// TestRequestTracker_ValidateAndConsume_EmptyID tests that empty request IDs are rejected.
func TestRequestTracker_ValidateAndConsume_EmptyID(t *testing.T) {
	rt := newRequestTracker(15 * time.Minute)
	defer rt.Stop()

	// Try to validate an empty request ID
	valid := rt.ValidateAndConsume("", time.Hour)
	assert.False(t, valid, "Empty request ID should be rejected")
}

// TestRequestTracker_Expiration tests that expired request IDs are rejected.
func TestRequestTracker_Expiration(t *testing.T) {
	rt := newRequestTracker(15 * time.Minute)
	defer rt.Stop()

	requestID := "expired-request-id"

	// Manually add an expired request ID
	rt.mutex.Lock()
	rt.requestIDs[requestID] = time.Now().Add(-2 * time.Hour) // Expired (older than 1 hour)
	rt.mutex.Unlock()

	// Try to validate expired request ID
	valid := rt.ValidateAndConsume(requestID, time.Hour)
	assert.False(t, valid, "Expired request ID should be rejected")

	// Verify it was cleaned up
	rt.mutex.RLock()
	_, exists := rt.requestIDs[requestID]
	rt.mutex.RUnlock()

	assert.False(t, exists, "Expired request ID should be removed")
}

// TestRequestTracker_ConcurrentAccess tests thread-safety of request tracker.
func TestRequestTracker_ConcurrentAccess(t *testing.T) {
	rt := newRequestTracker(15 * time.Minute)
	defer rt.Stop()

	// Concurrently track and validate request IDs
	done := make(chan bool, 100)
	for i := 0; i < 50; i++ {
		go func(id int) {
			requestID := fmt.Sprintf("concurrent-request-%d", id)
			rt.Track(requestID)
			done <- true
		}(i)
	}

	for i := 0; i < 50; i++ {
		go func(id int) {
			requestID := fmt.Sprintf("concurrent-request-%d", id)
			rt.ValidateAndConsume(requestID, time.Hour)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify no data races occurred (test should complete without panics)
	assert.True(t, true, "Concurrent access should be safe")
}

// TestValidateAssertionExpiration_NilAssertion tests nil assertion handling.
func TestValidateAssertionExpiration_NilAssertion(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
	}

	err := p.validateAssertionExpiration(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "assertion is nil")
}

// TestValidateAssertionExpiration_NotBefore tests NotBefore validation.
func TestValidateAssertionExpiration_NotBefore(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
		MaxAssertionAge:    1 * time.Hour,
	}

	// Create assertion with NotBefore in the future (beyond clock skew)
	assertion := &saml.Assertion{
		Conditions: &saml.Conditions{
			NotBefore: time.Now().Add(10 * time.Minute), // 10 minutes in future
		},
		IssueInstant: time.Now(),
	}

	err := p.validateAssertionExpiration(assertion)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet valid")
}

// TestValidateAssertionExpiration_NotBefore_WithClockSkew tests NotBefore with clock skew tolerance.
func TestValidateAssertionExpiration_NotBefore_WithClockSkew(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
		MaxAssertionAge:    1 * time.Hour,
	}

	// Create assertion with NotBefore slightly in the future (within clock skew)
	assertion := &saml.Assertion{
		Conditions: &saml.Conditions{
			NotBefore: time.Now().Add(2 * time.Minute), // 2 minutes in future (within 5 min skew)
		},
		IssueInstant: time.Now(),
	}

	err := p.validateAssertionExpiration(assertion)
	assert.NoError(t, err, "Assertion should be valid within clock skew tolerance")
}

// TestValidateAssertionExpiration_NotOnOrAfter tests NotOnOrAfter validation.
func TestValidateAssertionExpiration_NotOnOrAfter(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
		MaxAssertionAge:    1 * time.Hour,
	}

	// Create assertion with NotOnOrAfter in the past (beyond clock skew)
	assertion := &saml.Assertion{
		Conditions: &saml.Conditions{
			NotOnOrAfter: time.Now().Add(-10 * time.Minute), // 10 minutes ago
		},
		IssueInstant: time.Now().Add(-15 * time.Minute),
	}

	err := p.validateAssertionExpiration(assertion)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// TestValidateAssertionExpiration_NotOnOrAfter_WithClockSkew tests NotOnOrAfter with clock skew tolerance.
func TestValidateAssertionExpiration_NotOnOrAfter_WithClockSkew(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
		MaxAssertionAge:    1 * time.Hour,
	}

	// Create assertion with NotOnOrAfter slightly in the past (within clock skew)
	assertion := &saml.Assertion{
		Conditions: &saml.Conditions{
			NotOnOrAfter: time.Now().Add(-2 * time.Minute), // 2 minutes ago (within 5 min skew)
		},
		IssueInstant: time.Now().Add(-5 * time.Minute),
	}

	err := p.validateAssertionExpiration(assertion)
	assert.NoError(t, err, "Assertion should be valid within clock skew tolerance")
}

// TestValidateAssertionExpiration_IssueInstant_TooOld tests IssueInstant age validation.
func TestValidateAssertionExpiration_IssueInstant_TooOld(t *testing.T) {
	p := &Provider{
		MaxAssertionAge:    time.Hour, // 1 hour max age
		AssertionClockSkew: 5 * time.Minute,
	}

	// Create assertion with IssueInstant too old
	assertion := &saml.Assertion{
		IssueInstant: time.Now().Add(-2 * time.Hour), // 2 hours ago (max is 1 hour)
	}

	err := p.validateAssertionExpiration(assertion)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too old")
}

// TestValidateAssertionExpiration_IssueInstant_Future tests IssueInstant future validation.
func TestValidateAssertionExpiration_IssueInstant_Future(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
		MaxAssertionAge:    1 * time.Hour,
	}

	// Create assertion with IssueInstant in the future (beyond clock skew)
	assertion := &saml.Assertion{
		IssueInstant: time.Now().Add(10 * time.Minute), // 10 minutes in future
	}

	err := p.validateAssertionExpiration(assertion)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from the future")
}

// TestValidateAssertionExpiration_IssueInstant_WithClockSkew tests IssueInstant with clock skew tolerance.
func TestValidateAssertionExpiration_IssueInstant_WithClockSkew(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
		MaxAssertionAge:    1 * time.Hour,
	}

	// Create assertion with IssueInstant slightly in the future (within clock skew)
	assertion := &saml.Assertion{
		IssueInstant: time.Now().Add(2 * time.Minute), // 2 minutes in future (within 5 min skew)
	}

	err := p.validateAssertionExpiration(assertion)
	assert.NoError(t, err, "Assertion should be valid within clock skew tolerance")
}

// TestValidateAssertionExpiration_ValidAssertion tests a valid assertion.
func TestValidateAssertionExpiration_ValidAssertion(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
		MaxAssertionAge:    1 * time.Hour,
	}

	// Create valid assertion
	assertion := &saml.Assertion{
		Conditions: &saml.Conditions{
			NotBefore:    time.Now().Add(-1 * time.Minute),
			NotOnOrAfter: time.Now().Add(30 * time.Minute),
		},
		IssueInstant: time.Now().Add(-5 * time.Minute),
	}

	err := p.validateAssertionExpiration(assertion)
	assert.NoError(t, err, "Valid assertion should pass validation")
}

// TestExtractIdPCertificatesFromMetadata_Security tests certificate extraction from metadata.
func TestExtractIdPCertificatesFromMetadata_Security(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
	}

	// Create test metadata with certificate
	metadata := createTestMetadata(t)
	p.metadataMutex.Lock()
	p.idpMetadata = metadata
	p.metadataMutex.Unlock()

	// Extract certificates
	certificates, err := p.extractIdPCertificatesFromMetadata()

	require.NoError(t, err)
	assert.NotEmpty(t, certificates, "Should extract at least one certificate")
	assert.Equal(t, 1, len(certificates), "Should extract exactly one certificate")
}

// TestExtractIdPCertificatesFromMetadata_NoMetadata tests extraction when metadata is not loaded.
func TestExtractIdPCertificatesFromMetadata_NoMetadata(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
	}

	// Don't load metadata
	certificates, err := p.extractIdPCertificatesFromMetadata()

	assert.Error(t, err)
	assert.Nil(t, certificates)
	assert.Contains(t, err.Error(), "metadata not loaded")
}

// TestExtractIdPCertificatesFromMetadata_NoDescriptors tests extraction when no IDPSSODescriptors exist.
func TestExtractIdPCertificatesFromMetadata_NoDescriptors(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
	}

	// Create metadata without IDPSSODescriptors
	metadata := &saml.EntityDescriptor{
		EntityID:          "https://test-idp.example.com",
		IDPSSODescriptors: []saml.IDPSSODescriptor{},
	}
	p.metadataMutex.Lock()
	p.idpMetadata = metadata
	p.metadataMutex.Unlock()

	// Extract certificates
	certificates, err := p.extractIdPCertificatesFromMetadata()

	require.NoError(t, err)
	assert.Empty(t, certificates, "Should return empty slice when no descriptors")
}

// TestExtractIdPCertificatesFromMetadata_MultipleCertificates_Security tests extraction of multiple certificates.
func TestExtractIdPCertificatesFromMetadata_MultipleCertificates_Security(t *testing.T) {
	p := &Provider{
		AssertionClockSkew: 5 * time.Minute,
	}

	// Create test certificates
	cert1 := createTestCertificate(t)
	cert2 := createTestCertificate(t)

	cert1DER := cert1.Raw
	cert1Base64 := base64.StdEncoding.EncodeToString(cert1DER)
	cert2DER := cert2.Raw
	cert2Base64 := base64.StdEncoding.EncodeToString(cert2DER)

	// Create metadata with multiple certificates
	metadata := &saml.EntityDescriptor{
		EntityID: "https://test-idp.example.com",
		IDPSSODescriptors: []saml.IDPSSODescriptor{
			{
				SSODescriptor: saml.SSODescriptor{
					RoleDescriptor: saml.RoleDescriptor{
						KeyDescriptors: []saml.KeyDescriptor{
							{
								Use: "signing",
								KeyInfo: saml.KeyInfo{
									X509Data: saml.X509Data{
										X509Certificates: []saml.X509Certificate{
											{Data: cert1Base64},
											{Data: cert2Base64},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	p.metadataMutex.Lock()
	p.idpMetadata = metadata
	p.metadataMutex.Unlock()

	// Extract certificates
	certificates, err := p.extractIdPCertificatesFromMetadata()

	require.NoError(t, err)
	assert.Equal(t, 2, len(certificates), "Should extract both certificates")
}

// TestShouldRefreshMetadata_ValidUntil tests metadata refresh based on ValidUntil.
func TestShouldRefreshMetadata_ValidUntil(t *testing.T) {
	p := &Provider{
		IdPMetadataURL:               "https://test-idp.example.com/metadata",
		MetadataRefreshCheckInterval: 1 * time.Hour,
		MetadataRefreshInterval:      24 * time.Hour,
	}

	// Create metadata that expires soon
	metadata := &saml.EntityDescriptor{
		EntityID:   "https://test-idp.example.com",
		ValidUntil: time.Now().Add(30 * time.Minute), // Expires within check interval
	}
	p.metadataMutex.Lock()
	p.idpMetadata = metadata
	p.metadataLastRefresh = time.Now()
	p.metadataMutex.Unlock()

	shouldRefresh := p.shouldRefreshMetadata()
	assert.True(t, shouldRefresh, "Should refresh when ValidUntil is approaching")
}

// TestShouldRefreshMetadata_CacheDuration tests metadata refresh based on CacheDuration.
func TestShouldRefreshMetadata_CacheDuration(t *testing.T) {
	p := &Provider{
		IdPMetadataURL:               "https://test-idp.example.com/metadata",
		MetadataRefreshCheckInterval: 1 * time.Hour,
		MetadataRefreshInterval:      24 * time.Hour,
	}

	// Create metadata with CacheDuration
	metadata := &saml.EntityDescriptor{
		EntityID:      "https://test-idp.example.com",
		CacheDuration: 1 * time.Hour,
	}
	p.metadataMutex.Lock()
	p.idpMetadata = metadata
	p.metadataLastRefresh = time.Now().Add(-2 * time.Hour) // Refreshed 2 hours ago
	p.metadataMutex.Unlock()

	shouldRefresh := p.shouldRefreshMetadata()
	assert.True(t, shouldRefresh, "Should refresh when CacheDuration has passed")
}

// TestShouldRefreshMetadata_DefaultInterval tests metadata refresh based on default interval.
func TestShouldRefreshMetadata_DefaultInterval(t *testing.T) {
	p := &Provider{
		IdPMetadataURL:               "https://test-idp.example.com/metadata",
		MetadataRefreshCheckInterval: 1 * time.Hour,
		MetadataRefreshInterval:      24 * time.Hour,
	}

	// Create metadata without ValidUntil or CacheDuration
	metadata := &saml.EntityDescriptor{
		EntityID: "https://test-idp.example.com",
	}
	p.metadataMutex.Lock()
	p.idpMetadata = metadata
	p.metadataLastRefresh = time.Now().Add(-25 * time.Hour) // Refreshed 25 hours ago
	p.metadataMutex.Unlock()

	shouldRefresh := p.shouldRefreshMetadata()
	assert.True(t, shouldRefresh, "Should refresh after default interval (24 hours)")
}

// TestShouldRefreshMetadata_NoRefreshNeeded tests when refresh is not needed.
func TestShouldRefreshMetadata_NoRefreshNeeded(t *testing.T) {
	p := &Provider{
		IdPMetadataURL:               "https://test-idp.example.com/metadata",
		MetadataRefreshCheckInterval: 1 * time.Hour,
		MetadataRefreshInterval:      24 * time.Hour,
	}

	// Create metadata that doesn't need refresh
	metadata := &saml.EntityDescriptor{
		EntityID:   "https://test-idp.example.com",
		ValidUntil: time.Now().Add(2 * time.Hour), // Expires in 2 hours
	}
	p.metadataMutex.Lock()
	p.idpMetadata = metadata
	p.metadataLastRefresh = time.Now().Add(-1 * time.Hour) // Refreshed 1 hour ago
	p.metadataMutex.Unlock()

	shouldRefresh := p.shouldRefreshMetadata()
	assert.False(t, shouldRefresh, "Should not refresh when not needed")
}

// TestShouldRefreshMetadata_NoMetadata tests when metadata is not loaded.
func TestShouldRefreshMetadata_NoMetadata(t *testing.T) {
	p := &Provider{
		IdPMetadataURL:               "https://test-idp.example.com/metadata",
		MetadataRefreshCheckInterval: 1 * time.Hour,
		MetadataRefreshInterval:      24 * time.Hour,
	}

	// Don't load metadata
	shouldRefresh := p.shouldRefreshMetadata()
	assert.False(t, shouldRefresh, "Should not refresh when metadata is not loaded")
}
