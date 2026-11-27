package services

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"terralist/internal/server/models/oauth"
	"terralist/internal/server/repositories"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/database"
	"terralist/pkg/database/factory"
	"terralist/pkg/database/sqlite"
	"terralist/pkg/secret"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRefresher is an upstream provider that rotates its refresh tokens.
type fakeRefresher struct {
	user        auth.User
	refreshErr  error
	refreshes   int
	lastGrant   auth.Grant
	grantPrefix string
}

func (f *fakeRefresher) Name() string                  { return "Fake" }
func (f *fakeRefresher) GetAuthorizeUrl(string) string { return "" }

func (f *fakeRefresher) GetUserDetails(code string, user *auth.User) error {
	_, err := f.ExchangeCode(code, user)
	return err
}

func (f *fakeRefresher) ExchangeCode(_ string, user *auth.User) (*auth.Grant, error) {
	*user = f.user
	return &auth.Grant{AccessToken: "upstream-access", RefreshToken: f.grantPrefix + "0"}, nil
}

func (f *fakeRefresher) Refresh(grant auth.Grant, user *auth.User) (*auth.Grant, error) {
	f.lastGrant = grant
	if f.refreshErr != nil {
		return nil, f.refreshErr
	}

	f.refreshes++
	*user = f.user

	return &auth.Grant{AccessToken: "upstream-access", RefreshToken: fmt.Sprintf("%s%d", f.grantPrefix, f.refreshes)}, nil
}

type refreshFixture struct {
	service    *DefaultLoginService
	provider   *fakeRefresher
	repository *repositories.DefaultOAuthRefreshTokenRepository
	jwt        jwt.JWT
	now        time.Time
}

func newRefreshFixture(t *testing.T) *refreshFixture {
	t.Helper()

	engine, err := factory.NewDatabase(database.SQLITE, &sqlite.Config{
		Path: filepath.Join(t.TempDir(), "test.db"),
	})
	require.NoError(t, err)
	require.NoError(t, engine.Handler().AutoMigrate(&oauth.Code{}, &oauth.RefreshToken{}))

	sealer, err := secret.NewSealer("oauth-refresh-secret-of-at-least-32-chars")
	require.NoError(t, err)

	jwtManager, err := jwt.New("refresh-test-signing-secret")
	require.NoError(t, err)

	f := &refreshFixture{
		provider: &fakeRefresher{
			user:        auth.User{Name: "jdoe", Email: "jdoe@example.com", Groups: []string{"platform"}},
			grantPrefix: "upstream-refresh-",
		},
		repository: &repositories.DefaultOAuthRefreshTokenRepository{Database: engine},
		jwt:        jwtManager,
		now:        time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC),
	}

	f.service = &DefaultLoginService{
		Provider: f.provider,
		JWT:      jwtManager,
		CodeRepository: &repositories.DefaultOAuthCodeRepository{
			Database: engine,
			TTL:      2 * time.Minute,
		},
		TokenExpirationSecs: 86400,
		Refresh: &RefreshOptions{
			Repository:                f.repository,
			Sealer:                    sealer,
			AccessTokenExpirationSecs: 3600,
			TokenExpiration:           30 * 24 * time.Hour,
			MaxSessionAge:             90 * 24 * time.Hour,
			Now:                       func() time.Time { return f.now },
		},
	}

	return f
}

const testVerifier = "verifier-verifier-verifier-verifier-verifier"

func testChallenge() string {
	sum := sha256.Sum256([]byte(testVerifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// login runs the authorization code flow for the given client and returns
// the token response.
func (f *refreshFixture) login(t *testing.T, clientID string) *oauth.TokenResponse {
	t.Helper()

	r := &oauth.Request{
		ClientID:            clientID,
		CodeChallenge:       testChallenge(),
		CodeChallengeMethod: "S256",
		RedirectURI:         "http://localhost:10000/login",
		State:               "state",
	}

	components, erro := f.service.UnpackCode("upstream-code", r)
	require.Nil(t, erro)

	resp, erro := f.service.ValidateToken(components, testVerifier)
	require.Nil(t, erro)

	return resp
}

func TestRefresh_TerraformLoginGetsNoRefreshToken(t *testing.T) {
	f := newRefreshFixture(t)

	resp := f.login(t, oauth.TerraformClientID)

	assert.Empty(t, resp.RefreshToken)
	assert.Equal(t, 86400, resp.ExpiresIn)
}

func TestRefresh_OtherClientGetsRefreshToken(t *testing.T) {
	f := newRefreshFixture(t)

	resp := f.login(t, "oauth-client")

	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, 3600, resp.ExpiresIn)

	entry, err := f.repository.FindByHash(hashRefreshToken(resp.RefreshToken))
	require.NoError(t, err)
	assert.Equal(t, "oauth-client", entry.ClientID)
	assert.Equal(t, "jdoe", entry.Subject)
	assert.NotContains(t, entry.UpstreamGrant, "upstream-refresh-0", "the upstream grant must be sealed")
	require.NotNil(t, entry.SessionExpiresAt)
	assert.Equal(t, f.now.Add(90*24*time.Hour), entry.SessionExpiresAt.UTC())
}

func TestRefresh_RotatesTokensAndUpstreamGrant(t *testing.T) {
	f := newRefreshFixture(t)
	first := f.login(t, "oauth-client")

	f.provider.user.Groups = []string{"platform", "admins"}

	second, erro := f.service.RefreshToken(first.RefreshToken, "oauth-client")
	require.Nil(t, erro)
	assert.NotEqual(t, first.RefreshToken, second.RefreshToken)
	assert.Equal(t, "upstream-refresh-0", f.provider.lastGrant.RefreshToken)

	claims, err := f.jwt.Extract(second.AccessToken)
	require.NoError(t, err)
	assert.Contains(t, fmt.Sprintf("%v", claims), "admins", "the access token must carry the refreshed groups")

	third, erro := f.service.RefreshToken(second.RefreshToken, "oauth-client")
	require.Nil(t, erro)
	assert.NotEmpty(t, third.RefreshToken)
	assert.Equal(t, "upstream-refresh-1", f.provider.lastGrant.RefreshToken, "the rotated upstream grant must be kept")
}

func TestRefresh_ReusedTokenRevokesSession(t *testing.T) {
	f := newRefreshFixture(t)
	first := f.login(t, "oauth-client")

	second, erro := f.service.RefreshToken(first.RefreshToken, "oauth-client")
	require.Nil(t, erro)

	_, erro = f.service.RefreshToken(first.RefreshToken, "oauth-client")
	require.NotNil(t, erro)
	assert.Equal(t, oauth.InvalidGrant, erro.Kind())

	_, erro = f.service.RefreshToken(second.RefreshToken, "oauth-client")
	require.NotNil(t, erro, "the current token must stop working once the old one is reused")
	assert.Equal(t, oauth.InvalidGrant, erro.Kind())
}

func TestRefresh_WrongClientIsRefused(t *testing.T) {
	f := newRefreshFixture(t)
	first := f.login(t, "oauth-client")

	_, erro := f.service.RefreshToken(first.RefreshToken, "other-app")
	require.NotNil(t, erro)
	assert.Equal(t, oauth.InvalidGrant, erro.Kind())

	_, erro = f.service.RefreshToken(first.RefreshToken, "oauth-client")
	assert.Nil(t, erro, "a wrong client must not revoke the session")
}

func TestRefresh_UpstreamRefusalRevokesSession(t *testing.T) {
	f := newRefreshFixture(t)
	first := f.login(t, "oauth-client")

	f.provider.refreshErr = errors.New("user is not a member of the required groups")

	_, erro := f.service.RefreshToken(first.RefreshToken, "oauth-client")
	require.NotNil(t, erro)
	assert.Equal(t, oauth.InvalidGrant, erro.Kind())

	f.provider.refreshErr = nil

	_, erro = f.service.RefreshToken(first.RefreshToken, "oauth-client")
	require.NotNil(t, erro)
	assert.Equal(t, oauth.InvalidGrant, erro.Kind())
}

func TestRefresh_UpstreamOutageKeepsSession(t *testing.T) {
	f := newRefreshFixture(t)
	first := f.login(t, "oauth-client")

	f.provider.refreshErr = fmt.Errorf("%w: status 502", auth.ErrUpstreamUnavailable)

	_, erro := f.service.RefreshToken(first.RefreshToken, "oauth-client")
	require.NotNil(t, erro)
	assert.Equal(t, oauth.TemporarilyUnavailable, erro.Kind())

	f.provider.refreshErr = nil

	_, erro = f.service.RefreshToken(first.RefreshToken, "oauth-client")
	assert.Nil(t, erro, "the same token must work again once the provider is back")
}

func TestRefresh_AnotherUpstreamUserRevokesSession(t *testing.T) {
	f := newRefreshFixture(t)
	first := f.login(t, "oauth-client")

	f.provider.user.Name = "someone-else"

	_, erro := f.service.RefreshToken(first.RefreshToken, "oauth-client")
	require.NotNil(t, erro)
	assert.Equal(t, oauth.InvalidGrant, erro.Kind())

	_, err := f.repository.FindByHash(hashRefreshToken(first.RefreshToken))
	assert.ErrorIs(t, err, repositories.ErrNotFound)
}

func TestRefresh_Expiry(t *testing.T) {
	t.Run("an unused token expires", func(t *testing.T) {
		f := newRefreshFixture(t)
		first := f.login(t, "oauth-client")

		f.now = f.now.Add(31 * 24 * time.Hour)

		_, erro := f.service.RefreshToken(first.RefreshToken, "oauth-client")
		require.NotNil(t, erro)
		assert.Equal(t, oauth.InvalidGrant, erro.Kind())
	})

	t.Run("a session ends at its maximum age however often it is refreshed", func(t *testing.T) {
		f := newRefreshFixture(t)
		token := f.login(t, "oauth-client").RefreshToken

		for range 4 {
			f.now = f.now.Add(25 * 24 * time.Hour)

			resp, erro := f.service.RefreshToken(token, "oauth-client")
			if f.now.After(time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC).Add(90 * 24 * time.Hour)) {
				require.NotNil(t, erro)
				assert.Equal(t, oauth.InvalidGrant, erro.Kind())
				return
			}

			require.Nil(t, erro)
			token = resp.RefreshToken
		}

		t.Fatal("the session outlived its maximum age")
	})
}

func TestRevokeToken(t *testing.T) {
	f := newRefreshFixture(t)
	first := f.login(t, "oauth-client")

	assert.Nil(t, f.service.RevokeToken(first.RefreshToken, "other-app"))
	_, err := f.repository.FindByHash(hashRefreshToken(first.RefreshToken))
	require.NoError(t, err, "another client must not revoke the session")

	assert.Nil(t, f.service.RevokeToken(first.RefreshToken, "oauth-client"))

	_, erro := f.service.RefreshToken(first.RefreshToken, "oauth-client")
	require.NotNil(t, erro)
	assert.Equal(t, oauth.InvalidGrant, erro.Kind())

	assert.Nil(t, f.service.RevokeToken("unknown", "oauth-client"))
}

func TestRefresh_DisabledService(t *testing.T) {
	service := &DefaultLoginService{Provider: &fakeRefresher{}}

	assert.False(t, service.RefreshEnabled())

	_, erro := service.RefreshToken("token", "oauth-client")
	require.NotNil(t, erro)
	assert.Equal(t, oauth.UnsupportedGrantType, erro.Kind())
}
