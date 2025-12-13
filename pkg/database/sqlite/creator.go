package sqlite

import (
	"fmt"
	"sync"

	"terralist/pkg/database"
	"terralist/pkg/database/logger"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
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
		return nil, fmt.Errorf("unsupported configurator: %T", config)
	}

	dsn := cfg.DSN()

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: &logger.Logger{},
	})

	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Using SQLite database at %s", cfg.Path)

	return &database.DefaultEngine{
		Handle: db,
	}, nil
}
