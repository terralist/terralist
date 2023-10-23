package oidc

import (
	"fmt"
)

// Config implements auth.Configurator interface and
// handles the configuration parameters for bitbucket authentication
type Config struct {
	ClientID                   string
	ClientSecret               string
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
