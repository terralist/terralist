package database

import (
	"github.com/valentindeaconu/terralist/model/module"
	"github.com/valentindeaconu/terralist/model/provider"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	Handler *gorm.DB
}

var AppDatabase Database = Database{
	Handler: nil,
}

func (m *Database) Open() error {
	db, err := gorm.Open(
		sqlite.Open("registry.db"),
		&gorm.Config{},
	)

	m.Handler = db

	return err
}

func (m *Database) Init() error {
	if m.Handler == nil {
		if err := m.Open(); err != nil {
			return err
		}
	}

	m.Handler.AutoMigrate(
		&module.Dependency{},
		&module.Provider{},
		&module.Submodule{},
		&module.Version{},
		&module.Module{},
	)
	m.Handler.AutoMigrate(
		&provider.GpgPublicKey{},
		&provider.Platform{},
		&provider.Version{},
		&provider.Provider{},
	)

	return nil
}
