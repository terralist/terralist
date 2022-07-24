package repositories

import (
	"errors"
	"fmt"
	"sort"

	"terralist/internal/server/models/module"
	"terralist/pkg/database"
	"terralist/pkg/storage"
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
	Delete(*module.Module) error

	// DeleteVersion removes a version from a module
	DeleteVersion(m *module.Module, version string) error
}

// DefaultModuleRepository is a concrete implementation of ModuleRepository
type DefaultModuleRepository struct {
	Database database.Engine
	Resolver storage.Resolver
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
		key, err := r.Resolver.Find(v.Location)
		if err != nil {
			return nil, fmt.Errorf("could not find url for key %v: %v", v.Location, err)
		}

		m.Versions[i].Location = key
	}

	sort.Slice(m.Versions, func(i, j int) bool {
		lhs := version.Version(m.Versions[i].Version)
		rhs := version.Version(m.Versions[j].Version)

		return version.Compare(lhs, rhs) <= 0
	})

	return &m, nil
}

func (r *DefaultModuleRepository) FindVersion(
	namespace string,
	name string,
	provider string,
	version string,
) (*module.Version, error) {
	m, err := r.Find(namespace, name, provider)

	if err != nil {
		return nil, err
	}

	var ver *module.Version = nil
	for _, v := range m.Versions {
		if v.Version == version {
			ver = &v
			break
		}
	}

	if ver == nil {
		return nil, fmt.Errorf("no version found")
	}

	ver.Location, err = r.Resolver.Find(ver.Location)
	if err != nil {
		return nil, fmt.Errorf("could not resolve location: %v", err)
	}

	return ver, nil
}

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

	// Create a transaction to revert db changes if the resolver fails
	// to store the file
	// If the database fails to create the object, the call to store
	// the object will never be called
	if err := r.Database.Handler().Transaction(func(tx *gorm.DB) error {
		if len(toUpsert.Versions) != 1 {
			if err := tx.Save(toUpsert).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(toUpsert).Error; err != nil {
				return err
			}
		}

		toUpload := &toUpsert.Versions[len(toUpsert.Versions)-1]
		toUpload.Location, err = r.Resolver.Store(&storage.StoreInput{
			URL:     toUpload.Location,
			Archive: true,
			KeyPrefix: fmt.Sprintf(
				"modules/%s/%s/%s/%s",
				toUpsert.Namespace,
				toUpsert.Name,
				toUpsert.Provider,
				toUpload.Version,
			),
			FileName: fmt.Sprintf("v%s.zip", toUpload.Version),
		})
		if err != nil {
			return fmt.Errorf("could store the new version: %v", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return toUpsert, nil
}

func (r *DefaultModuleRepository) Delete(m *module.Module) error {
	for _, ver := range m.Versions {
		if err := r.Resolver.Purge(ver.Location); err != nil {
			log.Warn().
				AnErr("Error", err).
				Str("Module", m.String()).
				Str("Version", ver.Version).
				Str("Key", ver.Location).
				Msg("Could not purge, require manual clean-up")
		}
	}

	if err := r.Database.Handler().Delete(&m).Error; err != nil {
		return err
	}

	return nil
}

func (r *DefaultModuleRepository) DeleteVersion(m *module.Module, version string) error {
	var toDelete *module.Version = nil
	for _, v := range m.Versions {
		if v.Version == version {
			toDelete = &v
			break
		}
	}

	if toDelete == nil {
		return fmt.Errorf("no version found")
	}

	if len(m.Versions) == 1 {
		return r.Delete(m)
	}

	if err := r.Resolver.Purge(toDelete.Location); err != nil {
		log.Warn().
			AnErr("Error", err).
			Str("Module", m.String()).
			Str("Version", toDelete.Version).
			Str("Key", toDelete.Location).
			Msg("Could not purge, require manual clean-up")
	}

	return r.Database.Handler().Delete(toDelete).Error
}
