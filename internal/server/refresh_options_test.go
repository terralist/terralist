package server

import (
	"path/filepath"
	"testing"

	"terralist/internal/server/models/oauth"
	"terralist/pkg/auth"
	"terralist/pkg/database"
	"terralist/pkg/database/factory"
	dbSqlite "terralist/pkg/database/sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type plainProvider struct{}

func (plainProvider) Name() string                            { return "SAML" }
func (plainProvider) GetAuthorizeUrl(string) string           { return "" }
func (plainProvider) GetUserDetails(string, *auth.User) error { return nil }

type refreshingProvider struct{ plainProvider }

func (refreshingProvider) ExchangeCode(string, *auth.User) (*auth.Grant, error) { return nil, nil }
func (refreshingProvider) Refresh(auth.Grant, *auth.User) (*auth.Grant, error)  { return nil, nil }

func TestNewRefreshOptions(t *testing.T) {
	engine, err := factory.NewDatabase(database.SQLITE, &dbSqlite.Config{
		Path: filepath.Join(t.TempDir(), "test.db"),
	})
	require.NoError(t, err)
	require.NoError(t, engine.Handler().AutoMigrate(&oauth.RefreshToken{}))

	valid := func() UserConfig {
		return UserConfig{
			TokenSigningSecret:                "token-signing-secret",
			OAuthStateSecret:                  "oauth-state-secret",
			OAuthRefreshEnabled:               true,
			OAuthRefreshSecret:                "oauth-refresh-secret-of-at-least-32-chars",
			OAuthRefreshAccessTokenExpiration: "1h",
			OAuthRefreshTokenExpiration:       "720h",
			OAuthRefreshMaxSessionAge:         "2160h",
		}
	}
	config := Config{Database: engine, Provider: refreshingProvider{}}

	t.Run("accepts a valid configuration", func(t *testing.T) {
		options, err := newRefreshOptions(valid(), config)
		require.NoError(t, err)
		assert.Equal(t, 3600, options.AccessTokenExpirationSecs)
	})

	t.Run("removes the refresh sessions when disabled", func(t *testing.T) {
		require.NoError(t, engine.Handler().Create(&oauth.RefreshToken{
			ID: "session", TokenHash: "hash", ClientID: "app", Subject: "jdoe", UpstreamGrant: "sealed",
		}).Error)

		userConfig := valid()
		userConfig.OAuthRefreshEnabled = false

		options, err := newRefreshOptions(userConfig, config)
		require.NoError(t, err)
		assert.Nil(t, options)

		var count int64
		require.NoError(t, engine.Handler().Model(&oauth.RefreshToken{}).Count(&count).Error)
		assert.Zero(t, count)
	})

	for name, tc := range map[string]struct {
		change   func(*UserConfig)
		provider auth.Provider
	}{
		"a provider that cannot refresh": {provider: plainProvider{}},
		"a missing secret":               {change: func(c *UserConfig) { c.OAuthRefreshSecret = "" }},
		"a short secret":                 {change: func(c *UserConfig) { c.OAuthRefreshSecret = "too-short" }},
		"the token signing secret": {change: func(c *UserConfig) {
			c.TokenSigningSecret = "shared-secret-of-at-least-32-characters"
			c.OAuthRefreshSecret = c.TokenSigningSecret
		}},
		"an invalid access token expiration":  {change: func(c *UserConfig) { c.OAuthRefreshAccessTokenExpiration = "soon" }},
		"a refresh token shorter than access": {change: func(c *UserConfig) { c.OAuthRefreshTokenExpiration = "30m" }},
		"a negative max session age":          {change: func(c *UserConfig) { c.OAuthRefreshMaxSessionAge = "-1h" }},
	} {
		t.Run("refuses "+name, func(t *testing.T) {
			userConfig := valid()
			if tc.change != nil {
				tc.change(&userConfig)
			}

			c := config
			if tc.provider != nil {
				c.Provider = tc.provider
			}

			_, err := newRefreshOptions(userConfig, c)
			assert.Error(t, err)
		})
	}
}
