package services

import (
	"fmt"

	"github.com/valentindeaconu/terralist/database"
	"github.com/valentindeaconu/terralist/models/module"
)

func ModuleFind(namespace string, name string, provider string) (module.Module, error) {
	p := module.Module{}

	h := database.Handler().Where(module.Module{
		Namespace: namespace,
		Name:      name,
		Provider:  provider,
	}).
		Preload("Versions.Providers").
		Preload("Versions.Dependencies").
		Preload("Versions.Submodules").
		Preload("Versions.Submodules.Providers").
		Preload("Versions.Submodules.Dependencies").
		Find(&p)

	if h.Error != nil {
		return p, fmt.Errorf("no module found with given arguments (source %s/%s/%s)", namespace, name, provider)
	}

	return p, nil
}

func ModuleFindVersion(namespace string, name string, provider string, version string) (module.Version, error) {
	p, err := ModuleFind(namespace, name, provider)

	if err != nil {
		return module.Version{}, err
	}

	for _, v := range p.Versions {
		if v.Version == version {
			return v, nil
		}
	}

	return module.Version{}, fmt.Errorf("no version found")
}

func ModuleUpsert(new module.Module) (module.Module, error) {
	existing, err := ModuleFind(new.Namespace, new.Name, new.Provider)

	if err == nil {
		newVersion := new.Versions[0].Version

		for _, version := range existing.Versions {
			if version.Version == newVersion {
				return module.Module{}, fmt.Errorf("version %s already exists", newVersion)
			}
		}

		existing.Versions = append(existing.Versions, new.Versions[0])

		if result := database.Handler().Save(&existing); result.Error != nil {
			return module.Module{}, result.Error
		}

		return existing, nil
	}

	if result := database.Handler().Create(&new); result.Error != nil {
		return module.Module{}, result.Error
	}

	return new, nil
}

func ModuleDelete(namespace string, name string, provider string) error {
	m, err := ModuleFind(namespace, name, provider)

	if err == nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", namespace, name, provider)
	}

	if result := database.Handler().Delete(&m); result.Error != nil {
		return result.Error
	}

	return nil
}

func ModuleVersionDelete(namespace string, name string, provider string, version string) error {
	m, err := ModuleFind(namespace, name, provider)

	if err == nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", namespace, name, provider)
	}

	q := false
	for idx, ver := range m.Versions {
		if ver.Version == version {
			m.Versions = append(m.Versions[:idx], m.Versions[idx+1:]...)
			q = true
		}
	}

	if q {
		if result := database.Handler().Save(&m); result.Error != nil {
			return result.Error
		}
	}

	return fmt.Errorf("no version found")
}
