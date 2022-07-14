package services

import (
	"errors"
	"fmt"
	"sort"
	"terralist/pkg/version"

	"gorm.io/gorm"

	"terralist/internal/server/models/module"
	"terralist/pkg/database"
)

type ModuleService struct {
	Database database.Engine
}

func (s *ModuleService) Find(namespace string, name string, provider string) (*module.Module, error) {
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
		First(&m).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no module found with given arguments (source %s/%s/%s)", namespace, name, provider)
		} else {
			return nil, fmt.Errorf("error while querying the database: %v", err)
		}
	}

	sort.Slice(m.Versions, func(i, j int) bool {
		lhs := version.Version(m.Versions[i].Version)
		rhs := version.Version(m.Versions[j].Version)

		return version.Compare(lhs, rhs) <= 0
	})

	return &m, nil
}

func (s *ModuleService) FindVersion(namespace string, name string, provider string, version string) (*module.Version, error) {
	m, err := s.Find(namespace, name, provider)

	if err != nil {
		return nil, err
	}

	for _, v := range m.Versions {
		if v.Version == version {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("no version found")
}

func (s *ModuleService) Upsert(n module.Module) (*module.Module, error) {
	m, err := s.Find(n.Namespace, n.Name, n.Provider)

	if err == nil {
		newVersion := n.Versions[0].Version

		for _, ver := range n.Versions {
			if ver.Version == newVersion {
				return nil, fmt.Errorf("version %s already exists", newVersion)
			}
		}

		m.Versions = append(m.Versions, n.Versions[0])

		if err := s.Database.Handler().Save(m).Error; err != nil {
			return nil, err
		}

		return m, nil
	}

	if result := s.Database.Handler().Create(&n); result.Error != nil {
		return nil, result.Error
	}

	return &n, nil
}

func (s *ModuleService) Delete(namespace string, name string, provider string) error {
	m, err := s.Find(namespace, name, provider)
	if err != nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", namespace, name, provider)
	}

	if err := s.Database.Handler().Delete(&m).Error; err != nil {
		return err
	}

	return nil
}

func (s *ModuleService) DeleteVersion(namespace string, name string, provider string, version string) error {
	m, err := s.Find(namespace, name, provider)
	if err != nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", namespace, name, provider)
	}

	var toDelete *module.Version = nil
	for _, v := range m.Versions {
		if v.Version == version {
			toDelete = &v
			break
		}
	}

	if toDelete != nil {
		if len(m.Versions) == 1 {
			if err := s.Database.Handler().Delete(m).Error; err != nil {
				return err
			}
		} else {
			if err := s.Database.Handler().Delete(toDelete).Error; err != nil {
				return err
			}
		}

		return nil
	}

	return fmt.Errorf("no version found")
}
