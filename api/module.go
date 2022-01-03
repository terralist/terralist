package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/model/module"
	"github.com/valentindeaconu/terralist/service"
	"github.com/valentindeaconu/terralist/settings"
)

type ModuleController struct {
	Router        *gin.Engine
	ModuleService *service.ModuleService
}

func CreateModuleController(
	router *gin.Engine,
	moduleService *service.ModuleService,
) *ModuleController {
	p := new(ModuleController)
	p.Router = router
	p.ModuleService = moduleService

	return p
}

func (m *ModuleController) PrepareRoutes() {
	// https://www.terraform.io/docs/internals/module-registry-protocol.html#list-available-versions-for-a-specific-module
	m.Router.GET(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/versions",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		func(c *gin.Context) {
			namespace := c.Param("namespace")
			name := c.Param("name")
			provider := c.Param("provider")

			p, err := m.ModuleService.Find(namespace, name, provider)

			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"errors": []string{
						"Requested module was not found",
						err.Error(),
					},
				})
			} else {
				c.JSON(http.StatusOK, p.ToVersionListResponse())
			}
		},
	)

	// https://www.terraform.io/docs/internals/module-registry-protocol.html#download-source-code-for-a-specific-module-version
	m.Router.GET(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/:version/download",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		func(c *gin.Context) {
			namespace := c.Param("namespace")
			name := c.Param("name")
			provider := c.Param("provider")
			version := c.Param("version")

			v, err := m.ModuleService.FindVersion(namespace, name, provider, version)

			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"errors": []string{"Requested module was not found"},
				})
			} else {
				c.Header("X-Terraform-Get", v.Location)
				c.JSON(http.StatusOK, gin.H{
					"errors": []string{},
				})
			}
		},
	)

	// Upload a new module
	m.Router.POST(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/:version/upload",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		func(c *gin.Context) {
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

			if _, err := m.ModuleService.Upsert(request); err != nil {
				c.JSON(http.StatusConflict, gin.H{
					"errors": []string{err.Error()},
				})
			} else {
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
		},
	)
}
