package provider

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Platform struct {
	entity.Entity
	VersionID    uuid.UUID
	System       string `gorm:"not null"`
	Architecture string `gorm:"not null"`
	Location     string `gorm:"not null"`
	ShaSum       string `gorm:"not null"`
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
