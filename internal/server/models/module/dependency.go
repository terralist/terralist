package module

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Dependency struct {
	entity.Entity
	ParentID uuid.UUID
}

func (Dependency) TableName() string {
	return "module_dependencies"
}

type DependencyDTO struct {
}

func (d DependencyDTO) ToDependency() Dependency {
	return Dependency{}
}
