package mysql

import (
	"fmt"
	"sync"
	"terralist/pkg/database"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Creator struct{}

var (
	lock = &sync.Mutex{}
)

func (t *Creator) New(config database.Configurator, migrator database.Migrator) (database.Engine, error) {
	lock.Lock()
	defer lock.Unlock()

	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("wrong database configuration")
	}

	dsn := cfg.DSN()

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // data source name
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	}), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	return &database.DefaultEngine{
		Handle:   db,
		Migrator: migrator,
	}, nil
}
