package factory

import (
	"fmt"

	"terralist/pkg/cache"
	"terralist/pkg/cache/memory"
)

func NewCache(backend cache.Backend, config cache.Configurator) (cache.Cache, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("could not create a new cache with invalid configuration: %v", err)
	}

	config.SetDefaults()

	var creator cache.Creator
	var backendName string

	switch backend {
	case cache.MEMORY:
		creator = &memory.Creator{}
		backendName = "memory"
	default:
		return nil, fmt.Errorf("unrecognized backend type")
	}

	c, err := creator.New(config)
	if err != nil {
		return nil, err
	}

	return &cache.MetricsCache{
		Cache:   c,
		Backend: backendName,
	}, nil
}
