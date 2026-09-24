package registry

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	openpgpErrors "github.com/ProtonMail/go-crypto/openpgp/errors"
	"github.com/rs/zerolog/log"
)

var (
	// ErrUnknownIssuer is returned when none of the offered keys issued the
	// signature.
	ErrUnknownIssuer = errors.New("signature from unknown issuer")
)

// ParseShaSums reads a SHA256SUMS file into a map from file name to hex
// digest. A leading asterisk on the file name, the binary mode marker of
// sha256sum, is stripped.
func ParseShaSums(r io.Reader) (map[string]string, error) {
	sums := map[string]string{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("malformed SHA256SUMS line %q", line)
		}

		sums[strings.TrimPrefix(fields[1], "*")] = fields[0]
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("could not read the SHA256SUMS file: %w", err)
	}

	return sums, nil
}

// ErrKeyExpired is returned when the signature is valid but the key that
// issued it has passed its validity period and the verifier does not accept
// expired keys.
var ErrKeyExpired = errors.New("signing key expired")

// SignatureVerifier checks SHA256SUMS signatures against the keys a registry
// advertises, the way Terraform does before trusting the document.
type SignatureVerifier struct {
	// AcceptExpiredKeys accepts a valid signature issued by a key whose
	// validity period has ended. Registries keep advertising the key that
	// signed a release after it expired and do not re-sign old releases, so
	// rejecting expired keys refuses every release signed before the key
	// expired. Terraform accepts them, see
	// https://github.com/hashicorp/terraform/blob/a2391dea749166090ea5f6e817ca151128dc1899/internal/getproviders/package_authentication.go#L460-L479
	AcceptExpiredKeys bool
}

// VerifyShaSums checks the detached signature of a SHA256SUMS document against
// the offered keys and returns the id of the key that issued the signature.
// The signature may be binary or armored. The key expiry is checked last by
// the OpenPGP library, so an expired key is only ever reported for a signature
// that passed every other check.
func (v SignatureVerifier) VerifyShaSums(document, signature []byte, keys []GPGPublicKey) (string, error) {
	for _, key := range keys {
		keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(key.ASCIIArmor))
		if err != nil {
			return "", fmt.Errorf("could not decode signing key %s: %w", key.KeyID, err)
		}

		entity, err := checkDetachedSignature(keyring, document, signature)
		if errors.Is(err, openpgpErrors.ErrUnknownIssuer) {
			continue
		}

		if errors.Is(err, openpgpErrors.ErrKeyExpired) {
			return v.acceptExpired(entity)
		}

		if err != nil {
			return "", fmt.Errorf("could not verify the SHA256SUMS signature: %w", err)
		}

		return entity.PrimaryKey.KeyIdString(), nil
	}

	return "", ErrUnknownIssuer
}

// acceptExpired decides what to do with a valid signature from an expired key.
func (v SignatureVerifier) acceptExpired(entity *openpgp.Entity) (string, error) {
	keyID := entity.PrimaryKey.KeyIdString()

	if !v.AcceptExpiredKeys {
		return "", fmt.Errorf("%w: %s", ErrKeyExpired, keyID)
	}

	log.Warn().
		Str("key_id", keyID).
		Msg("Accepted a SHA256SUMS signature issued by an expired signing key.")

	return keyID, nil
}

// checkDetachedSignature verifies a binary or armored detached signature.
func checkDetachedSignature(keyring openpgp.KeyRing, document, signature []byte) (*openpgp.Entity, error) {
	entity, err := openpgp.CheckDetachedSignature(keyring, bytes.NewReader(document), bytes.NewReader(signature), nil)
	if err != nil && !errors.Is(err, openpgpErrors.ErrUnknownIssuer) && !errors.Is(err, openpgpErrors.ErrKeyExpired) {
		entity, err = openpgp.CheckArmoredDetachedSignature(keyring, bytes.NewReader(document), bytes.NewReader(signature), nil)
	}

	return entity, err
}
