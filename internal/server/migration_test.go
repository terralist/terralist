package server

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"terralist/internal/server/models/apikey"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type legacyModuleVersion struct {
	ID            uuid.UUID `gorm:"primary_key;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ModuleID      uuid.UUID
	Version       string `gorm:"not null"`
	Location      string `gorm:"not null"`
	Documentation string `gorm:"not null;default:''"`
}

func (legacyModuleVersion) TableName() string {
	return "module_versions"
}

type tableInfo struct {
	Name      string         `gorm:"column:name"`
	DfltValue sql.NullString `gorm:"column:dflt_value"`
}

func TestInitialMigrationDropsModuleDocumentationDefault(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	if err := db.AutoMigrate(&legacyModuleVersion{}); err != nil {
		t.Fatalf("failed to create legacy schema: %v", err)
	}

	before, err := documentationColumnDefault(db)
	if err != nil {
		t.Fatalf("failed to inspect legacy schema: %v", err)
	}
	if !before.Valid {
		t.Fatal("expected legacy schema to have a default for module_versions.documentation")
	}

	if err := (&InitialMigration{}).Migrate(db); err != nil {
		t.Fatalf("failed to run initial migration: %v", err)
	}

	after, err := documentationColumnDefault(db)
	if err != nil {
		t.Fatalf("failed to inspect migrated schema: %v", err)
	}
	if after.Valid {
		t.Fatalf("expected module_versions.documentation default to be removed, got %q", after.String)
	}
}

func documentationColumnDefault(db *gorm.DB) (sql.NullString, error) {
	var columns []tableInfo
	if err := db.Raw("PRAGMA table_info('module_versions')").Scan(&columns).Error; err != nil {
		return sql.NullString{}, err
	}

	for _, column := range columns {
		if column.Name == "documentation" {
			return column.DfltValue, nil
		}
	}

	return sql.NullString{}, gorm.ErrRecordNotFound
}

func TestInitialMigrationDropsMirrorTables(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	tables := []string{"mirror_providers", "mirror_versions", "mirror_platforms"}
	for _, table := range tables {
		if err := db.Exec("CREATE TABLE " + table + " (id TEXT PRIMARY KEY)").Error; err != nil {
			t.Fatalf("failed to create table %s: %v", table, err)
		}
	}

	if err := (&InitialMigration{}).Migrate(db); err != nil {
		t.Fatalf("failed to run initial migration: %v", err)
	}

	for _, table := range tables {
		if db.Migrator().HasTable(table) {
			t.Errorf("expected table %s to be dropped", table)
		}
	}
}

type legacyApiKey struct {
	ID        uuid.UUID `gorm:"primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string `gorm:"not null"`
	Scope     string `gorm:"not null"`
	CreatedBy string `gorm:"not null"`
}

func (legacyApiKey) TableName() string {
	return "api_keys"
}

func TestInitialMigrationHashesLegacyApiKeys(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	if err := db.AutoMigrate(&legacyApiKey{}, &apikey.Policy{}); err != nil {
		t.Fatalf("failed to create legacy tables: %v", err)
	}

	legacyID := uuid.New()
	if err := db.Create(&legacyApiKey{ID: legacyID, Name: "ci", Scope: "team-a", CreatedBy: "ci@example.com"}).Error; err != nil {
		t.Fatalf("failed to insert legacy key: %v", err)
	}

	policy := apikey.Policy{ApiKeyID: legacyID, Resource: "modules", Action: "get", Object: "*", Effect: "allow"}
	if err := db.Create(&policy).Error; err != nil {
		t.Fatalf("failed to insert legacy policy: %v", err)
	}

	if err := (&InitialMigration{}).Migrate(db); err != nil {
		t.Fatalf("failed to run initial migration: %v", err)
	}

	var keys []apikey.ApiKey
	if err := db.Preload("Policies").Find(&keys).Error; err != nil {
		t.Fatalf("failed to load api keys: %v", err)
	}

	if len(keys) != 1 {
		t.Fatalf("expected exactly one api key after migration, got %d", len(keys))
	}

	migrated := keys[0]

	if migrated.ID == legacyID {
		t.Errorf("expected the migrated key to receive a fresh id")
	}

	if migrated.Hash != apikey.HashSecret(legacyID.String()) {
		t.Errorf("expected the legacy id to be stored as the key hash")
	}

	if migrated.Name != "ci" || migrated.Scope != "team-a" || migrated.CreatedBy != "ci@example.com" {
		t.Errorf("expected key attributes to be preserved, got %+v", migrated)
	}

	if len(migrated.Policies) != 1 || migrated.Policies[0].ID != policy.ID {
		t.Errorf("expected the policy to follow the migrated key, got %+v", migrated.Policies)
	}
}

func TestInitialMigrationLeavesHashedApiKeysAlone(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	if err := (&InitialMigration{}).Migrate(db); err != nil {
		t.Fatalf("failed to run initial migration: %v", err)
	}

	key := apikey.ApiKey{Hash: apikey.HashSecret("tlk_example"), Name: "ci", Scope: "team-a", CreatedBy: "ci@example.com"}
	if err := db.Create(&key).Error; err != nil {
		t.Fatalf("failed to insert hashed key: %v", err)
	}

	if err := (&InitialMigration{}).Migrate(db); err != nil {
		t.Fatalf("failed to re-run initial migration: %v", err)
	}

	var found apikey.ApiKey
	if err := db.Where("id = ?", key.ID).First(&found).Error; err != nil {
		t.Fatalf("expected the hashed key to keep its id: %v", err)
	}
}

func TestInitialMigrationRefusesDuplicateArtifacts(t *testing.T) {
	for _, tc := range []struct {
		table   string
		columns string
		values  string
	}{
		{"providers", "authority_id TEXT, name TEXT", "'a1', 'null'"},
		{"provider_versions", "provider_id TEXT, version TEXT", "'p1', '3.2.4'"},
		{"modules", "authority_id TEXT, name TEXT, provider TEXT", "'a1', 'vpc', 'aws'"},
		{"module_versions", "module_id TEXT, version TEXT", "'m1', '1.0.0'"},
	} {
		t.Run(tc.table, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
			if err != nil {
				t.Fatalf("failed to open sqlite database: %v", err)
			}

			if err := db.Exec("CREATE TABLE " + tc.table + " (id TEXT PRIMARY KEY, " + tc.columns + ")").Error; err != nil {
				t.Fatalf("failed to create table: %v", err)
			}

			for _, id := range []string{"1", "2"} {
				if err := db.Exec("INSERT INTO " + tc.table + " VALUES ('" + id + "', " + tc.values + ")").Error; err != nil {
					t.Fatalf("failed to insert row: %v", err)
				}
			}

			err = (&InitialMigration{}).Migrate(db)
			if err == nil {
				t.Fatal("expected the migration to refuse the duplicate rows")
			}

			if !strings.Contains(err.Error(), tc.table) {
				t.Errorf("expected the error to name the %s table, got %v", tc.table, err)
			}
		})
	}
}

type legacyAuthorityApiKey struct {
	ID          uuid.UUID `gorm:"primary_key;"`
	AuthorityID uuid.UUID
	Name        string
}

func (legacyAuthorityApiKey) TableName() string {
	return "authority_api_keys"
}

func TestInitialMigrationDropsAuthorityApiKeys(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	if err := db.AutoMigrate(&legacyAuthorityApiKey{}); err != nil {
		t.Fatalf("failed to create legacy table: %v", err)
	}

	if err := (&InitialMigration{}).Migrate(db); err != nil {
		t.Fatalf("failed to run initial migration: %v", err)
	}

	if db.Migrator().HasTable("authority_api_keys") {
		t.Errorf("expected authority_api_keys table to be dropped")
	}
}
