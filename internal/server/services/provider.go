package services

import (
	"fmt"

	"github.com/valentindeaconu/terralist/internal/server/database"
	models "github.com/valentindeaconu/terralist/internal/server/models/provider"
)

type ProviderService struct {
	Database database.Engine
}

func (p *ProviderService) Find(namespace string, name string) (models.Provider, error) {
	provider := models.Provider{}

	h := p.Database.Handler().Where(models.Provider{
		Name:      name,
		Namespace: namespace,
	}).
		Preload("Versions.Platforms.SigningKeys").
		Find(&provider)

	if h.Error != nil {
		return provider, fmt.Errorf("no provider found with given arguments (provider %s/%s)", namespace, name)
	}

	return provider, nil
}

func (p *ProviderService) FindVersion(namespace string, name string, version string) (models.Version, error) {
	provider, err := p.Find(namespace, name)

	if err != nil {
		return models.Version{}, err
	}

	for _, v := range provider.Versions {
		if v.Version == version {
			return v, nil
		}
	}

	return models.Version{}, fmt.Errorf("no version found")
}

func (p *ProviderService) Upsert(new models.Provider) (models.Provider, error) {
	existing, err := p.Find(new.Namespace, new.Name)

	if err == nil {
		newVersion := new.Versions[0].Version

		for _, version := range existing.Versions {
			if version.Version == newVersion {
				return models.Provider{}, fmt.Errorf("version %s already exists", newVersion)
			}
		}

		existing.Versions = append(existing.Versions, new.Versions[0])

		if result := p.Database.Handler().Save(&existing); result.Error != nil {
			return models.Provider{}, err
		}

		return existing, nil

	}

	if result := p.Database.Handler().Create(&new); result.Error != nil {
		return models.Provider{}, err
	}

	return new, nil
}

func (p *ProviderService) Delete(namespace string, name string) error {
	provider, err := p.Find(namespace, name)

	if err == nil {
		return fmt.Errorf("provider %s/%s is not uploaded to this registry", namespace, name)
	}

	if result := p.Database.Handler().Delete(&provider); result.Error != nil {
		return result.Error
	}

	return nil
}

func (p *ProviderService) DeleteVersion(namespace string, name string, version string) error {
	provider, err := p.Find(namespace, name)

	if err == nil {
		return fmt.Errorf("provider %s/%s is not uploaded to this registry", namespace, name)
	}

	q := false
	for idx, ver := range provider.Versions {
		if ver.Version == version {
			provider.Versions = append(provider.Versions[:idx], provider.Versions[idx+1:]...)
			q = true
		}
	}

	if q {
		if result := p.Database.Handler().Save(&provider); result.Error != nil {
			return result.Error
		}
	}

	return fmt.Errorf("no version found")
}
