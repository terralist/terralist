package repositories

import (
	"errors"
	"fmt"
	"sort"

	"terralist/internal/server/models/module"
	"terralist/pkg/database"
	"terralist/pkg/storage/resolver"
	"terralist/pkg/version"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ModuleRepository describes a service that can interact with the modules database
type ModuleRepository interface {
	// Find searches for a specific module
	Find(namespace string, name string, provider string) (*module.Module, error)

	// FindVersion searches for a specific module version
	FindVersion(namespace string, name string, provider string, version string) (*module.Version, error)

	// Upsert either updates or creates a new (if it does not already exist) module
	Upsert(n module.Module) (*module.Module, error)

	// Delete removes a module with all its data (versions)
	Delete(namespace string, name string, provider string) error

	// DeleteVersion removes a version from a module
	DeleteVersion(namespace string, name string, provider string, version string) error
}

// DefaultModuleRepository is a concrete implementation of ModuleRepository
type DefaultModuleRepository struct {
	Database database.Engine
	Resolver resolver.Resolver
}

func (r *DefaultModuleRepository) Find(namespace string, name string, provider string) (*module.Module, error) {
	m := module.Module{}

	err := r.Database.Handler().Where(module.Module{
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

	for i, v := range m.Versions {
		key, err := r.Resolver.Find(v.FetchKey)
		if err != nil {
			return nil, fmt.Errorf("could not find url for key %v: %v", v.FetchKey, err)
		}

		m.Versions[i].FetchKey = key
	}

	sort.Slice(m.Versions, func(i, j int) bool {
		lhs := version.Version(m.Versions[i].Version)
		rhs := version.Version(m.Versions[j].Version)

		return version.Compare(lhs, rhs) <= 0
	})

	return &m, nil
}

func (r *DefaultModuleRepository) FindVersion(namespace string, name string, provider string, version string) (*module.Version, error) {
	m, err := r.Find(namespace, name, provider)

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

// Upsert is designed to upload an entire module, but in reality,
// it will only upload a single version
func (r *DefaultModuleRepository) Upsert(n module.Module) (*module.Module, error) {
	var toUpsert *module.Module

	m, err := r.Find(n.Namespace, n.Name, n.Provider)
	if err == nil {
		newVersion := version.Version(n.Versions[0].Version)

		for _, ver := range n.Versions {
			if version.Compare(version.Version(ver.Version), newVersion) == 0 {
				return nil, fmt.Errorf("version %s already exists", newVersion)
			}
		}

		m.Versions = append(m.Versions, n.Versions[0])

		toUpsert = m
	} else {
		toUpsert = &n
	}

	toUpload := &toUpsert.Versions[len(toUpsert.Versions)-1]
	toUpload.FetchKey, err = r.Resolver.Store(toUpload.FetchKey, true)
	if err != nil {
		return nil, fmt.Errorf("could store the new version: %v", err)
	}

	if len(toUpsert.Versions) != 1 {
		if err := r.Database.Handler().Save(toUpsert).Error; err != nil {
			return nil, err
		}
	} else {
		if err := r.Database.Handler().Create(toUpsert).Error; err != nil {
			return nil, err
		}
	}

	return toUpsert, nil
}

func (r *DefaultModuleRepository) Delete(namespace string, name string, provider string) error {
	m, err := r.Find(namespace, name, provider)
	if err != nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", namespace, name, provider)
	}

	for _, ver := range m.Versions {
		if err := r.Resolver.Purge(ver.FetchKey); err != nil {
			log.Warn().
				AnErr("Error", err).
				Str("Module", fmt.Sprintf("%s/%s/%s", namespace, name, provider)).
				Str("Version", ver.Version).
				Str("Key", ver.FetchKey).
				Msg("Could not purge, require manual clean-up")
		}
	}

	if err := r.Database.Handler().Delete(&m).Error; err != nil {
		return err
	}

	return nil
}

func (r *DefaultModuleRepository) DeleteVersion(namespace string, name string, provider string, version string) error {
	m, err := r.Find(namespace, name, provider)
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
			for _, ver := range m.Versions {
				if err := r.Resolver.Purge(ver.FetchKey); err != nil {
					log.Warn().
						AnErr("Error", err).
						Str("Module", fmt.Sprintf("%s/%s/%s", namespace, name, provider)).
						Str("Version", ver.Version).
						Str("Key", ver.FetchKey).
						Msg("Could not purge, require manual clean-up")
				}
			}

			if err := r.Database.Handler().Delete(m).Error; err != nil {
				return err
			}
		} else {
			if err := r.Resolver.Purge(toDelete.FetchKey); err != nil {
				log.Warn().
					AnErr("Error", err).
					Str("Module", fmt.Sprintf("%s/%s/%s", namespace, name, provider)).
					Str("Version", toDelete.Version).
					Str("Key", toDelete.FetchKey).
					Msg("Could not purge, require manual clean-up")
			}

			if err := r.Database.Handler().Delete(toDelete).Error; err != nil {
				return err
			}
		}

		return nil
	}

	return fmt.Errorf("no version found")
}
