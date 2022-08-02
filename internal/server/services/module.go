package services

import (
	"fmt"

	"terralist/internal/server/models/module"
	"terralist/internal/server/repositories"
	"terralist/pkg/version"

	"github.com/google/uuid"
)

// ModuleService describes a service that holds the business logic for modules registry
type ModuleService interface {
	// Get returns a specific module
	Get(namespace string, name string, provider string) (*module.ListResponseDTO, error)

	// GetVersion returns a public URL from which a specific a module version can be
	// downloaded
	GetVersion(namespace string, name string, provider string, version string) (*string, error)

	// Upload loads a new module version to the system
	// If the module does not exist, it will be created
	Upload(*module.CreateDTO) error

	// Delete removes a module with all its data from the system
	Delete(authorityID uuid.UUID, name string, provider string) error

	// DeleteVersion removes a module version from the system
	// If the version removed is the only module version available, the entire
	// module will be removed
	DeleteVersion(authorityID uuid.UUID, name string, provider string, version string) error
}

// DefaultModuleService is the concrete implementation of ModuleService
type DefaultModuleService struct {
	ModuleRepository *repositories.DefaultModuleRepository
	AuthorityService AuthorityService
}

func (s *DefaultModuleService) Get(namespace string, name string, provider string) (*module.ListResponseDTO, error) {
	m, err := s.ModuleRepository.Find(namespace, name, provider)
	if err != nil {
		return nil, err
	}

	dto := m.ToListResponseDTO()
	return &dto, nil
}

func (s *DefaultModuleService) GetVersion(
	namespace string,
	name string,
	provider string,
	version string,
) (*string, error) {
	url, err := s.ModuleRepository.FindVersionLocation(namespace, name, provider, version)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (s *DefaultModuleService) Upload(d *module.CreateDTO) error {
	if semVer := version.Version(d.Version); !semVer.Valid() {
		return fmt.Errorf("version should respect the semantic versioning standard (semver.org)")
	}

	m := d.ToModule()

	a, err := s.AuthorityService.Get(m.AuthorityID)
	if err != nil {
		return err
	}

	if _, err := s.ModuleRepository.Upsert(a.Name, m); err != nil {
		return err
	}

	return nil
}

func (s *DefaultModuleService) Delete(authorityID uuid.UUID, name string, provider string) error {
	a, err := s.AuthorityService.Get(authorityID)
	if err != nil {
		return err
	}

	m, err := s.ModuleRepository.Find(a.Name, name, provider)
	if err != nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", a.Name, name, provider)
	}

	if err := s.ModuleRepository.Delete(m); err != nil {
		return err
	}

	return nil
}

func (s *DefaultModuleService) DeleteVersion(authorityID uuid.UUID, name string, provider string, version string) error {
	a, err := s.AuthorityService.Get(authorityID)
	if err != nil {
		return err
	}

	m, err := s.ModuleRepository.Find(a.Name, name, provider)
	if err != nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", a.Name, name, provider)
	}

	if err := s.ModuleRepository.DeleteVersion(m, version); err != nil {
		return err
	}

	return nil
}
