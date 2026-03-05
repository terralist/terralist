package saml

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSanitizeError tests the error sanitization function.
func TestSanitizeError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedMsg    string
		shouldSanitize bool
	}{
		{
			name:           "certificate error",
			err:            errors.New("failed to parse certificate: x509: certificate has expired"),
			expectedMsg:    "Certificate configuration error. Please contact your administrator.",
			shouldSanitize: true,
		},
		{
			name:           "file path error",
			err:            errors.New("failed to open file: /etc/secrets/cert.pem: permission denied"),
			expectedMsg:    "Certificate configuration error. Please contact your administrator.", // Contains "pem" so matches certificate pattern first
			shouldSanitize: true,
		},
		{
			name:           "network error",
			err:            errors.New("failed to fetch metadata: dial tcp 192.168.1.1:443: connection refused"),
			expectedMsg:    "Network error occurred. Please try again later.",
			shouldSanitize: true,
		},
		{
			name:           "XML parsing error",
			err:            errors.New("failed to parse XML: invalid character"),
			expectedMsg:    "Invalid SAML response format.",
			shouldSanitize: true,
		},
		{
			name:           "replay attack error",
			err:            errors.New("invalid or replayed SAML request ID"),
			expectedMsg:    "Invalid SAML response format.", // Contains "invalid" so matches XML/parsing pattern first
			shouldSanitize: true,
		},
		{
			name:           "attribute error (safe to expose)",
			err:            errors.New("name attribute not found in SAML response"),
			expectedMsg:    "name attribute not found in SAML response",
			shouldSanitize: false, // Should not sanitize - safe user-facing error
		},
		{
			name:           "relayState error (safe to expose)",
			err:            errors.New("relayState exceeds maximum size: 100 bytes (max 80 bytes)"),
			expectedMsg:    "relayState exceeds maximum size: 100 bytes (max 80 bytes)",
			shouldSanitize: false, // Should not sanitize - already sanitized
		},
		{
			name:           "generic unknown error",
			err:            errors.New("some unknown internal error with sensitive details"),
			expectedMsg:    "Authentication failed. Please try again or contact your administrator.",
			shouldSanitize: true,
		},
		{
			name:           "nil error",
			err:            nil,
			expectedMsg:    "",
			shouldSanitize: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizedMsg, originalErr := SanitizeError(tt.err)
			assert.Equal(t, tt.expectedMsg, sanitizedMsg, "Sanitized message should match expected")
			if tt.err != nil {
				assert.Equal(t, tt.err, originalErr, "Original error should be preserved for logging")
			}
		})
	}
}

// TestSanitizeErrorForURL tests the exported function for URL sanitization.
func TestSanitizeErrorForURL(t *testing.T) {
	// Test certificate error
	certErr := errors.New("failed to parse certificate: x509: invalid certificate")
	sanitized := SanitizeErrorForURL(certErr)
	assert.Contains(t, sanitized, "Certificate configuration error")
	assert.NotContains(t, sanitized, "x509")
	assert.NotContains(t, sanitized, "invalid certificate")

	// Test file path error (without certificate-related keywords)
	fileErr := errors.New("failed to open file: /etc/secrets/config.json: permission denied")
	sanitized = SanitizeErrorForURL(fileErr)
	assert.Contains(t, sanitized, "Configuration file error")
	assert.NotContains(t, sanitized, "/etc/secrets/config.json")

	// Test nil error
	sanitized = SanitizeErrorForURL(nil)
	assert.Empty(t, sanitized)
}

// TestSanitizeErrorForLogging tests that original errors are preserved for logging.
func TestSanitizeErrorForLogging(t *testing.T) {
	originalErr := errors.New("failed to parse certificate: x509: certificate has expired")
	loggingErr := SanitizeErrorForLogging(originalErr)

	// Original error should be preserved for logging
	assert.Equal(t, originalErr, loggingErr)
	assert.Contains(t, loggingErr.Error(), "x509")
	assert.Contains(t, loggingErr.Error(), "certificate has expired")
}

// TestSanitizeError_Categories tests different error categories.
func TestSanitizeError_Categories(t *testing.T) {
	tests := []struct {
		category string
		err      error
		contains string
	}{
		{
			category: "certificate errors",
			err:      errors.New("certificate parsing failed"),
			contains: "Certificate configuration error",
		},
		{
			category: "file system errors",
			err:      errors.New("file not found"),
			contains: "Configuration file error",
		},
		{
			category: "network errors",
			err:      errors.New("network timeout"),
			contains: "Network error occurred",
		},
		{
			category: "XML parsing errors",
			err:      errors.New("XML parse error"),
			contains: "Invalid SAML response format",
		},
		{
			category: "authentication errors",
			err:      errors.New("replay attack detected"),
			contains: "Authentication failed: invalid or expired response",
		},
		{
			category: "metadata errors",
			err:      errors.New("metadata load failed"),
			contains: "Identity provider configuration error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			sanitized, _ := SanitizeError(tt.err)
			assert.Contains(t, sanitized, tt.contains, "Error should be sanitized to appropriate category")
		})
	}
}
