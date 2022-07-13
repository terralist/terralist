package sqlite

// Config implements database.Configurator interface and
// handles the default configuration parameters of the sqlite database
type Config struct {
	Path string
}

func (c *Config) SetDefaults() {
	if c.Path == "" {
		c.Path = "storage.db"
	}

	return
}

func (c *Config) Validate() error {
	return nil
}
