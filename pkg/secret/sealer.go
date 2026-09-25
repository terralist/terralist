// Package secret encrypts values that must be stored but never read back in
// plaintext by anyone without the server secret, such as upstream tokens.
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// MinSecretLength is the shortest secret a Sealer accepts.
const MinSecretLength = 32

// keySalt separates the keys derived by the Sealer from any other use of the
// same secret.
var keySalt = []byte("terralist-sealer")

// Sealer encrypts and decrypts values with a key derived from a secret.
type Sealer struct {
	key []byte
}

// NewSealer derives an AES-256 key from the secret with Argon2id, refusing
// secrets shorter than MinSecretLength.
func NewSealer(secret string) (*Sealer, error) {
	if len(secret) < MinSecretLength {
		return nil, fmt.Errorf("the secret must be at least %d characters long", MinSecretLength)
	}

	return &Sealer{key: argon2.IDKey([]byte(secret), keySalt, 3, 64*1024, 4, 32)}, nil
}

// Seal encrypts the value with AES-GCM under a random nonce, bound to the
// scope, and returns the nonce and ciphertext as one base64 string. The value
// only opens under the same scope.
func (s *Sealer) Seal(value, scope string) (string, error) {
	aead, err := s.aead()
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("could not generate nonce: %w", err)
	}

	sealed := aead.Seal(nonce, nonce, []byte(value), []byte(scope))

	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Open decrypts a value produced by Seal under the same scope.
func (s *Sealer) Open(sealed, scope string) (string, error) {
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

	value, err := aead.Open(nil, nonce, ciphertext, []byte(scope))
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
