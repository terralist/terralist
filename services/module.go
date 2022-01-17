package services

import (
	"fmt"
	"sort"
	"strings"

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

// Create a new module
func ModuleCreate(new module.Module) (module.Module, error) {
	_, err := ModuleFind(new.Namespace, new.Name, new.Provider)

	if err == nil {
		return module.Module{}, fmt.Errorf("module %s/%s/%s already exists", new.Namespace, new.Name, new.Provider)
	}

	if result := database.Handler().Create(&new); result.Error != nil {
		return module.Module{}, result.Error
	}

	return new, nil
}

// Add a version to an existing module
func ModuleAddVersion(namespace string, name string, provider string, version module.Version) (module.Module, error) {
	m, err := ModuleFind(namespace, name, provider)

	if err != nil {
		return module.Module{}, err
	}

	m.Versions = append(m.Versions, version)

	sort.SliceStable(m.Versions, func(i, j int) bool {
		return strings.Compare(m.Versions[i].Version, m.Versions[j].Version) >= 0
	})

	if result := database.Handler().Save(&m); result.Error != nil {
		return module.Module{}, result.Error
	}

	return m, nil
}

// Delete a module
func ModuleDelete(namespace string, name string, provider string) error {
	module, err := ModuleFind(namespace, name, provider)

	if err == nil {
		database.Handler().Delete(&module)
	}

	return err
}

// Delete a version from an existing module
func ModuleDeleteVersion(namespace string, name string, provider string, version string) error {
	module, err := ModuleFind(namespace, name, provider)

	q := false
	if err == nil {
		for idx, ver := range module.Versions {
			if ver.Version == version {
				q = true
				module.Versions = append(module.Versions[:idx], module.Versions[:idx+1]...)
				database.Handler().Save(&module)
			}
		}

		if !q {
			return fmt.Errorf("no version %s was found for module %s/%s/%s", version, namespace, name, provider)
		}
	}

	return err
}
