package repositories

import (
	"errors"
	"path/filepath"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/pkg/database"
	"terralist/pkg/database/factory"
	"terralist/pkg/database/sqlite"
)

func newTestModuleRepository(t *testing.T) (*DefaultModuleRepository, *authority.Authority) {
	t.Helper()

	engine, err := factory.NewDatabase(database.SQLITE, &sqlite.Config{
		Path: filepath.Join(t.TempDir(), "test.db"),
	})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := engine.Handler().AutoMigrate(
		&authority.Authority{},
		&module.Module{},
		&module.Version{},
		&module.Submodule{},
		&module.Provider{},
		&module.Dependency{},
	); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	for _, index := range module.UniqueIndexes {
		if err := index.Create(engine.Handler()); err != nil {
			t.Fatalf("failed to create index %s: %v", index.Name, err)
		}
	}

	a := &authority.Authority{Name: "hashicorp", PolicyURL: "https://example.com", Owner: "owner@example.com"}
	if err := engine.Handler().Create(a).Error; err != nil {
		t.Fatalf("failed to create authority: %v", err)
	}

	return &DefaultModuleRepository{Database: engine}, a
}

func TestModuleRepository_UpsertRejectsDuplicateModule(t *testing.T) {
	repo, a := newTestModuleRepository(t)

	for i, tc := range []struct {
		name, system string
		want         error
	}{
		{"dir", "template", nil},
		{"dir", "template", ErrAlreadyExists},
		{"Dir", "Template", ErrAlreadyExists},
	} {
		_, err := repo.Upsert(module.Module{
			AuthorityID: a.ID,
			Name:        tc.name,
			Provider:    tc.system,
			Versions:    []module.Version{{Version: "1.0.0", Location: "1.0.0.zip"}},
		})
		if !errors.Is(err, tc.want) {
			t.Fatalf("upsert %d of %s/%s: expected %v, got %v", i, tc.name, tc.system, tc.want, err)
		}
	}
}

func TestModuleRepository_UpsertRejectsDuplicateVersion(t *testing.T) {
	repo, a := newTestModuleRepository(t)

	if _, err := repo.Upsert(module.Module{
		AuthorityID: a.ID,
		Name:        "dir",
		Provider:    "template",
		Versions:    []module.Version{{Version: "1.0.0", Location: "1.0.0.zip"}},
	}); err != nil {
		t.Fatalf("failed to create module: %v", err)
	}

	current, err := repo.Find("hashicorp", "dir", "template")
	if err != nil {
		t.Fatalf("failed to find module: %v", err)
	}

	current.Versions = append(current.Versions, module.Version{Version: "1.0.0", Location: "other.zip"})

	if _, err := repo.Upsert(*current); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected a second 1.0.0 version to be rejected as existing, got %v", err)
	}
}

func TestModuleRepository_UpsertAddsVersion(t *testing.T) {
	repo, a := newTestModuleRepository(t)

	if _, err := repo.Upsert(module.Module{
		AuthorityID: a.ID,
		Name:        "dir",
		Provider:    "template",
		Versions:    []module.Version{{Version: "1.0.0", Location: "1.0.0.zip"}},
	}); err != nil {
		t.Fatalf("failed to create module: %v", err)
	}

	current, err := repo.Find("hashicorp", "dir", "template")
	if err != nil {
		t.Fatalf("failed to find module: %v", err)
	}

	current.Versions = append(current.Versions, module.Version{Version: "1.0.1", Location: "1.0.1.zip"})

	if _, err := repo.Upsert(*current); err != nil {
		t.Fatalf("expected the version to be added to the existing module, got %v", err)
	}

	found, err := repo.Find("hashicorp", "dir", "template")
	if err != nil {
		t.Fatalf("failed to find module: %v", err)
	}

	if len(found.Versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(found.Versions))
	}
}
