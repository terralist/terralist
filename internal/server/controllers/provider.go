package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/internal/server/models/provider"
	"github.com/valentindeaconu/terralist/internal/server/services"
)

type ProviderController struct {
	ProviderService *services.ProviderService
}

func (p *ProviderController) Get() func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		provider, err := p.ProviderService.Find(namespace, name)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{
					"Requested provider was not found",
					err.Error(),
				},
			})
			return
		}
		c.JSON(http.StatusOK, provider.ToVersionListProvider())
	}
}

func (p *ProviderController) GetVersion() func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		version := c.Param("version")
		system := c.Param("os")
		arch := c.Param("arch")

		v, err := p.ProviderService.FindVersion(namespace, name, version)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{"Requested provider was not found"},
			})
			return
		}

		response, err := v.ToDownloadProvider(system, arch)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{err.Error()},
			})
		} else {
			c.JSON(http.StatusOK, response)
		}
	}
}

func (p *ProviderController) Upload() func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		version := c.Param("version")

		var providerVersion provider.CreateProviderDTO

		if err := c.BindJSON(&providerVersion); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{err.Error()},
			})
			return
		}

		providerVersion.Namespace = namespace
		providerVersion.Name = name
		providerVersion.Version = version

		request := provider.FromCreateDTO(providerVersion)

		if _, err := p.ProviderService.Upsert(request); err != nil {
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

func (p *ProviderController) Delete() func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		if err := p.ProviderService.Delete(namespace, name); err != nil {
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

func (p *ProviderController) DeleteVersion() func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		version := c.Param("version")

		if err := p.ProviderService.DeleteVersion(namespace, name, version); err != nil {
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
