package services

import (
	"fmt"

	"terralist/internal/server/models/provider"
	"terralist/pkg/database"
)

type ProviderService struct {
	Database database.Engine
}

func (s *ProviderService) Find(namespace string, name string) (provider.Provider, error) {
	p := provider.Provider{}

	if err := s.Database.Handler().Where(provider.Provider{
		Name:      name,
		Namespace: namespace,
	}).
		Preload("Versions.Platforms.SigningKeys").
		Find(&p).
		Error; err != nil {
		return provider.Provider{}, fmt.Errorf("no provider found with given arguments (provider %s/%s)", namespace, name)
	}

	return p, nil
}

func (s *ProviderService) FindVersion(namespace string, name string, version string) (provider.Version, error) {
	p, err := s.Find(namespace, name)

	if err != nil {
		return provider.Version{}, err
	}

	for _, v := range p.Versions {
		if v.Version == version {
			return v, nil
		}
	}

	return provider.Version{}, fmt.Errorf("no version found")
}

func (s *ProviderService) Upsert(new provider.Provider) (provider.Provider, error) {
	existing, err := s.Find(new.Namespace, new.Name)

	if err == nil {
		newVersion := new.Versions[0].Version

		for _, version := range existing.Versions {
			if version.Version == newVersion {
				return provider.Provider{}, fmt.Errorf("version %s already exists", newVersion)
			}
		}

		existing.Versions = append(existing.Versions, new.Versions[0])

		if result := s.Database.Handler().Save(&existing); result.Error != nil {
			return provider.Provider{}, err
		}

		return existing, nil

	}

	if result := s.Database.Handler().Create(&new); result.Error != nil {
		return provider.Provider{}, err
	}

	return new, nil
}

func (s *ProviderService) Delete(namespace string, name string) error {
	p, err := s.Find(namespace, name)

	if err == nil {
		return fmt.Errorf("provider %s/%s is not uploaded to this registry", namespace, name)
	}

	if result := s.Database.Handler().Delete(&p); result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *ProviderService) DeleteVersion(namespace string, name string, version string) error {
	p, err := s.Find(namespace, name)

	if err == nil {
		return fmt.Errorf("provider %s/%s is not uploaded to this registry", namespace, name)
	}

	q := false
	for idx, ver := range p.Versions {
		if ver.Version == version {
			p.Versions = append(p.Versions[:idx], p.Versions[idx+1:]...)
			q = true
		}
	}

	if q {
		if result := s.Database.Handler().Delete(&p); result.Error != nil {
			return result.Error
		}
	}

	return fmt.Errorf("no version found")
}
