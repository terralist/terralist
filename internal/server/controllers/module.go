package controllers

import (
	"net/http"
	"terralist/pkg/version"

	"terralist/internal/server/models/module"
	"terralist/internal/server/services"

	"github.com/gin-gonic/gin"
)

type ModuleController struct {
	ModuleService *services.ModuleService
}

func (m *ModuleController) Get() func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		provider := c.Param("provider")

		mod, err := m.ModuleService.Find(namespace, name, provider)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{
					"Requested module was not found",
					err.Error(),
				},
			})
			return
		}
		c.JSON(http.StatusOK, mod.ToListResponseDTO())
	}
}

func (m *ModuleController) GetVersion() func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		provider := c.Param("provider")
		ver := c.Param("version")

		v, err := m.ModuleService.FindVersion(namespace, name, provider, ver)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{"Requested module was not found"},
			})
			return
		}

		c.Header("X-Terraform-Get", v.FetchKey)
		c.JSON(http.StatusOK, gin.H{
			"errors": []string{},
		})
	}
}

func (m *ModuleController) Upload() func(c *gin.Context) {
	return func(c *gin.Context) {
		ver := c.Param("version")
		if semVer := version.Version(ver); !semVer.Valid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{"version should respect the semantic versioning standard (semver.org)"},
			})
		}

		namespace := c.Param("namespace")
		name := c.Param("name")
		provider := c.Param("provider")

		var body module.CreateDTO
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{err.Error()},
			})
		}

		body.Namespace = namespace
		body.Name = name
		body.Provider = provider
		body.Version = ver

		request := body.ToModule()

		if _, err := m.ModuleService.Upsert(request); err != nil {
			c.JSON(http.StatusConflict, gin.H{
				"errors": []string{err.Error()},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"errors": []string{},
		})
	}
}

func (m *ModuleController) Delete() func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		provider := c.Param("provider")

		if err := m.ModuleService.Delete(namespace, name, provider); err != nil {
			c.JSON(http.StatusConflict, gin.H{
				"errors": []string{err.Error()},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"errors": []string{},
		})
	}
}

func (m *ModuleController) DeleteVersion() func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		provider := c.Param("provider")
		ver := c.Param("version")

		if err := m.ModuleService.DeleteVersion(namespace, name, provider, ver); err != nil {
			c.JSON(http.StatusConflict, gin.H{
				"errors": []string{err.Error()},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"errors": []string{},
		})
	}
}
