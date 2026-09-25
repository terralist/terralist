package services

import (
	"errors"
	"fmt"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/internal/server/repositories"
	"terralist/pkg/file"
	"terralist/pkg/metrics"

	"github.com/samber/lo"
)

// upstreamAuthority returns the authority when it has an enabled upstream and
// the caller may fetch from it. Fetched modules must be stored, so without a
// storage resolver the upstream is never consulted.
func (s *DefaultModuleService) upstreamAuthority(namespace string, withUpstream bool) *authority.Authority {
	if !withUpstream || s.Upstream == nil || s.Resolver == nil {
		return nil
	}

	a, err := s.AuthorityService.GetByName(namespace)
	if err != nil || !a.UpstreamEnabled {
		return nil
	}

	return a
}

// upstreamModuleVersions lists the upstream versions an authority allows for a
// module, or nothing when the upstream cannot be consulted.
func (s *DefaultModuleService) upstreamModuleVersions(a *authority.Authority, name, system string) []string {
	if a == nil {
		return nil
	}

	versions, err := s.Upstream.ModuleVersions(a, name, system)
	if err != nil {
		metrics.RecordError("upstream", "error")
		return nil
	}

	return versions
}

// upstreamArchiveURL points a download at the archive route, which fetches the
// version from the upstream on first request, when the upstream offers it.
func (s *DefaultModuleService) upstreamArchiveURL(a *authority.Authority, name, system, version string) (*string, error) {
	if !lo.Contains(s.upstreamModuleVersions(a, name, system), version) {
		return nil, fmt.Errorf("version %s of %s/%s/%s: %w", version, a.Name, name, system, repositories.ErrNotFound)
	}

	url := fmt.Sprintf("%s/%s/%s/%s/%s/archive", s.ArchiveBaseURL, a.Name, name, system, version)

	metrics.RecordRequest(a.Name, "download")
	metrics.RecordArtifactDownload("module", a.Name)

	return &url, nil
}

func (s *DefaultModuleService) Download(namespace, name, provider, version string, allowFetch bool) (string, error) {
	location, err := s.ModuleRepository.FindVersionLocation(namespace, name, provider, version)
	if err == nil {
		return s.locationURL(*location)
	}

	if !errors.Is(err, repositories.ErrNotFound) {
		return "", err
	}

	if !allowFetch {
		return "", ErrFetchRequiresCreate
	}

	a := s.upstreamAuthority(namespace, true)
	if a == nil {
		return "", fmt.Errorf("version %s of %s/%s/%s: %w", version, namespace, name, provider, repositories.ErrNotFound)
	}

	key := fmt.Sprintf("%s/%s/%s/%s", namespace, name, provider, version)
	stored, err, _ := s.fetches.Do(key, func() (any, error) {
		return s.fetchVersion(a, name, provider, version)
	})
	if err != nil {
		return "", err
	}

	return s.locationURL(stored.(string)) //nolint:forcetypeassert
}

// fetchVersion downloads a module version from its upstream source, stores it
// the way an upload is stored and returns the storage key.
func (s *DefaultModuleService) fetchVersion(a *authority.Authority, name, system, version string) (string, error) {
	location, err := s.Upstream.ModuleLocation(a, name, system, version)
	if err != nil {
		return "", err
	}

	if stored, err := s.ModuleRepository.FindVersionLocation(a.Name, name, system, version); err == nil {
		return *stored, nil
	}

	dto := module.CreateDTO{
		AuthorityID:      a.ID,
		Name:             name,
		Provider:         system,
		Origin:           module.OriginUpstream,
		VersionCreateDTO: module.VersionCreateDTO{Version: version},
	}

	if err := s.Upload(&dto, file.NewRemoteFile(location, nil)); err != nil {
		return "", fmt.Errorf("could not fetch %s/%s/%s %s from %s: %v", a.Name, name, system, version, location, err)
	}

	stored, err := s.ModuleRepository.FindVersionLocation(a.Name, name, system, version)
	if err != nil {
		return "", err
	}

	return *stored, nil
}

func (s *DefaultModuleService) Fetch(namespace, name, provider, version string) error {
	_, err := s.Download(namespace, name, provider, version, true)

	return err
}

// locationURL resolves a storage key to a download URL, or returns it as is
// without a resolver.
func (s *DefaultModuleService) locationURL(location string) (string, error) {
	if s.Resolver == nil {
		return location, nil
	}

	url, err := s.Resolver.Find(location)
	if err != nil {
		return "", fmt.Errorf("could not resolve location: %v", err)
	}

	return url, nil
}
