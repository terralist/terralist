package oidc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
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
	ScopesSupported       []string `json:"scopes_supported"`
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
		c.applyDiscoveredConfiguration()
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

func (c *Config) applyDiscoveredConfiguration() {
	document, err := c.discoverConfiguration()
	if err != nil {
		log.Warn().
			Str("oidc_host", c.Host).
			Err(err).
			Msg("OIDC discovery failed; falling back to manual endpoint configuration")
		return
	}

	c.applyDiscoveredEndpoint("authorization_endpoint", document.AuthorizationEndpoint, &c.AuthorizeUrl)
	c.applyDiscoveredEndpoint("token_endpoint", document.TokenEndpoint, &c.TokenUrl)
	c.applyDiscoveredEndpoint("userinfo_endpoint", document.UserInfoEndpoint, &c.UserInfoUrl)

	c.warnOnScopeAdvertisement(document)
}

func (c *Config) discoverConfiguration() (*discoveryDocument, error) {
	discoveryURL, err := buildDiscoveryURL(c.Host)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create OIDC discovery request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	res, err := discoveryHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oidc discovery request responded with status %d", res.StatusCode)
	}

	var document discoveryDocument
	if err := json.NewDecoder(res.Body).Decode(&document); err != nil {
		return nil, fmt.Errorf("decode OIDC discovery response: %w", err)
	}

	return &document, nil
}

func (c *Config) applyDiscoveredEndpoint(endpointName string, discoveredValue string, manualValue *string) {
	if discoveredValue == "" {
		return
	}

	if *manualValue != "" {
		log.Warn().
			Str("oidc_host", c.Host).
			Str("endpoint", endpointName).
			Msg("OIDC discovery provided an endpoint; ignoring the manual override")
	}

	*manualValue = discoveredValue
}

func (c *Config) warnOnScopeAdvertisement(document *discoveryDocument) {
	if len(document.ScopesSupported) == 0 {
		log.Warn().
			Str("oidc_host", c.Host).
			Msg("OIDC provider does not advertise supported scopes; authentication may not work as expected")
		return
	}

	missingScopes := []string{}
	for _, requiredScope := range requiredScopes {
		if !slices.Contains(document.ScopesSupported, requiredScope) {
			missingScopes = append(missingScopes, requiredScope)
		}
	}

	if len(missingScopes) > 0 {
		log.Warn().
			Str("oidc_host", c.Host).
			Strs("missing_scopes", missingScopes).
			Msg("OIDC provider does not advertise all required scopes; authentication may not work as expected")
	}
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
