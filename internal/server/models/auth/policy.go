package auth

import (
	"terralist/internal/server/models/auth/action"
	"terralist/pkg/database/entity"
	"terralist/pkg/matcher"
	"terralist/pkg/types"

	"github.com/google/uuid"
)

type Permission = uint8

const (
	PermissionGranted   = 0
	PermissionDenied    = 1
	PermissionUncertain = 2
)

type Effect = string

const (
	EffectAllow = "Allow"
	EffectDeny  = "Deny"
)

type Policy struct {
	entity.Entity
	ApiKeyID  uuid.UUID
	Label     string            `gorm:"not null" yaml:"label" json:"label"`
	Effect    Effect            `gorm:"not null" yaml:"effect" json:"effect"`
	Actions   types.StringArray `gorm:"not null" yaml:"actions" json:"actions"`
	Resources types.StringArray `gorm:"not null" yaml:"resources" json:"resources"`
}

func (Policy) TableName() string {
	return "policies"
}

// Evaluate checks whether a given action on a given resource is either allowed or
// denied by the current policy.
//
// If the policy cannot answer to the question, it will return an uncertain response.
func (p *Policy) Evaluate(action action.Action, resource Resource) Permission {
	hasAction := p.Actions.Any(func(act string) bool {
		return matcher.Match(action.String(), act)
	})

	hasResource := p.Resources.Any(func(res string) bool {
		return matcher.Match(resource, res)
	})

	if hasAction && hasResource {
		switch p.Effect {
		case EffectAllow:
			return PermissionGranted
		case EffectDeny:
			return PermissionDenied
		default:
			return PermissionUncertain
		}
	}

	return PermissionUncertain
}
