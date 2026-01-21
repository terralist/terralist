package services

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"terralist/internal/server/models/module"
	"terralist/internal/server/repositories"
	"terralist/pkg/docs"
	"terralist/pkg/file"
	"terralist/pkg/storage"
	"terralist/pkg/version"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ModuleService describes a service that holds the business logic for modules registry.
type ModuleService interface {
	// Get returns a specific module.
	Get(namespace, name, provider string) (*module.ListResponseDTO, error)

	// GetVersion returns a module version.
	GetVersion(namespace, name, provider, version string) (*module.VersionDTO, error)

	// GetSubmoduleDocumentation returns documentation for a specific submodule within a module version.
	GetSubmoduleDocumentation(namespace, name, provider, version, submodulePath string) (string, error)

	// GetVersionURL returns a public URL from which a specific a module version can be
	// downloaded.
	GetVersionURL(namespace, name, provider, version string) (*string, error)

	// Upload loads a new module version to the system.
	// If the module does not exist, it will be created.
	Upload(dto *module.CreateDTO, url string, header http.Header) error

	// Delete removes a module with all its data from the system.
	Delete(authorityID uuid.UUID, name string, provider string) error

	// DeleteVersion removes a module version from the system.
	// If the version removed is the only module version available, the entire
	// module will be removed.
	DeleteVersion(authorityID uuid.UUID, name string, provider string, version string) error
}

// DefaultModuleService is the concrete implementation of ModuleService.
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

func (s *DefaultModuleService) GetVersion(namespace, name, provider, version string) (*module.VersionDTO, error) {
	v, err := s.ModuleRepository.FindVersion(namespace, name, provider, version)
	if err != nil {
		return nil, err
	}

	dto := &module.VersionDTO{Version: v.Version}

	if s.Resolver != nil {
		url, err := s.Resolver.Find(v.Documentation)
		if err != nil {
			log.Warn().
				Str("moduleSlug", fmt.Sprintf("%s/%s/%s/%s", namespace, name, provider, version)).
				Err(err).
				Msg("no documentation for module")

			return dto, nil
		}

		resp, err := http.Get(url)
		if err != nil {
			log.Warn().
				Str("moduleSlug", fmt.Sprintf("%s/%s/%s/%s", namespace, name, provider, version)).
				Str("url", url).
				Err(err).
				Msg("could not fetch module's documentation")

			return dto, nil
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Warn().
				Str("moduleSlug", fmt.Sprintf("%s/%s/%s/%s", namespace, name, provider, version)).
				Str("url", url).
				Err(err).
				Msg("could not read module documentation")

			return dto, nil
		}

		dto.Documentation = string(body)
	}

	return dto, nil
}

func (s *DefaultModuleService) GetSubmoduleDocumentation(namespace, name, provider, version, submodulePath string) (string, error) {
	v, err := s.ModuleRepository.FindVersion(namespace, name, provider, version)
	if err != nil {
		return "", err
	}

	// Check if the submodule exists
	var submoduleExists bool
	for _, sm := range v.Submodules {
		if sm.Path == submodulePath {
			submoduleExists = true
			break
		}
	}

	if !submoduleExists {
		return "", fmt.Errorf("submodule %s not found in module %s/%s/%s/%s", submodulePath, namespace, name, provider, version)
	}

	if s.Resolver == nil {
		return "", fmt.Errorf("no resolver configured to fetch submodule documentation")
	}

	// Construct the documentation file path
	docsFileName := fmt.Sprintf("%s_%s.md", version, strings.ReplaceAll(submodulePath, "/", "_"))
	docsKey := fmt.Sprintf("modules/%s/%s/%s/submodules/%s", namespace, name, provider, docsFileName)

	url, err := s.Resolver.Find(docsKey)
	if err != nil {
		log.Warn().
			Str("moduleSlug", fmt.Sprintf("%s/%s/%s/%s", namespace, name, provider, version)).
			Str("submodulePath", submodulePath).
			Err(err).
			Msg("no documentation for submodule")

		return "", nil
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Warn().
			Str("moduleSlug", fmt.Sprintf("%s/%s/%s/%s", namespace, name, provider, version)).
			Str("submodulePath", submodulePath).
			Str("url", url).
			Err(err).
			Msg("could not fetch submodule's documentation")

		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warn().
			Str("moduleSlug", fmt.Sprintf("%s/%s/%s/%s", namespace, name, provider, version)).
			Str("submodulePath", submodulePath).
			Str("url", url).
			Err(err).
			Msg("could not read submodule documentation")

		return "", err
	}

	return string(body), nil
}

func (s *DefaultModuleService) GetVersionURL(namespace, name, provider, version string) (*string, error) {
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

func (s *DefaultModuleService) Upload(d *module.CreateDTO, url string, header http.Header) error {
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

	// Download module files
	archive, err := s.Fetcher.Fetch(d.Version, url, header)
	if err != nil {
		return err
	}
	defer archive.Close()

	var mdDocs = ""
	var submoduleDocs = make(map[string]string)

	if archiveFile, ok := archive.(*file.ArchiveFile); ok {
		// Generate main module documentation
		markdown, err := docs.GetModuleDocumentation(archiveFile.FS(), "")
		if err != nil {
			log.Warn().
				Str("moduleSlug", fmt.Sprintf("%s/%s/%s", a.Name, m.Name, m.Provider)).
				Err(err).
				Msg("failed to generate module markdown documentation")
		}

		mdDocs = markdown

		// Scan and generate documentation for submodules
		submodules, err := docs.FindSubmodules(archiveFile.FS())
		if err != nil {
			log.Warn().
				Str("moduleSlug", fmt.Sprintf("%s/%s/%s", a.Name, m.Name, m.Provider)).
				Err(err).
				Msg("failed to scan for submodules")
		} else if len(submodules) > 0 {
			log.Info().
				Str("moduleSlug", fmt.Sprintf("%s/%s/%s", a.Name, m.Name, m.Provider)).
				Int("count", len(submodules)).
				Msg("found submodules in module")

			for _, sm := range submodules {
				submoduleDocs[sm.Path] = sm.Documentation
			}
		}
	} else {
		log.Warn().
			Str("moduleSlug", fmt.Sprintf("%s/%s/%s", a.Name, m.Name, m.Provider)).
			Msg("module is not archive, cannot be parsed to extract documentation")
	}

	if s.Resolver != nil {
		// Upload the module archive to the resolver datastore
		location, err := s.Resolver.Store(&storage.StoreInput{
			Reader:      archive,
			Size:        archive.Metadata().Size(),
			ContentType: file.ContentType(archive),
			KeyPrefix: fmt.Sprintf(
				"modules/%s/%s/%s",
				a.Name,
				m.Name,
				m.Provider,
			),
			FileName: archive.Name(),
		})
		if err != nil {
			return fmt.Errorf("could store the new version: %v", err)
		}

		// Update the module location
		m.Versions[0].Location = location

		// Upload the module documentation to the resolver datastore
		docsFile := file.NewInMemoryFile(fmt.Sprintf("%s.md", d.Version), []byte(mdDocs))
		docsLocation, err := s.Resolver.Store(&storage.StoreInput{
			Reader:      docsFile,
			Size:        docsFile.Metadata().Size(),
			ContentType: "text/markdown; charset=utf-8",
			KeyPrefix: fmt.Sprintf(
				"modules/%s/%s/%s",
				a.Name,
				m.Name,
				m.Provider,
			),
			FileName: docsFile.Name(),
		})
		if err != nil {
			return fmt.Errorf("could store the new version's documentation: %v", err)
		}

		// Update the module documentation location
		m.Versions[0].Documentation = docsLocation

		// Upload submodule documentation to the resolver datastore
		for submodulePath, submoduleDoc := range submoduleDocs {
			if submoduleDoc == "" {
				continue
			}

			submoduleDocsFile := file.NewInMemoryFile(
				fmt.Sprintf("%s_%s.md", d.Version, strings.ReplaceAll(submodulePath, "/", "_")),
				[]byte(submoduleDoc),
			)
			submoduleDocsLocation, err := s.Resolver.Store(&storage.StoreInput{
				Reader:      submoduleDocsFile,
				Size:        submoduleDocsFile.Metadata().Size(),
				ContentType: "text/markdown; charset=utf-8",
				KeyPrefix: fmt.Sprintf(
					"modules/%s/%s/%s/submodules",
					a.Name,
					m.Name,
					m.Provider,
				),
				FileName: submoduleDocsFile.Name(),
			})
			if err != nil {
				log.Warn().
					Str("moduleSlug", fmt.Sprintf("%s/%s/%s", a.Name, m.Name, m.Provider)).
					Str("submodulePath", submodulePath).
					Err(err).
					Msg("failed to store submodule documentation")
				continue
			}

			// Find or create the submodule in the version
			found := false
			for i := range m.Versions[0].Submodules {
				if m.Versions[0].Submodules[i].Path == submodulePath {
					found = true
					break
				}
			}

			if !found {
				m.Versions[0].Submodules = append(m.Versions[0].Submodules, module.Submodule{
					Path: submodulePath,
				})
			}

			log.Info().
				Str("moduleSlug", fmt.Sprintf("%s/%s/%s", a.Name, m.Name, m.Provider)).
				Str("submodulePath", submodulePath).
				Str("docsLocation", submoduleDocsLocation).
				Msg("stored submodule documentation")
		}
	} else {
		// Terralist is using a proxy provider.
		m.Versions[0].Location = url

		// The documentation of a module can get pretty large and since we have no place to store it
		// it will end up in the database, increasing the disk size enormously.
		// To avoid this, for now, it's better to not have documentation at all while using the proxy
		// provider.
		// m.Versions[0].Documentation = mdDocs
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
			s.deleteVersion(a.Name, &ver)
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
		s.deleteVersion(a.Name, v)
	}

	if len(m.Versions) == 1 {
		return s.ModuleRepository.Delete(m)
	}

	return s.ModuleRepository.DeleteVersion(v)
}

// deleteVersion removes the files for a specific module version.
func (s *DefaultModuleService) deleteVersion(namespace string, v *module.Version) {
	// Delete the module archive
	if err := s.Resolver.Purge(v.Location); err != nil {
		log.Warn().
			AnErr("Error", err).
			Str("Module", v.Module.String()).
			Str("Version", v.Version).
			Str("Key", v.Location).
			Msg("Could not purge module archive, require manual clean-up")
	}

	// Delete the module documentation
	if v.Documentation != "" {
		if err := s.Resolver.Purge(v.Documentation); err != nil {
			log.Warn().
				AnErr("Error", err).
				Str("Module", v.Module.String()).
				Str("Version", v.Version).
				Str("Key", v.Documentation).
				Msg("Could not purge module documentation, require manual clean-up")
		}
	}

	// Delete documentation for all submodules
	for _, sm := range v.Submodules {
		// Construct the documentation file path using the same convention as Upload
		docsFileName := fmt.Sprintf("%s_%s.md", v.Version, strings.ReplaceAll(sm.Path, "/", "_"))
		docsKey := fmt.Sprintf("modules/%s/%s/%s/submodules/%s",
			namespace,
			v.Module.Name,
			v.Module.Provider,
			docsFileName)

		if err := s.Resolver.Purge(docsKey); err != nil {
			log.Warn().
				AnErr("Error", err).
				Str("Module", v.Module.String()).
				Str("Version", v.Version).
				Str("SubmodulePath", sm.Path).
				Str("Key", docsKey).
				Msg("Could not purge submodule documentation, require manual clean-up")
		}
	}
}
