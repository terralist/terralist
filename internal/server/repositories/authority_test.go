package repositories

import (
	"path/filepath"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/pkg/database"
	"terralist/pkg/database/factory"
	"terralist/pkg/database/sqlite"
)

func newTestAuthorityRepository(t *testing.T) *DefaultAuthorityRepository {
	t.Helper()

	engine, err := factory.NewDatabase(database.SQLITE, &sqlite.Config{
		Path: filepath.Join(t.TempDir(), "test.db"),
	})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := engine.Handler().AutoMigrate(&authority.Authority{}, &authority.Key{}, &authority.ApiKey{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return &DefaultAuthorityRepository{Database: engine}
}

func upstreamAuthority(name, namespace string) authority.Authority {
	hostname := "registry.terraform.io"

	return authority.Authority{
		Name:              name,
		PolicyURL:         "https://example.com/" + name,
		Owner:             "owner@example.com",
		UpstreamHostname:  &hostname,
		UpstreamNamespace: &namespace,
	}
}

func TestAuthorityRepository_FindByUpstream(t *testing.T) {
	repo := newTestAuthorityRepository(t)

	if _, err := repo.Upsert(authority.Authority{Name: "local", PolicyURL: "https://example.com/local", Owner: "owner@example.com"}); err != nil {
		t.Fatalf("failed to create local authority: %v", err)
	}
	if _, err := repo.Upsert(upstreamAuthority("hashicorp-mirror", "hashicorp")); err != nil {
		t.Fatalf("failed to create upstream authority: %v", err)
	}

	got, err := repo.FindByUpstream("Registry.Terraform.io", "HashiCorp")
	if err != nil {
		t.Fatalf("expected the upstream authority to be found, got: %v", err)
	}
	if got.Name != "hashicorp-mirror" {
		t.Fatalf("unexpected authority: %s", got.Name)
	}

	if _, err := repo.FindByUpstream("registry.terraform.io", "local"); err == nil {
		t.Fatalf("expected no authority for an unknown upstream namespace")
	}
}

func TestAuthorityRepository_UpstreamIdentityIsUnique(t *testing.T) {
	repo := newTestAuthorityRepository(t)

	if _, err := repo.Upsert(upstreamAuthority("first", "hashicorp")); err != nil {
		t.Fatalf("failed to create first authority: %v", err)
	}

	if _, err := repo.Upsert(upstreamAuthority("second", "hashicorp")); err == nil {
		t.Fatalf("expected a second authority with the same upstream identity to be rejected")
	}

	if _, err := repo.Upsert(upstreamAuthority("third", "integrations")); err != nil {
		t.Fatalf("expected a different upstream namespace to be accepted, got: %v", err)
	}
}

func TestAuthorityRepository_ManyAuthoritiesWithoutUpstream(t *testing.T) {
	repo := newTestAuthorityRepository(t)

	for _, name := range []string{"one", "two"} {
		if _, err := repo.Upsert(authority.Authority{Name: name, PolicyURL: "https://example.com/" + name, Owner: "owner@example.com"}); err != nil {
			t.Fatalf("expected authority %s without upstream to be accepted, got: %v", name, err)
		}
	}
}
