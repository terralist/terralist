package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/models/provider"
	"github.com/valentindeaconu/terralist/services"
)

func ProviderGet(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	p, err := services.ProviderFind(namespace, name)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"errors": []string{
				"Requested provider was not found",
				err.Error(),
			},
		})
		return
	}
	c.JSON(http.StatusOK, p.ToVersionListProvider())
}

func ProviderGetVersion(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	version := c.Param("version")
	system := c.Param("os")
	arch := c.Param("arch")

	v, err := services.ProviderFindVersion(namespace, name, version)

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

func ProviderUpload(c *gin.Context) {
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

	if _, err := services.ProviderUpsert(request); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"errors": []string{err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errors": []string{},
	})
}

func ProviderDelete(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if err := services.ProviderDelete(namespace, name); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"errors": []string{err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errors": []string{},
	})
}

func ProviderVersionDelete(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	version := c.Param("version")

	if err := services.ProviderVersionDelete(namespace, name, version); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"errors": []string{err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errors": []string{},
	})
}
