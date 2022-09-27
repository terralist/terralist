package provider

import (
	"strings"

	"terralist/pkg/database/entity"
	"terralist/pkg/version"

	"github.com/google/uuid"
)

type Provider struct {
	entity.Entity
	AuthorityID uuid.UUID
	Name        string    `gorm:"not null;index"`
	Versions    []Version `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
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

func (p Provider) GetVersion(v string) *Version {
	vv := version.Version(v)

	for _, ver := range p.Versions {
		if version.Compare(version.Version(ver.Version), vv) == 0 {
			return &ver
		}
	}

	return nil
}

type CreateProviderDTO struct {
	AuthorityID uuid.UUID
	Name        string
	Version     string
	ShaSums     CreateProviderShaSumsDTO `json:"shasums"`
	Protocols   []string                 `json:"protocols"`
	Platforms   []CreatePlatformDTO      `json:"platforms"`
}

func (d CreateProviderDTO) ToProvider() Provider {
	var platforms []Platform
	for _, p := range d.Platforms {
		platforms = append(platforms, p.ToPlatform())
	}

	return Provider{
		AuthorityID: d.AuthorityID,
		Name:        d.Name,
		Versions: []Version{
			{
				Version:             d.Version,
				ShaSumsUrl:          d.ShaSums.URL,
				ShaSumsSignatureUrl: d.ShaSums.SignatureURL,
				Protocols:           strings.Join(d.Protocols, ","),
				Platforms:           platforms,
			},
		},
	}
}

type CreateProviderShaSumsDTO struct {
	URL          string `json:"url"`
	SignatureURL string `json:"signature_url"`
}

type VersionListProviderDTO struct {
	Versions []VersionListVersionDTO `json:"versions"`
}
