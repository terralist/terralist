package services

import (
	"errors"
	"fmt"
	"sort"

	"terralist/internal/server/models/provider"
	"terralist/pkg/database"
	"terralist/pkg/version"

	"gorm.io/gorm"
)

type ProviderService struct {
	Database database.Engine
}

func (s *ProviderService) Find(namespace string, name string) (*provider.Provider, error) {
	p := provider.Provider{}

	err := s.Database.Handler().Where(provider.Provider{
		Name:      name,
		Namespace: namespace,
	}).Preload("Versions").
		Preload("Versions.Platforms").
		Preload("Versions.Platforms.SigningKeys").
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
	p, err := s.Find(namespace, name)

	if err != nil {
		return nil, err
	}

	for _, v := range p.Versions {
		if v.Version == version {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("no version found")
}

func (s *ProviderService) Upsert(n provider.Provider) (*provider.Provider, error) {
	p, err := s.Find(n.Namespace, n.Name)

	if err == nil {
		newVersion := n.Versions[0].Version

		for _, v := range p.Versions {
			if v.Version == newVersion {
				return nil, fmt.Errorf("version %s already exists", newVersion)
			}
		}

		p.Versions = append(p.Versions, n.Versions[0])

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

	var toDelete *provider.Version = nil
	for _, v := range p.Versions {
		if v.Version == version {
			toDelete = &v
			break
		}
	}

	if toDelete != nil {
		if len(p.Versions) == 1 {
			if err := s.Database.Handler().Delete(p).Error; err != nil {
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
