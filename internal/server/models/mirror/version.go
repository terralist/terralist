package mirror

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

// Version is a single version of a mirrored provider.
type Version struct {
	entity.Entity
	ProviderID uuid.UUID
	Provider   Provider
	Version    string     `gorm:"not null"`
	Platforms  []Platform `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Version) TableName() string {
	return "mirror_versions"
}

// GetPlatform returns the platform matching the given os and architecture, or nil.
func (v Version) GetPlatform(system, architecture string) *Platform {
	for i := range v.Platforms {
		if v.Platforms[i].System == system && v.Platforms[i].Architecture == architecture {
			return &v.Platforms[i]
		}
	}

	return nil
}
