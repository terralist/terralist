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

func (m *ModuleCreateDTO) ToModule() Module {
	out := Module{
		Namespace: m.Namespace,
		Name:      m.Name,
		Provider:  m.Provider,
		Versions: []Version{
			{
				Version:      m.Version,
				Location:     m.DownloadUrl,
				Providers:    FromProviderDTOs(m.Root.Providers),
				Dependencies: FromDependencyDTOs(m.Root.Dependencies),
			},
		},
	}

	for _, submodule := range m.Submodules {
		s := Submodule{
			Path:         submodule.Path,
			Providers:    FromProviderDTOs(submodule.Providers),
			Dependencies: FromDependencyDTOs(submodule.Dependencies),
		}

		out.Versions[0].Submodules = append(out.Versions[0].Submodules, s)
	}

	return out
}

func (m *VersionDTO) ToVersion() Version {
	return Version{
		Version:      m.Version,
		Location:     m.DownloadUrl,
		Providers:    FromProviderDTOs(m.Root.Providers),
		Dependencies: FromDependencyDTOs(m.Root.Dependencies),
	}
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
