package database

import (
	"github.com/valentindeaconu/terralist/internal/server/database/config"
	"github.com/valentindeaconu/terralist/internal/server/database/sqlite"
	"gorm.io/gorm"
)

// Engine handles the database connection and operations
type Engine interface {
	Connect() error
	Handler() *gorm.DB
}

// DatabaseCreator creates the database
type DatabaseCreator interface {
	NewDatabase(backend string, config config.DatabaseConfig) (Engine, error)
}

// DefaultDatabaseCreator is the concrete implementation of DatabaseCreator
type DefaultDatabaseCreator struct{}

func (d *DefaultDatabaseCreator) NewDatabase(backend string, config config.DatabaseConfig) (Engine, error) {
	switch backend {
	case "sqlite":
		return sqlite.NewDatabase(config)
	default:
		return sqlite.NewDatabase(config)
	}
}
