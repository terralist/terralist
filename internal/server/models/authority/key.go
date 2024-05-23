package authority

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Key struct {
	entity.Entity
	AuthorityID    uuid.UUID `gorm:"size:256"`
	KeyId          string    `gorm:"size:256;not null"`
	AsciiArmor     string    `gorm:"size:10000,not null"`
	TrustSignature string    `gorm:"size:10000,not null"`
}

func (Key) TableName() string {
	return "authority_keys"
}

type KeyDTO struct {
	ID             string `json:"id"`
	KeyId          string `json:"key_id"`
	AsciiArmor     string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
}

func (k Key) ToKeyDTO() KeyDTO {
	return KeyDTO{
		ID:             k.ID.String(),
		KeyId:          k.KeyId,
		AsciiArmor:     k.AsciiArmor,
		TrustSignature: k.TrustSignature,
	}
}

func (d KeyDTO) ToKey() Key {
	return Key{
		KeyId:          d.KeyId,
		AsciiArmor:     d.AsciiArmor,
		TrustSignature: d.TrustSignature,
	}
}
