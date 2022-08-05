package module

import (
	"fmt"

	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Module struct {
	entity.Entity
	AuthorityID uuid.UUID
	Name        string    `gorm:"not null"`
	Provider    string    `gorm:"not null"`
	Versions    []Version `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Module) TableName() string {
	return "modules"
}

func (m Module) String() string {
	return fmt.Sprintf("%s/%s", m.Name, m.Provider)
}

func (m Module) ToListResponseDTO() ListResponseDTO {
	module := ModuleDTO{}

	for _, version := range m.Versions {
		v := VersionListDTO{
			Version: version.Version,
		}

		module.Versions = append(module.Versions, v)
	}

	return ListResponseDTO{
		Modules: []ModuleDTO{module},
	}
}

type ListResponseDTO struct {
	Modules []ModuleDTO `json:"modules"`
}

type ModuleDTO struct {
	Versions []VersionListDTO `json:"versions"`
}

type CreateDTO struct {
	VersionDTO
	AuthorityID uuid.UUID
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	DownloadUrl string `json:"download_url"`
}

func (d CreateDTO) ToModule() Module {
	var providers []Provider
	for _, p := range d.Root.Providers {
		providers = append(providers, p.ToProvider())
	}

	var dependencies []Dependency
	for _, dep := range d.Root.Dependencies {
		dependencies = append(dependencies, dep.ToDependency())
	}

	out := Module{
		AuthorityID: d.AuthorityID,
		Name:        d.Name,
		Provider:    d.Provider,
		Versions: []Version{
			{
				Version:      d.Version,
				Location:     d.DownloadUrl,
				Providers:    providers,
				Dependencies: dependencies,
			},
		},
	}

	for _, submodule := range d.Submodules {
		var submoduleProviders []Provider
		for _, p := range submodule.Providers {
			providers = append(providers, p.ToProvider())
		}

		var submoduleDependencies []Dependency
		for _, dep := range submodule.Dependencies {
			dependencies = append(dependencies, dep.ToDependency())
		}

		s := Submodule{
			Path:         submodule.Path,
			Providers:    submoduleProviders,
			Dependencies: submoduleDependencies,
		}

		out.Versions[0].Submodules = append(out.Versions[0].Submodules, s)
	}

	return out
}
