package gitlab

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"terralist/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAuthorizeUrl(t *testing.T) {
	p := &Provider{
		ClientID:           "client-id",
		RedirectURL:        "https://terralist.example.com/v1/api/auth/redirect?x=1",
		GitLabOAuthBaseURL: "https://gitlab.example.com/oauth",
	}

	authorizeURL, err := url.Parse(p.GetAuthorizeUrl("a+b/c=&d"))
	require.NoError(t, err)

	query := authorizeURL.Query()
	assert.Equal(t, "a+b/c=&d", query.Get("state"))
	assert.Equal(t, "client-id", query.Get("client_id"))
	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, "https://terralist.example.com/v1/api/auth/redirect?x=1", query.Get("redirect_uri"))
	assert.Equal(t, "email openid", query.Get("scope"))
}

func TestGetUserDetails(t *testing.T) {
	newServer := func(userinfo map[string]any) *httptest.Server {
		mux := http.NewServeMux()
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "token"})
		})
		mux.HandleFunc("/oauth/userinfo", func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(userinfo)
		})
		return httptest.NewServer(mux)
	}

	t.Run("uses the username, not the full name", func(t *testing.T) {
		srv := newServer(map[string]any{
			"name":               "role:admin",
			"preferred_username": "jdoe",
			"email":              "jdoe@example.com",
		})
		defer srv.Close()

		p := &Provider{GitLabOAuthBaseURL: srv.URL + "/oauth"}

		var user auth.User
		require.NoError(t, p.GetUserDetails("code", &user))
		assert.Equal(t, "jdoe", user.Name)
		assert.Equal(t, "jdoe@example.com", user.Email)
	})

	t.Run("fails without a username", func(t *testing.T) {
		srv := newServer(map[string]any{
			"name":  "John Doe",
			"email": "jdoe@example.com",
		})
		defer srv.Close()

		p := &Provider{GitLabOAuthBaseURL: srv.URL + "/oauth"}

		var user auth.User
		require.Error(t, p.GetUserDetails("code", &user))
	})
}
