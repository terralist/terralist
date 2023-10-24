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
	queryParams := url.Values{
		"client_id":     {p.ClientID},
		"state":         {state},
		"response_type": {"code"},
		"redirect_uri":  {p.RedirectUrl},
		"scope":         {strings.Join(scope, " ")},
	}
	return fmt.Sprintf(
		"%s?%s",
		p.AuthorizeUrl,
		queryParams.Encode(),
	)
}

func (p *Provider) GetUserDetails(code string, user *auth.User) error {
	var t tokenResponse
	if err := p.PerformAccessTokenRequest(code, &t); err != nil {
		return err
	}

	name, email, err := p.PerformUserInfoRequest(t)
	if err != nil {
		return err
	}

	user.Name = name
	user.Email = email

	return nil
}

func (p *Provider) PerformAccessTokenRequest(code string, t *tokenResponse) error {
	reqBody := url.Values{
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {p.RedirectUrl},
	}
	req, err := http.NewRequest(http.MethodPost, p.TokenUrl, strings.NewReader(reqBody.Encode()))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := httpClient.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("oidc token request responded with status %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(t); err != nil {
		return err
	}

	return nil
}

func (p *Provider) PerformUserInfoRequest(t tokenResponse) (string, string, error) {
	req, err := http.NewRequest(http.MethodGet, p.UserInfoUrl, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.AccessToken))

	res, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("oidc user info request responded with status %d", res.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", "", err
	}

	return data["sub"].(string), data["email"].(string), nil
}
