package oidc

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

	return &Provider{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		AuthorizeUrl: cfg.AuthorizeUrl,
		TokenUrl:     cfg.TokenUrl,
		UserInfoUrl:  cfg.UserInfoUrl,
		RedirectUrl:  strings.TrimSuffix(cfg.TerralistSchemeHostAndPort, "/") + "/v1/api/auth/redirect",
	}, nil
}
