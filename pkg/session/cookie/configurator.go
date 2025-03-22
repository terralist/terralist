package cookie

import "fmt"

// Config implements session.Configurator interface
// and handles the configuration parameters of the
// cookie session implementation.
type Config struct {
	Name   string
	Secret string
}

func (c *Config) SetDefaults() {
	if c.Name == "" {
		c.Name = "_session"
	}
}

func (c *Config) Validate() error {
	if c.Secret == "" {
		return fmt.Errorf("missing required attribute 'Secret'")
	}

	return nil
}
