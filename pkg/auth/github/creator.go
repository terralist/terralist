package github

import (
	"fmt"
	"strings"

	"terralist/pkg/auth"
)

type Creator struct{}

func (t *Creator) New(config auth.Configurator) (auth.Provider, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	// Use api/v3 if not github.com
	var apiEndpoint string
	if strings.EqualFold(cfg.Domain, "github.com") {
		apiEndpoint = "https://api.github.com"
	} else {
		apiEndpoint = fmt.Sprintf("https://%s/api/v3", cfg.Domain)
	}

	return &Provider{
		ClientID:      cfg.ClientID,
		ClientSecret:  cfg.ClientSecret,
		Organization:  cfg.Organization,
		Teams:         cfg.Teams,
		oauthEndpoint: fmt.Sprintf("https://%s/login/oauth", cfg.Domain),
		apiEndpoint:   apiEndpoint,
	}, nil
}
