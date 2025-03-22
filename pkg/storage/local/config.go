package local

import (
	"path"
)

const (
	registryDirName = "registry"
)

// Config implements storage.Configurator interface and
// handles the configuration parameters of the local resolver.
type Config struct {
	HomeDirectory string
}

func (c *Config) SetDefaults() {
	c.HomeDirectory = path.Join(c.HomeDirectory, registryDirName)
}

func (c *Config) Validate() error {
	return nil
}
