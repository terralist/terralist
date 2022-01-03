package service

import (
	"fmt"

	"github.com/valentindeaconu/terralist/database"
	"github.com/valentindeaconu/terralist/model/module"
)

type ModuleService struct {
}

func (m *ModuleService) Find(namespace string, name string, provider string) (module.Module, error) {
	p := module.Module{}

	h := database.AppDatabase.Handler.
		Where(module.Module{
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

	if h.RowsAffected > 0 {
		return p, nil
	}

	return p, fmt.Errorf("no module found with given arguments (source %s/%s/%s)", namespace, name, provider)
}

func (m *ModuleService) FindVersion(namespace string, name string, provider string, version string) (module.Version, error) {
	p, err := m.Find(namespace, name, provider)

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

func (m *ModuleService) Upsert(new module.Module) (module.Module, error) {
	existing, err := m.Find(new.Namespace, new.Name, new.Provider)

	if err == nil {
		newVersion := new.Versions[0].Version

		for _, version := range existing.Versions {
			if version.Version == newVersion {
				return module.Module{}, fmt.Errorf("version %s already exists", newVersion)
			}
		}

		existing.Versions = append(existing.Versions, new.Versions[0])

		if result := database.AppDatabase.Handler.Save(&existing); result.RowsAffected != 1 {
			return module.Module{}, result.Error
		} else {
			return existing, nil
		}
	} else {
		if result := database.AppDatabase.Handler.Create(&new); result.RowsAffected != 1 {
			return module.Module{}, result.Error
		} else {
			return new, nil
		}
	}
}
