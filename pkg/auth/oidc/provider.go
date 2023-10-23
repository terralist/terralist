package oidc

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
	AuthorizeUrl string
	TokenUrl     string
	UserInfoUrl  string
	RedirectUrl  string
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

var (
	httpClient          = &http.Client{}
	scope      []string = []string{"openid"}
)

func (p *Provider) Name() string {
	return "OIDC"
}

func (p *Provider) GetAuthorizeUrl(state string) string {
	return fmt.Sprintf(
		"%s?client_id=%s&state=%s&response_type=code&redirect_uri=%s&scope=%s",
		p.AuthorizeUrl,
		p.ClientID,
		state,
		p.RedirectUrl,
		strings.Join(scope, "+"),
	)
}

func (p *Provider) GetUserDetails(code string, user *auth.User) error {
	// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-group-users
	var t tokenResponse
	if err := p.PerformAccessTokenRequest(code, &t); err != nil {
		return err
	}
	/*
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
		}*/

	user.Name = "twichers"
	user.Email = "torben.wichers@swisslife.de"

	return nil
}

func (p *Provider) PerformAccessTokenRequest(code string, t *tokenResponse) error {
	reqBody := url.Values{
		"grant_type": {"authorization_code"},
		"code":       {code},
	}
	req, err := http.NewRequest(http.MethodPost, p.TokenUrl, strings.NewReader(reqBody.Encode()))
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

/*
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
*/
