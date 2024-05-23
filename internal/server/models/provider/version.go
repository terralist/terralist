package provider

import (
	"strings"

	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Version struct {
	entity.Entity
	ProviderID          uuid.UUID `gorm:"size:256"`
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

type VersionListVersionDTO struct {
	Version   string                   `json:"version"`
	Protocols []string                 `json:"protocols"`
	Platforms []VersionListPlatformDTO `json:"platforms"`
}
