package rbac

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"terralist/pkg/auth"
)

func TestProtect_UsesGroupSubjects(t *testing.T) {
	t.Parallel()

	policy := `
g, role:engineering, role:developer
p, role:developer, authorities, create, *, allow
`

	policyPath := filepath.Join(t.TempDir(), "policy.csv")
	if err := os.WriteFile(policyPath, []byte(policy), 0600); err != nil {
		t.Fatalf("failed to write policy file: %v", err)
	}

	enforcer, err := NewEnforcer(policyPath, "readonly")
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
	}

	groupMember := auth.User{
		Name:   "alice",
		Email:  "alice@example.com",
		Groups: []string{"engineering"},
	}

	if err := enforcer.Protect(groupMember, ResourceAuthorities, ActionCreate, "example-org"); err != nil {
		t.Fatalf("expected group-based authorization to pass, got: %v", err)
	}

	nonMember := auth.User{
		Name:  "bob",
		Email: "bob@example.com",
	}

	err = enforcer.Protect(nonMember, ResourceAuthorities, ActionCreate, "example-org")
	if !errors.Is(err, ErrUnauthorizedSubject) {
		t.Fatalf("expected unauthorized for user without group role, got: %v", err)
	}
}

func TestProtect_SettingsRequiresExplicitPolicy(t *testing.T) {
	t.Parallel()

	policyPath := filepath.Join(t.TempDir(), "policy.csv")
	if err := os.WriteFile(policyPath, []byte(""), 0600); err != nil {
		t.Fatalf("failed to write policy file: %v", err)
	}

	enforcer, err := NewEnforcer(policyPath, "readonly")
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
	}

	readonlyUser := auth.User{
		Name:  "bob",
		Email: "bob@example.com",
	}

	err = enforcer.Protect(readonlyUser, ResourceSettings, ActionGet, "page")
	if !errors.Is(err, ErrUnauthorizedSubject) {
		t.Fatalf("expected readonly user to be unauthorized for settings without policy, got: %v", err)
	}

	adminPolicyPath := filepath.Join(t.TempDir(), "admin-policy.csv")
	if err := os.WriteFile(adminPolicyPath, []byte("g, alice@example.com, role:admin"), 0600); err != nil {
		t.Fatalf("failed to write admin policy file: %v", err)
	}

	adminEnforcer, err := NewEnforcer(adminPolicyPath, "readonly")
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
	}

	adminUser := auth.User{
		Name:   "alice",
		Email:  "alice@example.com",
	}

	if err := adminEnforcer.Protect(adminUser, ResourceSettings, ActionGet, "page"); err != nil {
		t.Fatalf("expected admin user to be allowed for settings, got: %v", err)
	}
}
