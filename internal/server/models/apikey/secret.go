package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// SecretPrefix marks API key secrets so they are recognisable in logs and by
// secret scanners.
const SecretPrefix = "tlk_"

// NewSecret generates a random API key secret. Only its hash is stored, so the
// value is shown to the user exactly once.
func NewSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("could not generate api key secret: %w", err)
	}

	return SecretPrefix + base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashSecret returns the hex encoded SHA-256 digest under which a secret is
// stored and looked up.
func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))

	return hex.EncodeToString(sum[:])
}
