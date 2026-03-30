package local

import (
	"fmt"
	"os"
	"strings"

	"terralist/pkg/auth/jwt"
	"terralist/pkg/storage"
)

type Creator struct{}

func (t *Creator) New(config storage.Configurator) (storage.Resolver, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	if err := os.MkdirAll(cfg.RegistryDirectory, 0700); err != nil {
		return nil, fmt.Errorf("could not create the registry dir: %w", err)
	}

	jwt, err := jwt.New(cfg.TokenSigningSecret)
	if err != nil {
		return nil, fmt.Errorf("could not create jwt handler: %w", err)
	}

	return &Resolver{
		RegistryDir: cfg.RegistryDirectory,
		LinkExpire:  cfg.LinkExpire * 60,
		URLFormat:   buildURLFormat(cfg.BaseURL, cfg.FilesEndpoint),

		JWT: jwt,
	}, nil
}

func buildURLFormat(baseURL, filesEndpoint string) string {
	base := strings.TrimRight(baseURL, "/")
	endpoint := strings.Trim(filesEndpoint, "/")

	if endpoint == "" {
		return fmt.Sprintf("%s/%%s?token=%%s", base)
	}

	return fmt.Sprintf("%s/%s/%%s?token=%%s", base, endpoint)
}
