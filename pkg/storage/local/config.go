package local

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

const (
	DataStoreDirName = "terralist.d"
)

// Config implements storage.Configurator interface and
// handles the configuration parameters of the local resolver
type Config struct {
	DataStorePath string
}

func (c *Config) SetDefaults() {
	if c.DataStorePath == "" {
		homeDir, _ := os.UserHomeDir()
		c.DataStorePath = path.Join(homeDir, DataStoreDirName)
	}

	c.DataStorePath, _ = filepath.Abs(c.DataStorePath)
}

func (c *Config) Validate() error {
	if !path.IsAbs(c.DataStorePath) {
		return fmt.Errorf("the datastore path should be absolute")
	}

	return nil
}
