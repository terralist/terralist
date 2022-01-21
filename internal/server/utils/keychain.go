package utils

import (
	"math/rand"
)

const (
	library = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_-+={}[]:;<>,."
)

// Keychain holds randomly generated runtime secrets
type Keychain struct {
	EncryptSalt        string
	CodeExchangeKey    string
	TokenSigningSecret []byte
}

// NewKeychain creates and initializez a Keychain instance
func NewKeychain() *Keychain {
	encryptSalt := generateRandomString()
	codeExchangeKey := generateRandomString()
	tokenSigningSecret := generateRandomString()

	return &Keychain{
		EncryptSalt:        string(encryptSalt),
		CodeExchangeKey:    string(codeExchangeKey),
		TokenSigningSecret: tokenSigningSecret,
	}
}

// generateRandomString generates a 256-bit random string
func generateRandomString() []byte {
	b := make([]byte, 32)
	for i := range b {
		b[i] = library[rand.Intn(len(library))]
	}

	return b
}
