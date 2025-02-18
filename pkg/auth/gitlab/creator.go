package gitlab

import (
	"fmt"
	"slices"
	"strings"

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
		RedirectURL:        strings.TrimSuffix(cfg.TerralistSchemeHostAndPort, "/") + "/v1/api/auth/redirect",
		Groups: slices.DeleteFunc(strings.Split(cfg.Groups, ","), func(e string) bool {
			return e == ""
		}),
	}, nil
}
