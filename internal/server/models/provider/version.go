package provider

import (
	"fmt"

	"terralist/pkg/database/entity"
	"terralist/pkg/database/types/array"
	"terralist/pkg/database/types/uuid"
)

type Version struct {
	entity.Entity
	ProviderID uuid.ID           `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Version    string            `gorm:"not null"`
	Protocols  array.StringArray `gorm:"not null"`
	Platforms  []Platform
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
		Protocols: v.Protocols,
		Platforms: platforms,
	}
}

type VersionListVersionDTO struct {
	Version   string                   `json:"version"`
	Protocols []string                 `json:"protocols"`
	Platforms []VersionListPlatformDTO `json:"platforms"`
}

func (v Version) ToDownloadProviderDTO(system string, architecture string) (DownloadProviderDTO, error) {
	out := DownloadProviderDTO{
		System:       system,
		Architecture: architecture,
	}

	for _, platform := range v.Platforms {
		if platform.System == system && platform.Architecture == architecture {
			out.FileName = platform.FileName
			out.DownloadUrl = platform.DownloadUrl
			out.ShaSumsUrl = platform.ShaSumsUrl
			out.ShaSumsSignatureUrl = platform.ShaSumsSignatureUrl
			out.ShaSum = platform.ShaSum
			out.Protocols = v.Protocols

			var signingKeys []GpgPublicKeyDTO
			for _, signingKey := range platform.SigningKeys {
				signingKeys = append(signingKeys, signingKey.ToGpgPublicKeyDTO())
			}

			out.SigningKeys.GpgPublicKeys = signingKeys

			return out, nil
		}
	}

	return out, fmt.Errorf("no platform found for %s_%s machine", system, architecture)
}
