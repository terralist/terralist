package module

import (
	"terralist/pkg/database/entity"
	"terralist/pkg/database/types/uuid"
)

type Dependency struct {
	entity.Entity
	ParentID uuid.ID `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Dependency) TableName() string {
	return "module_dependencies"
}

type DependencyDTO struct {
}

func (d DependencyDTO) ToDependency() Dependency {
	return Dependency{}
}
