package database

import (
	"github.com/valentindeaconu/terralist/internal/server/database/sqlite"
	"gorm.io/gorm"
)

type BackendType int

const (
	Sqlite = iota
)

// Engine handles the database connection and operations
type Engine interface {
	Connect() error
	Handler() *gorm.DB
}

// DatabaseCreator creates the database
type DatabaseCreator interface {
	NewDatabase(t BackendType) (Engine, error)
}

// DefaultDatabaseCreator is the concrete implementation of DatabaseCreator
type DefaultDatabaseCreator struct{}

func (d *DefaultDatabaseCreator) NewDatabase(t BackendType) (Engine, error) {
	switch t {
	case Sqlite:
		return sqlite.NewDatabase()
	default:
		return sqlite.NewDatabase()
	}
}
