package provider

import (
	"fmt"
	"strings"

	"terralist/internal/server/models/authority"
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

func (v Version) ToDownloadVersionDTO(os string, arch string) (DownloadVersionDTO, error) {
	filename := fmt.Sprintf(
		"terraform-provider-%s_%s_%s_%s.zip",
		v.Provider.Name,
		v.Version,
		os,
		arch,
	)

	var out DownloadVersionDTO
	for _, platform := range v.Platforms {
		if platform.System == os && platform.Architecture == arch {
			out.System = os
			out.Architecture = arch
			out.FileName = filename
			out.DownloadUrl = platform.Location
			out.ShaSumsUrl = v.ShaSumsUrl
			out.ShaSumsSignatureUrl = v.ShaSumsSignatureUrl
			out.ShaSum = platform.ShaSum
			out.Protocols = strings.Split(v.Protocols, ",")
			out.SigningKeys = v.Provider.Authority.ToAuthorityKeysDTO()

			return out, nil
		}
	}

	return out, fmt.Errorf("no platform found for %s_%s machine", os, arch)
}

type VersionListVersionDTO struct {
	Version   string                   `json:"version"`
	Protocols []string                 `json:"protocols"`
	Platforms []VersionListPlatformDTO `json:"platforms"`
}

type DownloadVersionDTO struct {
	Protocols           []string                   `json:"protocols"`
	System              string                     `json:"os"`
	Architecture        string                     `json:"arch"`
	FileName            string                     `json:"filename"`
	DownloadUrl         string                     `json:"download_url"`
	ShaSumsUrl          string                     `json:"shasums_url"`
	ShaSumsSignatureUrl string                     `json:"shasums_signature_url"`
	ShaSum              string                     `json:"shasum"`
	SigningKeys         authority.AuthorityKeysDTO `json:"signing_keys"`
}
