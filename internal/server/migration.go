package server

import (
	"terralist/internal/server/models/apikey"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/mirror"
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/oauth"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database"

	"github.com/google/uuid"
)

type InitialMigration struct{}

func (*InitialMigration) Migrate(db *database.DB) error {
	if err := db.AutoMigrate(
		&authority.Authority{},
		&authority.Key{},
		&authority.ApiKey{},
		&apikey.ApiKey{},
		&apikey.Policy{},
		&provider.Provider{},
		&provider.Version{},
		&provider.Platform{},
		&module.Module{},
		&module.Version{},
		&module.Submodule{},
		&module.Provider{},
		&module.Dependency{},
		&oauth.Code{},
		&mirror.Provider{},
		&mirror.Version{},
		&mirror.Platform{},
	); err != nil {
		return err
	}

	// Remove default empty string column in Version.Documentation
	if err := db.Migrator().AlterColumn(&module.Version{}, "Documentation"); err != nil {
		return err
	}

	return hashLegacyApiKeys(db)
}

// hashLegacyApiKeys converts API keys that still use their row id as the
// secret: the id is hashed into the hash column and the row receives a fresh
// id, so existing key values keep working while the plaintext disappears.
func hashLegacyApiKeys(db *database.DB) error {
	var legacy []apikey.ApiKey
	if err := db.Where("hash IS NULL OR hash = ''").Find(&legacy).Error; err != nil {
		return err
	}

	for _, old := range legacy {
		err := db.Transaction(func(tx *database.DB) error {
			migrated := old
			migrated.ID = uuid.New()
			migrated.Hash = apikey.HashSecret(old.ID.String())
			migrated.Policies = nil

			if err := tx.Create(&migrated).Error; err != nil {
				return err
			}

			if err := tx.Model(&apikey.Policy{}).
				Where("api_key_id = ?", old.ID).
				Update("api_key_id", migrated.ID).
				Error; err != nil {
				return err
			}

			return tx.Where("id = ?", old.ID).Delete(&apikey.ApiKey{}).Error
		})
		if err != nil {
			return err
		}
	}

	return nil
}
