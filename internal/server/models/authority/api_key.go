package authority

import (
	"time"

	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type ApiKey struct {
	entity.Entity
	AuthorityID uuid.UUID
	Expiration  *time.Time
}

func (ApiKey) TableName() string {
	return "authority_api_keys"
}

type ApiKeyDTO struct {
	ID string `json:"id"`
}

func (a ApiKey) ToDTO() ApiKeyDTO {
	return ApiKeyDTO{
		ID: a.ID.String(),
	}
}
