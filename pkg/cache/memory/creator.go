package memory

import (
	"fmt"
	"time"

	"terralist/pkg/cache"
)

type Creator struct{}

func (t *Creator) New(config cache.Configurator) (cache.Cache, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	c := &Cache{
		entries: map[string]entry{},
		now:     time.Now,
	}

	go c.run(cfg.SweepInterval)

	return c, nil
}
