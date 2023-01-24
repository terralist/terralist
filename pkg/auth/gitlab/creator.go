package gitlab

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
		ClientID:           cfg.ClientID,
		ClientSecret:       cfg.ClientSecret,
		GitLabOAuthBaseURL: fmt.Sprintf("https://%s/oauth", cfg.GitlabHostWithOptionalPort),
		RedirectURL:        cfg.TerralistHostAndPort + "/v1/api/auth/redirect",
	}, nil
}
