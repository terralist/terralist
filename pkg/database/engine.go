package database

import (
	"fmt"

	"gorm.io/gorm"
)

type DB = gorm.DB

// Engine handles the database connection and operations.
type Engine interface {
	WithMigration(Migrator) error
	Handler() *DB
}

// DefaultEngine is the default concrete implementation of database.Engine.
type DefaultEngine struct {
	Handle *gorm.DB
}

func (d *DefaultEngine) WithMigration(m Migrator) error {
	if m == nil {
		return fmt.Errorf("cannot migrate a nil migrator")
	}

	return m.Migrate(d.Handle)
}

func (d *DefaultEngine) Handler() *gorm.DB {
	return d.Handle
}
