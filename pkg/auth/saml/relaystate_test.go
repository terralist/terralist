package saml

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateRelayState tests the ValidateRelayState function.
func TestValidateRelayState(t *testing.T) {
	tests := []struct {
		name       string
		relayState string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid RelayState (within size limit)",
			relayState: base64.StdEncoding.EncodeToString([]byte("test-payload")),
			wantErr:    false,
		},
		{
			name:       "empty RelayState",
			relayState: "",
			wantErr:    true,
			errMsg:     "relayState cannot be empty",
		},
		{
			name:       "RelayState exceeds max size",
			relayState: base64.StdEncoding.EncodeToString(make([]byte, 600)), // 600 bytes encoded will be > 512 bytes
			wantErr:    true,
			errMsg:     "relayState exceeds maximum size",
		},
		{
			name:       "RelayState with invalid characters",
			relayState: "invalid@characters!",
			wantErr:    true,
			errMsg:     "relayState contains invalid characters",
		},
		{
			name:       "RelayState at max size boundary",
			relayState: base64.StdEncoding.EncodeToString(make([]byte, 384)), // 512 bytes when base64 encoded
			wantErr:    false,
		},
		{
			name:       "RelayState just over max size",
			relayState: base64.StdEncoding.EncodeToString(make([]byte, 385)), // ~514 bytes when base64 encoded
			wantErr:    true,
			errMsg:     "relayState exceeds maximum size",
		},
		{
			name:       "valid base64 string within limit",
			relayState: "dGVzdC1wYXlsb2Fk", // base64 for "test-payload"
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRelayState(tt.relayState)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateRelayState_SizeBoundary tests the exact size boundary.
func TestValidateRelayState_SizeBoundary(t *testing.T) {
	// Create a RelayState that is exactly 512 bytes
	relayState512 := base64.StdEncoding.EncodeToString(make([]byte, 384)) // 384 bytes -> 512 bytes base64
	if len(relayState512) <= 512 {
		err := ValidateRelayState(relayState512)
		assert.NoError(t, err, "RelayState at or below 512 bytes should be valid")
	}

	// Create a RelayState that exceeds 512 bytes
	relayState513 := base64.StdEncoding.EncodeToString(make([]byte, 385)) // 385 bytes -> ~514 bytes base64
	if len(relayState513) > 512 {
		err := ValidateRelayState(relayState513)
		assert.Error(t, err, "RelayState over 512 bytes should be rejected")
		assert.Contains(t, err.Error(), "exceeds maximum size")
	}
}

// TestValidateRelayState_Base64Format tests base64 format validation.
func TestValidateRelayState_Base64Format(t *testing.T) {
	// Valid base64 characters
	validChars := []string{
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"abcdefghijklmnopqrstuvwxyz",
		"0123456789",
		"+/=",
		"dGVzdA==", // valid base64
	}

	for _, valid := range validChars {
		err := ValidateRelayState(valid)
		// Note: These might fail size check, but should pass format check
		if len(valid) <= 80 {
			assert.NoError(t, err, "Valid base64 characters should pass format validation: %s", valid)
		}
	}

	// Invalid characters
	invalidChars := []string{
		"test@invalid",
		"test!invalid",
		"test#invalid",
		"test$invalid",
		"test%invalid",
	}

	for _, invalid := range invalidChars {
		if len(invalid) <= 80 {
			err := ValidateRelayState(invalid)
			assert.Error(t, err, "Invalid characters should be rejected: %s", invalid)
			assert.Contains(t, err.Error(), "invalid characters")
		}
	}
}
