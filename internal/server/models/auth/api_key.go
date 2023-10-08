package auth

import (
	"time"

	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type ApiKey struct {
	entity.Entity
	Label      string `gorm:"not null"`
	Expiration *time.Time
	Policies   []Policy `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}

type ApiKeyDTO struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Expiration string `json:"expiration"`
}

func (a ApiKey) ToDTO() ApiKeyDTO {
	var exp string = ""
	if a.Expiration != nil {
		exp = a.Expiration.Format("2006-01-02T15:04:05")
	}

	return ApiKeyDTO{
		ID:         a.ID.String(),
		Label:      a.Label,
		Expiration: exp,
	}
}

func (d ApiKeyDTO) ToApiKey() ApiKey {
	var exp *time.Time
	{
		expiration, err := time.Parse("2006-01-02T15:04:05", d.Expiration)
		if err != nil {
			exp = nil
		} else {
			exp = &expiration
		}
	}

	return ApiKey{
		Entity: entity.Entity{
			ID: uuid.MustParse(d.ID),
		},
		Label:      d.Label,
		Expiration: exp,
	}
}
