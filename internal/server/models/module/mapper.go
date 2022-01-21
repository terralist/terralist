package module

func (m *Module) ToVersionListResponse() ListResponseDTO {
	out := ListResponseDTO{}

	module := ModuleDTO{}

	for _, version := range m.Versions {
		v := VersionListDTO{
			Version: version.Version,
		}

		module.Versions = append(module.Versions, v)
	}

	out.Modules = append(out.Modules, module)

	return out
}

func FromCreateDTO(dto ModuleCreateDTO) Module {
	out := Module{
		Namespace: dto.Namespace,
		Name:      dto.Name,
		Provider:  dto.Provider,
		Versions: []Version{
			{
				Version:      dto.Version,
				Location:     dto.DownloadUrl,
				Providers:    FromProviderDTOs(dto.Root.Providers),
				Dependencies: FromDependencyDTOs(dto.Root.Dependencies),
			},
		},
	}

	for _, submodule := range dto.Submodules {
		s := Submodule{
			Path:         submodule.Path,
			Providers:    FromProviderDTOs(submodule.Providers),
			Dependencies: FromDependencyDTOs(submodule.Dependencies),
		}

		out.Versions[0].Submodules = append(out.Versions[0].Submodules, s)
	}

	return out
}

// Helpers
func ToProviderDTOs(providers []Provider) []ProviderDTO {
	result := make([]ProviderDTO, 0)

	for _, provider := range providers {
		p := ProviderDTO{
			Name:      provider.Name,
			Namespace: provider.Namespace,
			Source:    provider.Source,
			Version:   provider.Version,
		}

		result = append(result, p)
	}

	return result
}

func ToDependencyDTOs(dependencies []Dependency) []DependencyDTO {
	// TODO: Replace this with a for each when dependency contains fields
	return make([]DependencyDTO, len(dependencies))
}

func FromProviderDTOs(providers []ProviderDTO) []Provider {
	result := make([]Provider, 0)

	for _, provider := range providers {
		p := Provider{
			Name:      provider.Name,
			Namespace: provider.Namespace,
			Source:    provider.Source,
			Version:   provider.Version,
		}

		result = append(result, p)
	}

	return result
}

func FromDependencyDTOs(dependencies []DependencyDTO) []Dependency {
	// TODO: Replace this with a for each when dependency contains fields
	return make([]Dependency, len(dependencies))
}
