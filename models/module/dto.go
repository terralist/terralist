package module

type ListResponseDTO struct {
	Modules []ModuleDTO `json:"modules"`
}

type ModuleDTO struct {
	Versions []VersionListDTO `json:"versions"`
}

type VersionListDTO struct {
	Version string `json:"version"`
}

type VersionDTO struct {
	Version     string         `json:"version"`
	DownloadUrl string         `json:"download_url"`
	Root        RootDTO        `json:"root"`
	Submodules  []SubmoduleDTO `json:"submodules"`
}

type ModuleCreateDTO struct {
	VersionDTO
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
}

type SubmoduleDTO struct {
	Path         string          `json:"path"`
	Providers    []ProviderDTO   `json:"providers"`
	Dependencies []DependencyDTO `json:"dependencies"`
}

type RootDTO struct {
	Providers    []ProviderDTO   `json:"providers"`
	Dependencies []DependencyDTO `json:"dependencies"`
}

type ProviderDTO struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Source    string `json:"source"`
	Version   string `json:"version"`
}

type DependencyDTO struct {
}
