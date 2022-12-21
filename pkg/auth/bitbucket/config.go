package bitbucket

import (
	"fmt"
)

// Config implements auth.Configurator interface and
// handles the configuration parameters for bitbucket authentication
type Config struct {
	ClientID     string
	ClientSecret string
	Workspace    string
}

func (c *Config) SetDefaults() {}

func (c *Config) Validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("missing required client ID")
	}

	if c.ClientSecret == "" {
		return fmt.Errorf("missing required client secret")
	}

	return nil
}
