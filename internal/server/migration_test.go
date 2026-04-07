package server

import (
	"database/sql"
	"testing"
	"time"

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
