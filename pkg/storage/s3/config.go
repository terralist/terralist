package s3

import "fmt"

// Config implements storage.Configurator interface and
// handles the configuration parameters of the s3 resolver
type Config struct {
	BucketName      string
	BucketRegion    string
	AccessKeyID     string
	SecretAccessKey string

	LinkExpire         int
	DefaultCredentials bool
}

func (c *Config) SetDefaults() {}

func (c *Config) Validate() error {
	if c.BucketName == "" {
		return fmt.Errorf("missing required attribute 'BucketName'")
	}

	if c.AccessKeyID == "" || c.SecretAccessKey == "" {
		c.DefaultCredentials = true
	} else {
		c.DefaultCredentials = false
	}

	if c.SecretAccessKey == "" {
		return fmt.Errorf("missing required attribute 'SecretAccessKey'")
	}

	if c.LinkExpire <= 0 {
		return fmt.Errorf("the expire time for links must be positive > 0")
	}

	return nil
}
