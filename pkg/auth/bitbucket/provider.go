package bitbucket

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
	Workspace    string
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

var (
	oauthEndpoint = "https://bitbucket.org/site/oauth2"
	apiEndpoint   = "https://api.bitbucket.org/2.0"
	httpClient    = &http.Client{}
)

func (p *Provider) Name() string {
	return "BitBucket"
}

func (p *Provider) GetAuthorizeUrl(state string) string {
	return fmt.Sprintf(
		"%s/authorize?client_id=%s&state=%s&response_type=code",
		oauthEndpoint,
		p.ClientID,
		state,
	)
}

func (p *Provider) GetUserDetails(code string, user *auth.User) error {
	// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-group-users
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

	if p.Workspace != "" {
		if err := p.PerformCheckUserMemberInWorkspace(t); err != nil {
			return err
		}
	}

	user.Name = name
	user.Email = email

	return nil
}

func (p *Provider) PerformAccessTokenRequest(code string, t *tokenResponse) error {
	// https://developer.atlassian.com/cloud/bitbucket/oauth-2/
	accessTokenUrl := fmt.Sprintf(
		"%s/access_token",
		oauthEndpoint,
	)

	reqBody := url.Values{
		"grant_type": {"authorization_code"},
		"code":       {code},
	}
	req, err := http.NewRequest(http.MethodPost, accessTokenUrl, strings.NewReader(reqBody.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(p.ClientID, p.ClientSecret)

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

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.AccessToken))

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("BitBucket responded with status %d", res.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", err
	}

	return data["username"].(string), nil
}

func (p *Provider) PerformUserEmailRequest(t tokenResponse) (string, error) {
	emailsEndpoint := fmt.Sprintf("%s/user/emails", apiEndpoint)

	req, err := http.NewRequest(http.MethodGet, emailsEndpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.AccessToken))

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("BitBucket responded with status %d", res.StatusCode)
	}

	type bbEmailResponse struct {
		Values []struct {
			Type      string `json:"type,omitempty"`
			Email     string `json:"email,omitempty"`
			IsPrimary bool   `json:"is_primary,omitempty"`
		} `json:"values"`
	}

	var data bbEmailResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", err
	}

	var verifiedEmail string
	for _, e := range data.Values {
		if verifiedEmail == "" && e.IsPrimary {
			verifiedEmail = e.Email
		}
	}

	if verifiedEmail == "" {
		return "", fmt.Errorf("access could not be granted, no email information found")
	}

	return verifiedEmail, nil
}

func (p *Provider) PerformCheckUserMemberInWorkspace(t tokenResponse) error {
	orgEndpoint := fmt.Sprintf("%s/user/permissions/workspaces", apiEndpoint)

	req, err := http.NewRequest(http.MethodGet, orgEndpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.AccessToken))

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("could not access %s endpoint: %d", orgEndpoint, res.StatusCode)
	}

	var data struct {
		Values []map[string]interface{}
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return err
	}

	var isMember bool
	for _, e := range data.Values {
		workspace, ok := e["workspace"].(map[string]interface{})
		if !ok {
			continue
		}

		slug, _ := workspace["slug"]
		wsName, _ := workspace["name"]
		if slug == p.Workspace || wsName == p.Workspace {
			isMember = true
			break
		}
	}

	if !isMember {
		return fmt.Errorf("user is not member of workspace %s", p.Workspace)
	}

	return nil
}
