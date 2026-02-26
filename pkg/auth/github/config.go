package github

import (
	"fmt"
)

// Config implements auth.Configurator interface and
// handles the configuration parameters of the sqlite database.
type Config struct {
	ClientID     string
	ClientSecret string
	Organization string
	Teams        string
	Domain       string
}

func (c *Config) SetDefaults() {}

func (c *Config) Validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("missing required client ID")
	}

	if c.ClientSecret == "" {
		return fmt.Errorf("missing required client secret")
	}

	if c.Teams != "" && c.Organization == "" {
		return fmt.Errorf("missing organization when using teams")
	}

	return nil
}
