package services

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"sort"
	"terralist/pkg/storage"

	"terralist/internal/server/models/provider"
	"terralist/pkg/database"
	"terralist/pkg/version"

	"gorm.io/gorm"
)

type ProviderService struct {
	Database database.Engine
	Resolver storage.Resolver
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

		// At this point, all versions passes the check, we can start resolving
		var stored []string
		for i, plat := range toUpsertVersion.Platforms {
			url, err := s.Resolver.Store(plat.DownloadUrl)
			if err != nil {
				log.Error().
					Str("Provider", fmt.Sprintf("%s/%s", n.Namespace, n.Name)).
					Str("Version", toUpsertVersion.Version).
					Str("Platform", fmt.Sprintf("%s_%s", plat.System, plat.Architecture)).
					AnErr("Error", err).
					Msg("Error while uploading a new provider version platform.")

				break
			}

			toUpsertVersion.Platforms[i].DownloadUrl = url
			stored = append(stored, url)
		}

		if len(stored) != len(toUpsertVersion.Platforms) {
			// Not all platforms where uploaded, rollback stored ones and return with
			// error
			for _, url := range stored {
				err := s.Resolver.Purge(url)
				if err != nil {
					log.Error().
						Str("Provider", fmt.Sprintf("%s/%s", n.Namespace, n.Name)).
						Str("Version", toUpsertVersion.Version).
						Str("Remote URL", url).
						AnErr("Error", err).
						Msg("Error while purging a a provider version platform.")
				}
			}

			// All provider files where removed, return with error
			return nil, fmt.Errorf("could not upload all provider platforms")
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
