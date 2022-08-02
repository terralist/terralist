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

// ProviderRepository describes a service that can interact with the providers database
type ProviderRepository interface {
	// Find searches for a specific provider
	Find(namespace string, name string) (*provider.Provider, error)

	// FindVersion searches for a specific version of a provider
	FindVersion(namespace string, name string, version string) (*provider.Version, error)

	// Upsert either updates or creates a new (if it does not already exist) provider
	Upsert(string, provider.Provider) (*provider.Provider, error)

	// Delete removes a provider with all its data (versions)
	Delete(*provider.Provider) error

	// DeleteVersion removes a version from a provider
	DeleteVersion(p *provider.Provider, version string) error
}

// DefaultProviderRepository is a concrete implementation of ProviderRepository
type DefaultProviderRepository struct {
	Database database.Engine
}

func (r *DefaultProviderRepository) Find(namespace string, name string) (*provider.Provider, error) {
	p := provider.Provider{}

	atn := (authority.Authority{}).TableName()
	ptn := (provider.Provider{}).TableName()

	err := r.Database.Handler().
		Where(provider.Provider{
			Name: name,
		}).
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

func (r *DefaultProviderRepository) FindVersion(namespace string, name string, version string) (*provider.Version, error) {
	v := provider.Version{}

	atn := (authority.Authority{}).TableName()
	ptn := (provider.Provider{}).TableName()
	vtn := (provider.Version{}).TableName()

	err := r.Database.Handler().
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
		Where(&provider.Version{
			Version: version,
		}).
		Preload("Platforms").
		Preload("Provider").
		First(&v).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no version found with given arguments (provider %s/%s; version %s)", namespace, name, version)
		} else {
			return nil, fmt.Errorf("error while querying the database: %v", err)
		}
	}

	return &v, nil
}

// Upsert is designed to upload an entire provider, but in reality,
// it will only upload a single version at a time
func (r *DefaultProviderRepository) Upsert(namespace string, n provider.Provider) (*provider.Provider, error) {
	p, err := r.Find(namespace, n.Name)
	if err == nil {
		// The provider already exists, check if for version conflicts
		toUpsertVersion := &n.Versions[0]

		for _, v := range p.Versions {
			if version.Compare(version.Version(v.Version), version.Version(toUpsertVersion.Version)) == 0 {
				return nil, fmt.Errorf("version %s already exists", v.Version)
			}
		}

		p.Versions = append(p.Versions, *toUpsertVersion)

		if err := r.Database.Handler().Save(p).Error; err != nil {
			return nil, err
		}

		return p, nil
	}

	if err := r.Database.Handler().Create(&n).Error; err != nil {
		return nil, err
	}

	return &n, nil
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
