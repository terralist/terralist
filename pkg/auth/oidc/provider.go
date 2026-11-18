package oidc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"terralist/pkg/auth"

	"github.com/rs/zerolog/log"
)

// Provider is the concrete implementation of oauth.Engine.
type Provider struct {
	ClientID     string
	ClientSecret string
	AuthorizeUrl string
	TokenUrl     string
	UserInfoUrl  string
	RedirectUrl  string

	// Scopes are requested in addition to the required OIDC scopes.
	Scopes []string
	// GroupsClaim is the claim that holds the user's groups. Groups are not
	// extracted when it is empty.
	GroupsClaim string
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
}

var (
	httpClient = &http.Client{}
)

func (p *Provider) Name() string {
	return "OIDC"
}

func (p *Provider) GetAuthorizeUrl(state string) string {
	queryParams := url.Values{
		"client_id":     []string{p.ClientID},
		"state":         []string{state},
		"response_type": []string{"code"},
		"redirect_uri":  []string{p.RedirectUrl},
		"scope":         []string{strings.Join(p.scopes(), " ")},
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

	claims, err := p.PerformUserInfoRequest(t)
	if err != nil {
		return err
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return fmt.Errorf("no user provided")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return fmt.Errorf("no email provided")
	}

	user.Name = sub
	user.Email = email

	if p.GroupsClaim != "" {
		groups, err := p.extractGroups(claims, t.IDToken)
		if err != nil {
			return err
		}
		user.Groups = groups
	}

	return nil
}

func (p *Provider) scopes() []string {
	scopes := append([]string{}, requiredScopes...)
	for _, scope := range p.Scopes {
		if !slices.Contains(scopes, scope) {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

// extractGroups reads the groups claim from the userinfo response, falling
// back to the ID token, since some providers only include groups in the latter.
func (p *Provider) extractGroups(userInfo map[string]any, idToken string) ([]string, error) {
	if value, ok := lookupClaim(userInfo, p.GroupsClaim); ok {
		return parseGroups(p.GroupsClaim, value)
	}

	if idToken != "" {
		idClaims, err := decodeIDTokenClaims(idToken)
		if err != nil {
			return nil, err
		}

		if value, ok := lookupClaim(idClaims, p.GroupsClaim); ok {
			return parseGroups(p.GroupsClaim, value)
		}
	}

	log.Debug().
		Str("groups_claim", p.GroupsClaim).
		Msg("OIDC groups claim not found in userinfo response or ID token")

	return nil, nil
}

// lookupClaim resolves a claim by its exact name first, so names containing
// dots (e.g. URLs) work, then as a dot-separated path into nested objects.
func lookupClaim(claims map[string]any, name string) (any, bool) {
	if value, ok := claims[name]; ok {
		return value, true
	}

	if !strings.Contains(name, ".") {
		return nil, false
	}

	var current any = claims
	for _, part := range strings.Split(name, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}

		if current, ok = object[part]; !ok {
			return nil, false
		}
	}

	return current, true
}

func parseGroups(claim string, value any) ([]string, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case string:
		if v == "" {
			return nil, nil
		}
		return []string{v}, nil
	case []any:
		groups := make([]string, 0, len(v))
		for _, item := range v {
			group, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("oidc claim %q must contain only strings", claim)
			}
			groups = append(groups, group)
		}
		return groups, nil
	case map[string]any:
		groups := make([]string, 0, len(v))
		for group := range v {
			groups = append(groups, group)
		}
		slices.Sort(groups)
		return groups, nil
	default:
		return nil, fmt.Errorf("oidc claim %q has unsupported type %T", claim, value)
	}
}

// decodeIDTokenClaims reads the payload of an ID token without verifying its
// signature. This is only safe because the token is received directly from
// the token endpoint over TLS (OIDC Core 1.0, section 3.1.3.7).
func decodeIDTokenClaims(idToken string) (map[string]any, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed oidc id token")
	}

	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	if err != nil {
		return nil, fmt.Errorf("decode oidc id token payload: %w", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("decode oidc id token claims: %w", err)
	}

	return claims, nil
}

func (p *Provider) PerformAccessTokenRequest(code string, t *tokenResponse) error {
	reqBody := url.Values{
		"client_id":     []string{p.ClientID},
		"client_secret": []string{p.ClientSecret},
		"grant_type":    []string{"authorization_code"},
		"code":          []string{code},
		"redirect_uri":  []string{p.RedirectUrl},
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

func (p *Provider) PerformUserInfoRequest(t tokenResponse) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, p.UserInfoUrl, nil)
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
		return nil, fmt.Errorf("oidc user info request responded with status %d", res.StatusCode)
	}

	var claims map[string]any
	if err := json.NewDecoder(res.Body).Decode(&claims); err != nil {
		return nil, err
	}

	return claims, nil
}
