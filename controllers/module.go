package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/models/module"
	"github.com/valentindeaconu/terralist/services"
	"github.com/valentindeaconu/terralist/settings"
)

func ModuleGet(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	provider := c.Param("provider")

	p, err := services.ModuleFind(namespace, name, provider)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"errors": []string{
				"Requested module was not found",
				err.Error(),
			},
		})
		return
	}
	c.JSON(http.StatusOK, p.ToVersionListResponse())
}

func ModuleGetVersion(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	provider := c.Param("provider")
	version := c.Param("version")

	v, err := services.ModuleFindVersion(namespace, name, provider, version)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"errors": []string{"Requested module was not found"},
		})
		return
	}
	c.Header("X-Terraform-Get", v.Location)
	c.JSON(http.StatusOK, gin.H{
		"errors": []string{},
	})
}

func ModuleCreate(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	provider := c.Param("provider")
	version := c.Param("version")

	var body module.ModuleCreateDTO
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []string{err.Error()},
		})
	}

	body.Namespace = namespace
	body.Name = name
	body.Provider = provider
	body.Version = version

	request := module.FromCreateDTO(body)

	if _, err := services.ModuleUpsert(request); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"errors": []string{err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"download_uri": fmt.Sprintf(
			"%s/%s/%s/%s/%s/download",
			settings.ServiceDiscovery.ModuleEndpoint,
			namespace,
			name,
			provider,
			version,
		),
		"errors": []string{},
	})
}
