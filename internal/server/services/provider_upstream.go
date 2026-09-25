package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/repositories"
	"terralist/pkg/file"
	"terralist/pkg/metrics"
	"terralist/pkg/registry"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type (
	registryPlatform = registry.Platform
	registryKey      = registry.GPGPublicKey
)

var (
	// ErrFetchRequiresCreate is returned when serving a package would fetch it
	// from the upstream registry and the caller may only read.
	ErrFetchRequiresCreate = errors.New("fetching a package from the upstream registry requires the create permission")
)

// upstreamAuthority returns the authority when it has an enabled upstream and
// the caller may fetch from it. Fetched packages must be stored, so without a
// storage resolver the upstream is never consulted.
func (s *DefaultProviderService) upstreamAuthority(namespace string, withUpstream bool) *authority.Authority {
	if !withUpstream || s.Upstream == nil || s.Resolver == nil {
		return nil
	}

	a, err := s.AuthorityService.GetByName(namespace)
	if err != nil || !a.UpstreamEnabled {
		return nil
	}

	return a
}

// upstreamVersions lists the upstream versions an authority allows for a
// provider, or reports the upstream unavailable.
func (s *DefaultProviderService) upstreamVersions(a *authority.Authority, name string) ([]UpstreamVersion, error) {
	if a == nil {
		return nil, nil
	}

	versions, err := s.Upstream.ProviderVersions(a, name)
	if err != nil {
		metrics.RecordError("upstream", "error")
		log.Warn().Err(err).Str("authority", a.Name).Str("provider", name).Msg("Could not list the upstream versions.")

		return nil, upstreamFailure(err)
	}

	return versions, nil
}

// localVersions returns the set of versions a provider holds, whether the
// registry protocol can serve them or not. Uploaded versions always win over
// the upstream ones.
func localVersions(p *provider.Provider) map[string]struct{} {
	return lo.SliceToMap(p.Versions, func(v provider.Version) (string, struct{}) {
		return v.Version, struct{}{}
	})
}

// mergeUpstreamVersions appends the upstream versions the provider does not
// hold locally to the registry version list.
func mergeUpstreamVersions(dto *provider.VersionListProviderDTO, local map[string]struct{}, upstream []UpstreamVersion) {
	for _, v := range upstream {
		if _, ok := local[v.Version]; ok {
			continue
		}

		dto.Versions = append(dto.Versions, provider.VersionListVersionDTO{
			Version:   v.Version,
			Protocols: v.Protocols,
			Platforms: lo.Map(v.Platforms, func(p registryPlatform, _ int) provider.VersionListPlatformDTO {
				return provider.VersionListPlatformDTO{System: p.OS, Architecture: p.Arch}
			}),
		})
	}
}

// mergeUpstreamArchives lists the upstream platforms of a version that are
// not stored, pointing their download back at the mirror document location.
func mergeUpstreamArchives(dto *provider.MirrorArchivesDTO, name, version string, upstream []UpstreamVersion, metadata *UpstreamVersionMetadata) {
	v, found := lo.Find(upstream, func(v UpstreamVersion) bool { return v.Version == version })
	if !found || metadata == nil {
		return
	}

	for _, p := range v.Platforms {
		key := fmt.Sprintf("%s_%s", p.OS, p.Arch)
		if _, ok := dto.Archives[key]; ok {
			continue
		}

		fileName := provider.PackageFileName(name, version, p.OS, p.Arch)
		digest, ok := metadata.ShaSums[fileName]
		if !ok {
			continue
		}

		dto.Archives[key] = provider.MirrorArchiveDTO{
			URL:    fileName,
			Hashes: []string{fmt.Sprintf("zh:%s", digest)},
		}
	}
}

// ensureUpstreamVersion returns the local version row, creating it from the
// upstream metadata when the provider does not hold the version yet: the
// SHA256SUMS document and its signature are stored and the upstream signing
// keys recorded, so the registry protocol can serve the version. A version
// uploaded by an operator is never completed from the upstream.
func (s *DefaultProviderService) ensureUpstreamVersion(a *authority.Authority, name, version string) (*provider.Version, error) {
	current, err := s.ProviderRepository.Find(a.Name, name)
	if err != nil {
		current = &provider.Provider{AuthorityID: a.ID, Name: name}
	}

	if v := current.GetVersion(version); v != nil {
		if v.Origin != provider.OriginUpstream {
			return nil, fmt.Errorf("version %s of %s/%s was uploaded: %w", version, a.Name, name, repositories.ErrNotFound)
		}

		return v, nil
	}

	metadata, err := s.Upstream.ProviderVersion(a, name, version)
	if err != nil {
		return nil, upstreamFailure(err)
	}

	keys, err := json.Marshal(provider.SigningKeysDTO{
		Keys: lo.Map(metadata.SigningKeys, func(k registryKey, _ int) provider.PublicKeyDTO {
			return provider.PublicKeyDTO{
				KeyId:          k.KeyID,
				AsciiArmor:     k.ASCIIArmor,
				TrustSignature: k.TrustSignature,
				Source:         k.Source,
				SourceURL:      k.SourceURL,
			}
		}),
	})
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("terraform-provider-%s_%s", name, version)
	stored, err := s.uploadFiles(a.Name, name, version, map[string]file.File{
		shaSumsKey:    file.NewInMemoryFile(prefix+"_SHA256SUMS", metadata.ShaSumsDocument),
		shaSumsSigKey: file.NewInMemoryFile(prefix+"_SHA256SUMS.sig", metadata.Signature),
	})
	if err != nil {
		return nil, err
	}

	current.Versions = append(current.Versions, provider.Version{
		Version:             version,
		Protocols:           strings.Join(metadata.Protocols, ","),
		ShaSumsUrl:          stored[shaSumsKey],
		ShaSumsSignatureUrl: stored[shaSumsSigKey],
		SigningKeys:         string(keys),
		Origin:              provider.OriginUpstream,
	})

	if _, err := s.ProviderRepository.Upsert(*current); err != nil {
		if errors.Is(err, repositories.ErrAlreadyExists) {
			// Another request created the version meanwhile.
			return s.storedUpstreamVersion(a, name, version, err)
		}

		return nil, err
	}

	return &current.Versions[len(current.Versions)-1], nil
}

// storedUpstreamVersion re-reads a version pulled from the upstream by another
// request, failing with cause when there is none.
func (s *DefaultProviderService) storedUpstreamVersion(a *authority.Authority, name, version string, cause error) (*provider.Version, error) {
	current, err := s.ProviderRepository.Find(a.Name, name)
	if err != nil {
		return nil, cause
	}

	v := current.GetVersion(version)
	if v == nil || v.Origin != provider.OriginUpstream {
		return nil, cause
	}

	return v, nil
}

// upstreamDownloadDTO describes a platform that is not stored yet, pointing
// Terraform at the mirror route that fetches it on first download.
func (s *DefaultProviderService) upstreamDownloadDTO(a *authority.Authority, v *provider.Version, name, system, architecture string) (*provider.DownloadPlatformDTO, error) {
	metadata, err := s.Upstream.ProviderVersion(a, name, v.Version)
	if err != nil {
		return nil, upstreamFailure(err)
	}

	fileName := provider.PackageFileName(name, v.Version, system, architecture)
	digest, ok := metadata.ShaSums[fileName]
	if !ok {
		return nil, fmt.Errorf("platform %s_%s of %s/%s %s: %w", system, architecture, a.Name, name, v.Version, repositories.ErrNotFound)
	}

	keys, ok := v.SigningKeysDTO()
	if !ok {
		keys = s.authorityKeys(a)
	}

	dto := &provider.DownloadPlatformDTO{
		Protocols:           strings.Split(v.Protocols, ","),
		System:              system,
		Architecture:        architecture,
		FileName:            fileName,
		DownloadUrl:         fmt.Sprintf("%s/%s/%s/%s", s.MirrorBaseURL, a.Name, name, fileName),
		ShaSumsUrl:          v.ShaSumsUrl,
		ShaSumsSignatureUrl: v.ShaSumsSignatureUrl,
		ShaSum:              digest,
		SigningKeys:         keys,
	}

	if s.Resolver != nil {
		if dto.ShaSumsUrl, err = s.Resolver.Find(v.ShaSumsUrl); err != nil {
			return nil, fmt.Errorf("could not resolve shasums location: %v", err)
		}

		if dto.ShaSumsSignatureUrl, err = s.Resolver.Find(v.ShaSumsSignatureUrl); err != nil {
			return nil, fmt.Errorf("could not resolve shasums signature location: %v", err)
		}
	}

	return dto, nil
}

// authorityKeys maps the keys of an authority to the registry protocol format.
func (s *DefaultProviderService) authorityKeys(a *authority.Authority) provider.SigningKeysDTO {
	return provider.SigningKeysDTO{
		Keys: lo.Map(a.Keys, func(k authority.Key, _ int) provider.PublicKeyDTO {
			return provider.PublicKeyDTO{
				KeyId:          k.KeyId,
				AsciiArmor:     k.AsciiArmor,
				TrustSignature: k.TrustSignature,
				Source:         a.Name,
				SourceURL:      a.PolicyURL,
			}
		}),
	}
}

func (s *DefaultProviderService) Download(namespace, name, version, system, architecture string, allowFetch bool) (string, error) {
	p, err := s.ProviderRepository.FindVersionPlatform(namespace, name, version, system, architecture)
	if err == nil {
		return s.locationURL(p.Location)
	}

	if !errors.Is(err, repositories.ErrNotFound) {
		return "", err
	}

	a := s.upstreamAuthority(namespace, true)
	if a == nil {
		return "", fmt.Errorf("platform %s_%s of %s/%s %s: %w", system, architecture, namespace, name, version, repositories.ErrNotFound)
	}

	if !allowFetch {
		return "", ErrFetchRequiresCreate
	}

	key := fmt.Sprintf("%s/%s/%s/%s_%s", namespace, name, version, system, architecture)
	location, err, _ := s.fetches.Do(key, func() (any, error) {
		return s.fetchPackage(a, name, version, system, architecture)
	})
	if err != nil {
		return "", err
	}

	return s.locationURL(location.(string)) //nolint:forcetypeassert
}

// fetchPackage downloads one platform package from the upstream registry with
// its digest enforced, stores it and records the platform, returning the
// storage key.
func (s *DefaultProviderService) fetchPackage(a *authority.Authority, name, version, system, architecture string) (string, error) {
	start := time.Now()

	pkg, err := s.Upstream.ProviderPackage(a, name, version, system, architecture)
	if err != nil {
		return "", err
	}

	v, err := s.ensureUpstreamVersion(a, name, version)
	if err != nil {
		return "", err
	}

	if existing := v.GetPlatform(system, architecture); existing != nil {
		return existing.Location, nil
	}

	archive, cleanup, err := s.Fetcher.FetchFileChecksum(pkg.FileName, pkg.URL, pkg.ShaSum, nil)
	if err != nil {
		return "", fmt.Errorf("could not fetch %s from the upstream registry: %v", pkg.FileName, err)
	}
	defer cleanup()

	platformKey := fmt.Sprintf("%s_%s", system, architecture)
	stored, err := s.uploadFiles(a.Name, name, version, map[string]file.File{platformKey: archive})
	if err != nil {
		return "", err
	}

	current, err := s.ProviderRepository.Find(a.Name, name)
	if err != nil {
		return "", err
	}

	storedVersion := current.GetVersion(version)
	if storedVersion == nil {
		return "", fmt.Errorf("version %s of %s/%s disappeared while fetching", version, a.Name, name)
	}

	storedVersion.Platforms = append(storedVersion.Platforms, provider.Platform{
		System:       system,
		Architecture: architecture,
		Location:     stored[platformKey],
		ShaSum:       pkg.ShaSum,
		Origin:       provider.OriginUpstream,
	})

	if _, err := s.ProviderRepository.Upsert(*current); err != nil {
		if errors.Is(err, repositories.ErrAlreadyExists) {
			// Another request stored the platform meanwhile.
			if v, findErr := s.storedUpstreamVersion(a, name, version, err); findErr == nil {
				if p := v.GetPlatform(system, architecture); p != nil {
					return p.Location, nil
				}
			}
		}

		return "", err
	}

	metrics.RecordUpstreamFetch(lo.FromPtr(a.UpstreamHostname), time.Since(start).Seconds())
	metrics.RecordArtifactUpload("provider", a.Name)

	return stored[platformKey], nil
}

func (s *DefaultProviderService) Fetch(namespace, name, version string, platforms []string) []provider.FetchResultDTO {
	results := make([]provider.FetchResultDTO, 0, len(platforms))

	for _, platform := range platforms {
		result := provider.FetchResultDTO{Platform: platform}

		system, architecture, ok := strings.Cut(platform, "_")
		if !ok || system == "" || architecture == "" {
			result.Error = fmt.Sprintf("invalid platform %q, expected os_arch", platform)
		} else if _, err := s.Download(namespace, name, version, system, architecture, true); err != nil {
			result.Error = err.Error()
		}

		results = append(results, result)
	}

	return results
}

// locationURL resolves a storage key to a download URL, or returns it as is
// without a resolver.
func (s *DefaultProviderService) locationURL(location string) (string, error) {
	if s.Resolver == nil {
		return location, nil
	}

	url, err := s.Resolver.Find(location)
	if err != nil {
		return "", fmt.Errorf("could not resolve package location: %v", err)
	}

	return url, nil
}
