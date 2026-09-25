package repositories

import (
	"errors"
	"path/filepath"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database"
	"terralist/pkg/database/factory"
	"terralist/pkg/database/sqlite"
)

func newTestProviderRepository(t *testing.T) (*DefaultProviderRepository, *authority.Authority) {
	t.Helper()

	engine, err := factory.NewDatabase(database.SQLITE, &sqlite.Config{
		Path: filepath.Join(t.TempDir(), "test.db"),
	})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := engine.Handler().AutoMigrate(&authority.Authority{}, &provider.Provider{}, &provider.Version{}, &provider.Platform{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	a := &authority.Authority{Name: "hashicorp", PolicyURL: "https://example.com", Owner: "owner@example.com"}
	if err := engine.Handler().Create(a).Error; err != nil {
		t.Fatalf("failed to create authority: %v", err)
	}

	return &DefaultProviderRepository{Database: engine}, a
}

func TestProviderRepository_UpsertAddsPlatformToSingleVersion(t *testing.T) {
	repo, a := newTestProviderRepository(t)

	created, err := repo.Upsert(provider.Provider{
		AuthorityID: a.ID,
		Name:        "random",
		Versions: []provider.Version{{
			Version:   "3.6.2",
			Protocols: "5.0",
			Platforms: []provider.Platform{{System: "linux", Architecture: "amd64", Location: "linux.zip", ShaSum: "aaaa"}},
		}},
	})
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	current, err := repo.Find("hashicorp", "random")
	if err != nil {
		t.Fatalf("failed to find provider: %v", err)
	}

	v := current.GetVersion("3.6.2")
	v.Platforms = append(v.Platforms, provider.Platform{System: "darwin", Architecture: "arm64", Location: "darwin.zip", ShaSum: "bbbb"})

	if _, err := repo.Upsert(*current); err != nil {
		t.Fatalf("expected the platform to be added to the existing provider, got: %v", err)
	}

	found, err := repo.Find("hashicorp", "random")
	if err != nil {
		t.Fatalf("failed to find provider: %v", err)
	}
	if found.ID != created.ID {
		t.Fatalf("expected the provider row to be kept")
	}
	if platforms := found.GetVersion("3.6.2").Platforms; len(platforms) != 2 {
		t.Fatalf("expected 2 platforms, got %d", len(platforms))
	}
}

func TestProviderRepository_UpsertRejectsDuplicatePlatform(t *testing.T) {
	repo, a := newTestProviderRepository(t)

	if _, err := repo.Upsert(provider.Provider{
		AuthorityID: a.ID,
		Name:        "random",
		Versions: []provider.Version{{
			Version:   "3.6.2",
			Protocols: "5.0",
			Platforms: []provider.Platform{{System: "linux", Architecture: "amd64", Location: "linux.zip", ShaSum: "aaaa"}},
		}},
	}); err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	current, _ := repo.Find("hashicorp", "random")
	v := current.GetVersion("3.6.2")
	v.Platforms = append(v.Platforms, provider.Platform{System: "linux", Architecture: "amd64", Location: "other.zip", ShaSum: "cccc"})

	if _, err := repo.Upsert(*current); err == nil {
		t.Fatalf("expected a second platform for the same system and architecture to be rejected")
	}
}

func TestProviderRepository_UpsertRejectsDuplicateProvider(t *testing.T) {
	repo, a := newTestProviderRepository(t)

	for i, want := range []error{nil, ErrAlreadyExists} {
		_, err := repo.Upsert(provider.Provider{AuthorityID: a.ID, Name: "random"})
		if !errors.Is(err, want) {
			t.Fatalf("upsert %d: expected %v, got %v", i, want, err)
		}
	}
}

func TestProviderRepository_UpsertRejectsDuplicateVersion(t *testing.T) {
	repo, a := newTestProviderRepository(t)

	if _, err := repo.Upsert(provider.Provider{
		AuthorityID: a.ID,
		Name:        "random",
		Versions:    []provider.Version{{Version: "3.6.2", Protocols: "5.0"}},
	}); err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	current, _ := repo.Find("hashicorp", "random")
	current.Versions = append(current.Versions, provider.Version{Version: "3.6.2", Protocols: "5.0"})

	if _, err := repo.Upsert(*current); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected a second 3.6.2 version to be rejected as existing, got %v", err)
	}
}
