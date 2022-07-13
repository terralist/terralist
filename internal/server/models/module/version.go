package module

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Version struct {
	entity.Entity
	ModuleID     uuid.UUID    `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Version      string       `gorm:"not null"`
	Location     string       `gorm:"not null"`
	Providers    []Provider   `gorm:"foreignKey:ParentID;references:ID"`
	Dependencies []Dependency `gorm:"foreignKey:ParentID;references:ID"`
	Submodules   []Submodule
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
	Root       RootDTO        `json:"root"`
	Submodules []SubmoduleDTO `json:"submodules"`
}

type VersionListDTO struct {
	Version string `json:"version"`
}
