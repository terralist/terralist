package database

import (
	"fmt"

	db "terralist/pkg/database"
)

// Config implements session.Configurator interface for database-backed sessions.
type Config struct {
	Database   db.Engine
	CookieName string
	Secret     string
	MaxAge     int // session max age in seconds, 0 means 24 hours
}

func (c *Config) SetDefaults() {
	if c.CookieName == "" {
		c.CookieName = "_session_id"
	}
	if c.MaxAge == 0 {
		c.MaxAge = 86400 // 24 hours
	}
}

func (c *Config) Validate() error {
	if c.Database == nil {
		return fmt.Errorf("missing required attribute 'Database'")
	}
	if c.Secret == "" {
		return fmt.Errorf("missing required attribute 'Secret'")
	}
	return nil
}
