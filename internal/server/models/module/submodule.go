package module

import (
	"terralist/pkg/database/entity"
	"terralist/pkg/database/types/uuid"
)

type Submodule struct {
	entity.Entity
	VersionID    uuid.ID      `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Path         string       `gorm:"not null"`
	Providers    []Provider   `gorm:"foreignKey:ParentID;references:ID"`
	Dependencies []Dependency `gorm:"foreignKey:ParentID;references:ID"`
}

func (Submodule) TableName() string {
	return "module_submodules"
}

type SubmoduleDTO struct {
	Path         string          `json:"path"`
	Providers    []ProviderDTO   `json:"providers"`
	Dependencies []DependencyDTO `json:"dependencies"`
}
