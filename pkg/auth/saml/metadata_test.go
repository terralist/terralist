package saml

import (
	"testing"

	"github.com/crewjam/saml"
	"github.com/stretchr/testify/assert"
)

// TestGetSSOURL tests the getSSOURL function which extracts SSO URL from metadata.
// This tests our code logic for selecting the appropriate SSO URL.
func TestGetSSOURL(t *testing.T) {
	p := &Provider{}

	// Test case 1: SSO URL with HTTP-Redirect binding (preferred)
	metadata1 := createTestMetadata(t)
	metadata1.IDPSSODescriptors[0].SingleSignOnServices = []saml.Endpoint{
		{
			Binding:  saml.HTTPRedirectBinding,
			Location: "https://test-idp.example.com/sso/redirect",
		},
		{
			Binding:  saml.HTTPPostBinding,
			Location: "https://test-idp.example.com/sso/post",
		},
	}
	p.metadataMutex.Lock()
	p.idpMetadata = metadata1
	p.metadataMutex.Unlock()

	ssoURL := p.getSSOURL()
	assert.Equal(t, "https://test-idp.example.com/sso/redirect", ssoURL, "Should prefer HTTP-Redirect binding")

	// Test case 2: Only HTTP-POST binding (fallback to first)
	metadata2 := createTestMetadata(t)
	metadata2.IDPSSODescriptors[0].SingleSignOnServices = []saml.Endpoint{
		{
			Binding:  saml.HTTPPostBinding,
			Location: "https://test-idp.example.com/sso/post",
		},
	}
	p.metadataMutex.Lock()
	p.idpMetadata = metadata2
	p.metadataMutex.Unlock()

	ssoURL = p.getSSOURL()
	assert.Equal(t, "https://test-idp.example.com/sso/post", ssoURL, "Should fallback to first SSO service if redirect binding not found")

	// Test case 3: No metadata loaded
	p.metadataMutex.Lock()
	p.idpMetadata = nil
	p.metadataMutex.Unlock()

	ssoURL = p.getSSOURL()
	assert.Empty(t, ssoURL, "Should return empty string when metadata is not loaded")

	// Test case 4: No IDPSSODescriptors
	metadata3 := &saml.EntityDescriptor{
		EntityID:          "https://test-idp.example.com",
		IDPSSODescriptors: []saml.IDPSSODescriptor{},
	}
	p.metadataMutex.Lock()
	p.idpMetadata = metadata3
	p.metadataMutex.Unlock()

	ssoURL = p.getSSOURL()
	assert.Empty(t, ssoURL, "Should return empty string when no IDPSSODescriptors")
}

// TestMetadataEntityID tests that EntityID is correctly accessible from metadata.
// This is a basic sanity check that our metadata structure is correct.
func TestMetadataEntityID(t *testing.T) {
	metadata := createTestMetadata(t)
	expectedEntityID := "https://test-idp.example.com"

	assert.Equal(t, expectedEntityID, metadata.EntityID, "EntityID should be correctly set")
	assert.NotEmpty(t, metadata.IDPSSODescriptors, "Should have at least one IDPSSODescriptor")
}
