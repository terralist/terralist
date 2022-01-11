package database

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/valentindeaconu/terralist/model/module"
	"github.com/valentindeaconu/terralist/model/provider"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB = *gorm.DB

type contextFunc func(DB) (bool, interface{})

var lock = &sync.Mutex{}

var (
	db DB
)

func EnsureConnection() {
	if db == nil {
		if e := Open(); e == nil {
			log.Fatal(e.Error())
		}
	}
}

func Open() error {
	lock.Lock()
	defer lock.Unlock()

	d, err := gorm.Open(
		sqlite.Open("registry.db"),
		&gorm.Config{},
	)

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

func Create(value interface{}) error {
	lock.Lock()
	defer lock.Unlock()

	result := db.Create(value)

	return result.Error
}

func Save(value interface{}) error {
	lock.Lock()
	defer lock.Unlock()

	result := db.Save(value)

	return result.Error
}

func Delete(value interface{}) error {
	lock.Lock()
	defer lock.Unlock()

	result := db.Delete(value)

	return result.Error
}

func Run(fn contextFunc) (bool, interface{}) {
	lock.Lock()
	defer lock.Unlock()

	success, result := fn(db)

	return success, result
}
