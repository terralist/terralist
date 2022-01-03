package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/model/provider"
	"github.com/valentindeaconu/terralist/service"
	"github.com/valentindeaconu/terralist/settings"
)

type ProviderController struct {
	Router          *gin.Engine
	ProviderService *service.ProviderService
}

func CreateProviderController(
	router *gin.Engine,
	providerService *service.ProviderService,
) *ProviderController {
	p := new(ProviderController)

	p.Router = router
	p.ProviderService = providerService

	return p
}

func (m *ProviderController) PrepareRoutes() {
	// https://www.terraform.io/docs/internals/provider-registry-protocol.html#list-available-versions
	m.Router.GET(
		fmt.Sprintf(
			"%s/:namespace/:name/versions",
			settings.ServiceDiscovery.ProviderEndpoint,
		),
		func(c *gin.Context) {
			namespace := c.Param("namespace")
			name := c.Param("name")

			p, err := m.ProviderService.Find(namespace, name)

			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"errors": []string{
						"Requested provider was not found",
						err.Error(),
					},
				})
			} else {
				c.JSON(http.StatusOK, p.ToVersionListProvider())
			}
		},
	)

	// https://www.terraform.io/docs/internals/provider-registry-protocol.html#find-a-provider-package
	m.Router.GET(
		fmt.Sprintf(
			"%s/:namespace/:name/:version/download/:os/:arch",
			settings.ServiceDiscovery.ProviderEndpoint,
		),
		func(c *gin.Context) {
			namespace := c.Param("namespace")
			name := c.Param("name")
			version := c.Param("version")
			system := c.Param("os")
			arch := c.Param("arch")

			v, err := m.ProviderService.FindVersion(namespace, name, version)

			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"errors": []string{"Requested provider was not found"},
				})
			} else {
				response, err := v.ToDownloadProvider(system, arch)

				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{
						"errors": []string{err.Error()},
					})
				} else {
					c.JSON(http.StatusOK, response)
				}
			}
		},
	)

	// Upload a new provider
	m.Router.POST(
		fmt.Sprintf(
			"%s/:namespace/:name/:version/upload",
			settings.ServiceDiscovery.ProviderEndpoint,
		),
		func(c *gin.Context) {
			namespace := c.Param("namespace")
			name := c.Param("name")
			version := c.Param("version")

			var providerVersion provider.CreateProviderDTO
			if err := c.BindJSON(&providerVersion); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
			}

			providerVersion.Namespace = namespace
			providerVersion.Name = name
			providerVersion.Version = version

			request := provider.FromCreateDTO(providerVersion)

			if _, err := m.ProviderService.Upsert(request); err != nil {
				c.JSON(http.StatusConflict, gin.H{
					"errors": []string{err.Error()},
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"download_uri": fmt.Sprintf(
						"%s/%s/%s/%s/download/:system/:arch",
						settings.ServiceDiscovery.ProviderEndpoint,
						namespace,
						name,
						version,
					),
					"errors": []string{},
				})
			}
		},
	)
}
