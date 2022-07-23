package authority

import (
	"terralist/pkg/database/entity"
	"time"

	"github.com/google/uuid"
)

type ApiKey struct {
	entity.Entity
	AuthorityID uuid.UUID
	Expiration  time.Time
}

func (ApiKey) TableName() string {
	return "authority_api_keys"
}
