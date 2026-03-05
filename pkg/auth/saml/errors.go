package saml

import (
	"errors"
	"strings"
)

// Error types for categorization.
var (
	// ErrAuthenticationFailed is a generic authentication failure error.
	ErrAuthenticationFailed = errors.New("authentication failed")
	// ErrInvalidResponse is returned for invalid SAML responses.
	ErrInvalidResponse = errors.New("invalid SAML response")
	// ErrConfigurationError is returned for configuration-related errors.
	ErrConfigurationError = errors.New("configuration error")
	// ErrInternalError is returned for internal system errors.
	ErrInternalError = errors.New("internal error")
)

// SanitizeError sanitizes an error message for user-facing display.
// It categorizes errors and returns safe, generic messages while preserving
// detailed error information for server-side logging.
// Returns: (sanitized user message, original error for logging).
func SanitizeError(err error) (string, error) {
	if err == nil {
		return "", nil
	}

	errMsg := err.Error()
	errMsgLower := strings.ToLower(errMsg)

	// Certificate-related errors - don't expose certificate details
	if strings.Contains(errMsgLower, "certificate") ||
		strings.Contains(errMsgLower, "cert") ||
		strings.Contains(errMsgLower, "pem") ||
		strings.Contains(errMsgLower, "x509") ||
		strings.Contains(errMsgLower, "private key") ||
		strings.Contains(errMsgLower, "decrypt") {
		return "Certificate configuration error. Please contact your administrator.", err
	}

	// File system errors - don't expose file paths
	if strings.Contains(errMsgLower, "file") ||
		strings.Contains(errMsgLower, "path") ||
		strings.Contains(errMsgLower, "directory") ||
		strings.Contains(errMsgLower, "open") ||
		strings.Contains(errMsgLower, "read") ||
		strings.Contains(errMsgLower, "permission denied") ||
		strings.Contains(errMsgLower, "no such file") {
		return "Configuration file error. Please contact your administrator.", err
	}

	// Network/URL errors - sanitize URLs but keep generic message
	if strings.Contains(errMsgLower, "url") ||
		strings.Contains(errMsgLower, "http") ||
		strings.Contains(errMsgLower, "fetch") ||
		strings.Contains(errMsgLower, "network") ||
		strings.Contains(errMsgLower, "connection") ||
		strings.Contains(errMsgLower, "timeout") ||
		strings.Contains(errMsgLower, "dial") {
		return "Network error occurred. Please try again later.", err
	}

	// XML/parsing errors - generic message
	if strings.Contains(errMsgLower, "xml") ||
		strings.Contains(errMsgLower, "parse") ||
		strings.Contains(errMsgLower, "unmarshal") ||
		strings.Contains(errMsgLower, "decode") ||
		strings.Contains(errMsgLower, "invalid") {
		return "Invalid SAML response format.", err
	}

	// Authentication/validation errors - can be more specific but still safe
	if strings.Contains(errMsgLower, "replay") ||
		strings.Contains(errMsgLower, "expired") ||
		strings.Contains(errMsgLower, "expiration") ||
		strings.Contains(errMsgLower, "invalid or replayed") ||
		strings.Contains(errMsgLower, "request id") {
		return "Authentication failed: invalid or expired response.", err
	}

	// Attribute errors - safe to expose
	if strings.Contains(errMsgLower, "attribute") ||
		strings.Contains(errMsgLower, "name") ||
		strings.Contains(errMsgLower, "email") {
		return errMsg, err // These are safe user-facing errors
	}

	// RelayState errors - safe to expose (already sanitized)
	if strings.Contains(errMsgLower, "relaystate") ||
		strings.Contains(errMsgLower, "relay state") {
		return errMsg, err // Already sanitized in ValidateRelayState
	}

	// Metadata errors - generic message
	if strings.Contains(errMsgLower, "metadata") {
		return "Identity provider configuration error. Please contact your administrator.", err
	}

	// Service provider initialization errors
	if strings.Contains(errMsgLower, "service provider") ||
		strings.Contains(errMsgLower, "sp init") ||
		strings.Contains(errMsgLower, "initialize") {
		return "Service provider configuration error. Please contact your administrator.", err
	}

	// Default: generic error message for unknown errors
	return "Authentication failed. Please try again or contact your administrator.", err
}

// SanitizeErrorForURL sanitizes an error for use in URL query parameters.
// This is used for error_description in OAuth error redirects.
// This is exported so it can be used by other packages (e.g., server package).
func SanitizeErrorForURL(err error) string {
	if err == nil {
		return ""
	}
	sanitized, _ := SanitizeError(err)
	return sanitized
}

// sanitizeErrorForLogging returns the original error for server-side logging.
// This preserves all error details for debugging while keeping user messages safe.
func SanitizeErrorForLogging(err error) error {
	_, originalErr := SanitizeError(err)
	return originalErr
}
