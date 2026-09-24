package provider

import (
	"fmt"
	"strings"
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Platform struct {
	entity.Entity
	VersionID    uuid.UUID `gorm:"uniqueIndex:idx_provider_platforms_platform"`
	Version      Version
	System       string `gorm:"not null;uniqueIndex:idx_provider_platforms_platform"`
	Architecture string `gorm:"not null;uniqueIndex:idx_provider_platforms_platform"`
	Location     string `gorm:"not null"`
	ShaSum       string `gorm:"not null"`
	Origin       string `gorm:"not null;default:manual"`

	// H1 is the package hash in Terraform's h1 scheme, when the uploader
	// provided it. Terralist never computes it.
	H1 string
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
	H1           string `json:"h1,omitempty"`
}

func (d CreatePlatformDTO) ToPlatform() Platform {
	return Platform{
		System:       d.System,
		Architecture: d.Architecture,
		Location:     d.Location,
		ShaSum:       d.ShaSum,
		H1:           d.H1,
	}
}

type VersionListPlatformDTO struct {
	System       string `json:"os"`
	Architecture string `json:"arch"`
}

// PackageFileName returns the file name of a provider package as published by
// registries and by `terraform providers mirror`.
func PackageFileName(name, version, system, architecture string) string {
	return fmt.Sprintf("terraform-provider-%s_%s_%s_%s.zip", name, version, system, architecture)
}

// FetchResultDTO reports the outcome of fetching one platform from the
// upstream registry.
type FetchResultDTO struct {
	Platform string `json:"platform"`
	Error    string `json:"error,omitempty"`
}

func (p Platform) ToDownloadPlatformDTO(keys SigningKeysDTO) DownloadPlatformDTO {
	fileName := PackageFileName(p.Version.Provider.Name, p.Version.Version, p.System, p.Architecture)

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
