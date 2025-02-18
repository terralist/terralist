package gitlab

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
	//ClientID for the provider
	ClientID string

	//Client secret for the provider
	ClientSecret string

	//RedirectURL must be exactly the same as configured in Gitlab
	RedirectURL string

	//GitLabOAuthBaseURL contains the hostname and an optional port
	GitLabOAuthBaseURL string

	//Groups is a list of groups the user must be a member of
	Groups []string
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

var (
	httpClient          = &http.Client{}
	scope      []string = []string{"email", "openid"}
)

func (p *Provider) Name() string {
	return "GitLab"
}

func (p *Provider) GetAuthorizeUrl(state string) string {
	return fmt.Sprintf(
		"%s/authorize?client_id=%s&state=%s&response_type=code&redirect_uri=%s&scope=%s",
		p.GitLabOAuthBaseURL,
		p.ClientID,
		state,
		p.RedirectURL,
		strings.Join(scope, "+"),
	)
}

func (p *Provider) GetUserDetails(code string, user *auth.User) error {
	// https://docs.gitlab.com/ee/integration/openid_connect_provider.html
	var t tokenResponse
	if err := p.PerformAccessTokenRequest(code, &t); err != nil {
		return err
	}
	userdata, err := p.PerformUserProfileRequest(t)
	if err != nil {
		return err
	}
	if name, ok := userdata["name"].(string); ok {
		user.Name = name
	} else {
		return fmt.Errorf("name not found in user data")
	}

	if email, ok := userdata["email"].(string); ok {
		user.Email = email
	} else {
		return fmt.Errorf("email not found in user data")
	}

	// Check if the user is a member of the required groups from GitLab
	if len(p.Groups) > 0 {
		userGroups := userdata["groups"].([]interface{})
		for _, group := range p.Groups {
			for _, userGroup := range userGroups {
				if group == userGroup.(string) {
					return nil
				}
			}
		}
		return fmt.Errorf("user is not a member of the required groups")
	}

	return nil
}

func (p *Provider) PerformAccessTokenRequest(code string, t *tokenResponse) error {
	accessTokenUrl := fmt.Sprintf(
		"%s/token",
		p.GitLabOAuthBaseURL,
	)

	reqBody := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {p.RedirectURL},
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

func (p *Provider) PerformUserProfileRequest(t tokenResponse) (map[string]interface{}, error) {
	userEndpoint := fmt.Sprintf("%s/userinfo", p.GitLabOAuthBaseURL)

	req, err := http.NewRequest(http.MethodGet, userEndpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.AccessToken))

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Gitlab responded with status %d", res.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}
