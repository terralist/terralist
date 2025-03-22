package repositories

import (
	"errors"
	"fmt"
	"sort"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database"
	"terralist/pkg/version"

	"gorm.io/gorm"
)

// ProviderRepository describes a service that can interact with the providers database.
type ProviderRepository interface {
	// Find searches for a specific provider.
	Find(namespace, name string) (*provider.Provider, error)

	// FindVersionPlatform searches for a specific platform binary metadata
	// of a provider version.
	FindVersionPlatform(namespace, name, version, os, arch string) (*provider.Platform, error)

	// Upsert either updates or creates a new (if it does not already exist) provider.
	Upsert(provider.Provider) (*provider.Provider, error)

	// Delete removes a provider with all its data (versions).
	Delete(*provider.Provider) error

	// DeleteVersion removes a version from a provider.
	DeleteVersion(p *provider.Provider, version string) error
}

// DefaultProviderRepository is a concrete implementation of ProviderRepository.
type DefaultProviderRepository struct {
	Database database.Engine
}

func (r *DefaultProviderRepository) Find(namespace, name string) (*provider.Provider, error) {
	p := provider.Provider{}

	atn := (authority.Authority{}).TableName()
	ptn := (provider.Provider{}).TableName()

	err := r.Database.Handler().
		Where(
			fmt.Sprintf("LOWER(%s.name) = LOWER(?)", ptn),
			name,
		).
		Joins(
			fmt.Sprintf(
				"JOIN %s ON %s.id = %s.authority_id AND LOWER(%s.name) = LOWER(?)",
				atn,
				atn,
				ptn,
				atn,
			),
			namespace,
		).
		Preload("Versions").
		Preload("Versions.Platforms").
		First(&p).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no provider found with given arguments (provider %s/%s)", namespace, name)
		} else {
			return nil, fmt.Errorf("error while querying the database: %v", err)
		}
	}

	sort.Slice(p.Versions, func(i, j int) bool {
		lhs := version.Version(p.Versions[i].Version)
		rhs := version.Version(p.Versions[j].Version)

		return version.Compare(lhs, rhs) <= 0
	})

	return &p, nil
}

func (r *DefaultProviderRepository) FindVersionPlatform(
	namespace, name, version, os, arch string,
) (*provider.Platform, error) {
	p := provider.Platform{}

	atn := (authority.Authority{}).TableName()
	ptn := (provider.Provider{}).TableName()
	vtn := (provider.Version{}).TableName()
	pltn := (provider.Platform{}).TableName()

	err := r.Database.Handler().
		Joins(
			fmt.Sprintf(
				"JOIN %s ON %s.id = %s.version_id AND %s.version = ?",
				vtn,
				vtn,
				pltn,
				vtn,
			),
			version,
		).
		Joins(
			fmt.Sprintf(
				"JOIN %s ON %s.id = %s.provider_id AND LOWER(%s.name) = LOWER(?)",
				ptn,
				ptn,
				vtn,
				ptn,
			),
			name,
		).
		Joins(
			fmt.Sprintf(
				"JOIN %s ON %s.id = %s.authority_id AND LOWER(%s.name) = LOWER(?)",
				atn,
				atn,
				ptn,
				atn,
			),
			namespace,
		).
		Where(fmt.Sprintf("%s.system = ? AND %s.architecture = ?", pltn, pltn), os, arch).
		Preload("Version.Provider").
		First(&p).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		} else {
			return nil, fmt.Errorf("error while querying the database: %v", err)
		}
	}

	return &p, nil
}

func (r *DefaultProviderRepository) Upsert(p provider.Provider) (*provider.Provider, error) {
	if len(p.Versions) != 1 {
		if err := r.Database.Handler().Save(&p).Error; err != nil {
			return nil, err
		}
	} else {
		if err := r.Database.Handler().Create(&p).Error; err != nil {
			return nil, err
		}
	}

	return &p, nil
}

func (r *DefaultProviderRepository) Delete(p *provider.Provider) error {
	if err := r.Database.Handler().Delete(p).Error; err != nil {
		return err
	}

	return nil
}

func (r *DefaultProviderRepository) DeleteVersion(p *provider.Provider, version string) error {
	var toDeleteVersion *provider.Version = nil
	for _, v := range p.Versions {
		if v.Version == version {
			toDeleteVersion = &v
			break
		}
	}

	if toDeleteVersion != nil {
		var toDelete any
		if len(p.Versions) == 1 {
			toDelete = &p
		} else {
			toDelete = toDeleteVersion
		}

		if err := r.Database.Handler().Delete(toDelete).Error; err != nil {
			return err
		}
	}

	return nil
}
