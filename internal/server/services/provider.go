package services

import (
	"fmt"

	"terralist/internal/server/models/provider"
	"terralist/internal/server/repositories"
	"terralist/pkg/version"

	"github.com/google/uuid"
)

// ProviderService describes a service that holds the business logic for providers registry
type ProviderService interface {
	// Get returns a specific provider
	Get(namespace string, name string) (*provider.VersionListProviderDTO, error)

	// GetVersion returns a specific installation for a provider
	GetVersion(
		namespace string,
		name string,
		version string,
		system string,
		architecture string,
	) (*provider.DownloadVersionDTO, error)

	// Upload loads a new provider version into the system
	// If the provider does not already exist, it will create a new one
	Upload(*provider.CreateProviderDTO) error

	// Delete removes a provider from the system with all its data (versions)
	Delete(authorityID uuid.UUID, namespace string, name string) error

	// DeleteVersion removes a specific version from the system with all its data (installations)
	// If the removed version is the only version available in the system, the entire
	// provider will be removed
	DeleteVersion(authorityID uuid.UUID, namespace string, name string, version string) error
}

// DefaultProviderService is the concrete implementation of ProviderService
type DefaultProviderService struct {
	ProviderRepository repositories.ProviderRepository
}

func (s *DefaultProviderService) Get(namespace string, name string) (*provider.VersionListProviderDTO, error) {
	// Find the provider
	p, err := s.ProviderRepository.Find(namespace, name)

	if err != nil {
		return nil, fmt.Errorf("requested provider was not found: %v", err)
	}

	// Map to response DTO
	dto := p.ToVersionListProviderDTO()

	return &dto, nil
}

func (s *DefaultProviderService) GetVersion(
	namespace string,
	name string,
	version string,
	system string,
	architecture string,
) (*provider.DownloadVersionDTO, error) {
	p, err := s.ProviderRepository.FindVersion(namespace, name, version)
	if err != nil {
		return nil, err
	}

	dto, err := p.ToDownloadVersionDTO(system, architecture)
	if err != nil {
		return nil, err
	}

	return &dto, nil
}

func (s *DefaultProviderService) Upload(d *provider.CreateProviderDTO) error {
	if semVer := version.Version(d.Version); !semVer.Valid() {
		return fmt.Errorf("version should respect the semantic versioning standard (semver.org)")
	}

	p := d.ToProvider()
	if _, err := s.ProviderRepository.Upsert(p); err != nil {
		return err
	}

	return nil
}

func (s *DefaultProviderService) Delete(authorityID uuid.UUID, namespace string, name string) error {
	p, err := s.ProviderRepository.Find(namespace, name)
	if err != nil {
		return err
	}

	if p.Authority.ID != authorityID {
		return fmt.Errorf("authority does not match")
	}

	if err := s.ProviderRepository.Delete(p); err != nil {
		return err
	}

	return nil
}

func (s *DefaultProviderService) DeleteVersion(
	authorityID uuid.UUID,
	namespace string,
	name string,
	version string,
) error {
	p, err := s.ProviderRepository.Find(namespace, name)
	if err != nil {
		return err
	}

	if p.Authority.ID != authorityID {
		return fmt.Errorf("authority does not match")
	}

	if err := s.ProviderRepository.DeleteVersion(p, version); err != nil {
		return err
	}

	return nil
}
