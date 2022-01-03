package service

import (
	"fmt"

	"github.com/valentindeaconu/terralist/database"
	"github.com/valentindeaconu/terralist/model/provider"
)

type ProviderService struct {
}

func (m *ProviderService) Find(namespace string, name string) (provider.Provider, error) {
	p := provider.Provider{}

	h := database.AppDatabase.Handler.
		Where(provider.Provider{
			Name:      name,
			Namespace: namespace,
		}).
		Preload("Versions.Platforms.SigningKeys").
		Find(&p)

	if h.RowsAffected > 0 {
		return p, nil
	}

	return p, fmt.Errorf("no provider found with given arguments (provider %s/%s)", namespace, name)
}

func (m *ProviderService) FindVersion(namespace string, name string, version string) (provider.Version, error) {
	p, err := m.Find(namespace, name)

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

func (m *ProviderService) Upsert(new provider.Provider) (provider.Provider, error) {
	existing, err := m.Find(new.Namespace, new.Name)

	if err == nil {
		newVersion := new.Versions[0].Version

		for _, version := range existing.Versions {
			if version.Version == newVersion {
				return provider.Provider{}, fmt.Errorf("version %s already exists", newVersion)
			}
		}

		existing.Versions = append(existing.Versions, new.Versions[0])

		if result := database.AppDatabase.Handler.Save(&existing); result.RowsAffected != 1 {
			return provider.Provider{}, result.Error
		} else {
			return existing, nil
		}
	} else {
		if result := database.AppDatabase.Handler.Create(&new); result.RowsAffected != 1 {
			return provider.Provider{}, result.Error
		} else {
			return new, nil
		}
	}
}
