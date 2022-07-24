package factory

import (
	"fmt"

	"terralist/pkg/database"
	"terralist/pkg/database/postgresql"
	"terralist/pkg/database/sqlite"
)

func NewDatabase(backend database.Backend, config database.Configurator) (database.Engine, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("could not create a new database with invalid configuration: %v", err)
	}

	// Set DB defaults
	config.SetDefaults()

	var creator database.Creator

	switch backend {
	case database.SQLITE:
		creator = &sqlite.Creator{}
	case database.POSTGRESQL:
		creator = &postgresql.Creator{}
	default:
		return nil, fmt.Errorf("unrecognized backend type")
	}

	return creator.New(config)
}
