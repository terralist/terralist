package postgresql

import (
	"fmt"
	"sync"

	"terralist/pkg/database"

	gormzerolog "github.com/mpalmer/gorm-zerolog"
	"gorm.io/driver/postgres"
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

	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("wrong database configuration")
	}

	dsn := cfg.DSN()

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{
		Logger: (gormzerolog.Logger{}).LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	return &database.DefaultEngine{
		Handle: db,
	}, nil
}
