package services

import (
	"testing"
	"time"

	"terralist/internal/server/models/oauth"
)

func TestInMemoryOAuthCodeStore_TakeSingleUse(t *testing.T) {
	store := NewInMemoryOAuthCodeStore(2 * time.Minute)

	components := oauth.CodeComponents{
		UserName:  "alice",
		UserEmail: "alice@example.com",
	}

	code, err := store.Put(components)
	if err != nil {
		t.Fatalf("expected no error generating code, got: %v", err)
	}

	got, ok := store.Take(code)
	if !ok || got == nil {
		t.Fatalf("expected stored code to be retrievable")
	}

	if got.UserEmail != "alice@example.com" {
		t.Fatalf("unexpected user email: %s", got.UserEmail)
	}

	if _, ok := store.Take(code); ok {
		t.Fatalf("expected code to be single use")
	}
}

func TestInMemoryOAuthCodeStore_Expires(t *testing.T) {
	store := NewInMemoryOAuthCodeStore(1 * time.Second)
	inMemoryStore := store.(*inMemoryOAuthCodeStore)

	base := time.Unix(1000, 0)
	inMemoryStore.now = func() time.Time {
		return base
	}

	code, err := store.Put(oauth.CodeComponents{UserName: "alice"})
	if err != nil {
		t.Fatalf("expected no error generating code, got: %v", err)
	}

	inMemoryStore.now = func() time.Time {
		return base.Add(2 * time.Second)
	}

	if _, ok := store.Take(code); ok {
		t.Fatalf("expected code to be expired")
	}
}
