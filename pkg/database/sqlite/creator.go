package sqlite

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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

	// See https://gitlab.com/cznic/sqlite/-/issues/47
	dsn := cfg.Path
	q := make(url.Values)
	q.Set("_time_format", "sqlite")
	dsn += "?" + q.Encode()

	if err := os.MkdirAll(filepath.Dir(cfg.Path), os.ModePerm); err != nil {
		return nil, fmt.Errorf("could not prepare sqlite directory: %w", err)
	}

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
