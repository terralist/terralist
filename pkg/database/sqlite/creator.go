package sqlite

import (
	"fmt"
	"net/url"
	"sync"

	"terralist/pkg/database"
	"terralist/pkg/database/logger"

	"github.com/glebarez/sqlite"
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

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: &logger.Logger{},
	})

	if err != nil {
		return nil, err
	}

	return &database.DefaultEngine{
		Handle: db,
	}, nil
}
