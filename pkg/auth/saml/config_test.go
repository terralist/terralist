package saml

import (
	"testing"
	"time"

	"terralist/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig_SetDefaults tests the SetDefaults method.
func TestConfig_SetDefaults(t *testing.T) {
	cfg := &Config{}

	// Test that defaults are set when empty
	cfg.SetDefaults()
	assert.Equal(t, "displayName", cfg.NameAttribute, "NameAttribute should default to displayName")
	assert.Equal(t, "email", cfg.EmailAttribute, "EmailAttribute should default to email")

	// Test that existing values are not overwritten
	cfg = &Config{
		NameAttribute:  "customName",
		EmailAttribute: "customEmail",
	}
	cfg.SetDefaults()
	assert.Equal(t, "customName", cfg.NameAttribute, "Existing NameAttribute should not be overwritten")
	assert.Equal(t, "customEmail", cfg.EmailAttribute, "Existing EmailAttribute should not be overwritten")
}

// TestConfig_Validate tests the Validate method with various scenarios.
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with metadata URL",
			config: &Config{
				IdPMetadataURL:             "https://idp.example.com/metadata",
				TerralistSchemeHostAndPort: "https://terralist.example.com",
			},
			wantErr: false,
		},
		{
			name: "valid config with metadata file",
			config: &Config{
				IdPMetadataFile:            "/path/to/metadata.xml",
				TerralistSchemeHostAndPort: "https://terralist.example.com",
			},
			wantErr: false,
		},
		{
			name: "valid config with direct IdP config",
			config: &Config{
				IdPEntityID:                "https://idp.example.com",
				IdPSSOURL:                  "https://idp.example.com/sso",
				IdPSSOCertificate:          "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
				TerralistSchemeHostAndPort: "https://terralist.example.com",
			},
			wantErr: false,
		},
		{
			name: "missing all IdP config",
			config: &Config{
				TerralistSchemeHostAndPort: "https://terralist.example.com",
			},
			wantErr: true,
			errMsg:  "missing required IdP configuration",
		},
		{
			name: "both metadata URL and file",
			config: &Config{
				IdPMetadataURL:             "https://idp.example.com/metadata",
				IdPMetadataFile:            "/path/to/metadata.xml",
				TerralistSchemeHostAndPort: "https://terralist.example.com",
			},
			wantErr: true,
			errMsg:  "both IdPMetadataURL and IdPMetadataFile cannot be set",
		},
		{
			name: "both metadata source and direct config",
			config: &Config{
				IdPMetadataURL:             "https://idp.example.com/metadata",
				IdPEntityID:                "https://idp.example.com",
				IdPSSOURL:                  "https://idp.example.com/sso",
				TerralistSchemeHostAndPort: "https://terralist.example.com",
			},
			wantErr: true,
			errMsg:  "cannot specify both metadata source",
		},
		{
			name: "IdPEntityID without IdPSSOURL",
			config: &Config{
				IdPEntityID:                "https://idp.example.com",
				TerralistSchemeHostAndPort: "https://terralist.example.com",
			},
			wantErr: true,
			errMsg:  "missing required IdP configuration", // Validation checks for missing config first
		},
		{
			name: "IdPSSOURL without IdPEntityID",
			config: &Config{
				IdPSSOURL:                  "https://idp.example.com/sso",
				TerralistSchemeHostAndPort: "https://terralist.example.com",
			},
			wantErr: true,
			errMsg:  "missing required IdP configuration", // Validation checks for missing config first
		},
		{
			name: "missing Terralist URL",
			config: &Config{
				IdPMetadataURL: "https://idp.example.com/metadata",
			},
			wantErr: true,
			errMsg:  "missing required Terralist scheme host and port",
		},
		{
			name: "HTTP scheme (should fail)",
			config: &Config{
				IdPMetadataURL:             "https://idp.example.com/metadata",
				TerralistSchemeHostAndPort: "http://terralist.example.com",
			},
			wantErr: true,
			errMsg:  "SAML requires HTTPS transport",
		},
		{
			name: "invalid URL",
			config: &Config{
				IdPMetadataURL:             "https://idp.example.com/metadata",
				TerralistSchemeHostAndPort: "://invalid-url",
			},
			wantErr: true,
			errMsg:  "invalid Terralist URL",
		},
		{
			name: "cert file without key file",
			config: &Config{
				IdPMetadataURL:             "https://idp.example.com/metadata",
				TerralistSchemeHostAndPort: "https://terralist.example.com",
				CertFile:                   "/path/to/cert.pem",
			},
			wantErr: true,
			errMsg:  "cert file specified but key file is missing",
		},
		{
			name: "key file without cert file",
			config: &Config{
				IdPMetadataURL:             "https://idp.example.com/metadata",
				TerralistSchemeHostAndPort: "https://terralist.example.com",
				KeyFile:                    "/path/to/key.pem",
			},
			wantErr: true,
			errMsg:  "key file specified but cert file is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set defaults first to ensure duration fields are valid
			tt.config.SetDefaults()
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProvider_Name tests the Name method.
func TestProvider_Name(t *testing.T) {
	p := &Provider{}
	assert.Equal(t, "SAML", p.Name())
}

// TestProvider_mapAttributesToUser tests the mapAttributesToUser method.
func TestProvider_mapAttributesToUser(t *testing.T) {
	tests := []struct {
		name         string
		provider     *Provider
		attributes   map[string][]string
		expectedErr  bool
		expectedUser *auth.User
	}{
		{
			name: "successful mapping with custom attributes",
			provider: &Provider{
				NameAttribute:  "displayName",
				EmailAttribute: "email",
			},
			attributes: map[string][]string{
				"displayName": {"John Doe"},
				"email":       {"john@example.com"},
			},
			expectedErr: false,
			expectedUser: &auth.User{
				Name:  "John Doe",
				Email: "john@example.com",
			},
		},
		{
			name: "successful mapping with default attributes",
			provider: &Provider{
				NameAttribute:  "displayName",
				EmailAttribute: "email",
			},
			attributes: map[string][]string{
				"displayName": {"Jane Smith"},
				"email":       {"jane@example.com"},
				"groups":      {"admin", "user"},
			},
			expectedErr: false,
			expectedUser: &auth.User{
				Name:  "Jane Smith",
				Email: "jane@example.com",
			},
		},
		{
			name: "mapping with groups attribute",
			provider: &Provider{
				NameAttribute:   "displayName",
				EmailAttribute:  "email",
				GroupsAttribute: "groups",
			},
			attributes: map[string][]string{
				"displayName": {"Bob User"},
				"email":       {"bob@example.com"},
				"groups":      {"admin", "developer"},
			},
			expectedErr: false,
			expectedUser: &auth.User{
				Name:  "Bob User",
				Email: "bob@example.com",
			},
		},
		{
			name: "fallback to common attribute names",
			provider: &Provider{
				NameAttribute:  "displayName",
				EmailAttribute: "email",
			},
			attributes: map[string][]string{
				"name": {"Alice Admin"},
				"mail": {"alice@example.com"},
			},
			expectedErr: false,
			expectedUser: &auth.User{
				Name:  "Alice Admin",
				Email: "alice@example.com",
			},
		},
		{
			name: "missing email attribute",
			provider: &Provider{
				NameAttribute:  "displayName",
				EmailAttribute: "email",
			},
			attributes: map[string][]string{
				"displayName": {"Test User"},
			},
			expectedErr: true,
		},
		{
			name: "missing name attribute",
			provider: &Provider{
				NameAttribute:  "displayName",
				EmailAttribute: "email",
			},
			attributes: map[string][]string{
				"email": {"test@example.com"},
			},
			expectedErr: true,
		},
		{
			name: "multiple values for attribute (should use first)",
			provider: &Provider{
				NameAttribute:  "displayName",
				EmailAttribute: "email",
			},
			attributes: map[string][]string{
				"displayName": {"First Name", "Second Name"},
				"email":       {"first@example.com", "second@example.com"},
			},
			expectedErr: false,
			expectedUser: &auth.User{
				Name:  "First Name",
				Email: "first@example.com",
			},
		},
		{
			name: "name attribute templating - combine first and last name",
			provider: &Provider{
				NameAttribute:  "{{.givenName}} {{.sn}}",
				EmailAttribute: "email",
			},
			attributes: map[string][]string{
				"givenName": {"John"},
				"sn":        {"Doe"},
				"email":     {"john.doe@example.com"},
			},
			expectedErr: false,
			expectedUser: &auth.User{
				Name:  "John Doe",
				Email: "john.doe@example.com",
			},
		},
		{
			name: "name attribute templating - with fallback",
			provider: &Provider{
				NameAttribute:  "{{.displayName}}",
				EmailAttribute: "email",
			},
			attributes: map[string][]string{
				"givenName":   {"John"},
				"sn":          {"Doe"},
				"displayName": {"John Doe"},
				"email":       {"john.doe@example.com"},
			},
			expectedErr: false,
			expectedUser: &auth.User{
				Name:  "John Doe",
				Email: "john.doe@example.com",
			},
		},
		{
			name: "name attribute templating - empty result falls back",
			provider: &Provider{
				NameAttribute:  "{{.nonexistent}}",
				EmailAttribute: "email",
			},
			attributes: map[string][]string{
				"name":  {"Fallback Name"},
				"email": {"test@example.com"},
			},
			expectedErr: false,
			expectedUser: &auth.User{
				Name:  "Fallback Name",
				Email: "test@example.com",
			},
		},
		{
			name: "name attribute templating - invalid template",
			provider: &Provider{
				NameAttribute:  "{{invalid syntax",
				EmailAttribute: "email",
			},
			attributes: map[string][]string{
				"email": {"test@example.com"},
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &auth.User{}
			err := tt.provider.mapAttributesToUser(tt.attributes, user)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedUser != nil {
					assert.Equal(t, tt.expectedUser.Name, user.Name)
					assert.Equal(t, tt.expectedUser.Email, user.Email)
				}
			}
		})
	}
}

// TestCreator_New tests the Creator.New method.
func TestCreator_New(t *testing.T) {
	creator := &Creator{}

	// Test with valid config
	cfg := &Config{
		IdPMetadataURL:             "https://idp.example.com/metadata",
		TerralistSchemeHostAndPort: "https://terralist.example.com",
		NameAttribute:              "displayName",
		EmailAttribute:             "email",
	}

	cfg.SetDefaults() // Set defaults to ensure duration fields are valid

	provider, err := creator.New(cfg)
	require.NoError(t, err)
	require.NotNil(t, provider)

	samlProvider, ok := provider.(*Provider)
	require.True(t, ok, "Provider should be of type *Provider")

	assert.Equal(t, cfg.IdPMetadataURL, samlProvider.IdPMetadataURL)
	assert.Equal(t, "https://terralist.example.com/v1/api/auth/saml/metadata", samlProvider.SPEntityID)
	assert.Equal(t, cfg.NameAttribute, samlProvider.NameAttribute)
	assert.Equal(t, cfg.EmailAttribute, samlProvider.EmailAttribute)
	assert.Equal(t, "https://terralist.example.com/v1/api/auth/saml/acs", samlProvider.ACSUrl)
	assert.Equal(t, "https://terralist.example.com/v1/api/auth/saml/metadata", samlProvider.MetadataURL)
	assert.NotNil(t, samlProvider.requestTracker, "Request tracker should be initialized")
	assert.NotNil(t, samlProvider.stopMetadataRefresh, "Stop channel should be initialized")

	// Test with invalid configurator type (nil)
	_, err = creator.New(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported configurator")

	// Test URL trimming (trailing slash)
	cfg2 := &Config{
		IdPMetadataURL:             "https://idp.example.com/metadata",
		TerralistSchemeHostAndPort: "https://terralist.example.com/",
	}

	cfg2.SetDefaults() // Set defaults to ensure duration fields are valid
	provider2, err := creator.New(cfg2)
	require.NoError(t, err)
	samlProvider2, ok := provider2.(*Provider)
	require.True(t, ok, "provider should be of type *Provider")
	assert.Equal(t, "https://terralist.example.com/v1/api/auth/saml/acs", samlProvider2.ACSUrl)
	assert.Equal(t, "https://terralist.example.com/v1/api/auth/saml/metadata", samlProvider2.MetadataURL)
}

// TestProvider_GetSPMetadata tests the GetSPMetadata method.
func TestProvider_GetSPMetadata(t *testing.T) {
	p := &Provider{
		ACSUrl:                     "https://terralist.example.com/v1/api/auth/saml/acs",
		MetadataURL:                "https://terralist.example.com/v1/api/auth/saml/metadata",
		TerralistSchemeHostAndPort: "https://terralist.example.com",
		requestTracker:             newRequestTracker(15 * time.Minute),
		stopMetadataRefresh:        make(chan struct{}),
	}
	defer p.requestTracker.Stop()

	// Load test metadata
	metadata := createTestMetadata(t)
	p.metadataMutex.Lock()
	p.idpMetadata = metadata
	p.metadataMutex.Unlock()

	// Initialize service provider (this will generate self-signed cert)
	err := p.initializeServiceProvider()
	require.NoError(t, err, "Should initialize service provider successfully")

	// Get SP metadata
	metadataXML, err := p.GetSPMetadata()
	require.NoError(t, err, "Should generate SP metadata successfully")
	require.NotEmpty(t, metadataXML, "SP metadata should not be empty")

	// Verify it's valid XML
	assert.Contains(t, string(metadataXML), "<?xml", "Should contain XML declaration")
	assert.Contains(t, string(metadataXML), "EntityDescriptor", "Should contain EntityDescriptor")
	assert.Contains(t, string(metadataXML), p.SPEntityID, "Should contain SP EntityID")
}

// TestProvider_GetSPMetadata_NotInitialized tests GetSPMetadata when service provider is not initialized.
func TestProvider_GetSPMetadata_NotInitialized(t *testing.T) {
	p := &Provider{
		ACSUrl:                     "https://terralist.example.com/v1/api/auth/saml/acs",
		MetadataURL:                "https://terralist.example.com/v1/api/auth/saml/metadata",
		TerralistSchemeHostAndPort: "https://terralist.example.com",
		requestTracker:             newRequestTracker(15 * time.Minute),
		stopMetadataRefresh:        make(chan struct{}),
	}
	defer p.requestTracker.Stop()

	// Load test metadata
	metadata := createTestMetadata(t)
	p.metadataMutex.Lock()
	p.idpMetadata = metadata
	p.metadataMutex.Unlock()

	// Get SP metadata without initializing (should auto-initialize)
	metadataXML, err := p.GetSPMetadata()
	require.NoError(t, err, "Should auto-initialize and generate SP metadata")
	require.NotEmpty(t, metadataXML, "SP metadata should not be empty")
}
