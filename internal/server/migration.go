package server

import (
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database"
)

type InitialMigration struct{}

func (*InitialMigration) Migrate(db *database.DB) error {
	if err := db.AutoMigrate(
		&module.Module{},
		&module.Version{},
		&module.Submodule{},
		&module.Provider{},
		&module.Dependency{},
	); err != nil {
		return err
	}

	err := db.Debug().AutoMigrate(
		&provider.Provider{},
		&provider.Version{},
		&provider.Platform{},
		&provider.GpgPublicKey{},
	)

	return err
}
