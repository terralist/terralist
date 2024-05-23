package mssql

import (
	"fmt"
	"sync"

	"terralist/pkg/database"
	"terralist/pkg/database/logger"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
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

	db, err := gorm.Open(sqlserver.New(sqlserver.Config{
		DSN: dsn,
	}), &gorm.Config{
		Logger: &logger.Logger{},
	})

	if err != nil {
		return nil, err
	}

	return &database.DefaultEngine{
		Handle: db.Debug(),
	}, nil
}
