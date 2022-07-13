package factory

import (
	"fmt"
	"terralist/pkg/database"
	"terralist/pkg/database/mysql"
	"terralist/pkg/database/sqlite"
)

func NewDatabase(backend database.Backend, config database.Configurator, migrator database.Migrator) (database.Engine, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("could not create a new database with invalid configuration: %v", err)
	}

	// Set DB defaults
	config.SetDefaults()

	switch backend {
	case database.SQLITE:
		creator := sqlite.Creator{}
		return creator.New(config, migrator)
	case database.MYSQL:
		creator := mysql.Creator{}
		return creator.New(config, migrator)
	default:
		return nil, fmt.Errorf("unrecognized backend type")
	}
}
