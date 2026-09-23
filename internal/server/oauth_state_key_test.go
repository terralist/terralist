package server

import (
	"bytes"
	"testing"

	"terralist/internal/server/models/oauth"
)

func TestOAuthStateKey(t *testing.T) {
	t.Run("uses the dedicated secret when configured", func(t *testing.T) {
		key := oauthStateKey(UserConfig{
			OAuthStateSecret:   "state-secret",
			TokenSigningSecret: "token-secret",
		})

		if !bytes.Equal(key, []byte("state-secret")) {
			t.Fatalf("expected the configured state secret, got %q", key)
		}
	})

	t.Run("derives the key from the token signing secret when unset", func(t *testing.T) {
		key := oauthStateKey(UserConfig{
			TokenSigningSecret: "token-secret",
		})

		if !bytes.Equal(key, oauth.DeriveStateKey("token-secret")) {
			t.Fatalf("expected a key derived from the token signing secret, got %q", key)
		}
	})
}
