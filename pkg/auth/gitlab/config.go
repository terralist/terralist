package gitlab

import (
	"fmt"
)

const (
	defaultGitLabHost = "gitlab.com"
)

// Config implements auth.Configurator interface and
// handles the configuration parameters for gitlab authentication
type Config struct {
	ClientID                   string
	ClientSecret               string
	TerralistSchemeHostAndPort string
	GitlabHostWithOptionalPort string
}

func (c *Config) SetDefaults() {
	if c.GitlabHostWithOptionalPort == "" {
		c.GitlabHostWithOptionalPort = defaultGitLabHost
	}
}

func (c *Config) Validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("missing required client ID")
	}

	if c.ClientSecret == "" {
		return fmt.Errorf("missing required client secret")
	}

	return nil
}
