package proxy

// Config implements storage.Configurator interface and
// handles the configuration parameters of the proxy resolver
type Config struct{}

func (c *Config) SetDefaults() {}

func (c *Config) Validate() error {
	return nil
}
