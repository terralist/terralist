package local

import (
	"fmt"
	"terralist/pkg/storage/resolver"
)

type Creator struct{}

func (t *Creator) New(config resolver.Configurator) (resolver.Resolver, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	// TODO: Make sure registry dir exists

	return &Resolver{
		RegistryDir: cfg.HomeDirectory,
	}, nil
}
