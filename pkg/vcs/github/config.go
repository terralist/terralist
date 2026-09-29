package github

import (
	"fmt"
)

type Config struct {
	WebhookSecret string

	AccessToken string

	AppID             int
	AppInstallationID int
	AppPrivateKeyPath string

	BaseURL string
}

func (c *Config) SetDefaults() {}

func (c *Config) Validate() error {
	if c.WebhookSecret == "" {
		return fmt.Errorf("missing required webhook secret")
	}

	if c.BaseURL == "" {
		return fmt.Errorf("missing required base URL")
	}

	return nil
}
