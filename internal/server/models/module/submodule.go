package module

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Submodule struct {
	entity.Entity
	VersionID    uuid.UUID    `gorm:"size:256"`
	Path         string       `gorm:"not null"`
	Providers    []Provider   `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Dependencies []Dependency `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Submodule) TableName() string {
	return "module_submodules"
}

type SubmoduleDTO struct {
	Path         string          `json:"path"`
	Providers    []ProviderDTO   `json:"providers"`
	Dependencies []DependencyDTO `json:"dependencies"`
}
