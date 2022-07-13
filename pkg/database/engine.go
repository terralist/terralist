package database

import (
	"gorm.io/gorm"
)

type DB = gorm.DB

// Engine handles the database connection and operations
type Engine interface {
	Connect() error
	Handler() *DB
}

// DefaultEngine is the default concrete implementation of database.Engine
type DefaultEngine struct {
	Handle   *gorm.DB
	Migrator Migrator
}

func (d *DefaultEngine) Connect() error {
	if d.Migrator != nil {
		return d.Migrator.Migrate(d.Handle)
	}

	return nil
}

func (d *DefaultEngine) Handler() *gorm.DB {
	return d.Handle
}
