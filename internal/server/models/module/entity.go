package module

import (
	"github.com/google/uuid"
	"github.com/valentindeaconu/terralist/internal/server/models"
)

// ORM
type Module struct {
	models.Base
	Namespace string
	Name      string
	Provider  string
	Versions  []Version
}

func (Module) TableName() string {
	return "modules"
}

type Version struct {
	models.Base
	ModuleID     uuid.UUID
	Version      string
	Location     string
	Providers    []Provider   `gorm:"foreignKey:ParentID;references:ID"`
	Dependencies []Dependency `gorm:"foreignKey:ParentID;references:ID"`
	Submodules   []Submodule
}

func (Version) TableName() string {
	return "module_versions"
}

type Submodule struct {
	models.Base
	VersionID    uuid.UUID
	Path         string
	Providers    []Provider   `gorm:"foreignKey:ParentID;references:ID"`
	Dependencies []Dependency `gorm:"foreignKey:ParentID;references:ID"`
}

func (Submodule) TableName() string {
	return "module_submodules"
}

type Provider struct {
	models.Base
	ParentID  uuid.UUID
	Name      string
	Namespace string
	Source    string
	Version   string
}

func (Provider) TableName() string {
	return "module_providers"
}

type Dependency struct {
	models.Base
	ParentID uuid.UUID
}

func (Dependency) TableName() string {
	return "module_dependencies"
}
