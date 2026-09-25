package services

import (
	"bytes"
	"testing"
	"time"

	"terralist/internal/server/models/authority"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// testSigner is a signing identity an authority can hold the key of.
type testSigner struct {
	entity *openpgp.Entity
	// signedAt is when signatures are made: within the lifetime of an
	// expired key, now otherwise.
	signedAt func() time.Time
}

// newTestSigner generates a signing identity, whose key expires after
// lifetime when it is positive.
func newTestSigner(t *testing.T, lifetime time.Duration) *testSigner {
	t.Helper()

	config := &packet.Config{}
	signedAt := time.Now
	if lifetime > 0 {
		// Backdate the key so that it is already expired, and sign while it
		// was still valid.
		created := time.Now().Add(-2 * lifetime)
		config.Time = func() time.Time { return created }
		config.KeyLifetimeSecs = uint32(lifetime.Seconds())
		signedAt = func() time.Time { return created.Add(time.Second) }
	}

	entity, err := openpgp.NewEntity("Signer", "", "signer@example.com", config)
	if err != nil {
		t.Fatalf("could not generate key: %v", err)
	}

	return &testSigner{entity: entity, signedAt: signedAt}
}

// key returns the public key of the signer as an authority key.
func (s *testSigner) key(t *testing.T) authority.Key {
	t.Helper()

	var armored bytes.Buffer
	encoder, err := armor.Encode(&armored, openpgp.PublicKeyType, nil)
	if err != nil {
		t.Fatalf("could not armor key: %v", err)
	}
	if err := s.entity.Serialize(encoder); err != nil {
		t.Fatalf("could not serialize key: %v", err)
	}
	if err := encoder.Close(); err != nil {
		t.Fatalf("could not close armor: %v", err)
	}

	return authority.Key{KeyId: s.entity.PrimaryKey.KeyIdString(), AsciiArmor: armored.String()}
}

// sign returns a binary detached signature of the document.
func (s *testSigner) sign(t *testing.T, document []byte) []byte {
	t.Helper()

	var signature bytes.Buffer
	config := &packet.Config{Time: s.signedAt}
	if err := openpgp.DetachSign(&signature, s.entity, bytes.NewReader(document), config); err != nil {
		t.Fatalf("could not sign: %v", err)
	}

	return signature.Bytes()
}
