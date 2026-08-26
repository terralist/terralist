package e2e

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// terraformCallbackURL is the CLI-side redirect URI presented during the login
// protocol. Its host differs from the server host so the server issues an
// authorization code to be exchanged, rather than establishing a session.
const terraformCallbackURL = "http://127.0.0.1:65000/terraform-callback"

type oauthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// randomURLSafe returns nBytes of random data encoded as base64url.
func randomURLSafe(t *testing.T, nBytes int) string {
	t.Helper()

	buf := make([]byte, nBytes)
	_, err := rand.Read(buf)
	require.NoError(t, err)

	return base64.RawURLEncoding.EncodeToString(buf)
}

// obtainAuthorizationCode drives the full OAuth 2.0 login flow against the
// running server and its OIDC provider, returning the opaque authorization code
// handed back to the CLI callback together with the PKCE verifier that unlocks
// it.
func obtainAuthorizationCode(t *testing.T) (code, verifier string) {
	t.Helper()

	verifier = randomURLSafe(t, 32)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	callback, err := url.Parse(terraformCallbackURL)
	require.NoError(t, err)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	client := &http.Client{
		Jar:     jar,
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Stop before following the redirect to the CLI callback and read
			// the authorization code from its query instead.
			if req.URL.Host == callback.Host && req.URL.Path == callback.Path {
				code = req.URL.Query().Get("code")
				return http.ErrUseLastResponse
			}

			if len(via) >= 10 {
				return fmt.Errorf("too many redirects while obtaining authorization code")
			}

			return nil
		},
	}

	query := url.Values{}
	query.Set("client_id", "e2e-test-client")
	query.Set("code_challenge", challenge)
	query.Set("code_challenge_method", "S256")
	query.Set("redirect_uri", terraformCallbackURL)
	query.Set("response_type", "code")
	query.Set("state", randomURLSafe(t, 16))

	resp, err := client.Get(apiURL("/v1/auth/authorization") + "?" + query.Encode())
	require.NoError(t, err)
	defer resp.Body.Close()

	require.NotEmpty(t, code, "expected to capture an authorization code from the login flow")

	return code, verifier
}

// exchangeCode posts an authorization code to the token endpoint without
// following redirects, so both a successful token response and a redirected
// error response are observable.
func exchangeCode(t *testing.T, code, verifier string) *http.Response {
	t.Helper()

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("code_verifier", verifier)
	form.Set("redirect_uri", terraformCallbackURL)
	form.Set("client_id", "e2e-test-client")

	req, err := http.NewRequest(http.MethodPost, apiURL("/v1/auth/token"), strings.NewReader(form.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

// assertCodeRejected asserts that a token exchange was denied: no token, and an
// OAuth error redirect rather than a 200.
func assertCodeRejected(t *testing.T, resp *http.Response) {
	t.Helper()

	require.NotEqual(t, http.StatusOK, resp.StatusCode, "a rejected exchange must not return a token")
	require.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Location"), "error=invalid_request")
}

func TestAuthorizationCodeExchange(t *testing.T) {
	code, verifier := obtainAuthorizationCode(t)

	resp := exchangeCode(t, code, verifier)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var token oauthToken
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&token))
	require.NotEmpty(t, token.AccessToken)
	assert.Equal(t, "bearer", token.TokenType)

	// The minted token must be accepted as a registry credential.
	authResp := doRequest(t, http.MethodGet, apiURL("/v1/api/authorities/"), nil, map[string]string{
		"Authorization": "Bearer " + token.AccessToken,
	})
	defer authResp.Body.Close()

	assert.Equal(t, http.StatusOK, authResp.StatusCode)
}

func TestAuthorizationCodeIsSingleUse(t *testing.T) {
	code, verifier := obtainAuthorizationCode(t)

	first := exchangeCode(t, code, verifier)
	defer first.Body.Close()
	require.Equal(t, http.StatusOK, first.StatusCode)

	second := exchangeCode(t, code, verifier)
	defer second.Body.Close()
	assertCodeRejected(t, second)
}

func TestForgedAuthorizationCodeRejected(t *testing.T) {
	verifier := "terralist-e2e-poc-verifier"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	claims := map[string]any{
		"code_challenge":        challenge,
		"code_challenge_method": "S256",
		"user_name":             "attacker",
		"user_email":            "attacker@evil.test",
		"user_groups":           []string{"admins"},
	}
	data, err := json.Marshal(claims)
	require.NoError(t, err)

	// The original exploit: a self-crafted payload prefixed with filler bytes
	// of the salt's length. It must no longer resolve to a token.
	forged := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("A", 32) + "/" + string(data)))

	resp := exchangeCode(t, forged, verifier)
	defer resp.Body.Close()

	assertCodeRejected(t, resp)
}
