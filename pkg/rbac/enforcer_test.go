package rbac

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"terralist/pkg/auth"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEvaluateInline(t *testing.T) {
	Convey("Subject: Evaluating inline policies", t, func() {
		Convey("Given an allow policy for modules get", func() {
			policies := []auth.Policy{
				{Resource: "modules", Action: "get", Object: "*", Effect: "allow"},
			}

			Convey("When checking a matching request", func() {
				result := EvaluateInline(policies, "modules", "get", "my-authority/my-module/aws")

				Convey("Then it should be allowed", func() {
					So(result, ShouldBeTrue)
				})
			})

			Convey("When checking a non-matching resource", func() {
				result := EvaluateInline(policies, "providers", "get", "my-authority/my-provider")

				Convey("Then it should be denied", func() {
					So(result, ShouldBeFalse)
				})
			})

			Convey("When checking a non-matching action", func() {
				result := EvaluateInline(policies, "modules", "create", "my-authority/my-module/aws")

				Convey("Then it should be denied", func() {
					So(result, ShouldBeFalse)
				})
			})
		})

		Convey("Given a wildcard policy", func() {
			policies := []auth.Policy{
				{Resource: "*", Action: "*", Object: "*", Effect: "allow"},
			}

			Convey("When checking any request", func() {
				result := EvaluateInline(policies, "api-keys", "delete", "some-key")

				Convey("Then it should be allowed", func() {
					So(result, ShouldBeTrue)
				})
			})
		})

		Convey("Given an allow and a deny policy for the same resource", func() {
			policies := []auth.Policy{
				{Resource: "modules", Action: "*", Object: "*", Effect: "allow"},
				{Resource: "modules", Action: "delete", Object: "*", Effect: "deny"},
			}

			Convey("When checking the denied action", func() {
				result := EvaluateInline(policies, "modules", "delete", "my-authority/my-module/aws")

				Convey("Then it should be denied (deny takes precedence)", func() {
					So(result, ShouldBeFalse)
				})
			})

			Convey("When checking a non-denied action", func() {
				result := EvaluateInline(policies, "modules", "get", "my-authority/my-module/aws")

				Convey("Then it should be allowed", func() {
					So(result, ShouldBeTrue)
				})
			})
		})

		Convey("Given a scoped object policy", func() {
			policies := []auth.Policy{
				{Resource: "modules", Action: "get", Object: "my-authority/*", Effect: "allow"},
			}

			Convey("When checking a matching object", func() {
				result := EvaluateInline(policies, "modules", "get", "my-authority/my-module/aws")

				Convey("Then it should be allowed", func() {
					So(result, ShouldBeTrue)
				})
			})

			Convey("When checking a non-matching object", func() {
				result := EvaluateInline(policies, "modules", "get", "other-authority/my-module/aws")

				Convey("Then it should be denied", func() {
					So(result, ShouldBeFalse)
				})
			})
		})

		Convey("Given no policies", func() {
			Convey("When checking any request", func() {
				result := EvaluateInline(nil, "modules", "get", "anything")

				Convey("Then it should be denied", func() {
					So(result, ShouldBeFalse)
				})
			})
		})
	})
}

func TestProtect_InlinePolicies(t *testing.T) {
	Convey("Subject: Protect with inline policies", t, func() {
		enforcer, err := NewEnforcer("", "readonly")
		So(err, ShouldBeNil)

		Convey("Given a user with inline policies", func() {
			user := auth.User{
				Name:  "apikey:some-uuid",
				Email: "creator@example.com",
				InlinePolicies: []auth.Policy{
					{Resource: "modules", Action: "get", Object: "*", Effect: "allow"},
				},
			}

			Convey("When checking a matching request", func() {
				err := enforcer.Protect(user, "modules", "get", "my-authority/my-module")

				Convey("Then it should be allowed", func() {
					So(err, ShouldBeNil)
				})
			})

			Convey("When checking a non-matching request", func() {
				err := enforcer.Protect(user, "modules", "create", "my-authority/my-module")

				Convey("Then it should be denied", func() {
					So(err, ShouldEqual, ErrUnauthorizedSubject)
				})
			})

			Convey("When checking a resource not in inline policies", func() {
				err := enforcer.Protect(user, "api-keys", "get", "*")

				Convey("Then it should be denied (inline policies bypass global policies)", func() {
					So(err, ShouldEqual, ErrUnauthorizedSubject)
				})
			})
		})

		Convey("Given a user without inline policies", func() {
			user := auth.User{
				Name:  "regular-user",
				Email: "user@example.com",
			}

			Convey("When checking a get request on modules", func() {
				err := enforcer.Protect(user, "modules", "get", "my-authority/my-module")

				Convey("Then it should use global policies (readonly allows get)", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})
}

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
		Name:  "alice",
		Email: "alice@example.com",
	}

	if err := adminEnforcer.Protect(adminUser, ResourceSettings, ActionGet, "page"); err != nil {
		t.Fatalf("expected admin user to be allowed for settings, got: %v", err)
	}
}
