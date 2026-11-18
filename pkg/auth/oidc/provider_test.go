package oidc

import (
	"encoding/base64"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"terralist/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeIDToken(t *testing.T, claims map[string]any) string {
	t.Helper()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	require.NoError(t, err)

	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
}

func newTestProvider(t *testing.T, userInfo map[string]any, idToken string) *Provider {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"access_token": "access-token",
			"id_token":     idToken,
		})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer access-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(userInfo)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	originalClient := httpClient
	httpClient = server.Client()
	t.Cleanup(func() {
		httpClient = originalClient
	})

	return &Provider{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenUrl:     server.URL + "/token",
		UserInfoUrl:  server.URL + "/userinfo",
		RedirectUrl:  "https://terralist.example.com/v1/api/auth/redirect",
	}
}

func TestProviderGetAuthorizeURL_AppendsAdditionalScopes(t *testing.T) {
	provider := &Provider{
		ClientID:     "client-id",
		AuthorizeUrl: "https://issuer.example.com/auth",
		RedirectUrl:  "https://terralist.example.com/v1/api/auth/redirect",
		Scopes:       []string{"groups", "email", "profile"},
	}

	parsed, err := url.Parse(provider.GetAuthorizeUrl("state-value"))
	require.NoError(t, err)

	assert.Equal(t, "openid email groups profile", parsed.Query().Get("scope"))
}

func TestProviderGetUserDetails_WithoutGroupsClaim(t *testing.T) {
	provider := newTestProvider(t, map[string]any{
		"sub":    "user-1",
		"email":  "user@example.com",
		"groups": []string{"admins"},
	}, "")

	var user auth.User
	require.NoError(t, provider.GetUserDetails("code", &user))

	assert.Equal(t, "user-1", user.Name)
	assert.Equal(t, "user@example.com", user.Email)
	assert.Empty(t, user.Groups)
}

func TestProviderGetUserDetails_MissingRequiredClaims(t *testing.T) {
	t.Run("sub", func(t *testing.T) {
		provider := newTestProvider(t, map[string]any{"email": "user@example.com"}, "")
		var user auth.User
		assert.EqualError(t, provider.GetUserDetails("code", &user), "no user provided")
	})

	t.Run("email", func(t *testing.T) {
		provider := newTestProvider(t, map[string]any{"sub": "user-1"}, "")
		var user auth.User
		assert.EqualError(t, provider.GetUserDetails("code", &user), "no email provided")
	})
}

func TestProviderGetUserDetails_Groups(t *testing.T) {
	tests := []struct {
		name          string
		groupsClaim   string
		userInfo      map[string]any
		idTokenClaims map[string]any
		want          []string
	}{
		{
			name:        "array in userinfo",
			groupsClaim: "groups",
			userInfo:    map[string]any{"groups": []string{"admins", "developers"}},
			want:        []string{"admins", "developers"},
		},
		{
			name:        "single string in userinfo",
			groupsClaim: "groups",
			userInfo:    map[string]any{"groups": "admins"},
			want:        []string{"admins"},
		},
		{
			name:        "claim name containing dots",
			groupsClaim: "https://example.com/groups",
			userInfo:    map[string]any{"https://example.com/groups": []string{"admins"}},
			want:        []string{"admins"},
		},
		{
			name:        "nested claim path",
			groupsClaim: "realm_access.roles",
			userInfo: map[string]any{
				"realm_access": map[string]any{"roles": []string{"admins", "developers"}},
			},
			want: []string{"admins", "developers"},
		},
		{
			name:        "object claim uses its keys",
			groupsClaim: "urn:zitadel:iam:org:project:roles",
			userInfo: map[string]any{
				"urn:zitadel:iam:org:project:roles": map[string]any{
					"developers": map[string]any{"123": "example.com"},
					"admins":     map[string]any{"123": "example.com"},
				},
			},
			want: []string{"admins", "developers"},
		},
		{
			name:          "falls back to id token",
			groupsClaim:   "cognito:groups",
			idTokenClaims: map[string]any{"cognito:groups": []string{"admins"}},
			want:          []string{"admins"},
		},
		{
			name:          "userinfo takes precedence over id token",
			groupsClaim:   "groups",
			userInfo:      map[string]any{"groups": []string{"from-userinfo"}},
			idTokenClaims: map[string]any{"groups": []string{"from-id-token"}},
			want:          []string{"from-userinfo"},
		},
		{
			name:        "missing claim yields no groups",
			groupsClaim: "groups",
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userInfo := map[string]any{"sub": "user-1", "email": "user@example.com"}
			maps.Copy(userInfo, tt.userInfo)

			idToken := ""
			if tt.idTokenClaims != nil {
				idToken = fakeIDToken(t, tt.idTokenClaims)
			}

			provider := newTestProvider(t, userInfo, idToken)
			provider.GroupsClaim = tt.groupsClaim

			var user auth.User
			require.NoError(t, provider.GetUserDetails("code", &user))

			assert.Equal(t, tt.want, user.Groups)
		})
	}
}

func TestProviderGetUserDetails_InvalidGroups(t *testing.T) {
	t.Run("non-string array item", func(t *testing.T) {
		provider := newTestProvider(t, map[string]any{
			"sub":    "user-1",
			"email":  "user@example.com",
			"groups": []any{"admins", 42},
		}, "")
		provider.GroupsClaim = "groups"

		var user auth.User
		assert.ErrorContains(t, provider.GetUserDetails("code", &user), `oidc claim "groups" must contain only strings`)
	})

	t.Run("unsupported type", func(t *testing.T) {
		provider := newTestProvider(t, map[string]any{
			"sub":    "user-1",
			"email":  "user@example.com",
			"groups": 42,
		}, "")
		provider.GroupsClaim = "groups"

		var user auth.User
		assert.ErrorContains(t, provider.GetUserDetails("code", &user), `oidc claim "groups" has unsupported type`)
	})

	t.Run("malformed id token", func(t *testing.T) {
		provider := newTestProvider(t, map[string]any{
			"sub":   "user-1",
			"email": "user@example.com",
		}, "not-a-jwt")
		provider.GroupsClaim = "groups"

		var user auth.User
		assert.EqualError(t, provider.GetUserDetails("code", &user), "malformed oidc id token")
	})
}

func TestCreator_ParsesScopesAndGroupsClaim(t *testing.T) {
	provider, err := (&Creator{}).New(&Config{
		ClientID:                   "client-id",
		ClientSecret:               "client-secret",
		Scopes:                     "groups, profile,,offline_access",
		GroupsClaim:                " groups ",
		TerralistSchemeHostAndPort: "https://terralist.example.com/",
	})
	require.NoError(t, err)

	oidcProvider, ok := provider.(*Provider)
	require.True(t, ok)

	assert.Equal(t, []string{"groups", "profile", "offline_access"}, oidcProvider.Scopes)
	assert.Equal(t, "groups", oidcProvider.GroupsClaim)
	assert.Equal(t, "https://terralist.example.com/v1/api/auth/redirect", oidcProvider.RedirectUrl)
}

func TestProviderGetAuthorizeURL_OfflineAccess(t *testing.T) {
	scope := func(p *Provider) string {
		parsed, err := url.Parse(p.GetAuthorizeUrl("state"))
		require.NoError(t, err)
		return parsed.Query().Get("scope")
	}

	assert.Equal(t, "openid email", scope(&Provider{AuthorizeUrl: "https://issuer.example.com/auth"}))
	assert.Equal(t, "openid email groups offline_access", scope(&Provider{
		AuthorizeUrl:  "https://issuer.example.com/auth",
		Scopes:        []string{"groups"},
		OfflineAccess: true,
	}))
	assert.Equal(t, "openid email offline_access", scope(&Provider{
		AuthorizeUrl:  "https://issuer.example.com/auth",
		Scopes:        []string{"offline_access"},
		OfflineAccess: true,
	}), "offline_access must not be requested twice")
	assert.Equal(t, []string{"openid", "email"}, requiredScopes, "the required scopes must not be changed")
}

func TestProviderRefresh(t *testing.T) {
	var received url.Values

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		received = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"access_token":  "new-access",
			"refresh_token": "new-refresh",
			"id_token":      fakeIDToken(t, map[string]any{"groups": []string{"platform"}}),
		})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer new-access", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"sub": "user-1", "email": "user@example.com"})
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	provider := &Provider{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenUrl:     server.URL + "/token",
		UserInfoUrl:  server.URL + "/userinfo",
		GroupsClaim:  "groups",
	}

	var user auth.User
	grant, err := provider.Refresh(auth.Grant{RefreshToken: "old-refresh"}, &user)
	require.NoError(t, err)

	assert.Equal(t, "refresh_token", received.Get("grant_type"))
	assert.Equal(t, "old-refresh", received.Get("refresh_token"))
	assert.Equal(t, "new-refresh", grant.RefreshToken)
	assert.Equal(t, "user-1", user.Name)
	assert.Equal(t, []string{"platform"}, user.Groups, "groups must be read from the refreshed ID token")
}
