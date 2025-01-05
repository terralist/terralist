package local

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	registryDirName = "registry"
)

// Config implements storage.Configurator interface and
// handles the configuration parameters of the local resolver
type Config struct {
	HomeDirectory     string
	RegistryDirectory string
	BaseURL           string
	FilesEndpoint     string

	TokenSigningSecret string
	LinkExpire         int
}

func (c *Config) SetDefaults() {
	if c.RegistryDirectory == "" {
		c.RegistryDirectory = filepath.Join(sanitizePath(c.HomeDirectory), registryDirName)
	} else {
		c.RegistryDirectory = sanitizePath(c.RegistryDirectory)
	}
}

func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("local resolver needs to know the base URL")
	}

	if c.FilesEndpoint == "" {
		return fmt.Errorf("local resolver needs to know the files endpoint")
	}

	if c.TokenSigningSecret == "" {
		return fmt.Errorf("a secret for signing tokens is required")
	}

	if c.LinkExpire <= 0 {
		return fmt.Errorf("the expire time for links must be positive > 0")
	}

	return nil
}

func sanitizePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	usr, _ := user.Current()
	dir := usr.HomeDir
	cwd, _ := os.Getwd()

	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(dir, path[2:])
	} else if path == "." {
		path = cwd
	} else if strings.HasPrefix(path, "./") {
		path = filepath.Join(cwd, path[2:])
	}

	return path
}
