package services

import (
	"fmt"

	"terralist/internal/server/models/module"
	"terralist/internal/server/repositories"
	"terralist/pkg/file"
	"terralist/pkg/storage"
	"terralist/pkg/version"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ModuleService describes a service that holds the business logic for modules registry
type ModuleService interface {
	// Get returns a specific module
	Get(namespace, name, provider string) (*module.ListResponseDTO, error)

	// GetVersion returns a public URL from which a specific a module version can be
	// downloaded
	GetVersion(namespace, name, provider, version string) (*string, error)

	// Upload loads a new module version to the system
	// If the module does not exist, it will be created
	Upload(dto *module.CreateDTO, url string) error

	// Delete removes a module with all its data from the system
	Delete(authorityID uuid.UUID, name string, provider string) error

	// DeleteVersion removes a module version from the system
	// If the version removed is the only module version available, the entire
	// module will be removed
	DeleteVersion(authorityID uuid.UUID, name string, provider string, version string) error
}

// DefaultModuleService is the concrete implementation of ModuleService
type DefaultModuleService struct {
	ModuleRepository repositories.ModuleRepository
	AuthorityService AuthorityService
	Resolver         storage.Resolver
	Fetcher          file.Fetcher
}

func (s *DefaultModuleService) Get(namespace, name, provider string) (*module.ListResponseDTO, error) {
	m, err := s.ModuleRepository.Find(namespace, name, provider)
	if err != nil {
		return nil, err
	}

	dto := m.ToListResponseDTO()
	return &dto, nil
}

func (s *DefaultModuleService) GetVersion(namespace, name, provider, version string) (*string, error) {
	location, err := s.ModuleRepository.FindVersionLocation(namespace, name, provider, version)
	if err != nil {
		return nil, err
	}

	if s.Resolver != nil {
		url, err := s.Resolver.Find(*location)
		if err != nil {
			return nil, fmt.Errorf("could not resolve location: %v", err)
		}

		return &url, nil
	}

	return location, nil
}

func (s *DefaultModuleService) Upload(d *module.CreateDTO, url string) error {
	// Validate version
	if semVer := version.Version(d.Version); !semVer.Valid() {
		return fmt.Errorf("version should respect the semantic versioning standard (semver.org)")
	}

	// Map the DTO
	m := d.ToModule()

	// Find the authority
	a, err := s.AuthorityService.GetByID(m.AuthorityID)
	if err != nil {
		return err
	}

	// Check if the module already exists and has this version
	current, err := s.ModuleRepository.Find(a.Name, m.Name, m.Provider)
	if err == nil {
		if current.GetVersion(d.Version) != nil {
			return fmt.Errorf("version %s already exists", d.Version)
		}
	}

	if s.Resolver != nil {
		// Download module files
		archive, err := s.Fetcher.Fetch(d.Version, url)
		if err != nil {
			return err
		}

		// Upload the module archive to the resolver datastore
		location, err := s.Resolver.Store(&storage.StoreInput{
			Content: archive.Content,
			KeyPrefix: fmt.Sprintf(
				"modules/%s/%s/%s",
				a.Name,
				m.Name,
				m.Provider,
			),
			FileName: archive.Name,
		})
		if err != nil {
			return fmt.Errorf("could store the new version: %v", err)
		}

		// Update the module location
		m.Versions[0].Location = location
	}

	// Only add the new version if the module already exists
	var toUpload *module.Module
	if current != nil {
		current.Versions = append(current.Versions, m.Versions[0])

		toUpload = current
	} else {
		toUpload = &m
	}

	if _, err := s.ModuleRepository.Upsert(*toUpload); err != nil {
		return err
	}

	return nil
}

func (s *DefaultModuleService) Delete(authorityID uuid.UUID, name string, provider string) error {
	a, err := s.AuthorityService.GetByID(authorityID)
	if err != nil {
		return err
	}

	m, err := s.ModuleRepository.Find(a.Name, name, provider)
	if err != nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", a.Name, name, provider)
	}

	if s.Resolver != nil {
		for _, ver := range m.Versions {
			s.deleteVersion(&ver)
		}
	}

	if err := s.ModuleRepository.Delete(m); err != nil {
		return err
	}

	return nil
}

func (s *DefaultModuleService) DeleteVersion(authorityID uuid.UUID, name string, provider string, version string) error {
	a, err := s.AuthorityService.GetByID(authorityID)
	if err != nil {
		return err
	}

	m, err := s.ModuleRepository.Find(a.Name, name, provider)
	if err != nil {
		return fmt.Errorf("module %s/%s/%s is not uploaded to this registry", a.Name, name, provider)
	}

	v := m.GetVersion(version)
	if v == nil {
		return fmt.Errorf("module %s/%s/%s does not contain version %s", a.Name, name, provider, version)
	}

	if s.Resolver != nil {
		s.deleteVersion(v)
	}

	if len(m.Versions) == 1 {
		return s.ModuleRepository.Delete(m)
	}

	return s.ModuleRepository.DeleteVersion(v)
}

// deleteVersion removes the files for a specific module version
func (s *DefaultModuleService) deleteVersion(v *module.Version) {
	if err := s.Resolver.Purge(v.Location); err != nil {
		log.Warn().
			AnErr("Error", err).
			Str("Module", v.Module.String()).
			Str("Version", v.Version).
			Str("Key", v.Location).
			Msg("Could not purge, require manual clean-up")
	}
}
