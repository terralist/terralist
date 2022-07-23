package services

import (
	"fmt"
	
	"terralist/internal/server/models/module"
	"terralist/internal/server/repositories"
	"terralist/pkg/version"
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
	Delete(namespace string, name string, provider string) error

	// DeleteVersion removes a module version from the system
	// If the version removed is the only module version available, the entire
	// module will be removed
	DeleteVersion(namespace string, name string, provider string, version string) error
}

// DefaultModuleService is the concrete implementation of ModuleService
type DefaultModuleService struct {
	ModuleRepository *repositories.DefaultModuleRepository
}

func (s *DefaultModuleService) Get(namespace string, name string, provider string) (*module.ListResponseDTO, error) {
	m, err := s.ModuleRepository.Find(namespace, name, provider)
	if err != nil {
		return nil, err
	}

	dto := m.ToListResponseDTO()
	return &dto, nil
}

//
//func (m *DefaultModuleService) Get() func(c *gin.Context) {
//	return func(c *gin.Context) {
//		namespace := c.Param("namespace")
//		name := c.Param("name")
//		provider := c.Param("provider")
//
//		mod, err := m.ModuleRepository.Find(namespace, name, provider)
//
//		if err != nil {
//			c.JSON(http.StatusNotFound, gin.H{
//				"errors": []string{
//					"Requested module was not found",
//					err.Error(),
//				},
//			})
//			return
//		}
//		c.JSON(http.StatusOK, mod.ToListResponseDTO())
//	}
//}

func (s *DefaultModuleService) GetVersion(
	namespace string,
	name string,
	provider string,
	version string,
) (*string, error) {
	v, err := s.ModuleRepository.FindVersion(namespace, name, provider, version)
	if err != nil {
		return nil, err
	}

	return &v.FetchKey, nil
}

//
//
//func (m *DefaultModuleService) GetVersion() func(c *gin.Context) {
//	return func(c *gin.Context) {
//		namespace := c.Param("namespace")
//		name := c.Param("name")
//		provider := c.Param("provider")
//		ver := c.Param("version")
//
//		v, err := m.ModuleRepository.FindVersion(namespace, name, provider, ver)
//
//		if err != nil {
//			c.JSON(http.StatusNotFound, gin.H{
//				"errors": []string{"Requested module was not found"},
//			})
//			return
//		}
//
//		c.Header("X-Terraform-Get", v.FetchKey)
//		c.JSON(http.StatusOK, gin.H{
//			"errors": []string{},
//		})
//	}
//}

func (s *DefaultModuleService) Upload(d *module.CreateDTO) error {
	if semVer := version.Version(d.Version); !semVer.Valid() {
		return fmt.Errorf("version should respect the semantic versioning standard (semver.org)")
	}

	m := d.ToModule()
	if _, err := s.ModuleRepository.Upsert(m); err != nil {
		return err
	}

	return nil
}

//func (m *DefaultModuleService) Upload() func(c *gin.Context) {
//	return func(c *gin.Context) {
//		ver := c.Param("version")
//		if semVer := version.Version(ver); !semVer.Valid() {
//			c.JSON(http.StatusBadRequest, gin.H{
//				"errors": []string{"version should respect the semantic versioning standard (semver.org)"},
//			})
//		}
//
//		namespace := c.Param("namespace")
//		name := c.Param("name")
//		provider := c.Param("provider")
//
//		var body module.CreateDTO
//		if err := c.BindJSON(&body); err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{
//				"errors": []string{err.Error()},
//			})
//		}
//
//		body.Namespace = namespace
//		body.Name = name
//		body.Provider = provider
//		body.Version = ver
//
//		request := body.ToModule()
//
//		if _, err := m.ModuleRepository.Upsert(request); err != nil {
//			c.JSON(http.StatusConflict, gin.H{
//				"errors": []string{err.Error()},
//			})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{
//			"errors": []string{},
//		})
//	}
//}

func (s *DefaultModuleService) Delete(namespace string, name string, provider string) error {
	if err := s.ModuleRepository.Delete(namespace, name, provider); err != nil {
		return err
	}

	return nil
}

//
//func (m *DefaultModuleService) Delete() func(c *gin.Context) {
//	return func(c *gin.Context) {
//		namespace := c.Param("namespace")
//		name := c.Param("name")
//		provider := c.Param("provider")
//
//		if err := m.ModuleRepository.Delete(namespace, name, provider); err != nil {
//			c.JSON(http.StatusConflict, gin.H{
//				"errors": []string{err.Error()},
//			})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{
//			"errors": []string{},
//		})
//	}
//}

func (s *DefaultModuleService) DeleteVersion(namespace string, name string, provider string, version string) error {
	if err := s.ModuleRepository.DeleteVersion(namespace, name, provider, version); err != nil {
		return err
	}

	return nil
}

//
//
//func (m *DefaultModuleService) DeleteVersion() func(c *gin.Context) {
//	return func(c *gin.Context) {
//		namespace := c.Param("namespace")
//		name := c.Param("name")
//		provider := c.Param("provider")
//		ver := c.Param("version")
//
//		if err := m.ModuleRepository.DeleteVersion(namespace, name, provider, ver); err != nil {
//			c.JSON(http.StatusConflict, gin.H{
//				"errors": []string{err.Error()},
//			})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{
//			"errors": []string{},
//		})
//	}
//}
