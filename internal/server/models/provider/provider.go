package provider

import (
	"strings"

	"terralist/internal/server/models/artifact"
	"terralist/pkg/database/entity"
	"terralist/pkg/version"

	"github.com/google/uuid"
	"github.com/ssoroka/slice"
)

type Provider struct {
	entity.Entity
	AuthorityID uuid.UUID `gorm:"size:256"`
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

func (p Provider) ToArtifact() artifact.Artifact {
	return artifact.Artifact{
		ID:   p.ID.String(),
		Name: p.Name,
		Type: artifact.TypeProvider,
		Versions: slice.Map[Version, string](p.Versions, func(v Version) string {
			return v.Version
		}),
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05"),
	}
}

type CreateProviderDTO struct {
	AuthorityID uuid.UUID `gorm:"size:256"`
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
