package server

import (
	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database"
)

type InitialMigration struct{}

func (*InitialMigration) Migrate(db *database.DB) error {
	if err := db.AutoMigrate(
		&authority.Authority{},
		&authority.Key{},
		&authority.ApiKey{},
		&provider.Provider{},
		&provider.Version{},
		&provider.Platform{},
		&module.Module{},
		&module.Version{},
		&module.Submodule{},
		&module.Provider{},
		&module.Dependency{},
	); err != nil {
		return err
	}

	

	return nil
}
