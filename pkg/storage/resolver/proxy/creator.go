package proxy

import (
	"fmt"
	"terralist/pkg/storage/resolver"
)

type Creator struct{}

func (t *Creator) New(config resolver.Configurator) (resolver.Resolver, error) {
	_, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	return &Resolver{}, nil
}
