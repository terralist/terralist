package module

import (
	"fmt"

	"terralist/internal/server/models/artifact"
	"terralist/pkg/database/entity"
	"terralist/pkg/version"

	"github.com/google/uuid"
	"github.com/ssoroka/slice"
)

type Module struct {
	entity.Entity
	AuthorityID uuid.UUID `gorm:"size:256"`
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

func (m Module) ToArtifact() artifact.Artifact {
	return artifact.Artifact{
		ID:       m.ID.String(),
		Name:     m.Name,
		Provider: m.Provider,
		Type:     artifact.TypeModule,
		Versions: slice.Map[Version, string](m.Versions, func(v Version) string {
			return v.Version
		}),
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02T15:04:05"),
	}
}

func (m Module) GetVersion(v string) *Version {
	vv := version.Version(v)

	for _, ver := range m.Versions {
		if version.Compare(version.Version(ver.Version), vv) == 0 {
			return &ver
		}
	}

	return nil
}

type ListResponseDTO struct {
	Modules []ModuleDTO `json:"modules"`
}

type ModuleDTO struct {
	Versions []VersionListDTO `json:"versions"`
}

type CreateDTO struct {
	VersionDTO
	AuthorityID uuid.UUID `gorm:"size:256"`
	Name        string    `json:"name"`
	Provider    string    `json:"provider"`
}

type CreateFromURLDTO struct {
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
