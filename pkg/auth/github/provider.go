package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"terralist/pkg/auth"
)

// Provider is the concrete implementation of oauth.Engine
type Provider struct {
	ClientID     string
	ClientSecret string
	Organization string
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

var (
	oauthEndpoint = "https://github.com/login/oauth"
	apiEndpoint   = "https://api.github.com"
	httpClient    = &http.Client{}
)

func (p *Provider) Name() string {
	return "GitHub"
}

func (p *Provider) GetAuthorizeUrl(state string) string {
	scopes := []string{"read:user", "user:email"}

	if p.Organization != "" {
		scopes = append(scopes, "read:org")
	}

	scope := strings.Join(scopes, " ")

	return fmt.Sprintf(
		"%s/authorize?client_id=%s&state=%s&scope=%s",
		oauthEndpoint,
		p.ClientID,
		state,
		url.QueryEscape(scope),
	)
}

func (p *Provider) GetUserDetails(code string, user *auth.User) error {
	var t tokenResponse
	if err := p.PerformAccessTokenRequest(code, &t); err != nil {
		return err
	}

	name, err := p.PerformUserNameRequest(t)
	if err != nil {
		return err
	}

	email, err := p.PerformUserEmailRequest(t)
	if err != nil {
		return err
	}

	if p.Organization != "" {
		if err := p.PerformCheckUserMemberInOrganization(t); err != nil {
			return err
		}
	}

	user.Name = name
	user.Email = email

	return nil
}

func (p *Provider) PerformAccessTokenRequest(code string, t *tokenResponse) error {
	accessTokenUrl := fmt.Sprintf(
		"%s/access_token?client_id=%s&client_secret=%s&code=%s",
		oauthEndpoint,
		p.ClientID,
		p.ClientSecret,
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

func (p *Provider) PerformUserNameRequest(t tokenResponse) (string, error) {
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

	if name, ok := data["name"].(string); ok {
		return name, nil
	}

	if login, ok := data["login"].(string); ok {
		return login, nil
	}

	return "", fmt.Errorf("could not get the user or login name")
}

func (p *Provider) PerformUserEmailRequest(t tokenResponse) (string, error) {
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

func (p *Provider) PerformCheckUserMemberInOrganization(t tokenResponse) error {
	orgEndpoint := fmt.Sprintf("%s/user/memberships/orgs/%s", apiEndpoint, p.Organization)

	req, err := http.NewRequest(http.MethodGet, orgEndpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", t.AccessToken))

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("user is not a member in %s organization", p.Organization)
	}

	return nil
}
