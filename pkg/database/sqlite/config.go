package sqlite

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

// Config implements database.Configurator interface and
// handles the default configuration parameters of the sqlite database.
type Config struct {
	Path string
}

func (c *Config) SetDefaults() {}

func (c *Config) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("sqlite-path is required when using SQLite")
	}

	if err := os.MkdirAll(filepath.Dir(c.Path), os.ModePerm); err != nil {
		return fmt.Errorf("could not prepare sqlite directory: %w", err)
	}

	return nil
}

func (c *Config) DSN() string {
	url := &url.URL{
		Path: c.Path,
	}

	q := url.Query()

	// See https://gitlab.com/cznic/sqlite/-/issues/47
	q.Set("_time_format", "sqlite")

	url.RawQuery = q.Encode()

	return url.String()
}
