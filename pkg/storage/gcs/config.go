package gcs

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	prefixRegEx = regexp.MustCompile(`(?m)^[a-zA-Z0-9\(\)\'\*\.\-_\!\/]+$`)
)

// Config implements storage.Configurator interface and
// handles the configuration parameters of the s3 resolver.
type Config struct {
	BucketName                 string
	BucketPrefix               string
	ServiceAccountCredFilePath string

	LinkExpire         int
	DefaultCredentials bool
}

func (c *Config) SetDefaults() {}

func (c *Config) Validate() error {
	if c.BucketName == "" {
		return fmt.Errorf("missing required attribute 'BucketName'")
	}
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" && c.ServiceAccountCredFilePath == "" {
		c.ServiceAccountCredFilePath = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}

	if c.ServiceAccountCredFilePath != "" {
		c.DefaultCredentials = false
	} else {
		c.DefaultCredentials = true
	}

	if c.BucketPrefix != "" {
		if strings.HasPrefix(c.BucketPrefix, "/") {
			return fmt.Errorf("the prefix must not start with a slash ('/')")
		}

		if strings.HasSuffix(c.BucketPrefix, "/") {
			return fmt.Errorf("the prefix must not end with a slash ('/')")
		}

		if !prefixRegEx.MatchString(c.BucketPrefix) {
			return fmt.Errorf("the prefix contains invalid characters")
		}

		c.BucketPrefix = fmt.Sprintf("%s/", c.BucketPrefix)
	}

	if c.LinkExpire <= 0 {
		return fmt.Errorf("the expire time for links must be positive > 0")
	}

	return nil
}
