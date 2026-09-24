package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"terralist/pkg/auth/jwt"
)

// downloadTokenExpiry bounds how long a download link stays valid, in seconds.
// Terraform downloads an artifact right after reading the document that lists
// it.
const downloadTokenExpiry = 15 * 60

// DownloadTokens issues and checks the capability tokens carried by artifact
// download links. Terraform downloads artifacts without credentials, so the
// link itself proves that the document it came from was served to a caller
// allowed to read the artifact, and whether that caller may fetch it from the
// upstream registry. The subject names the artifact the token is bound to.
type DownloadTokens struct {
	jwt jwt.JWT
}

// downloadTokenPayload uses keys that no user token carries, so that the JWT
// helper never mistakes a download token for a user.
type downloadTokenPayload struct {
	Subject string `json:"dl_subject"`
	Fetch   bool   `json:"dl_fetch"`
}

// NewDownloadTokens derives a signing key for download tokens from the token
// signing secret, so that download tokens and user tokens never verify
// against each other.
func NewDownloadTokens(secret string) (*DownloadTokens, error) {
	key := sha256.Sum256([]byte("terralist-download-token:" + secret))

	signer, err := jwt.New(hex.EncodeToString(key[:]))
	if err != nil {
		return nil, err
	}

	return &DownloadTokens{jwt: signer}, nil
}

// Sign issues a token for one artifact, recording whether the bearer may
// fetch it from the upstream registry.
func (t *DownloadTokens) Sign(subject string, fetch bool) (string, error) {
	return t.jwt.Build(downloadTokenPayload{Subject: subject, Fetch: fetch}, downloadTokenExpiry)
}

// Verify checks a token against the artifact it is presented for and returns
// the fetch permission it carries.
func (t *DownloadTokens) Verify(token, subject string) (bool, bool) {
	data, err := t.jwt.Extract(token)
	if err != nil {
		return false, false
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return false, false
	}

	var payload downloadTokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return false, false
	}

	if payload.Subject == "" || payload.Subject != subject {
		return false, false
	}

	return payload.Fetch, true
}
