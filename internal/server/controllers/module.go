package controllers

import (
	"net/http"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/module"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	modulesTerraformApiBase = "/modules"
	modulesDefaultApiBase   = "/api/modules"
)

// ModuleController registers the routes that handles the modules
type ModuleController interface {
	api.RestController

	// TerraformApi returns the endpoint where Terraform can query
	// modules
	TerraformApi() string
}

// DefaultModuleController is a concrete implementation of ModuleController
type DefaultModuleController struct {
	ModuleService services.ModuleService
	ApiKeyService services.ApiKeyService
	JWT           jwt.JWT
}

func (c *DefaultModuleController) TerraformApi() string {
	return modulesTerraformApiBase + "/"
}

func (c *DefaultModuleController) Paths() []string {
	return []string{
		modulesTerraformApiBase,
		modulesDefaultApiBase,
	}
}

func (c *DefaultModuleController) Subscribe(apis ...*gin.RouterGroup) {
	// tfApi should be compliant with the Terraform Registry Protocol for
	// modules
	// Docs: https://www.terraform.io/docs/internals/module-registry-protocol.html#list-available-versions-for-a-specific-module
	tfApi := apis[0]
	tfApi.Use(handlers.Authorize(c.JWT, c.ApiKeyService))

	tfApi.GET(
		"/:namespace/:name/:provider/versions",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			provider := ctx.Param("provider")

			d, err := c.ModuleService.Get(namespace, name, provider)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusOK, d)
		},
	)

	// Docs: https://www.terraform.io/docs/internals/module-registry-protocol.html#download-source-code-for-a-specific-module-version
	tfApi.GET(
		"/:namespace/:name/:provider/:version/download",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

			location, err := c.ModuleService.GetVersion(namespace, name, provider, version)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.Header("X-Terraform-Get", *location)
			ctx.JSON(http.StatusOK, gin.H{
				"errors": []string{},
			})
		},
	)

	// api holds the routes that are not described by the Terraform protocol
	api := apis[1]
	api.Use(handlers.Authorize(c.JWT, c.ApiKeyService))

	// Upload a new module version
	api.POST(
		"/:name/:provider/:version/upload",
		handlers.RequireAuthority(),
		func(ctx *gin.Context) {
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

			authorityID, err := uuid.Parse(ctx.GetString("authority"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"invalid authority"},
				})
				return
			}

			dto := module.CreateDTO{
				AuthorityID: authorityID,
				Name:        name,
				Provider:    provider,
				VersionDTO: module.VersionDTO{
					Version: version,
				},
			}

			var body module.CreateFromURLDTO
			if err := ctx.BindJSON(&body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			if err := c.ModuleService.Upload(&dto, body.DownloadUrl); err != nil {
				ctx.JSON(http.StatusConflict, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"errors": []string{},
			})
		},
	)

	// Delete a provider
	api.DELETE(
		"/:name/:provider/remove",
		handlers.RequireAuthority(),
		func(ctx *gin.Context) {
			name := ctx.Param("name")
			provider := ctx.Param("provider")

			authorityID, err := uuid.Parse(ctx.GetString("authority"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"invalid authority"},
				})
				return
			}

			if err := c.ModuleService.Delete(authorityID, name, provider); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"errors": []string{},
			})
		},
	)

	// Delete a provider version
	api.DELETE(
		"/:name/:provider/:version/remove",
		handlers.RequireAuthority(),
		func(ctx *gin.Context) {
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

			authorityID, err := uuid.Parse(ctx.GetString("authority"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"invalid authority"},
				})
				return
			}

			if err := c.ModuleService.DeleteVersion(authorityID, name, provider, version); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"errors": []string{},
			})
		},
	)
}
