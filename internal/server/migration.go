package server

import (
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database"
)

type InitialMigration struct{}

func (*InitialMigration) Migrate(db *database.DB) error {
	if err := db.AutoMigrate(
		&module.Dependency{},
		&module.Provider{},
		&module.Submodule{},
		&module.Version{},
		&module.Module{},
	); err != nil {
		return err
	}

	err := db.AutoMigrate(
		&provider.GpgPublicKey{},
		&provider.Platform{},
		&provider.Version{},
		&provider.Provider{},
	)

	return err
}
