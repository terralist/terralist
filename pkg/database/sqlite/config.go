package sqlite

import (
	"os"
	"path/filepath"
)

// Config implements database.Configurator interface and
// handles the default configuration parameters of the sqlite database.
type Config struct {
	Path string
	Home string
}

func (c *Config) SetDefaults() {
	if c.Path != "" {
		return
	}

	c.Path = DefaultPath(c.Home)
}

func (c *Config) Validate() error {
	return nil
}

// DefaultPath returns the default sqlite path for the given home directory.
func DefaultPath(home string) string {
	if home == "" {
		home = os.Getenv("TERRALIST_HOME")
	}

	if home == "" {
		if userHome, err := os.UserHomeDir(); err == nil {
			home = userHome
		}
	}

	if home == "" {
		home = "."
	}

	return filepath.Join(home, "data", "storage.db")
}
