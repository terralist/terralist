package gitlab

import (
	"net/url"
	"testing"

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
