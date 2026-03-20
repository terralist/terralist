package apikey

import (
	"time"

	"terralist/pkg/database/entity"

	"github.com/samber/lo"
)

type ApiKey struct {
	entity.Entity
	Name       string `gorm:"not null"`
	Scope      string `gorm:"not null"`
	CreatedBy  string `gorm:"not null"`
	Expiration *time.Time
	Policies   []Policy `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}

type ApiKeyDTO struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Scope      string      `json:"scope"`
	CreatedBy  string      `json:"created_by"`
	Expiration string      `json:"expiration"`
	Policies   []PolicyDTO `json:"policies"`
}

type CreateApiKeyDTO struct {
	Name     string            `json:"name" binding:"required"`
	Scope    string            `json:"scope" binding:"required"`
	ExpireIn int               `json:"expire_in"`
	Policies []CreatePolicyDTO `json:"policies" binding:"required,min=1"`
}

func (a ApiKey) ToDTO() ApiKeyDTO {
	var exp string
	if a.Expiration != nil {
		exp = a.Expiration.Format("2006-01-02T15:04:05")
	}

	return ApiKeyDTO{
		ID:         a.ID.String(),
		Name:       a.Name,
		Scope:      a.Scope,
		CreatedBy:  a.CreatedBy,
		Expiration: exp,
		Policies: lo.Map(a.Policies, func(p Policy, _ int) PolicyDTO {
			return p.ToDTO()
		}),
	}
}
