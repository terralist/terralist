package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"terralist/internal/server/models/provider"
	"terralist/pkg/auth/jwt"
)

// packageTokenExpiry bounds how long a package download link stays valid, in
// seconds. Terraform downloads a package right after reading the document
// that lists it.
const packageTokenExpiry = 15 * 60

// PackageTokens issues and checks the capability tokens carried by package
// download links. Terraform downloads packages without credentials, so the
// link itself proves that the document it came from was served to a caller
// allowed to read the package, and whether that caller may fetch it from the
// upstream registry.
type PackageTokens struct {
	jwt jwt.JWT
}

// packageTokenPayload uses keys that no user token carries, so that the JWT
// helper never mistakes a package token for a user.
type packageTokenPayload struct {
	Namespace    string `json:"pkg_namespace"`
	Name         string `json:"pkg_name"`
	Version      string `json:"pkg_version"`
	System       string `json:"pkg_system"`
	Architecture string `json:"pkg_architecture"`
	Fetch        bool   `json:"pkg_fetch"`
}

// NewPackageTokens derives a signing key for package tokens from the token
// signing secret, so that package tokens and user tokens never verify against
// each other.
func NewPackageTokens(secret string) (*PackageTokens, error) {
	key := sha256.Sum256([]byte("terralist-package-token:" + secret))

	signer, err := jwt.New(hex.EncodeToString(key[:]))
	if err != nil {
		return nil, err
	}

	return &PackageTokens{jwt: signer}, nil
}

// Sign issues a token for one package of an authority, recording whether the
// bearer may fetch it from the upstream registry.
func (t *PackageTokens) Sign(namespace string, pkg provider.Package, fetch bool) (string, error) {
	return t.jwt.Build(packageTokenPayload{
		Namespace:    namespace,
		Name:         pkg.Name,
		Version:      pkg.Version,
		System:       pkg.System,
		Architecture: pkg.Architecture,
		Fetch:        fetch,
	}, packageTokenExpiry)
}

// Verify checks a token against the package it is presented for and returns
// the fetch permission it carries.
func (t *PackageTokens) Verify(token, namespace string, pkg provider.Package) (bool, bool) {
	data, err := t.jwt.Extract(token)
	if err != nil {
		return false, false
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return false, false
	}

	var payload packageTokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return false, false
	}

	if payload.Namespace != namespace || payload.Name != pkg.Name || payload.Version != pkg.Version ||
		payload.System != pkg.System || payload.Architecture != pkg.Architecture || payload.Name == "" {
		return false, false
	}

	return payload.Fetch, true
}
