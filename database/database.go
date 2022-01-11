package database

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/valentindeaconu/terralist/models/module"
	"github.com/valentindeaconu/terralist/models/provider"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB = *gorm.DB

var lock = &sync.Mutex{}

var (
	db DB
)

func EnsureConnection() {
	if db == nil {
		if e := Connect(); e == nil {
			log.Fatal(e.Error())
		}
	}
}

func Connect() error {
	lock.Lock()
	defer lock.Unlock()

	d, err := gorm.Open(
		sqlite.Open("registry.db"),
		&gorm.Config{},
	)

	if err != nil {
		return err
	}

	db = d

	if err := db.AutoMigrate(
		&module.Dependency{},
		&module.Provider{},
		&module.Submodule{},
		&module.Version{},
		&module.Module{},
	); err != nil {
		return err
	}

	err = db.AutoMigrate(
		&provider.GpgPublicKey{},
		&provider.Platform{},
		&provider.Version{},
		&provider.Provider{},
	)

	return err
}

func Handler() DB {
	EnsureConnection()

	return db
}
