package repositories

import (
	"errors"
	"fmt"
	"sort"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/pkg/database"
	"terralist/pkg/version"

	"gorm.io/gorm"
)

// ModuleRepository describes a service that can interact with the modules database.
type ModuleRepository interface {
	// Find searches for a specific module.
	Find(namespace, name, provider string) (*module.Module, error)

	// FindVersion searches for a specific module version.
	FindVersion(namespace, name, provider, version string) (*module.Version, error)

	// FindVersionLocation searches for a specific module version location.
	FindVersionLocation(namespace, name, provider, version string) (*string, error)

	// Upsert either updates or creates a new (if it does not already exist) module.
	Upsert(n module.Module) (*module.Module, error)

	// Delete removes a module with all its data (versions).
	Delete(*module.Module) error

	// DeleteVersion removes a version from a module.
	DeleteVersion(*module.Version) error
}

// DefaultModuleRepository is a concrete implementation of ModuleRepository.
type DefaultModuleRepository struct {
	Database database.Engine
}

func (r *DefaultModuleRepository) Find(namespace, name, provider string) (*module.Module, error) {
	m := module.Module{}

	atn := (authority.Authority{}).TableName()
	mtn := (module.Module{}).TableName()

	err := r.Database.Handler().
		Where(
			fmt.Sprintf("LOWER(%s.name) = LOWER(?) AND LOWER(%s.provider) = LOWER(?)", mtn, mtn),
			name,
			provider,
		).
		Joins(
			fmt.Sprintf(
				"JOIN %s ON %s.id = %s.authority_id AND LOWER(%s.name) = LOWER(?)",
				atn,
				atn,
				mtn,
				atn,
			),
			namespace,
		).
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

func (r *DefaultModuleRepository) FindVersion(namespace, name, provider, version string) (*module.Version, error) {
	ver := module.Version{}

	atn := (authority.Authority{}).TableName()
	mtn := (module.Module{}).TableName()
	vtn := (module.Version{}).TableName()

	err := r.Database.Handler().
		Table(vtn).
		Joins(
			fmt.Sprintf(
				"JOIN %s ON %s.id = %s.module_id AND LOWER(%s.name) = LOWER(?) AND LOWER(%s.provider) = LOWER(?)",
				mtn,
				mtn,
				vtn,
				mtn,
				mtn,
			),
			name,
			provider,
		).
		Joins(
			fmt.Sprintf(
				"JOIN %s ON %s.id = %s.authority_id AND LOWER(%s.name) = LOWER(?)",
				atn,
				atn,
				mtn,
				atn,
			),
			namespace,
		).
		Where(fmt.Sprintf("%s.version = ?", vtn), version).
		Preload("Providers").
		Preload("Dependencies").
		Preload("Submodules").
		Preload("Submodules.Providers").
		Preload("Submodules.Dependencies").
		First(&ver).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf(
				"no module version found with given arguments (module %s/%s/%s/%s)",
				namespace,
				name,
				provider,
				version,
			)
		} else {
			return nil, fmt.Errorf("error while querying the database: %v", err)
		}
	}

	return &ver, nil
}

func (r *DefaultModuleRepository) FindVersionLocation(namespace, name, provider, version string) (*string, error) {
	var location string

	atn := (authority.Authority{}).TableName()
	mtn := (module.Module{}).TableName()
	vtn := (module.Version{}).TableName()

	res := r.Database.Handler().
		Table(vtn).
		Select(fmt.Sprintf("%s.location", vtn)).
		Joins(
			fmt.Sprintf(
				"JOIN %s ON %s.id = %s.module_id AND LOWER(%s.name) = LOWER(?) AND LOWER(%s.provider) = LOWER(?)",
				mtn,
				mtn,
				vtn,
				mtn,
				mtn,
			),
			name,
			provider,
		).
		Joins(
			fmt.Sprintf(
				"JOIN %s ON %s.id = %s.authority_id AND LOWER(%s.name) = LOWER(?)",
				atn,
				atn,
				mtn,
				atn,
			),
			namespace,
		).
		Where(fmt.Sprintf("%s.version = ?", vtn), version).
		Scan(&location)

	if err := res.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, fmt.Errorf("error while querying the database: %v", err)
		}
	}

	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	return &location, nil
}

func (r *DefaultModuleRepository) Upsert(m module.Module) (*module.Module, error) {
	if len(m.Versions) != 1 {
		if err := r.Database.Handler().Session(&gorm.Session{FullSaveAssociations: true}).Save(&m).Error; err != nil {
			return nil, err
		}
	} else {
		if err := r.Database.Handler().Session(&gorm.Session{FullSaveAssociations: true}).Create(&m).Error; err != nil {
			return nil, err
		}
	}

	return &m, nil
}

func (r *DefaultModuleRepository) Delete(m *module.Module) error {
	if err := r.Database.Handler().Delete(&m).Error; err != nil {
		return err
	}

	return nil
}

func (r *DefaultModuleRepository) DeleteVersion(v *module.Version) error {
	return r.Database.Handler().Delete(v).Error
}
