package local

import (
	"fmt"
	"os"

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
		LinkExpire:  cfg.LinkExpire,
		URLFormat:   fmt.Sprintf("%s%s/%%s?token=%%s", cfg.BaseURL, cfg.FilesEndpoint),

		JWT: jwt,
	}, nil
}
