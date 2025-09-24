package provider

import (
	"strings"

	"terralist/internal/server/models/artifact"
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Version struct {
	entity.Entity
	ProviderID          uuid.UUID
	Provider            Provider
	Version             string     `gorm:"not null"`
	Protocols           string     `gorm:"not null"`
	ShaSumsUrl          string     `gorm:"shasums_url"`
	ShaSumsSignatureUrl string     `gorm:"shasums_signature_url"`
	Platforms           []Platform `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Version) TableName() string {
	return "provider_versions"
}

func (v Version) ToVersionListVersionDTO() VersionListVersionDTO {
	var platforms []VersionListPlatformDTO
	for _, p := range v.Platforms {
		platforms = append(platforms, p.ToVersionListPlatformDTO())
	}

	return VersionListVersionDTO{
		Version:   v.Version,
		Protocols: strings.Split(v.Protocols, ","),
		Platforms: platforms,
	}
}

func (v Version) ToArtifactVersion() artifact.Version {
	return artifact.Version{
		Tag: v.Version,
	}
}

type VersionListVersionDTO struct {
	Version   string                   `json:"version"`
	Protocols []string                 `json:"protocols"`
	Platforms []VersionListPlatformDTO `json:"platforms"`
}

type VersionAllPlatformsDTO struct {
	Version   string        `json:"version"`
	Protocols []string      `json:"protocols"`
	Platforms []PlatformDTO `json:"platforms"`
}

type PlatformDTO struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	DownloadURL string `json:"download_url"`
	Shasum      string `json:"shasum"`
}
