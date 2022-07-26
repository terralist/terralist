package module

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Version struct {
	entity.Entity
	ModuleID     uuid.UUID
	Version      string       `gorm:"not null"`
	Location     string       `gorm:"not null"`
	Providers    []Provider   `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Dependencies []Dependency `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Submodules   []Submodule  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Version) TableName() string {
	return "module_versions"
}

type RootDTO struct {
	Providers    []ProviderDTO   `json:"providers"`
	Dependencies []DependencyDTO `json:"dependencies"`
}

type VersionDTO struct {
	Version    string         `json:"version"`
	Root       RootDTO        `json:"root,omitempty"`
	Submodules []SubmoduleDTO `json:"submodules,omitempty"`
}

type VersionListDTO struct {
	Version string `json:"version"`
}
