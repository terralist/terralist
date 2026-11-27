package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RefreshTokenResponse is the part of an OAuth 2.0 token response needed to
// refresh a grant.
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	Error        string `json:"error"`
}

// RequestRefresh sends a refresh token request to a provider token endpoint.
// Network failures and server errors are reported as ErrUpstreamUnavailable,
// any other failure as a refused grant.
func RequestRefresh(client *http.Client, req *http.Request) (*RefreshTokenResponse, error) {
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusInternalServerError {
		return nil, fmt.Errorf("%w: token endpoint responded with status %d", ErrUpstreamUnavailable, res.StatusCode)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint refused the refresh with status %d", res.StatusCode)
	}

	var t RefreshTokenResponse
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		return nil, fmt.Errorf("could not decode the token response: %w", err)
	}

	if t.Error != "" {
		return nil, fmt.Errorf("token endpoint refused the refresh: %s", t.Error)
	}

	if t.AccessToken == "" {
		return nil, fmt.Errorf("token endpoint returned no access token")
	}

	return &t, nil
}

// NextGrant builds the grant to keep after a refresh. Providers that do not
// rotate refresh tokens leave the response without one, and the previous one
// stays valid.
func NextGrant(previous Grant, t *RefreshTokenResponse) *Grant {
	next := &Grant{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
	}

	if next.RefreshToken == "" {
		next.RefreshToken = previous.RefreshToken
	}

	return next
}
