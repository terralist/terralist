package provider

import (
	"fmt"
	"strings"
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Platform struct {
	entity.Entity
	VersionID    uuid.UUID
	Version      Version
	System       string `gorm:"not null"`
	Architecture string `gorm:"not null"`
	Location     string `gorm:"not null"`
	ShaSum       string `gorm:"not null"`
}

func (Platform) TableName() string {
	return "provider_platforms"
}

func (p Platform) String() string {
	return fmt.Sprintf("%s_%s", p.System, p.Architecture)
}

func (p Platform) ToVersionListPlatformDTO() VersionListPlatformDTO {
	return VersionListPlatformDTO{
		System:       p.System,
		Architecture: p.Architecture,
	}
}

type CreatePlatformDTO struct {
	System       string `json:"os"`
	Architecture string `json:"arch"`
	Location     string `json:"download_url"`
	ShaSum       string `json:"shasum"`
}

func (d CreatePlatformDTO) ToPlatform() Platform {
	return Platform{
		System:       d.System,
		Architecture: d.Architecture,
		Location:     d.Location,
		ShaSum:       d.ShaSum,
	}
}

type VersionListPlatformDTO struct {
	System       string `json:"os"`
	Architecture string `json:"arch"`
}

func (p Platform) ToDownloadPlatformDTO(keys SigningKeysDTO) DownloadPlatformDTO {
	fileName := fmt.Sprintf(
		"terraform-provider-%s_%s_%s_%s.zip",
		p.Version.Provider.Name,
		p.Version,
		p.System,
		p.Architecture,
	)

	return DownloadPlatformDTO{
		System:              p.System,
		Architecture:        p.Architecture,
		FileName:            fileName,
		DownloadUrl:         p.Location,
		ShaSumsUrl:          p.Version.ShaSumsUrl,
		ShaSumsSignatureUrl: p.Version.ShaSumsSignatureUrl,
		ShaSum:              p.ShaSum,
		Protocols:           strings.Split(p.Version.Protocols, ","),
		SigningKeys:         keys,
	}
}

type DownloadPlatformDTO struct {
	Protocols           []string       `json:"protocols"`
	System              string         `json:"os"`
	Architecture        string         `json:"arch"`
	FileName            string         `json:"filename"`
	DownloadUrl         string         `json:"download_url"`
	ShaSumsUrl          string         `json:"shasums_url"`
	ShaSumsSignatureUrl string         `json:"shasums_signature_url"`
	ShaSum              string         `json:"shasum"`
	SigningKeys         SigningKeysDTO `json:"signing_keys"`
}

type SigningKeysDTO struct {
	Keys []PublicKeyDTO `json:"gpg_public_keys"`
}

type PublicKeyDTO struct {
	KeyId          string `json:"key_id"`
	AsciiArmor     string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
	Source         string `json:"string"`
	SourceURL      string `json:"source_url"`
}
