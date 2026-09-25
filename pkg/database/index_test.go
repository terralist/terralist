package database_test

import (
	"testing"

	"terralist/pkg/database"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *database.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	if err := db.Exec("CREATE TABLE owners (id TEXT PRIMARY KEY, name TEXT, alias TEXT)").Error; err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

var ownersIndex = database.CaseInsensitiveUniqueIndex{Table: "owners", Name: "idx_owners_lower_name_alias", Columns: []string{"name", "alias"}}

func TestCaseInsensitiveUniqueIndexCreate(t *testing.T) {
	db := newTestDB(t)

	if err := ownersIndex.Create(db); err != nil {
		t.Fatalf("failed to create the index: %v", err)
	}

	if err := ownersIndex.Create(db); err != nil {
		t.Fatalf("expected creating an existing index to succeed, got: %v", err)
	}

	if err := db.Exec("INSERT INTO owners VALUES ('1', 'HashiCorp', 'hc')").Error; err != nil {
		t.Fatalf("failed to insert row: %v", err)
	}

	if err := db.Exec("INSERT INTO owners VALUES ('2', 'hashicorp', 'HC')").Error; err == nil {
		t.Fatal("expected a row differing only in case to be rejected")
	}

	for _, id := range []string{"3", "4"} {
		if err := db.Exec("INSERT INTO owners VALUES (?, 'hashicorp', NULL)", id).Error; err != nil {
			t.Fatalf("expected rows with a NULL column not to collide, got: %v", err)
		}
	}
}

func TestCaseInsensitiveUniqueIndexDuplicates(t *testing.T) {
	db := newTestDB(t)

	for _, row := range [][]any{
		{"1", "HashiCorp", "hc"},
		{"2", "hashicorp", "HC"},
		{"3", "other", "o"},
		{"4", "other", nil},
		{"5", "other", nil},
	} {
		if err := db.Exec("INSERT INTO owners VALUES (?, ?, ?)", row...).Error; err != nil {
			t.Fatalf("failed to insert row: %v", err)
		}
	}

	duplicates, err := ownersIndex.Duplicates(db)
	if err != nil {
		t.Fatalf("failed to list duplicates: %v", err)
	}

	if len(duplicates) != 1 {
		t.Fatalf("expected one duplicate, got %v", duplicates)
	}

	withMissingColumn := database.CaseInsensitiveUniqueIndex{Table: "owners", Name: "idx_owners_lower_name_email", Columns: []string{"name", "email"}}
	if duplicates, err := withMissingColumn.Duplicates(db); err != nil || len(duplicates) != 0 {
		t.Fatalf("expected no duplicates for a column the table does not have yet, got %v, %v", duplicates, err)
	}

	if err := db.Exec("DROP TABLE owners").Error; err != nil {
		t.Fatalf("failed to drop table: %v", err)
	}

	if duplicates, err := ownersIndex.Duplicates(db); err != nil || len(duplicates) != 0 {
		t.Fatalf("expected no duplicates for a missing table, got %v, %v", duplicates, err)
	}
}

func TestDropIndex(t *testing.T) {
	db := newTestDB(t)

	if err := db.Exec("CREATE UNIQUE INDEX idx_owners_name ON owners (name)").Error; err != nil {
		t.Fatalf("failed to create index: %v", err)
	}

	if err := database.DropIndex(db, "owners", "idx_owners_name"); err != nil {
		t.Fatalf("failed to drop the index: %v", err)
	}

	if db.Migrator().HasIndex("owners", "idx_owners_name") {
		t.Fatal("expected the index to be dropped")
	}

	if err := database.DropIndex(db, "owners", "idx_owners_name"); err != nil {
		t.Fatalf("expected dropping a missing index to succeed, got: %v", err)
	}
}

func TestCaseInsensitiveUniqueIndexDuplicatesNameTheColumns(t *testing.T) {
	db := newTestDB(t)

	for _, id := range []string{"1", "2"} {
		if err := db.Exec("INSERT INTO owners VALUES (?, 'HashiCorp', 'hc')", id).Error; err != nil {
			t.Fatalf("failed to insert row: %v", err)
		}
	}

	duplicates, err := ownersIndex.Duplicates(db)
	if err != nil || len(duplicates) != 1 {
		t.Fatalf("expected one duplicate, got %v, %v", duplicates, err)
	}

	if duplicates[0]["name"] != "hashicorp" || duplicates[0]["alias"] != "hc" {
		t.Fatalf("expected the duplicate to be keyed by column, got %v", duplicates[0])
	}
}
