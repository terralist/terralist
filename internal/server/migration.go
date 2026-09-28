package server

import (
	"fmt"
	"slices"
	"strings"

	"terralist/internal/server/models/apikey"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/oauth"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database"

	"github.com/google/uuid"
)

type InitialMigration struct{}

func (*InitialMigration) Migrate(db *database.DB) error {
	if err := refuseDuplicateArtifacts(db); err != nil {
		return err
	}

	if err := refuseDuplicatesIgnoringCase(db, uniqueIndexes()); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&authority.Authority{},
		&authority.Key{},
		&authority.Rule{},
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
	); err != nil {
		return err
	}

	if err := db.Migrator().DropTable("mirror_platforms", "mirror_versions", "mirror_providers"); err != nil {
		return err
	}

	// Authorities, providers and modules were held unique by case-sensitive
	// indexes, which the case-insensitive ones replace.
	for _, legacy := range []struct{ table, name string }{
		{"authorities", "idx_authorities_name"},
		{"authorities", "idx_authorities_upstream"},
		{"providers", "idx_providers_authority_name"},
		{"modules", "idx_modules_authority_name_provider"},
	} {
		if err := database.DropIndex(db, legacy.table, legacy.name); err != nil {
			return err
		}
	}

	for _, index := range uniqueIndexes() {
		if err := index.Create(db); err != nil {
			return err
		}
	}

	// Remove default empty string column in Version.Documentation
	if err := db.Migrator().AlterColumn(&module.Version{}, "Documentation"); err != nil {
		return err
	}

	if err := db.Migrator().DropTable("authority_api_keys"); err != nil {
		return err
	}

	return hashLegacyApiKeys(db)
}

// uniqueArtifacts lists the columns identifying a row of each artifact table,
// which the schema holds unique.
var uniqueArtifacts = []struct {
	table   string
	columns []string
}{
	{"provider_versions", []string{"provider_id", "version"}},
	{"module_versions", []string{"module_id", "version"}},
}

// uniqueIndexes lists the indexes holding artifacts unique regardless of case.
func uniqueIndexes() []database.CaseInsensitiveUniqueIndex {
	return slices.Concat(authority.UniqueIndexes, provider.UniqueIndexes, module.UniqueIndexes)
}

// refuseDuplicateArtifacts fails when an artifact table holds rows the unique
// indexes would reject, naming them so they can be removed before Terralist
// starts; no data is changed.
func refuseDuplicateArtifacts(db *database.DB) error {
	for _, unique := range uniqueArtifacts {
		if !db.Migrator().HasTable(unique.table) {
			continue
		}

		columns := strings.Join(unique.columns, ", ")

		var duplicates []map[string]any
		if err := db.Table(unique.table).
			Select(columns).
			Group(columns).
			Having("COUNT(*) > 1").
			Find(&duplicates).
			Error; err != nil {
			return err
		}

		if len(duplicates) > 0 {
			return fmt.Errorf("table %s holds duplicate rows for (%s): %v; remove the duplicates before starting Terralist", unique.table, columns, duplicates)
		}
	}

	return nil
}

// refuseDuplicatesIgnoringCase fails when a table holds rows that differ only
// in case where a case-insensitive unique index is to be created, naming them
// so they can be removed before Terralist starts; no data is changed.
func refuseDuplicatesIgnoringCase(db *database.DB, indexes []database.CaseInsensitiveUniqueIndex) error {
	for _, index := range indexes {
		duplicates, err := index.Duplicates(db)
		if err != nil {
			return err
		}

		if len(duplicates) > 0 {
			return fmt.Errorf("table %s holds rows differing only in case for (%s): %v; remove the duplicates before starting Terralist", index.Table, strings.Join(index.Columns, ", "), duplicates)
		}
	}

	return nil
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
