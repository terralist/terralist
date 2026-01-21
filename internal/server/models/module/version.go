package module

import (
	"terralist/internal/server/models/artifact"
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Version struct {
	entity.Entity
	ModuleID      uuid.UUID
	Module        Module
	Version       string       `gorm:"not null"`
	Location      string       `gorm:"not null"`
	Documentation string       `gorm:"not null;default:''"` // TODO: This adds backwards-compatibility, we should remove it in future versions
	Providers     []Provider   `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Dependencies  []Dependency `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Submodules    []Submodule  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Version) TableName() string {
	return "module_versions"
}

func (v Version) ToDTO() VersionDTO {
	var submodulesDTO []SubmoduleResponseDTO
	for _, sm := range v.Submodules {
		submodulesDTO = append(submodulesDTO, sm.ToDTO())
	}

	return VersionDTO{
		Version:       v.Version,
		Documentation: v.Documentation,
		Submodules:    submodulesDTO,
	}
}

type RootDTO struct {
	Providers    []ProviderDTO   `json:"providers"`
	Dependencies []DependencyDTO `json:"dependencies"`
}

type VersionDTO struct {
	Version       string                 `json:"version"`
	Documentation string                 `json:"documentation"`
	Submodules    []SubmoduleResponseDTO `json:"submodules,omitempty"`
}

func (v VersionDTO) ToArtifactVersion() artifact.Version {
	return artifact.Version{
		Tag:           v.Version,
		Documentation: v.Documentation,
	}
}

type VersionCreateDTO struct {
	Version    string         `json:"version"`
	Root       RootDTO        `json:"root,omitempty"`
	Submodules []SubmoduleDTO `json:"submodules,omitempty"`
}

type VersionListDTO struct {
	Version string `json:"version"`
}
