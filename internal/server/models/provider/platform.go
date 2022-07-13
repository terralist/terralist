package provider

import (
	"terralist/pkg/database/entity"
	"terralist/pkg/database/types/uuid"
)

type Platform struct {
	entity.Entity
	VersionID           uuid.ID `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	System              string  `gorm:"not null"`
	Architecture        string  `gorm:"not null"`
	FileName            string  `gorm:"not null"`
	DownloadUrl         string  `gorm:"not null"`
	ShaSumsUrl          string  `gorm:"not null"`
	ShaSumsSignatureUrl string  `gorm:"not null"`
	ShaSum              string  `gorm:"not null"`
	SigningKeys         []GpgPublicKey
}

func (Platform) TableName() string {
	return "provider_platforms"
}

func (p Platform) ToVersionListPlatformDTO() VersionListPlatformDTO {
	return VersionListPlatformDTO{
		System:       p.System,
		Architecture: p.Architecture,
	}
}

type CreatePlatformDTO struct {
	System              string         `json:"os"`
	Architecture        string         `json:"arch"`
	FileName            string         `json:"filename"`
	DownloadUrl         string         `json:"download_url"`
	ShaSumsUrl          string         `json:"shasums_url"`
	ShaSumsSignatureUrl string         `json:"shasums_signature_url"`
	ShaSum              string         `json:"shasum"`
	SigningKeys         SigningKeysDTO `json:"signing_keys"`
}

func (d CreatePlatformDTO) ToPlatform() Platform {
	var signingKeys []GpgPublicKey

	for _, signingKey := range d.SigningKeys.GpgPublicKeys {
		signingKeys = append(signingKeys, signingKey.ToGpgPublicKey())
	}

	return Platform{
		System:              d.System,
		Architecture:        d.Architecture,
		FileName:            d.FileName,
		DownloadUrl:         d.DownloadUrl,
		ShaSumsUrl:          d.ShaSumsUrl,
		ShaSumsSignatureUrl: d.ShaSumsSignatureUrl,
		ShaSum:              d.ShaSum,
		SigningKeys:         signingKeys,
	}
}

type VersionListPlatformDTO struct {
	System       string `json:"os"`
	Architecture string `json:"arch"`
}
