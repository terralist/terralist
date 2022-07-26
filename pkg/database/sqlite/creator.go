package sqlite

import (
	"sync"

	"terralist/pkg/database"
	"terralist/pkg/database/logger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Creator struct{}

var (
	lock = &sync.Mutex{}
)

func (t *Creator) New(config database.Configurator) (database.Engine, error) {
	lock.Lock()
	defer lock.Unlock()

	cfg := config.(*Config)

	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: &logger.Logger{},
	})

	if err != nil {
		return nil, err
	}

	return &database.DefaultEngine{
		Handle: db,
	}, nil
}
