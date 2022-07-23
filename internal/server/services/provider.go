package services

import (
	"errors"
	"fmt"
	"sort"

	"terralist/internal/server/models/provider"
	"terralist/pkg/database"
	"terralist/pkg/storage/resolver"
	"terralist/pkg/version"

	"gorm.io/gorm"
)

type ProviderService struct {
	Database database.Engine
	Resolver resolver.Resolver
}

func (s *ProviderService) Find(namespace string, name string) (*provider.Provider, error) {
	p := provider.Provider{}

	err := s.Database.Handler().Where(provider.Provider{
		Name:      name,
		Namespace: namespace,
	}).
		Preload("Authority").
		Preload("Authority.Keys").
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

func (s *ProviderService) FindVersion(namespace string, name string, version string) (*provider.Version, error) {
	v := provider.Version{}

	err := s.Database.Handler().
		Joins(
			"Provider",
			s.Database.Handler().Where(&provider.Provider{
				Name:      name,
				Namespace: namespace,
			}),
		).
		Where(&provider.Version{
			Version: version,
		}).
		Preload("Platforms").
		Preload("Provider.Authority.Keys").
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
func (s *ProviderService) Upsert(n provider.Provider) (*provider.Provider, error) {
	p, err := s.Find(n.Namespace, n.Name)
	if err == nil {
		// The provider already exists, check if for version conflicts
		toUpsertVersion := &n.Versions[0]

		for _, v := range p.Versions {
			if version.Compare(version.Version(v.Version), version.Version(toUpsertVersion.Version)) == 0 {
				return nil, fmt.Errorf("version %s already exists", v.Version)
			}
		}

		p.Versions = append(p.Versions, *toUpsertVersion)

		if err := s.Database.Handler().Save(p).Error; err != nil {
			return nil, err
		}

		return p, nil
	}

	if err := s.Database.Handler().Create(&n).Error; err != nil {
		return nil, err
	}

	return &n, nil
}

func (s *ProviderService) Delete(namespace string, name string) error {
	p, err := s.Find(namespace, name)
	if err != nil {
		return fmt.Errorf("provider %s/%s is not uploaded to this registry", namespace, name)
	}

	if err := s.Database.Handler().Delete(p).Error; err != nil {
		return err
	}

	return nil
}

func (s *ProviderService) DeleteVersion(namespace string, name string, version string) error {
	p, err := s.Find(namespace, name)
	if err != nil {
		return fmt.Errorf("provider %s/%s is not uploaded to this registry", namespace, name)
	}

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

		if err := s.Database.Handler().Delete(toDelete).Error; err != nil {
			return err
		}
	}

	return nil
}
