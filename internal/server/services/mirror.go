package services

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"terralist/internal/server/models/mirror"
	"terralist/internal/server/repositories"
	"terralist/pkg/file"
	"terralist/pkg/metrics"
	"terralist/pkg/storage"
	"terralist/pkg/version"

	"github.com/rs/zerolog/log"
	"golang.org/x/mod/sumdb/dirhash"
)

const (
	// mirrorStoragePrefix is the top-level directory under which all mirrored
	// packages are stored, keeping them apart from Terralist's own artifacts.
	mirrorStoragePrefix = "mirror"

	// h1HashPrefix marks a hash computed with the h1 scheme (the one Terraform
	// uses in its dependency lock file).
	h1HashPrefix = "h1:"
)

var (
	// hostnameRegexp accepts DNS hostnames with an optional port.
	hostnameRegexp = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*(:[0-9]{1,5})?$`)

	// identifierRegexp accepts provider namespaces and types.
	identifierRegexp = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
)

// MirrorService describes a service that holds the business logic for mirrored
// providers, exposed through the Terraform Provider Network Mirror Protocol.
type MirrorService interface {
	// ListVersions returns all mirrored versions of a provider.
	ListVersions(hostname, namespace, name string) (*mirror.VersionListDTO, error)

	// GetVersion returns the installation packages of a mirrored provider
	// version, with resolved download locations.
	GetVersion(hostname, namespace, name, version string) (*mirror.ArchivesDTO, error)

	// Upload stores the given provider packages for a version. Each archive
	// must be listed in the metadata and match its h1 hash. Archives listed in
	// the metadata but not uploaded are ignored.
	Upload(hostname, namespace, name, version string, metadata mirror.ArchivesDTO, archives []file.File) error

	// DeleteVersion removes a version of a mirrored provider.
	DeleteVersion(hostname, namespace, name, version string) error

	// Delete removes a mirrored provider with all its versions.
	Delete(hostname, namespace, name string) error

	// DeleteNamespace removes all mirrored providers under a namespace.
	DeleteNamespace(hostname, namespace string) error

	// DeleteHostname removes all mirrored providers under a hostname.
	DeleteHostname(hostname string) error
}

// DefaultMirrorService is a concrete implementation of MirrorService.
type DefaultMirrorService struct {
	MirrorRepository repositories.MirrorRepository
	Resolver         storage.Resolver
}

func (s *DefaultMirrorService) ListVersions(hostname, namespace, name string) (*mirror.VersionListDTO, error) {
	p, err := s.MirrorRepository.Find(hostname, namespace, name)
	if err != nil {
		return nil, err
	}

	dto := p.ToVersionListDTO()

	return &dto, nil
}

func (s *DefaultMirrorService) GetVersion(hostname, namespace, name, ver string) (*mirror.ArchivesDTO, error) {
	p, err := s.MirrorRepository.Find(hostname, namespace, name)
	if err != nil {
		return nil, err
	}

	v := p.GetVersion(ver)
	if v == nil {
		return nil, fmt.Errorf("provider %s/%s/%s does not contain version %s", hostname, namespace, name, ver)
	}

	dto := mirror.ArchivesDTO{
		Archives: make(map[string]mirror.ArchiveDTO, len(v.Platforms)),
	}

	for _, platform := range v.Platforms {
		url, err := s.Resolver.Find(platform.Location)
		if err != nil {
			return nil, fmt.Errorf("could not resolve package location for %s: %v", platform.String(), err)
		}

		dto.Archives[platform.String()] = platform.ToArchiveDTO(url)
	}

	metrics.RecordRequest(authorityLabel(hostname, namespace), "download")
	metrics.RecordArtifactDownload("mirror", authorityLabel(hostname, namespace))

	return &dto, nil
}

func (s *DefaultMirrorService) Upload(hostname, namespace, name, ver string, metadata mirror.ArchivesDTO, archives []file.File) error {
	hostname = strings.ToLower(hostname)

	if err := validateIdentity(hostname, namespace, name); err != nil {
		return err
	}

	if semVer := version.Version(ver); !semVer.Valid() {
		return fmt.Errorf("version should respect the semantic versioning standard (semver.org)")
	}

	if len(archives) == 0 {
		return fmt.Errorf("at least one provider package archive is required")
	}

	platforms, err := verifyArchives(metadata, archives)
	if err != nil {
		return err
	}

	current, err := s.MirrorRepository.Find(hostname, namespace, name)
	if err != nil {
		current = &mirror.Provider{
			Hostname:  hostname,
			Namespace: namespace,
			Name:      name,
		}
	}

	v := current.GetVersion(ver)
	if v == nil {
		current.Versions = append(current.Versions, mirror.Version{Version: ver})
		v = &current.Versions[len(current.Versions)-1]
	}

	for _, platform := range platforms {
		if v.GetPlatform(platform.System, platform.Architecture) != nil {
			return fmt.Errorf("platform %s already exists for version %s", platform.String(), ver)
		}
	}

	keys, err := s.storeArchives(hostname, namespace, name, ver, archives)
	if err != nil {
		return err
	}

	for i := range platforms {
		platforms[i].Location = keys[archives[i].Name()]
	}

	v.Platforms = append(v.Platforms, platforms...)

	if _, err := s.MirrorRepository.Upsert(*current); err != nil {
		s.purgeKeys(keys)
		return err
	}

	metrics.RecordArtifactUpload("mirror", authorityLabel(hostname, namespace))
	metrics.RecordRequest(authorityLabel(hostname, namespace), "upload")

	return nil
}

func (s *DefaultMirrorService) DeleteVersion(hostname, namespace, name, ver string) error {
	p, err := s.MirrorRepository.Find(hostname, namespace, name)
	if err != nil {
		return err
	}

	v := p.GetVersion(ver)
	if v == nil {
		return fmt.Errorf("provider %s/%s/%s does not contain version %s", hostname, namespace, name, ver)
	}

	s.purgeVersion(v)

	if err := s.MirrorRepository.DeleteVersion(p, ver); err != nil {
		return err
	}

	metrics.RecordArtifactDeletion("mirror", authorityLabel(hostname, namespace))

	return nil
}

func (s *DefaultMirrorService) Delete(hostname, namespace, name string) error {
	p, err := s.MirrorRepository.Find(hostname, namespace, name)
	if err != nil {
		return err
	}

	return s.deleteProvider(p)
}

func (s *DefaultMirrorService) DeleteNamespace(hostname, namespace string) error {
	providers, err := s.MirrorRepository.FindByNamespace(hostname, namespace)
	if err != nil {
		return err
	}

	if len(providers) == 0 {
		return fmt.Errorf("no mirrored provider found under %s/%s", hostname, namespace)
	}

	return s.deleteProviders(providers)
}

func (s *DefaultMirrorService) DeleteHostname(hostname string) error {
	providers, err := s.MirrorRepository.FindByHostname(hostname)
	if err != nil {
		return err
	}

	if len(providers) == 0 {
		return fmt.Errorf("no mirrored provider found under %s", hostname)
	}

	return s.deleteProviders(providers)
}

func (s *DefaultMirrorService) deleteProviders(providers []mirror.Provider) error {
	for i := range providers {
		if err := s.deleteProvider(&providers[i]); err != nil {
			return err
		}
	}

	return nil
}

// deleteProvider purges all packages of a provider and removes it.
func (s *DefaultMirrorService) deleteProvider(p *mirror.Provider) error {
	for i := range p.Versions {
		s.purgeVersion(&p.Versions[i])
	}

	if err := s.MirrorRepository.Delete(p); err != nil {
		return err
	}

	for range p.Versions {
		metrics.RecordArtifactDeletion("mirror", authorityLabel(p.Hostname, p.Namespace))
	}

	return nil
}

// storeArchives uploads the archives to the storage and returns the resulting
// keys, indexed by archive file name.
func (s *DefaultMirrorService) storeArchives(hostname, namespace, name, ver string, archives []file.File) (map[string]string, error) {
	prefix := fmt.Sprintf("%s/%s/%s/%s/%s", mirrorStoragePrefix, hostname, namespace, name, ver)
	keys := map[string]string{}

	for _, archive := range archives {
		key, err := s.Resolver.Store(&storage.StoreInput{
			Reader:      archive,
			Size:        archive.Metadata().Size(),
			ContentType: file.ContentType(archive),
			KeyPrefix:   prefix,
			FileName:    archive.Name(),
		})
		if err != nil {
			s.purgeKeys(keys)
			return nil, fmt.Errorf("could not upload %s: %v", archive.Name(), err)
		}

		keys[archive.Name()] = key
	}

	return keys, nil
}

// purgeVersion removes all packages of a version from the storage.
func (s *DefaultMirrorService) purgeVersion(v *mirror.Version) {
	for _, platform := range v.Platforms {
		if err := s.Resolver.Purge(platform.Location); err != nil {
			log.Warn().
				Str("location", platform.Location).
				Err(err).
				Msg("Could not purge mirrored provider package.")
		}
	}
}

// purgeKeys removes stored objects, best-effort.
func (s *DefaultMirrorService) purgeKeys(keys map[string]string) {
	for _, key := range keys {
		if err := s.Resolver.Purge(key); err != nil {
			log.Warn().
				Str("location", key).
				Err(err).
				Msg("Could not purge mirrored provider package.")
		}
	}
}

// verifyArchives matches each archive against the metadata entry with the same
// file name, checks its content against the declared h1 hash and returns the
// resulting platforms in the same order as the archives.
func verifyArchives(metadata mirror.ArchivesDTO, archives []file.File) ([]mirror.Platform, error) {
	platforms := make([]mirror.Platform, 0, len(archives))

	for _, archive := range archives {
		key, entry, ok := findArchiveEntry(metadata, archive.Name())
		if !ok {
			return nil, fmt.Errorf("archive %s is not listed in the version metadata", archive.Name())
		}

		platform, err := entry.ToPlatform(key)
		if err != nil {
			return nil, err
		}

		if !slices.ContainsFunc(entry.Hashes, func(h string) bool { return strings.HasPrefix(h, h1HashPrefix) }) {
			return nil, fmt.Errorf("archive %s has no h1 hash in the version metadata", archive.Name())
		}

		hash, err := computeH1(archive)
		if err != nil {
			return nil, fmt.Errorf("could not hash archive %s: %v", archive.Name(), err)
		}

		if !slices.Contains(entry.Hashes, hash) {
			return nil, fmt.Errorf("archive %s does not match its declared h1 hash", archive.Name())
		}

		platforms = append(platforms, platform)
	}

	return platforms, nil
}

// findArchiveEntry looks up the metadata entry whose url is the given file name.
func findArchiveEntry(metadata mirror.ArchivesDTO, fileName string) (string, mirror.ArchiveDTO, bool) {
	for key, entry := range metadata.Archives {
		if entry.URL == fileName {
			return key, entry, true
		}
	}

	return "", mirror.ArchiveDTO{}, false
}

// computeH1 computes the h1 hash of a zip archive, the same way Terraform does
// for its dependency lock file.
func computeH1(archive file.File) (string, error) {
	if _, err := archive.Seek(0, 0); err != nil {
		return "", err
	}

	tmp, err := file.SaveToTemp(archive)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tmp.Remove()
	}()

	if _, err := archive.Seek(0, 0); err != nil {
		return "", err
	}

	return dirhash.HashZip(tmp.Path(), dirhash.DefaultHash)
}

// validateIdentity checks the hostname, namespace and name of a mirrored provider.
func validateIdentity(hostname, namespace, name string) error {
	if !hostnameRegexp.MatchString(hostname) {
		return fmt.Errorf("invalid hostname %q", hostname)
	}

	if !identifierRegexp.MatchString(namespace) {
		return fmt.Errorf("invalid namespace %q", namespace)
	}

	if !identifierRegexp.MatchString(name) {
		return fmt.Errorf("invalid provider name %q", name)
	}

	return nil
}

// authorityLabel builds the metrics label identifying the owner of a mirrored
// provider.
func authorityLabel(hostname, namespace string) string {
	return fmt.Sprintf("%s/%s", hostname, namespace)
}
