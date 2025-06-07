package github

import (
	"fmt"

	"terralist/pkg/auth"
)

type Creator struct{}

func (t *Creator) New(config auth.Configurator) (auth.Provider, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	return &Provider{
		ClientID:      cfg.ClientID,
		ClientSecret:  cfg.ClientSecret,
		Organization:  cfg.Organization,
		Teams:         cfg.Teams,
		oauthEndpoint: fmt.Sprintf("https://%s/login/oauth", cfg.Domain),
		apiEndpoint:   fmt.Sprintf("https://api.%s", cfg.Domain),
	}, nil
}
