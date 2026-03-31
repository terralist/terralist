package oidc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

const wellKnownConfigurationPath = "/.well-known/openid-configuration"

var (
	discoveryHTTPClient = &http.Client{}
	requiredScopes      = []string{"openid", "email"}
)

type discoveryDocument struct {
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserInfoEndpoint      string   `json:"userinfo_endpoint"`
	SupportedScopes       []string `json:"supported_scopes"`
}

// Config implements auth.Configurator interface and
// handles the configuration parameters for OIDC authentication.
type Config struct {
	ClientID                   string
	ClientSecret               string
	Host                       string
	AuthorizeUrl               string
	TokenUrl                   string
	UserInfoUrl                string
	TerralistSchemeHostAndPort string
}

func (c *Config) SetDefaults() {}

func (c *Config) Validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("missing required client ID")
	}

	if c.ClientSecret == "" {
		return fmt.Errorf("missing required client secret")
	}

	if c.Host != "" {
		if c.AuthorizeUrl != "" || c.TokenUrl != "" || c.UserInfoUrl != "" {
			return fmt.Errorf("configure either oi-host or the manual OIDC endpoint URLs")
		}

		if err := c.discoverConfiguration(); err != nil {
			return err
		}
	}

	if c.AuthorizeUrl == "" {
		return fmt.Errorf("missing required authorize url")
	}

	if c.TokenUrl == "" {
		return fmt.Errorf("missing required token url")
	}

	if c.UserInfoUrl == "" {
		return fmt.Errorf("missing required userinfo url")
	}

	return nil
}

func (c *Config) discoverConfiguration() error {
	discoveryURL, err := buildDiscoveryURL(c.Host)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, discoveryURL, nil)
	if err != nil {
		return fmt.Errorf("create OIDC discovery request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	res, err := discoveryHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("oidc discovery request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("oidc discovery request responded with status %d", res.StatusCode)
	}

	var document discoveryDocument
	if err := json.NewDecoder(res.Body).Decode(&document); err != nil {
		return fmt.Errorf("decode OIDC discovery response: %w", err)
	}

	if document.AuthorizationEndpoint == "" {
		return fmt.Errorf("oidc discovery response missing authorization_endpoint")
	}

	if document.TokenEndpoint == "" {
		return fmt.Errorf("oidc discovery response missing token_endpoint")
	}

	if document.UserInfoEndpoint == "" {
		return fmt.Errorf("oidc discovery response missing userinfo_endpoint")
	}

	if len(document.SupportedScopes) == 0 {
		return fmt.Errorf("oidc discovery response missing supported_scopes")
	}

	for _, requiredScope := range requiredScopes {
		if !slices.Contains(document.SupportedScopes, requiredScope) {
			return fmt.Errorf("oidc provider does not support required scope %q", requiredScope)
		}
	}

	c.AuthorizeUrl = document.AuthorizationEndpoint
	c.TokenUrl = document.TokenEndpoint
	c.UserInfoUrl = document.UserInfoEndpoint

	return nil
}

func buildDiscoveryURL(host string) (string, error) {
	parsed, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("invalid oidc host %q: %w", host, err)
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid oidc host %q: expected an absolute http(s) URL", host)
	}

	parsed.RawQuery = ""
	parsed.Fragment = ""

	trimmedPath := strings.TrimRight(parsed.Path, "/")
	if !strings.HasSuffix(trimmedPath, wellKnownConfigurationPath) {
		trimmedPath += wellKnownConfigurationPath
	}

	parsed.Path = trimmedPath

	return parsed.String(), nil
}
