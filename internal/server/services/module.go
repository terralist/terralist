package services

import (
	"fmt"

	"terralist/internal/server/models/module"
	"terralist/pkg/database"
)

type ModuleService struct {
	Database database.Engine
}

func (s *ModuleService) Find(namespace string, name string, provider string) (module.Module, error) {
	m := module.Module{}

	err := s.Database.Handler().Where(module.Module{
		Namespace: namespace,
		Name:      name,
		Provider:  provider,
	}).
		Preload("Versions.Providers").
		Preload("Versions.Dependencies").
		Preload("Versions.Submodules").
		Preload("Versions.Submodules.Providers").
		Preload("Versions.Submodules.Dependencies").
		Find(&m).
		Error

	if err != nil {
		return m, fmt.Errorf("no module found with given arguments (source %s/%s/%s)", namespace, name, provider)
	}

	return m, nil
}

func (s *ModuleService) FindVersion(namespace string, name string, provider string, version string) (module.Version, error) {
	m, err := s.Find(namespace, name, provider)

	if err != nil {
		return module.Version{}, err
	}

	for _, v := range m.Versions {
		if v.Version == version {
			return v, nil
		}
	}

	return module.Version{}, fmt.Errorf("no version found")
}

func (s *ModuleService) Upsert(new module.Module) (module.Module, error) {
	existing, err := s.Find(new.Namespace, new.Name, new.Provider)

	if err == nil {
		newVersion := new.Versions[0].Version

		for _, version := range existing.Versions {
			if version.Version == newVersion {
				return module.Module{}, fmt.Errorf("version %s already exists", newVersion)
			}
		}

		existing.Versions = append(existing.Versions, new.Versions[0])

		if err := s.Database.Handler().Save(&existing).Error; err != nil {
			return module.Module{}, err
		}

		return existing, nil
	}

	if result := s.Database.Handler().Create(&new); result.Error != nil {
		return module.Module{}, result.Error
	}

	return new, nil
}

func (s *ModuleService) Delete(namespace string, name string, provider string) error {
	m, err := s.Find(namespace, name, provider)

	if err == nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", namespace, name, provider)
	}

	if err := s.Database.Handler().Delete(&m).Error; err != nil {
		return err
	}

	return nil
}

func (s *ModuleService) DeleteVersion(namespace string, name string, provider string, version string) error {
	m, err := s.Find(namespace, name, provider)

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
		if err := s.Database.Handler().Save(&s).Error; err != nil {
			return err
		}
	}

	return fmt.Errorf("no version found")
}
