package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformUserEmailRequest(t *testing.T) {
	emails := []map[string]any{
		{"email": "personal@gmail.com", "primary": true},
		{"email": "work@company.com", "primary": false},
		{"email": "alt@company.com", "primary": false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(emails)
	}))
	defer server.Close()

	t.Run("returns primary email when no preferred domain", func(t *testing.T) {
		p := &Provider{apiEndpoint: server.URL}
		email, err := p.PerformUserEmailRequest(tokenResponse{AccessToken: "test"})
		require.NoError(t, err)
		assert.Equal(t, "personal@gmail.com", email)
	})

	t.Run("returns preferred domain email when set", func(t *testing.T) {
		p := &Provider{apiEndpoint: server.URL, PreferredEmailDomain: "company.com"}
		email, err := p.PerformUserEmailRequest(tokenResponse{AccessToken: "test"})
		require.NoError(t, err)
		assert.Equal(t, "work@company.com", email)
	})

	t.Run("falls back to primary when preferred domain has no match", func(t *testing.T) {
		p := &Provider{apiEndpoint: server.URL, PreferredEmailDomain: "other.org"}
		email, err := p.PerformUserEmailRequest(tokenResponse{AccessToken: "test"})
		require.NoError(t, err)
		assert.Equal(t, "personal@gmail.com", email)
	})
}

func TestNextPageURL(t *testing.T) {
	t.Run("empty header returns empty string", func(t *testing.T) {
		assert.Equal(t, "", nextPageURL(""))
	})

	t.Run("header with next link returns the URL", func(t *testing.T) {
		header := `<https://api.github.com/orgs/my-org/teams?per_page=100&page=2>; rel="next", <https://api.github.com/orgs/my-org/teams?per_page=100&page=5>; rel="last"`
		assert.Equal(t, "https://api.github.com/orgs/my-org/teams?per_page=100&page=2", nextPageURL(header))
	})

	t.Run("header without next link returns empty string", func(t *testing.T) {
		header := `<https://api.github.com/orgs/my-org/teams?per_page=100&page=4>; rel="prev", <https://api.github.com/orgs/my-org/teams?per_page=100&page=1>; rel="first"`
		assert.Equal(t, "", nextPageURL(header))
	})

	t.Run("header with only next link", func(t *testing.T) {
		header := `<https://api.github.com/orgs/my-org/teams?per_page=100&page=2>; rel="next"`
		assert.Equal(t, "https://api.github.com/orgs/my-org/teams?per_page=100&page=2", nextPageURL(header))
	})
}
