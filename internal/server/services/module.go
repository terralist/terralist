package services

import (
	"fmt"

	"github.com/valentindeaconu/terralist/internal/server/database"
	models "github.com/valentindeaconu/terralist/internal/server/models/module"
)

type ModuleService struct {
	Database database.Engine
}

func (m *ModuleService) Find(namespace string, name string, provider string) (models.Module, error) {
	module := models.Module{}

	h := m.Database.Handler().Where(models.Module{
		Namespace: namespace,
		Name:      name,
		Provider:  provider,
	}).
		Preload("Versions.Providers").
		Preload("Versions.Dependencies").
		Preload("Versions.Submodules").
		Preload("Versions.Submodules.Providers").
		Preload("Versions.Submodules.Dependencies").
		Find(&module)

	if h.Error != nil {
		return module, fmt.Errorf("no module found with given arguments (source %s/%s/%s)", namespace, name, provider)
	}

	return module, nil
}

func (m *ModuleService) FindVersion(namespace string, name string, provider string, version string) (models.Version, error) {
	module, err := m.Find(namespace, name, provider)

	if err != nil {
		return models.Version{}, err
	}

	for _, v := range module.Versions {
		if v.Version == version {
			return v, nil
		}
	}

	return models.Version{}, fmt.Errorf("no version found")
}

func (m *ModuleService) Upsert(new models.Module) (models.Module, error) {
	existing, err := m.Find(new.Namespace, new.Name, new.Provider)

	if err == nil {
		newVersion := new.Versions[0].Version

		for _, version := range existing.Versions {
			if version.Version == newVersion {
				return models.Module{}, fmt.Errorf("version %s already exists", newVersion)
			}
		}

		existing.Versions = append(existing.Versions, new.Versions[0])

		if result := m.Database.Handler().Save(&existing); result.Error != nil {
			return models.Module{}, result.Error
		}

		return existing, nil
	}

	if result := m.Database.Handler().Create(&new); result.Error != nil {
		return models.Module{}, result.Error
	}

	return new, nil
}

func (m *ModuleService) Delete(namespace string, name string, provider string) error {
	module, err := m.Find(namespace, name, provider)

	if err == nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", namespace, name, provider)
	}

	if result := m.Database.Handler().Delete(&module); result.Error != nil {
		return result.Error
	}

	return nil
}

func (m *ModuleService) DeleteVersion(namespace string, name string, provider string, version string) error {
	module, err := m.Find(namespace, name, provider)

	if err == nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", namespace, name, provider)
	}

	q := false
	for idx, ver := range module.Versions {
		if ver.Version == version {
			module.Versions = append(module.Versions[:idx], module.Versions[idx+1:]...)
			q = true
		}
	}

	if q {
		if result := m.Database.Handler().Save(&m); result.Error != nil {
			return result.Error
		}
	}

	return fmt.Errorf("no version found")
}
