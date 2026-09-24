package memory

import (
	"fmt"
	"time"
)

const defaultSweepInterval = time.Minute

// Config implements cache.Configurator interface and handles the configuration
// parameters of the in-memory cache.
type Config struct {
	// SweepInterval is how often expired entries are removed from memory.
	// Expired entries are never served, whether swept yet or not.
	SweepInterval time.Duration
}

func (c *Config) SetDefaults() {
	if c.SweepInterval == 0 {
		c.SweepInterval = defaultSweepInterval
	}
}

func (c *Config) Validate() error {
	if c.SweepInterval < 0 {
		return fmt.Errorf("the sweep interval cannot be negative")
	}

	return nil
}
