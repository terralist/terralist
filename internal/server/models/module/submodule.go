package module

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Submodule struct {
	entity.Entity
	VersionID    uuid.UUID
	Path         string       `gorm:"not null"`
	Providers    []Provider   `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Dependencies []Dependency `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Submodule) TableName() string {
	return "module_submodules"
}

func (s Submodule) ToDTO() SubmoduleResponseDTO {
	return SubmoduleResponseDTO{
		Path: s.Path,
	}
}

type SubmoduleDTO struct {
	Path         string          `json:"path"`
	Providers    []ProviderDTO   `json:"providers"`
	Dependencies []DependencyDTO `json:"dependencies"`
}

type SubmoduleResponseDTO struct {
	Path string `json:"path"`
}
