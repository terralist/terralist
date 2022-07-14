package s3

import "fmt"

// Config implements storage.Configurator interface and
// handles the configuration parameters of the s3 resolver
type Config struct {
	BucketName      string
	AccessKeyID     string
	SecretAccessKey string
}

func (c *Config) SetDefaults() {}

func (c *Config) Validate() error {
	if c.BucketName == "" {
		return fmt.Errorf("missing required attribute 'BucketName'")
	}

	if c.AccessKeyID == "" {
		return fmt.Errorf("missing required attribute 'AccessKeyID'")
	}

	if c.SecretAccessKey == "" {
		return fmt.Errorf("missing required attribute 'SecretAccessKey'")
	}

	return nil
}
