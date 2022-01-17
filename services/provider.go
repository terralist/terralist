package services

import (
	"fmt"

	"github.com/valentindeaconu/terralist/database"
	"github.com/valentindeaconu/terralist/models/provider"
)

func ProviderFind(namespace string, name string) (provider.Provider, error) {
	p := provider.Provider{}

	h := database.Handler().Where(provider.Provider{
		Name:      name,
		Namespace: namespace,
	}).
		Preload("Versions.Platforms.SigningKeys").
		Find(&p)

	if h.Error != nil {
		return p, fmt.Errorf("no provider found with given arguments (provider %s/%s)", namespace, name)
	}

	return p, nil
}

func ProviderFindVersion(namespace string, name string, version string) (provider.Version, error) {
	p, err := ProviderFind(namespace, name)

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

func ProviderUpsert(new provider.Provider) (provider.Provider, error) {
	existing, err := ProviderFind(new.Namespace, new.Name)

	if err == nil {
		newVersion := new.Versions[0].Version

		for _, version := range existing.Versions {
			if version.Version == newVersion {
				return provider.Provider{}, fmt.Errorf("version %s already exists", newVersion)
			}
		}

		existing.Versions = append(existing.Versions, new.Versions[0])

		if result := database.Handler().Save(&existing); result.Error != nil {
			return provider.Provider{}, err
		}

		return existing, nil

	}

	if result := database.Handler().Create(&new); result.Error != nil {
		return provider.Provider{}, err
	}

	return new, nil
}

func ProviderDelete(namespace string, name string) error {
	p, err := ProviderFind(namespace, name)

	if err == nil {
		return fmt.Errorf("provider %s/%s is not uploaded to this registry", namespace, name)
	}

	if result := database.Handler().Delete(&p); result.Error != nil {
		return result.Error
	}

	return nil
}

func ProviderVersionDelete(namespace string, name string, version string) error {
	p, err := ProviderFind(namespace, name)

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
		if result := database.Handler().Save(&p); result.Error != nil {
			return result.Error
		}
	}

	return fmt.Errorf("no version found")
}
