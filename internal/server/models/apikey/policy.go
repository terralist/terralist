package apikey

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Policy struct {
	entity.Entity
	ApiKeyID uuid.UUID `gorm:"not null;index"`
	Resource string    `gorm:"not null"`
	Action   string    `gorm:"not null"`
	Object   string    `gorm:"not null"`
	Effect   string    `gorm:"not null"`
}

func (Policy) TableName() string {
	return "api_key_policies"
}

type PolicyDTO struct {
	ID       string `json:"id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Object   string `json:"object"`
	Effect   string `json:"effect"`
}

func (p Policy) ToDTO() PolicyDTO {
	return PolicyDTO{
		ID:       p.ID.String(),
		Resource: p.Resource,
		Action:   p.Action,
		Object:   p.Object,
		Effect:   p.Effect,
	}
}

type CreatePolicyDTO struct {
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
	Object   string `json:"object" binding:"required"`
	Effect   string `json:"effect" binding:"required"`
}

func (p CreatePolicyDTO) ToModel() Policy {
	return Policy{
		Resource: p.Resource,
		Action:   p.Action,
		Object:   p.Object,
		Effect:   p.Effect,
	}
}
