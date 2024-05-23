package module

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Version struct {
	entity.Entity
	ModuleID     uuid.UUID `gorm:"size:256"`
	Module       Module
	Version      string       `gorm:"not null"`
	Location     string       `gorm:"not null"`
	Providers    []Provider   `gorm:"foreignKey:ParentID;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION"`
	Dependencies []Dependency `gorm:"foreignKey:ParentID;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION"`
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
