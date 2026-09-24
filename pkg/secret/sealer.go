// Package secret encrypts values that must be stored but never read back in
// plaintext by anyone without the server secret, such as upstream tokens.
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// Sealer encrypts and decrypts values with a key derived from a secret.
type Sealer struct {
	key []byte
}

// NewSealer derives an AES-256 key from the secret.
func NewSealer(secret string) *Sealer {
	key := sha256.Sum256([]byte(secret))

	return &Sealer{key: key[:]}
}

// Seal encrypts the value with AES-GCM under a random nonce and returns the
// nonce and ciphertext as one base64 string.
func (s *Sealer) Seal(value string) (string, error) {
	aead, err := s.aead()
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("could not generate nonce: %w", err)
	}

	sealed := aead.Seal(nonce, nonce, []byte(value), nil)

	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Open decrypts a value produced by Seal.
func (s *Sealer) Open(sealed string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(sealed)
	if err != nil {
		return "", fmt.Errorf("sealed value is not base64: %w", err)
	}

	aead, err := s.aead()
	if err != nil {
		return "", err
	}

	if len(raw) < aead.NonceSize() {
		return "", fmt.Errorf("sealed value is too short")
	}

	nonce, ciphertext := raw[:aead.NonceSize()], raw[aead.NonceSize():]

	value, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("could not open sealed value: %w", err)
	}

	return string(value), nil
}

func (s *Sealer) aead() (cipher.AEAD, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}
