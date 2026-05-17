package github

import (
	"fmt"
	"terralist/pkg/vcs"
)

type Creator struct{}

func (t *Creator) New(config vcs.Configurator) (vcs.Provider, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	return &Provider{
		WebhookSecret:     cfg.WebhookSecret,
		AccessToken:       cfg.AccessToken,
		AppID:             cfg.AppID,
		AppInstallationID: cfg.AppInstallationID,
		AppPrivateKeyPath: cfg.AppPrivateKeyPath,
		BaseURL:           fmt.Sprintf("https://%s", cfg.BaseURL),
	}, nil
}
