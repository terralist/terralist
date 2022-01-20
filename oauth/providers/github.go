package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/valentindeaconu/terralist/oauth"
)

var (
	oauthEndpoint = "https://github.com/login/oauth"
	apiEndpoint   = "https://api.github.com"
	httpClient    = &http.Client{}
)

type GitHubOAuthProvider struct {
	ClientID     string
	ClientSecret string
	Organization string
}

type GitHubOAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func NewGitHubOAuthProvider(clientID string, clientSecret string, organization string) *GitHubOAuthProvider {
	return &GitHubOAuthProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Organization: organization,
	}
}

func (m *GitHubOAuthProvider) GetAuthorizeUrl(state string) string {
	scopes := []string{"read:user", "user:email"}

	if m.Organization != "" {
		scopes = append(scopes, "read:org")
	}

	scope := strings.Join(scopes, " ")

	return fmt.Sprintf(
		"%s/authorize?client_id=%s&state=%s&scope=%s",
		oauthEndpoint,
		m.ClientID,
		state,
		url.QueryEscape(scope),
	)
}

func (m *GitHubOAuthProvider) GetUserDetails(code string, user *oauth.UserDetails) error {
	var t GitHubOAuthTokenResponse
	if err := m.PerformAccessTokenRequest(code, &t); err != nil {
		return err
	}

	name, err := m.PerformUserNameRequest(t)
	if err != nil {
		return err
	}

	email, err := m.PerformUserEmailRequest(t)
	if err != nil {
		return err
	}

	user.Name = name
	user.Email = email

	return nil
}

func (m GitHubOAuthProvider) PerformAccessTokenRequest(code string, t *GitHubOAuthTokenResponse) error {
	accessTokenUrl := fmt.Sprintf(
		"%s/access_token?client_id=%s&client_secret=%s&code=%s",
		oauthEndpoint,
		m.ClientID,
		m.ClientSecret,
		code,
	)

	req, err := http.NewRequest(http.MethodPost, accessTokenUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(t); err != nil {
		return err
	}

	return nil
}

func (m GitHubOAuthProvider) PerformUserNameRequest(t GitHubOAuthTokenResponse) (string, error) {
	userEndpoint := fmt.Sprintf("%s/user", apiEndpoint)

	req, err := http.NewRequest(http.MethodGet, userEndpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", t.AccessToken))

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("GitHub responded with status %d", res.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", err
	}

	return data["name"].(string), nil
}

func (m GitHubOAuthProvider) PerformUserEmailRequest(t GitHubOAuthTokenResponse) (string, error) {
	emailsEndpoint := fmt.Sprintf("%s/user/emails", apiEndpoint)

	req, err := http.NewRequest(http.MethodGet, emailsEndpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", t.AccessToken))

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("GitHub responded with status %d", res.StatusCode)
	}

	var data []map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", err
	}

	var verifiedEmail string = ""
	for _, e := range data {
		if verifiedEmail == "" && e["primary"].(bool) {
			verifiedEmail = e["email"].(string)
		}
	}

	if verifiedEmail == "" {
		return "", fmt.Errorf("access could not be granted, no email information found")
	}

	return verifiedEmail, nil
}
