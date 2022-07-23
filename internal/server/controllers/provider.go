package controllers

import (
	"net/http"
	"terralist/pkg/version"

	"terralist/internal/server/models/provider"
	"terralist/internal/server/services"

	"github.com/gin-gonic/gin"
)

type ProviderController struct {
	ProviderService *services.ProviderService
}

func (p *ProviderController) Get() func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")

		prov, err := p.ProviderService.Find(namespace, name)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{
					"Requested provider was not found",
					err.Error(),
				},
			})
			return
		}
		c.JSON(http.StatusOK, prov.ToVersionListProviderDTO())
	}
}

func (p *ProviderController) GetVersion() func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace := c.Param("namespace")
		name := c.Param("name")
		ver := c.Param("version")
		system := c.Param("os")
		arch := c.Param("arch")

		v, err := p.ProviderService.FindVersion(namespace, name, ver)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"errors": []string{"Requested provider was not found"},
			})
			return
		}

		response, err := v.ToDownloadVersionDTO(system, arch)

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
		ver := c.Param("version")
		if semVer := version.Version(ver); !semVer.Valid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{"version should respect the semantic versioning standard (semver.org)"},
			})
		}

		namespace := c.Param("namespace")
		name := c.Param("name")

		var body provider.CreateProviderDTO

		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{err.Error()},
			})
			return
		}

		body.AuthorityID = c.GetString("issuer")
		body.Namespace = namespace
		body.Name = name
		body.Version = ver

		prov := body.ToProvider()

		if _, err := p.ProviderService.Upsert(prov); err != nil {
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
		ver := c.Param("version")

		if err := p.ProviderService.DeleteVersion(namespace, name, ver); err != nil {
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
