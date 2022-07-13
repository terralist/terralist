package provider

import (
	"strings"
	"terralist/pkg/database/entity"
)

type Provider struct {
	entity.Entity
	Name      string    `gorm:"not null"`
	Namespace string    `gorm:"not null"`
	Versions  []Version `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Provider) TableName() string {
	return "providers"
}

func (p Provider) ToVersionListProviderDTO() VersionListProviderDTO {
	var versions []VersionListVersionDTO
	for _, v := range p.Versions {
		versions = append(versions, v.ToVersionListVersionDTO())
	}

	return VersionListProviderDTO{
		Versions: versions,
	}
}

type CreateProviderDTO struct {
	Name      string              `json:"name"`
	Namespace string              `json:"namespace"`
	Version   string              `json:"version"`
	Protocols []string            `json:"protocols"`
	Platforms []CreatePlatformDTO `json:"platforms"`
}

func (d CreateProviderDTO) ToProvider() Provider {
	var platforms []Platform
	for _, p := range d.Platforms {
		platforms = append(platforms, p.ToPlatform())
	}

	return Provider{
		Name:      d.Name,
		Namespace: d.Namespace,
		Versions: []Version{
			{
				Version:   d.Version,
				Protocols: strings.Join(d.Protocols, ","),
				Platforms: platforms,
			},
		},
	}
}

type VersionListProviderDTO struct {
	Versions []VersionListVersionDTO `json:"versions"`
}

type DownloadProviderDTO struct {
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
