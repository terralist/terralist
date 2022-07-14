package local

import (
	"fmt"
	"terralist/pkg/storage"
)

type Creator struct{}

func (t *Creator) New(config storage.Configurator) (storage.Resolver, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	return &Resolver{
		DataStorePath: cfg.DataStorePath,
	}, nil
}
