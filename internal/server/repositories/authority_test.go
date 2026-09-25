package repositories

import (
	"path/filepath"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/pkg/database"
	"terralist/pkg/database/factory"
	"terralist/pkg/database/sqlite"

	"github.com/google/uuid"
)

func newTestAuthorityRepository(t *testing.T) *DefaultAuthorityRepository {
	t.Helper()

	engine, err := factory.NewDatabase(database.SQLITE, &sqlite.Config{
		Path: filepath.Join(t.TempDir(), "test.db"),
	})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := engine.Handler().AutoMigrate(&authority.Authority{}, &authority.Key{}, &authority.Rule{}); err != nil {
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

func TestAuthorityRepository_Rules(t *testing.T) {
	repo := newTestAuthorityRepository(t)

	a := upstreamAuthority("hashicorp", "hashicorp")
	a.Rules = []authority.Rule{{Kind: authority.RuleKindProvider, Name: "aws", Version: "*", Effect: authority.EffectDeny}}

	created, err := repo.Upsert(a)
	if err != nil {
		t.Fatalf("failed to create authority: %v", err)
	}
	if len(created.Rules) != 1 || created.Rules[0].ID == (uuid.UUID{}) {
		t.Fatalf("expected the rule to be stored with an id, got %+v", created.Rules)
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("failed to find authority: %v", err)
	}
	if len(found.Rules) != 1 || found.Rules[0].Name != "aws" {
		t.Fatalf("expected the rule to be loaded with the authority, got %+v", found.Rules)
	}

	byUpstream, err := repo.FindByUpstream("registry.terraform.io", "hashicorp")
	if err != nil {
		t.Fatalf("failed to find authority by upstream: %v", err)
	}
	if len(byUpstream.Rules) != 1 {
		t.Fatalf("expected the rule to be loaded by upstream lookup, got %+v", byUpstream.Rules)
	}

	if err := repo.DeleteRule(found.Rules[0].ID); err != nil {
		t.Fatalf("failed to delete rule: %v", err)
	}

	found, err = repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("failed to find authority: %v", err)
	}
	if len(found.Rules) != 0 {
		t.Fatalf("expected no rules after deletion, got %+v", found.Rules)
	}
}

func TestAuthorityRepository_UpsertCreatesWithGivenID(t *testing.T) {
	repo := newTestAuthorityRepository(t)

	a := upstreamAuthority("hashicorp", "hashicorp")
	a.ID = uuid.New()

	if _, err := repo.Upsert(a); err != nil {
		t.Fatalf("failed to create authority with a given ID: %v", err)
	}

	found, err := repo.FindByID(a.ID)
	if err != nil {
		t.Fatalf("expected the authority to be stored under its ID: %v", err)
	}

	if found.Name != "hashicorp" {
		t.Fatalf("expected the stored authority, got %+v", found)
	}
}

func TestAuthorityRepository_FindByNameIgnoresCase(t *testing.T) {
	repo := newTestAuthorityRepository(t)

	if _, err := repo.Upsert(authority.Authority{Name: "HashiCorp", PolicyURL: "https://example.com/hashicorp", Owner: "owner@example.com"}); err != nil {
		t.Fatalf("failed to create authority: %v", err)
	}

	for _, name := range []string{"HashiCorp", "hashicorp", "HASHICORP"} {
		found, err := repo.FindByName(name)
		if err != nil {
			t.Fatalf("expected %q to find the authority, got: %v", name, err)
		}

		if found.Name != "HashiCorp" {
			t.Fatalf("expected the name to be kept as stored, got %q", found.Name)
		}
	}
}
