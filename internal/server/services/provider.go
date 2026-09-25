package services

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/repositories"
	"terralist/pkg/file"
	"terralist/pkg/metrics"
	"terralist/pkg/registry"
	"terralist/pkg/storage"
	"terralist/pkg/version"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/sync/singleflight"
)

const (
	shaSumsKey    = "shaSums"
	shaSumsSigKey = "shaSumsSig"
)

// ProviderService describes a service that holds the business logic for providers registry.
type ProviderService interface {
	// Get returns a specific provider, with the versions the provider registry
	// protocol can serve. With withUpstream, the versions the authority's
	// upstream registry offers are merged in.
	Get(namespace, name string, withUpstream bool) (*provider.VersionListProviderDTO, error)

	// ListVersions returns every version of a provider, including the ones
	// served through the network mirror only.
	ListVersions(namespace, name string) ([]string, error)

	// GetVersion returns a specific installation for a provider. With
	// withUpstream, a platform the authority's upstream registry offers but
	// Terralist does not hold yet points its download at the mirror.
	GetVersion(namespace, name, version, system, architecture string, withUpstream bool) (*provider.DownloadPlatformDTO, error)

	// ListMirrorVersions returns the versions of a provider as the network
	// mirror protocol lists them, merged with the upstream ones when asked.
	ListMirrorVersions(namespace, name string, withUpstream bool) (*provider.MirrorVersionListDTO, error)

	// ListMirrorArchives returns the installation packages of a provider
	// version as the network mirror protocol lists them, with resolved
	// download locations. Upstream platforms not stored yet are listed with a
	// location relative to the document, served by Download.
	ListMirrorArchives(namespace, name, version string, withUpstream bool) (*provider.MirrorArchivesDTO, error)

	// Download resolves the download URL of a platform package, fetching it
	// from the upstream registry first when it is not stored and allowFetch
	// is set.
	Download(namespace, name, version, system, architecture string, allowFetch bool) (string, error)

	// Fetch downloads the given os_arch platforms of a version from the
	// upstream registry into storage and reports each outcome.
	Fetch(namespace, name, version string, platforms []string) []provider.FetchResultDTO

	// Upload loads a new provider version into the system.
	// If the provider does not already exist, it will create a new one.
	Upload(*provider.CreateProviderDTO) error

	// UploadPackages stores the packages of a new provider version uploaded
	// directly, as produced by `terraform providers mirror`. Without the
	// SHA256SUMS file and its signature the version is served through the
	// network mirror only.
	UploadPackages(*provider.PackagesUploadDTO) error

	// Delete removes a provider from the system with all its data (versions).
	Delete(authorityID uuid.UUID, name string) error

	// DeleteVersion removes a specific version from the system with all its data (installations).
	// If the removed version is the only version available in the system, the entire
	// provider will be removed.
	DeleteVersion(authorityID uuid.UUID, name string, version string) error
}

// DefaultProviderService is the concrete implementation of ProviderService.
type DefaultProviderService struct {
	ProviderRepository repositories.ProviderRepository
	AuthorityService   AuthorityService
	Resolver           storage.Resolver
	Fetcher            file.Fetcher
	Upstream           UpstreamService

	// MirrorBaseURL is the network mirror base of this Terralist instance
	// under its own hostname, where upstream packages are downloaded from.
	MirrorBaseURL string

	fetches singleflight.Group
}

func (s *DefaultProviderService) Get(namespace, name string, withUpstream bool) (*provider.VersionListProviderDTO, error) {
	dto := provider.VersionListProviderDTO{}
	local := map[string]struct{}{}

	p, err := s.ProviderRepository.Find(namespace, name)
	if err == nil {
		dto = p.ToVersionListProviderDTO()
		local = localVersions(p)
	}

	upstream, upstreamErr := s.upstreamVersions(s.upstreamAuthority(namespace, withUpstream), name)
	if err != nil && len(upstream) == 0 {
		if upstreamErr != nil {
			return nil, upstreamErr
		}

		return nil, fmt.Errorf("requested provider was not found: %v", err)
	}

	mergeUpstreamVersions(&dto, local, upstream)

	// Record list operation
	metrics.RecordRequest(namespace, "list")

	return &dto, nil
}

func (s *DefaultProviderService) ListVersions(namespace, name string) ([]string, error) {
	p, err := s.ProviderRepository.Find(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("requested provider was not found: %v", err)
	}

	versions := make([]string, 0, len(p.Versions))
	for _, v := range p.Versions {
		versions = append(versions, v.Version)
	}

	return versions, nil
}

func (s *DefaultProviderService) GetVersion(namespace, name, version, system, architecture string, withUpstream bool) (*provider.DownloadPlatformDTO, error) {
	p, err := s.ProviderRepository.FindVersionPlatform(namespace, name, version, system, architecture)
	if errors.Is(err, repositories.ErrNotFound) {
		if a := s.upstreamAuthority(namespace, withUpstream); a != nil {
			return s.upstreamVersionDownload(a, name, version, system, architecture)
		}
	}
	if err != nil {
		return nil, err
	}

	if p.Version.MirrorOnly() {
		return nil, fmt.Errorf("version %s of provider %s/%s is served through the network mirror only", version, namespace, name)
	}

	a, err := s.AuthorityService.GetByID(p.Version.Provider.AuthorityID)
	if err != nil {
		return nil, fmt.Errorf("could not find authority: %v", err)
	}

	keys, ok := p.Version.SigningKeysDTO()
	if !ok {
		keys = s.authorityKeys(a)
	}

	dto := p.ToDownloadPlatformDTO(keys)

	if s.Resolver != nil {
		if err := s.resolveLocations(&dto); err != nil {
			return nil, err
		}
	}

	// Record download metrics
	metrics.RecordRequest(namespace, "download")
	metrics.RecordArtifactDownload("provider", namespace)

	return &dto, nil
}

func (s *DefaultProviderService) ListMirrorVersions(namespace, name string, withUpstream bool) (*provider.MirrorVersionListDTO, error) {
	dto := provider.MirrorVersionListDTO{Versions: map[string]struct{}{}}

	p, err := s.ProviderRepository.Find(namespace, name)
	if err == nil {
		dto = p.ToMirrorVersionListDTO()
	}

	upstream, upstreamErr := s.upstreamVersions(s.upstreamAuthority(namespace, withUpstream), name)
	if err != nil && len(upstream) == 0 {
		if upstreamErr != nil {
			return nil, upstreamErr
		}

		return nil, fmt.Errorf("requested provider was not found: %v", err)
	}

	for _, v := range upstream {
		dto.Versions[v.Version] = struct{}{}
	}

	metrics.RecordRequest(namespace, "list")

	return &dto, nil
}

func (s *DefaultProviderService) ListMirrorArchives(namespace, name, version string, withUpstream bool) (*provider.MirrorArchivesDTO, error) {
	dto := provider.MirrorArchivesDTO{Archives: map[string]provider.MirrorArchiveDTO{}}

	// A version uploaded by an operator is never completed from the upstream.
	uploaded := false

	p, err := s.ProviderRepository.Find(namespace, name)
	if err == nil {
		if v := p.GetVersion(version); v != nil {
			dto = v.ToMirrorArchivesDTO()
			uploaded = v.Origin != provider.OriginUpstream
		}
	}

	if s.Resolver != nil {
		for key, archive := range dto.Archives {
			url, err := s.Resolver.Find(archive.URL)
			if err != nil {
				return nil, fmt.Errorf("could not resolve package location for %s: %v", key, err)
			}

			archive.URL = url
			dto.Archives[key] = archive
		}
	}

	if a := s.upstreamAuthority(namespace, withUpstream && !uploaded); a != nil {
		upstream, err := s.upstreamVersions(a, name)
		if err != nil && len(dto.Archives) == 0 {
			return nil, err
		}

		if lo.ContainsBy(upstream, func(v UpstreamVersion) bool { return v.Version == version }) {
			metadata, err := s.Upstream.ProviderVersion(a, name, version)
			if err != nil {
				metrics.RecordError("upstream", "error")
				log.Warn().Err(err).Str("authority", a.Name).Str("provider", name).Str("version", version).Msg("Could not read the upstream version.")

				if len(dto.Archives) == 0 {
					return nil, upstreamFailure(err)
				}
			}

			mergeUpstreamArchives(&dto, name, version, upstream, metadata)
		}
	}

	if len(dto.Archives) == 0 {
		return nil, fmt.Errorf("provider %s/%s does not contain version %s", namespace, name, version)
	}

	metrics.RecordRequest(namespace, "download")
	metrics.RecordArtifactDownload("provider", namespace)

	return &dto, nil
}

// upstreamVersionDownload answers a registry download request for a platform
// Terralist does not hold yet, creating the version from the upstream first.
func (s *DefaultProviderService) upstreamVersionDownload(a *authority.Authority, name, version, system, architecture string) (*provider.DownloadPlatformDTO, error) {
	v, err := s.ensureUpstreamVersion(a, name, version)
	if err != nil {
		return nil, err
	}

	dto, err := s.upstreamDownloadDTO(a, v, name, system, architecture)
	if err != nil {
		return nil, err
	}

	metrics.RecordRequest(a.Name, "download")
	metrics.RecordArtifactDownload("provider", a.Name)

	return dto, nil
}

func (s *DefaultProviderService) Upload(d *provider.CreateProviderDTO) error {
	// Validate version
	if semVer := version.Version(d.Version); !semVer.Valid() {
		return fmt.Errorf("version should respect the semantic versioning standard (semver.org)")
	}

	// Map the DTO
	p := d.ToProvider()

	// Find the authority
	a, err := s.AuthorityService.GetByID(p.AuthorityID)
	if err != nil {
		return err
	}

	// Check if the provider already exists and has this version
	current, err := s.ProviderRepository.Find(a.Name, p.Name)
	if err == nil {
		if current.GetVersion(d.Version) != nil {
			return fmt.Errorf("version %s already exists", d.Version)
		}
	}

	if s.Resolver != nil {
		// Download provider files
		files, cleanup, err := s.downloadFiles(d)
		if err != nil {
			return err
		}
		defer cleanup()

		if err := verifySignature(a, files[shaSumsKey], files[shaSumsSigKey]); err != nil {
			return err
		}

		// Upload provider files
		keys, err := s.uploadFiles(a.Name, p.Name, d.Version, files)
		if err != nil {
			return err
		}

		// Update provider locations
		p.Versions[0].ShaSumsUrl = keys[shaSumsKey]
		p.Versions[0].ShaSumsSignatureUrl = keys[shaSumsSigKey]

		for i, platform := range p.Versions[0].Platforms {
			p.Versions[0].Platforms[i].Location = keys[platform.String()]
		}
	}

	// Only add the new version if the provider already exists
	var toUpload *provider.Provider
	if current != nil {
		current.Versions = append(current.Versions, p.Versions[0])

		toUpload = current
	} else {
		toUpload = &p
	}

	if _, err := s.ProviderRepository.Upsert(*toUpload); err != nil {
		return err
	}

	// Record artifact upload metric
	metrics.RecordArtifactUpload("provider", a.Name)
	// Record upload request
	metrics.RecordRequest(a.Name, "upload")

	return nil
}

func (s *DefaultProviderService) UploadPackages(d *provider.PackagesUploadDTO) error {
	if semVer := version.Version(d.Version); !semVer.Valid() {
		return fmt.Errorf("version should respect the semantic versioning standard (semver.org)")
	}

	if s.Resolver == nil {
		return fmt.Errorf("uploading packages requires a providers storage resolver")
	}

	if len(d.Archives) == 0 {
		return fmt.Errorf("at least one provider package is required")
	}

	if (d.ShaSums == nil) != (d.ShaSumsSignature == nil) {
		return fmt.Errorf("the SHA256SUMS file and its signature must be uploaded together")
	}

	if d.Signed() && len(d.Protocols) == 0 {
		return fmt.Errorf("the provider protocols are required to serve a signed version through the registry")
	}

	a, err := s.AuthorityService.GetByID(d.AuthorityID)
	if err != nil {
		return err
	}

	current, err := s.ProviderRepository.Find(a.Name, d.Name)
	if err == nil && current.GetVersion(d.Version) != nil {
		return fmt.Errorf("version %s already exists", d.Version)
	}

	v := d.ToVersion()
	files := map[string]file.File{}

	if d.Signed() {
		if err := verifySignature(a, d.ShaSums, d.ShaSumsSignature); err != nil {
			return err
		}

		files[shaSumsKey] = d.ShaSums
		files[shaSumsSigKey] = d.ShaSumsSignature
	}

	v.Platforms, err = s.verifyPackages(d)
	if err != nil {
		return err
	}

	for i, archive := range d.Archives {
		files[v.Platforms[i].String()] = archive
	}

	keys, err := s.uploadFiles(a.Name, d.Name, d.Version, files)
	if err != nil {
		return err
	}

	v.ShaSumsUrl = keys[shaSumsKey]
	v.ShaSumsSignatureUrl = keys[shaSumsSigKey]
	for i := range v.Platforms {
		v.Platforms[i].Location = keys[v.Platforms[i].String()]
	}

	toUpload := current
	if toUpload == nil {
		toUpload = &provider.Provider{
			AuthorityID: d.AuthorityID,
			Name:        d.Name,
		}
	}
	toUpload.Versions = append(toUpload.Versions, v)

	if _, err := s.ProviderRepository.Upsert(*toUpload); err != nil {
		return err
	}

	metrics.RecordArtifactUpload("provider", a.Name)
	metrics.RecordRequest(a.Name, "upload")

	return nil
}

// verifyPackages matches every uploaded archive against the version document,
// computes its sha256 and, when a SHA256SUMS file is uploaded, checks the
// archive against it. It returns the platforms in the order of the archives.
func (s *DefaultProviderService) verifyPackages(d *provider.PackagesUploadDTO) ([]provider.Platform, error) {
	var sums map[string]string
	if d.Signed() {
		var err error
		if sums, err = parseShaSums(d.ShaSums); err != nil {
			return nil, err
		}
	}

	platforms := make([]provider.Platform, 0, len(d.Archives))
	for _, archive := range d.Archives {
		key, entry, ok := d.Metadata.FindArchive(archive.Name())
		if !ok {
			return nil, fmt.Errorf("package %s is not listed in the version document", archive.Name())
		}

		platform, err := entry.ToPlatform(key)
		if err != nil {
			return nil, err
		}

		if platform.ShaSum, err = sha256Sum(archive); err != nil {
			return nil, fmt.Errorf("could not hash package %s: %v", archive.Name(), err)
		}

		if sums != nil {
			expected, ok := sums[archive.Name()]
			if !ok {
				return nil, fmt.Errorf("package %s is not listed in the SHA256SUMS file", archive.Name())
			}

			if expected != platform.ShaSum {
				return nil, fmt.Errorf("package %s does not match its SHA256SUMS entry", archive.Name())
			}
		}

		platforms = append(platforms, platform)
	}

	return platforms, nil
}

// parseShaSums reads a SHA256SUMS file into a map from file name to hex digest
// and rewinds the file afterwards.
// verifySignature checks that the SHA256SUMS document is signed by a key of
// the authority, which the registry serves the version with. Expired keys are
// accepted, as they sign releases that cannot be signed again, such as the
// HashiCorp providers.
func verifySignature(a *authority.Authority, shaSums, signature file.File) error {
	document, err := readAndRewind(shaSums)
	if err != nil {
		return fmt.Errorf("could not read the SHA256SUMS file: %v", err)
	}

	sig, err := readAndRewind(signature)
	if err != nil {
		return fmt.Errorf("could not read the SHA256SUMS signature: %v", err)
	}

	keys := lo.Map(a.Keys, func(k authority.Key, _ int) registry.GPGPublicKey {
		return registry.GPGPublicKey{KeyID: k.KeyId, ASCIIArmor: k.AsciiArmor}
	})

	if _, err := (registry.SignatureVerifier{AcceptExpiredKeys: true}).VerifyShaSums(document, sig, keys); err != nil {
		return fmt.Errorf("the SHA256SUMS signature does not verify with the keys of authority %s: %w", a.Name, err)
	}

	return nil
}

// readAndRewind reads a file to the end and rewinds it for the next reader.
func readAndRewind(f file.File) ([]byte, error) {
	defer func() {
		_, _ = f.Seek(0, io.SeekStart)
	}()

	return io.ReadAll(f)
}

func parseShaSums(f file.File) (map[string]string, error) {
	defer func() {
		_, _ = f.Seek(0, io.SeekStart)
	}()

	sums := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}

		sums[strings.TrimPrefix(fields[1], "*")] = fields[0]
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("could not read the SHA256SUMS file: %v", err)
	}

	return sums, nil
}

// sha256Sum computes the hex sha256 digest of a file and rewinds it afterwards.
func sha256Sum(f file.File) (string, error) {
	defer func() {
		_, _ = f.Seek(0, io.SeekStart)
	}()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (s *DefaultProviderService) Delete(authorityID uuid.UUID, name string) error {
	a, err := s.AuthorityService.GetByID(authorityID)
	if err != nil {
		return err
	}

	p, err := s.ProviderRepository.Find(a.Name, name)
	if err != nil {
		return err
	}

	if p.AuthorityID != authorityID {
		return fmt.Errorf("authority does not match")
	}

	if s.Resolver != nil {
		for _, ver := range p.Versions {
			s.deleteVersion(&ver)
		}
	}

	if err := s.ProviderRepository.Delete(p); err != nil {
		return err
	}

	// Record artifact deletion metrics for all versions
	for range p.Versions {
		metrics.RecordArtifactDeletion("provider", a.Name)
	}

	return nil
}

func (s *DefaultProviderService) DeleteVersion(authorityID uuid.UUID, name string, version string) error {
	a, err := s.AuthorityService.GetByID(authorityID)
	if err != nil {
		return err
	}

	p, err := s.ProviderRepository.Find(a.Name, name)
	if err != nil {
		return err
	}

	if p.AuthorityID != authorityID {
		return fmt.Errorf("authority does not match")
	}

	if s.Resolver != nil {
		v := p.GetVersion(version)
		if v == nil {
			return fmt.Errorf("provider %s/%s does not contain version %s", a.Name, name, version)
		}

		s.deleteVersion(v)
	}

	if err := s.ProviderRepository.DeleteVersion(p, version); err != nil {
		return err
	}

	// Record artifact deletion metric
	metrics.RecordArtifactDeletion("provider", a.Name)

	return nil
}

// resolveLocations resolves the keys for a provider platform.
func (s *DefaultProviderService) resolveLocations(d *provider.DownloadPlatformDTO) error {
	var err error

	d.ShaSumsUrl, err = s.Resolver.Find(d.ShaSumsUrl)
	if err != nil {
		return fmt.Errorf("could not resolve shasums location: %v", err)
	}

	d.ShaSumsSignatureUrl, err = s.Resolver.Find(d.ShaSumsSignatureUrl)
	if err != nil {
		return fmt.Errorf("could not resolve shasums signature location: %v", err)
	}

	d.DownloadUrl, err = s.Resolver.Find(d.DownloadUrl)
	if err != nil {
		return fmt.Errorf("could not resolve binary location: %v", err)
	}

	return err
}

// downloadFiles fetches all provider files and returns a cleanup function
// that removes all temporary directories created during the download.
func (s *DefaultProviderService) downloadFiles(d *provider.CreateProviderDTO) (map[string]file.File, func(), error) {
	prefix := fmt.Sprintf("terraform-provider-%s_%s", d.Name, d.Version)

	headers := file.CreateHeader(d.Headers)

	var cleanups []func()
	cleanupAll := func() {
		for _, fn := range cleanups {
			fn()
		}
	}

	// Download provider files
	shaSums, shaSumsCleanup, err := s.Fetcher.FetchFile(fmt.Sprintf("%s_SHA256SUMS", prefix), d.ShaSums.URL, headers)
	if err != nil {
		cleanupAll()
		return nil, nil, fmt.Errorf("could not fetch shaSums file: %v", err)
	}
	cleanups = append(cleanups, shaSumsCleanup)

	shaSumsSig, shaSumsSigCleanup, err := s.Fetcher.FetchFile(fmt.Sprintf("%s_SHA256SUMS.sig", prefix), d.ShaSums.SignatureURL, headers)
	if err != nil {
		cleanupAll()
		return nil, nil, fmt.Errorf("could not fetch shaSums sig file: %v", err)
	}
	cleanups = append(cleanups, shaSumsSigCleanup)

	files := map[string]file.File{
		shaSumsKey:    shaSums,
		shaSumsSigKey: shaSumsSig,
	}

	for _, platform := range d.Platforms {
		p := platform.ToPlatform()
		osArch := p.String()

		binary, binaryCleanup, err := s.Fetcher.FetchFileChecksum(fmt.Sprintf("%s_%s.zip", prefix, osArch), p.Location, p.ShaSum, headers)
		if err != nil {
			cleanupAll()
			return nil, nil, fmt.Errorf("could not fetch %s file: %v", osArch, err)
		}
		cleanups = append(cleanups, binaryCleanup)

		files[osArch] = binary
	}

	return files, cleanupAll, nil
}

// uploadFiles uploads all stored provider files.
func (s *DefaultProviderService) uploadFiles(
	namespace, name, version string,
	files map[string]file.File,
) (map[string]string, error) {
	keys := map[string]string{}

	prefix := fmt.Sprintf("providers/%s/%s/%s", namespace, name, version)

	for k, v := range files {
		key, err := s.Resolver.Store(&storage.StoreInput{
			Reader:      v,
			Size:        v.Metadata().Size(),
			ContentType: file.ContentType(v),
			KeyPrefix:   prefix,
			FileName:    v.Name(),
		})
		if err != nil {
			return nil, fmt.Errorf("could not upload %s: %v", v.Name(), err)
		}

		keys[k] = key
	}

	return keys, nil
}

// deleteVersion removes all provider files for a specific version.
func (s *DefaultProviderService) deleteVersion(v *provider.Version) {
	for _, plat := range v.Platforms {
		if err := s.Resolver.Purge(plat.Location); err != nil {
			log.Warn().
				AnErr("Error", err).
				Str("Provider", v.Provider.Name).
				Str("Version", v.Version).
				Str("Platform", plat.String()).
				Str("Key", plat.Location).
				Msg("Could not purge, require manual clean-up")

		}
	}

	if err := s.Resolver.Purge(v.ShaSumsUrl); err != nil {
		log.Warn().
			AnErr("Error", err).
			Str("Provider", v.Provider.Name).
			Str("Version", v.Version).
			Str("Key", v.ShaSumsUrl).
			Msg("Could not purge, require manual clean-up")
	}

	if err := s.Resolver.Purge(v.ShaSumsSignatureUrl); err != nil {
		log.Warn().
			AnErr("Error", err).
			Str("Provider", v.Provider.Name).
			Str("Version", v.Version).
			Str("Key", v.ShaSumsSignatureUrl).
			Msg("Could not purge, require manual clean-up")
	}
}
