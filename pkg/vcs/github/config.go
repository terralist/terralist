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
	if c.AccessToken == "" && (c.AppID == 0 || c.AppInstallationID == 0 || c.AppPrivateKeyPath == "") {
		return fmt.Errorf("either access token or app ID, app installation ID, and app private key path must be provided")
	}

	if c.BaseURL == "" {
		return fmt.Errorf("missing required base URL")
	}

	return nil
}
