package mirror

import (
	"terralist/pkg/database/entity"
	"terralist/pkg/version"
)

// Provider is a provider mirrored from an upstream registry, identified by the
// upstream hostname, the namespace and the provider type.
type Provider struct {
	entity.Entity
	Hostname  string    `gorm:"not null;uniqueIndex:idx_mirror_provider"`
	Namespace string    `gorm:"not null;uniqueIndex:idx_mirror_provider"`
	Name      string    `gorm:"not null;uniqueIndex:idx_mirror_provider"`
	Versions  []Version `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Provider) TableName() string {
	return "mirror_providers"
}

// GetVersion returns the version matching the given semantic version, or nil.
func (p Provider) GetVersion(v string) *Version {
	vv := version.Version(v)

	for _, ver := range p.Versions {
		if version.Compare(version.Version(ver.Version), vv) == 0 {
			return &ver
		}
	}

	return nil
}

// ToVersionListDTO builds the network mirror protocol "List Available Versions"
// response.
func (p Provider) ToVersionListDTO() VersionListDTO {
	versions := make(map[string]struct{}, len(p.Versions))
	for _, v := range p.Versions {
		versions[v.Version] = struct{}{}
	}

	return VersionListDTO{
		Versions: versions,
	}
}

// VersionListDTO is the network mirror protocol "List Available Versions"
// response body. Each version maps to an empty object, as the protocol requires.
type VersionListDTO struct {
	Versions map[string]struct{} `json:"versions"`
}
