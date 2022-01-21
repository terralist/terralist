package sqlite

import (
	"sync"

	"github.com/valentindeaconu/terralist/internal/server/models/module"
	"github.com/valentindeaconu/terralist/internal/server/models/provider"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SqliteDatabase is the concrete implementation of database.Engine
type SqliteDatabase struct {
	Handle *gorm.DB
}

const (
	schemaFileName = "schema.db"
)

var (
	lock = &sync.Mutex{}
)

func NewDatabase() (*SqliteDatabase, error) {
	lock.Lock()
	defer lock.Unlock()

	db, err := gorm.Open(sqlite.Open(schemaFileName), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	return &SqliteDatabase{
		Handle: db,
	}, nil
}

func (d *SqliteDatabase) Connect() error {
	if err := d.Handle.AutoMigrate(
		&module.Dependency{},
		&module.Provider{},
		&module.Submodule{},
		&module.Version{},
		&module.Module{},
	); err != nil {
		return err
	}

	err := d.Handle.AutoMigrate(
		&provider.GpgPublicKey{},
		&provider.Platform{},
		&provider.Version{},
		&provider.Provider{},
	)

	return err
}

func (d *SqliteDatabase) Handler() *gorm.DB {
	return d.Handle
}
