package mappers

import (
	models "github.com/valentindeaconu/terralist/internal/server/models/module"
)

type ModuleMapper struct{}

func (m *ModuleMapper) ModuleToListResponseDTO(in models.Module) models.ListResponseDTO {
	module := models.ModuleDTO{}

	for _, version := range in.Versions {
		v := models.VersionListDTO{
			Version: version.Version,
		}

		module.Versions = append(module.Versions, v)
	}

	return models.ListResponseDTO{
		Modules: []models.ModuleDTO{module},
	}
}

func (m *ModuleMapper) ModuleCreateDTOToModule(in models.ModuleCreateDTO) models.Module {
	out := models.Module{
		Namespace: in.Namespace,
		Name:      in.Name,
		Provider:  in.Provider,
		Versions: []models.Version{
			{
				Version:      in.Version,
				Location:     in.DownloadUrl,
				Providers:    m.providerDTOsToProviders(in.Root.Providers),
				Dependencies: m.dependencyDTOsToDependencies(in.Root.Dependencies),
			},
		},
	}

	for _, submodule := range in.Submodules {
		s := models.Submodule{
			Path:         submodule.Path,
			Providers:    m.providerDTOsToProviders(submodule.Providers),
			Dependencies: m.dependencyDTOsToDependencies(submodule.Dependencies),
		}

		out.Versions[0].Submodules = append(out.Versions[0].Submodules, s)
	}

	return out
}

// Helpers
func (m *ModuleMapper) providerDTOsToProviders(providers []models.ProviderDTO) []models.Provider {
	result := make([]models.Provider, 0)

	for _, provider := range providers {
		p := models.Provider{
			Name:      provider.Name,
			Namespace: provider.Namespace,
			Source:    provider.Source,
			Version:   provider.Version,
		}

		result = append(result, p)
	}

	return result
}

func (m *ModuleMapper) dependencyDTOsToDependencies(dependencies []models.DependencyDTO) []models.Dependency {
	// TODO: Replace this with a for each when dependency contains fields
	return make([]models.Dependency, len(dependencies))
}
