package authority

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Key struct {
	entity.Entity
	AuthorityID    uuid.UUID
	KeyId          string `gorm:"not null"`
	AsciiArmor     string `gorm:"size:10000,not null"`
	TrustSignature string `gorm:"size:10000,not null"`
}

func (Key) TableName() string {
	return "authority_keys"
}

func (k Key) ToKeyDTO() KeyDTO {
	return KeyDTO{
		KeyId:          k.KeyId,
		AsciiArmor:     k.AsciiArmor,
		TrustSignature: k.TrustSignature,
	}
}

type KeyDTO struct {
	KeyId          string `json:"key_id"`
	AsciiArmor     string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
}

func (d KeyDTO) ToKey() Key {
	return Key{
		KeyId:          d.KeyId,
		AsciiArmor:     d.AsciiArmor,
		TrustSignature: d.TrustSignature,
	}
}
