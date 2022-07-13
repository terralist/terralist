package provider

import (
	"terralist/pkg/database/entity"
	"terralist/pkg/database/types/uuid"
)

type GpgPublicKey struct {
	entity.Entity
	PlatformID     uuid.ID `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	KeyId          string  `gorm:"not null"`
	AsciiArmor     string  `gorm:"not null"`
	TrustSignature string  `gorm:"not null"`
	Source         string  `gorm:"not null"`
	SourceUrl      string  `gorm:"not null"`
}

func (GpgPublicKey) TableName() string {
	return "provider_public_keys"
}

func (g GpgPublicKey) ToGpgPublicKeyDTO() GpgPublicKeyDTO {
	return GpgPublicKeyDTO{
		KeyId:          g.KeyId,
		AsciiArmor:     g.AsciiArmor,
		TrustSignature: g.TrustSignature,
		Source:         g.Source,
		SourceUrl:      g.SourceUrl,
	}
}

type SigningKeysDTO struct {
	GpgPublicKeys []GpgPublicKeyDTO `json:"gpg_public_keys"`
}

type GpgPublicKeyDTO struct {
	KeyId          string `json:"key_id"`
	AsciiArmor     string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
	Source         string `json:"source"`
	SourceUrl      string `json:"source_url"`
}

func (d GpgPublicKeyDTO) ToGpgPublicKey() GpgPublicKey {
	return GpgPublicKey{
		KeyId:          d.KeyId,
		AsciiArmor:     d.AsciiArmor,
		TrustSignature: d.TrustSignature,
		Source:         d.Source,
		SourceUrl:      d.SourceUrl,
	}
}
