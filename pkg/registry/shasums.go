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

// VerifyShaSums checks the detached signature of a SHA256SUMS document against
// the offered keys, the way Terraform does before trusting the document, and
// returns the id of the key that issued the signature. The signature may be
// binary or armored. An expired key is accepted, since the signature of an
// older release was valid when the registry vouched for it.
func VerifyShaSums(document, signature []byte, keys []GPGPublicKey) (string, error) {
	for _, key := range keys {
		keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(key.ASCIIArmor))
		if err != nil {
			return "", fmt.Errorf("could not decode signing key %s: %w", key.KeyID, err)
		}

		entity, err := checkDetachedSignature(keyring, document, signature)
		if errors.Is(err, openpgpErrors.ErrUnknownIssuer) {
			continue
		}

		if err != nil {
			return "", fmt.Errorf("could not verify the SHA256SUMS signature: %w", err)
		}

		return entity.PrimaryKey.KeyIdString(), nil
	}

	return "", ErrUnknownIssuer
}

// checkDetachedSignature verifies a binary or armored detached signature.
func checkDetachedSignature(keyring openpgp.KeyRing, document, signature []byte) (*openpgp.Entity, error) {
	entity, err := openpgp.CheckDetachedSignature(keyring, bytes.NewReader(document), bytes.NewReader(signature), nil)
	if err != nil && !errors.Is(err, openpgpErrors.ErrUnknownIssuer) && !errors.Is(err, openpgpErrors.ErrKeyExpired) {
		entity, err = openpgp.CheckArmoredDetachedSignature(keyring, bytes.NewReader(document), bytes.NewReader(signature), nil)
	}

	if errors.Is(err, openpgpErrors.ErrKeyExpired) && entity != nil {
		return entity, nil
	}

	return entity, err
}
