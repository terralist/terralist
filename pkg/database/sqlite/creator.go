package sqlite

import (
	"sync"

	"terralist/pkg/database"

	gormzerolog "github.com/mpalmer/gorm-zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
		Logger: (gormzerolog.Logger{}).LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	return &database.DefaultEngine{
		Handle: db,
	}, nil
}
