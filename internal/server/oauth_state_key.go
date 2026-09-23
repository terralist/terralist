package server

import (
	"terralist/internal/server/models/oauth"

	"github.com/rs/zerolog/log"
)

// oauthStateKey returns the key that signs the OAuth state payload. Falling
// back to a key derived from the token signing secret is deprecated and the
// oauth-state-secret flag will become required in the next version.
func oauthStateKey(config UserConfig) []byte {
	if config.OAuthStateSecret != "" {
		return []byte(config.OAuthStateSecret)
	}

	log.Warn().Msg("oauth-state-secret is not set, deriving it from token-signing-secret; this fallback is deprecated and the flag will become required in the next version")

	return oauth.DeriveStateKey(config.TokenSigningSecret)
}
