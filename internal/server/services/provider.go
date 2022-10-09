package services

import (
	"fmt"

	"terralist/internal/server/models/provider"
	"terralist/internal/server/repositories"
	"terralist/pkg/file"
	"terralist/pkg/storage"
	"terralist/pkg/version"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	shaSumsKey    = "shaSums"
	shaSumsSigKey = "shaSumsSig"
)

// ProviderService describes a service that holds the business logic for providers registry
type ProviderService interface {
	// Get returns a specific provider
	Get(namespace, name string) (*provider.VersionListProviderDTO, error)

	// GetVersion returns a specific installation for a provider
	GetVersion(namespace, name, version, system, architecture string) (*provider.DownloadPlatformDTO, error)

	// Upload loads a new provider version into the system
	// If the provider does not already exist, it will create a new one
	Upload(*provider.CreateProviderDTO) error

	// Delete removes a provider from the system with all its data (versions)
	Delete(authorityID uuid.UUID, name string) error

	// DeleteVersion removes a specific version from the system with all its data (installations)
	// If the removed version is the only version available in the system, the entire
	// provider will be removed
	DeleteVersion(authorityID uuid.UUID, name string, version string) error
}

// DefaultProviderService is the concrete implementation of ProviderService
type DefaultProviderService struct {
	ProviderRepository repositories.ProviderRepository
	AuthorityService   AuthorityService
	Resolver           storage.Resolver
}

func (s *DefaultProviderService) Get(namespace, name string) (*provider.VersionListProviderDTO, error) {
	// Find the provider
	p, err := s.ProviderRepository.Find(namespace, name)

	if err != nil {
		return nil, fmt.Errorf("requested provider was not found: %v", err)
	}

	// Map to response DTO
	dto := p.ToVersionListProviderDTO()

	return &dto, nil
}

func (s *DefaultProviderService) GetVersion(namespace, name, version, system, architecture string) (*provider.DownloadPlatformDTO, error) {
	p, err := s.ProviderRepository.FindVersionPlatform(namespace, name, version, system, architecture)
	if err != nil {
		return nil, err
	}

	a, err := s.AuthorityService.Get(p.Version.Provider.AuthorityID)
	if err != nil {
		return nil, fmt.Errorf("could not find authority: %v", err)
	}

	keys := []provider.PublicKeyDTO{}

	for _, k := range a.Keys {
		keys = append(keys, provider.PublicKeyDTO{
			KeyId:          k.KeyId,
			AsciiArmor:     k.AsciiArmor,
			TrustSignature: k.TrustSignature,
			Source:         a.Name,
			SourceURL:      a.PolicyURL,
		})
	}

	dto := p.ToDownloadPlatformDTO(provider.SigningKeysDTO{Keys: keys})

	if s.Resolver != nil {
		if err := s.resolveLocations(&dto); err != nil {
			return nil, err
		}
	}

	return &dto, nil
}

func (s *DefaultProviderService) Upload(d *provider.CreateProviderDTO) error {
	// Validate version
	if semVer := version.Version(d.Version); !semVer.Valid() {
		return fmt.Errorf("version should respect the semantic versioning standard (semver.org)")
	}

	// Map the DTO
	p := d.ToProvider()

	// Find the authority
	a, err := s.AuthorityService.Get(p.AuthorityID)
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
		files, err := s.downloadFiles(d)
		if err != nil {
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

	return nil
}

func (s *DefaultProviderService) Delete(authorityID uuid.UUID, name string) error {
	a, err := s.AuthorityService.Get(authorityID)
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

	return nil
}

func (s *DefaultProviderService) DeleteVersion(authorityID uuid.UUID, name string, version string) error {
	a, err := s.AuthorityService.Get(authorityID)
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
		s.deleteVersion(v)
	}

	if err := s.ProviderRepository.DeleteVersion(p, version); err != nil {
		return err
	}

	return nil
}

// resolveLocations resolves the keys for a provider platform
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

// downloadFiles fetches all provider files
func (s *DefaultProviderService) downloadFiles(d *provider.CreateProviderDTO) (map[string]*file.InMemoryFile, error) {
	prefix := fmt.Sprintf("terraform-provider-%s_%s", d.Name, d.Version)

	// Download provider files
	shaSums, err := file.FetchFile(fmt.Sprintf("%s_SHA256SUMS", prefix), d.ShaSums.URL)
	if err != nil {
		return nil, fmt.Errorf("could not fetch shaSums file: %v", err)
	}

	shaSumsSig, err := file.FetchFile(fmt.Sprintf("%s_SHA256SUMS.sig", prefix), d.ShaSums.SignatureURL)
	if err != nil {
		return nil, fmt.Errorf("could not fetch shaSums sig file: %v", err)
	}

	files := map[string]*file.InMemoryFile{
		shaSumsKey:    shaSums,
		shaSumsSigKey: shaSumsSig,
	}

	for _, platform := range d.Platforms {
		p := platform.ToPlatform()
		osArch := p.String()

		binary, err := file.FetchFileChecksum(
			fmt.Sprintf("%s_%s.zip", prefix, osArch),
			p.Location,
			p.ShaSum,
		)
		if err != nil {
			return nil, fmt.Errorf("could not fetch %s file: %v", osArch, err)
		}

		files[osArch] = binary
	}

	return files, nil
}

// uploadFiles uploads all stored provider files
func (s *DefaultProviderService) uploadFiles(
	namespace, name, version string,
	files map[string]*file.InMemoryFile,
) (map[string]string, error) {
	keys := map[string]string{}

	prefix := fmt.Sprintf("providers/%s/%s/%s", namespace, name, version)

	for k, v := range files {
		key, err := s.Resolver.Store(&storage.StoreInput{
			Content:   v.Content,
			KeyPrefix: prefix,
			FileName:  v.Name,
		})
		if err != nil {
			return nil, fmt.Errorf("could not upload %s: %v", v.Name, err)
		}

		keys[k] = key
	}

	return keys, nil
}

// deleteVersion removes all provider files for a specific version
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
