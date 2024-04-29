// FILE: pkg/storage/azure/config.go
package azure

import (
	"fmt"
)

// Config implements storage.Configurator interface and
// handles the configuration parameters of the Azure resolver
type Config struct {
	AccountName        string
	AccountKey         string
	ContainerName      string
	DefaultCredentials bool
}

func (c *Config) SetDefaults() {
	// Set any default values for your configuration here
}

func (c *Config) Validate() error {

	if c.AccountKey == "" {
		c.DefaultCredentials = true
	} else {
		c.DefaultCredentials = false
	}
	if c.AccountName == "" {
		return fmt.Errorf("missing required attribute 'AccountName'")
	}
	if c.ContainerName == "" {
		return fmt.Errorf("missing required attribute 'ContainerName'")
	}

	return nil
}
