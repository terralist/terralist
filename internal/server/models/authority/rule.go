package authority

import (
	"fmt"

	"terralist/pkg/database/entity"

	"github.com/gobwas/glob"
	"github.com/google/uuid"
)

const (
	RuleKindProvider = "provider"
	RuleKindModule   = "module"

	EffectAllow = "allow"
	EffectDeny  = "deny"

	PolicyAllow = "allow"
	PolicyDeny  = "deny"
)

// Rule allows or denies serving versions of upstream artifacts through an
// authority. Name and Version are globs; for modules the name is
// name/system.
type Rule struct {
	entity.Entity
	AuthorityID uuid.UUID
	Kind        string `gorm:"not null"`
	Name        string `gorm:"not null"`
	Version     string `gorm:"not null"`
	Effect      string `gorm:"not null"`
}

func (Rule) TableName() string {
	return "authority_rules"
}

// Validate checks the rule fields and compiles its globs.
func (r Rule) Validate() error {
	if r.Kind != RuleKindProvider && r.Kind != RuleKindModule {
		return fmt.Errorf("invalid rule kind %q, expected %s or %s", r.Kind, RuleKindProvider, RuleKindModule)
	}

	if r.Effect != EffectAllow && r.Effect != EffectDeny {
		return fmt.Errorf("invalid rule effect %q, expected %s or %s", r.Effect, EffectAllow, EffectDeny)
	}

	for field, pattern := range map[string]string{"name": r.Name, "version": r.Version} {
		if pattern == "" {
			return fmt.Errorf("the rule %s glob is required", field)
		}

		if _, err := glob.Compile(pattern); err != nil {
			return fmt.Errorf("invalid rule %s glob %q: %v", field, pattern, err)
		}
	}

	return nil
}

// Matches reports whether the rule applies to the given artifact version.
func (r Rule) Matches(kind, name, version string) bool {
	return r.Kind == kind && matchGlob(r.Name, name) && matchGlob(r.Version, version)
}

func matchGlob(pattern, value string) bool {
	compiled, err := glob.Compile(pattern)
	if err != nil {
		return false
	}

	return compiled.Match(value)
}

type RuleDTO struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Effect  string `json:"effect"`
}

func (r Rule) ToDTO() RuleDTO {
	return RuleDTO{
		ID:      r.ID.String(),
		Kind:    r.Kind,
		Name:    r.Name,
		Version: r.Version,
		Effect:  r.Effect,
	}
}

func (d RuleDTO) ToRule() Rule {
	return Rule{
		Kind:    d.Kind,
		Name:    d.Name,
		Version: d.Version,
		Effect:  d.Effect,
	}
}
