package services

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"sort"
	"terralist/pkg/storage/batch/find"
	"terralist/pkg/storage/batch/purge"
	"terralist/pkg/storage/batch/store"

	"terralist/internal/server/models/provider"
	"terralist/pkg/database"
	"terralist/pkg/storage/batch"
	batchFactory "terralist/pkg/storage/batch/factory"
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

	for i, v := range p.Versions {
		for j, plat := range v.Platforms {
			result, err := batchFactory.NewBatch(batch.FIND, s.Resolver).
				Add(&find.BatchInput{Key: plat.FetchKey}).
				Add(&find.BatchInput{Key: plat.ShaSumsFetchKey}).
				Add(&find.BatchInput{Key: plat.ShaSumsSignatureFetchKey}).
				Commit()

			if err != nil {
				return nil, err
			}
			res := result.(*find.BatchOutput)

			p.Versions[i].Platforms[j].FetchKey = res.URLs[0]
			p.Versions[i].Platforms[j].ShaSumsFetchKey = res.URLs[1]
			p.Versions[i].Platforms[j].ShaSumsSignatureFetchKey = res.URLs[2]
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
		for i, plat := range toUpsertVersion.Platforms {
			result, err := batchFactory.NewBatch(batch.STORE, s.Resolver).
				Add(&store.BatchInput{URL: plat.FetchKey, Archive: true}).
				Add(&store.BatchInput{URL: plat.ShaSumsFetchKey, Archive: false}).
				Add(&store.BatchInput{URL: plat.ShaSumsSignatureFetchKey, Archive: false}).
				Commit()

			if err != nil {
				return nil, err
			}
			res := result.(*store.BatchOutput)

			toUpsertVersion.Platforms[i].FetchKey = res.Keys[0]
			toUpsertVersion.Platforms[i].ShaSumsFetchKey = res.Keys[1]
			toUpsertVersion.Platforms[i].ShaSumsSignatureFetchKey = res.Keys[2]
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

	for _, ver := range p.Versions {
		for _, plat := range ver.Platforms {
			if _, err := batchFactory.NewBatch(batch.PURGE, s.Resolver).
				Add(&purge.BatchInput{Key: plat.FetchKey}).
				Add(&purge.BatchInput{Key: plat.ShaSumsFetchKey}).
				Add(&purge.BatchInput{Key: plat.ShaSumsSignatureFetchKey}).
				Commit(); err != nil {
				log.Warn().
					AnErr("Error", err).
					Str("Provider", fmt.Sprintf("%s/%s", namespace, name)).
					Str("Version", ver.Version).
					Str("Platform", fmt.Sprintf("%s_%s", plat.System, plat.Architecture)).
					Strs("Keys", []string{plat.FetchKey, plat.ShaSumsFetchKey, plat.ShaSumsSignatureFetchKey}).
					Msg("Could not purge, require manual clean-up")
			}
		}
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
			for _, ver := range p.Versions {
				for _, plat := range ver.Platforms {
					if _, err := batchFactory.NewBatch(batch.PURGE, s.Resolver).
						Add(&purge.BatchInput{Key: plat.FetchKey}).
						Add(&purge.BatchInput{Key: plat.ShaSumsFetchKey}).
						Add(&purge.BatchInput{Key: plat.ShaSumsSignatureFetchKey}).
						Commit(); err != nil {
						log.Error().
							AnErr("Error", err).
							Str("Provider", fmt.Sprintf("%s/%s", namespace, name)).
							Str("Version", ver.Version).
							Str("Platform", fmt.Sprintf("%s_%s", plat.System, plat.Architecture)).
							Strs("Keys", []string{plat.FetchKey, plat.ShaSumsFetchKey, plat.ShaSumsSignatureFetchKey}).
							Msg("Could not purge, require manual clean-up")
					}
				}
			}

			if err := s.Database.Handler().Delete(p).Error; err != nil {
				return err
			}
		} else {
			for _, plat := range toDelete.Platforms {
				if _, err := batchFactory.NewBatch(batch.PURGE, s.Resolver).
					Add(&purge.BatchInput{Key: plat.FetchKey}).
					Add(&purge.BatchInput{Key: plat.ShaSumsFetchKey}).
					Add(&purge.BatchInput{Key: plat.ShaSumsSignatureFetchKey}).
					Commit(); err != nil {
					log.Error().
						AnErr("Error", err).
						Str("Provider", fmt.Sprintf("%s/%s", namespace, name)).
						Str("Version", toDelete.Version).
						Str("Platform", fmt.Sprintf("%s_%s", plat.System, plat.Architecture)).
						Strs("Keys", []string{plat.FetchKey, plat.ShaSumsFetchKey, plat.ShaSumsSignatureFetchKey}).
						Msg("Could not purge, require manual clean-up")
				}
			}

			if err := s.Database.Handler().Delete(toDelete).Error; err != nil {
				return err
			}
		}

		return nil
	}

	return nil
}
