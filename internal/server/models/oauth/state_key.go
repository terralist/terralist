package oauth

// DeriveStateKey computes the key used to sign the OAuth state payload when
// no dedicated secret is configured, by keying HMAC-SHA256 with the token
// signing secret over a fixed label.
func DeriveStateKey(tokenSigningSecret string) []byte {
	return sign([]byte(tokenSigningSecret), []byte("oauth-state"))
}
