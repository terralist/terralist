package repositories

import (
	"path/filepath"
	"testing"
	"time"

	"terralist/internal/server/models/oauth"
	"terralist/pkg/database"
	"terralist/pkg/database/factory"
	"terralist/pkg/database/sqlite"
)

func newTestOAuthCodeRepository(t *testing.T, ttl time.Duration) *DefaultOAuthCodeRepository {
	t.Helper()

	engine, err := factory.NewDatabase(database.SQLITE, &sqlite.Config{
		Path: filepath.Join(t.TempDir(), "test.db"),
	})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := engine.Handler().AutoMigrate(&oauth.Code{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return &DefaultOAuthCodeRepository{
		Database: engine,
		TTL:      ttl,
	}
}

func TestOAuthCodeRepository_PutTakeRoundTrip(t *testing.T) {
	repo := newTestOAuthCodeRepository(t, 2*time.Minute)

	components := oauth.CodeComponents{
		UserName:   "alice",
		UserEmail:  "alice@example.com",
		UserGroups: []string{"engineering", "platform"},
	}

	code, err := repo.Put(components)
	if err != nil {
		t.Fatalf("expected no error putting code, got: %v", err)
	}
	if code == "" {
		t.Fatalf("expected a non-empty opaque code")
	}

	got, ok := repo.Take(code)
	if !ok || got == nil {
		t.Fatalf("expected stored code to be retrievable")
	}
	if got.UserEmail != "alice@example.com" {
		t.Fatalf("unexpected user email: %s", got.UserEmail)
	}
	if len(got.UserGroups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(got.UserGroups))
	}
}

func TestOAuthCodeRepository_TakeIsSingleUse(t *testing.T) {
	repo := newTestOAuthCodeRepository(t, 2*time.Minute)

	code, err := repo.Put(oauth.CodeComponents{UserName: "alice"})
	if err != nil {
		t.Fatalf("expected no error putting code, got: %v", err)
	}

	if _, ok := repo.Take(code); !ok {
		t.Fatalf("expected first take to succeed")
	}

	if _, ok := repo.Take(code); ok {
		t.Fatalf("expected code to be single use")
	}
}

func TestOAuthCodeRepository_TakeUnknownCode(t *testing.T) {
	repo := newTestOAuthCodeRepository(t, 2*time.Minute)

	if _, ok := repo.Take("nonexistent-code"); ok {
		t.Fatalf("expected unknown code to miss")
	}
}

func TestOAuthCodeRepository_TakeExpiredCode(t *testing.T) {
	repo := newTestOAuthCodeRepository(t, 1*time.Second)

	base := time.Unix(1000, 0)
	repo.now = func() time.Time {
		return base
	}

	code, err := repo.Put(oauth.CodeComponents{UserName: "alice"})
	if err != nil {
		t.Fatalf("expected no error putting code, got: %v", err)
	}

	repo.now = func() time.Time {
		return base.Add(2 * time.Second)
	}

	if _, ok := repo.Take(code); ok {
		t.Fatalf("expected code to be expired")
	}
}
