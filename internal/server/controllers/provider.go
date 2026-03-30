package controllers

import (
	"fmt"
	"net/http"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	providersTerraformApiBase = "/providers"
	providersDefaultApiBase   = "/api/providers"
)

// ProviderController registers the routes that handles the modules.
type ProviderController interface {
	api.RestController

	// TerraformApi returns the endpoint where Terraform can query
	// providers.
	TerraformApi() string
}

// DefaultProviderController is a concrete implementation of ProviderController.
type DefaultProviderController struct {
	ProviderService  services.ProviderService
	AuthorityService services.AuthorityService
	Authentication   *handlers.Authentication
	Authorization    *handlers.Authorization
	AnonymousRead    bool
}

func (c *DefaultProviderController) Paths() []string {
	return []string{
		providersTerraformApiBase,
		providersDefaultApiBase,
	}
}

func (c *DefaultProviderController) TerraformApi() string {
	return providersTerraformApiBase + "/"
}

func (c *DefaultProviderController) Subscribe(apis ...*gin.RouterGroup) {
	requireAuthorization := c.Authorization.RequireAuthorization(rbac.ResourceProviders)

	slugComposer := func(ctx *gin.Context) string {
		namespace := ctx.Param("namespace")
		name := ctx.Param("name")

		return fmt.Sprintf("%s/%s", namespace, name)
	}

	// tfApi should be compliant with the Terraform Registry Protocol for
	// providers
	// Docs: https://www.terraform.io/docs/internals/provider-registry-protocol.html#find-a-provider-package
	tfApi := apis[0]
	if !c.AnonymousRead {
		tfApi.Use(c.Authentication.AttemptAuthentication())
		tfApi.Use(requireAuthorization(rbac.ActionGet, slugComposer))
	}

	tfApi.GET(
		"/:namespace/:name/versions",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")

			d, err := c.ProviderService.Get(namespace, name)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusOK, d)
		},
	)

	// Docs: https://www.terraform.io/docs/internals/provider-registry-protocol.html#find-a-provider-package
	tfApi.GET(
		"/:namespace/:name/:version/download/:os/:arch",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			version := ctx.Param("version")
			os := ctx.Param("os")
			arch := ctx.Param("arch")

			dto, err := c.ProviderService.GetVersion(namespace, name, version, os, arch)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}
			ctx.JSON(http.StatusOK, dto)
		},
	)

	// api holds the routes that are not described by the Terraform protocol
	api := apis[1]
	api.Use(c.Authentication.AttemptAuthentication())

	// This is a protected endpoint, every request should be authenticated.
	api.Use(c.Authentication.RequireAuthentication())

	// Upload a new provider version
	api.POST(
		"/:namespace/:name/:version/upload",
		requireAuthorization(rbac.ActionCreate, slugComposer),
		func(ctx *gin.Context) {
			authorityID, ok := c.resolveAuthorityID(ctx)
			if !ok {
				return
			}

			name := ctx.Param("name")
			version := ctx.Param("version")

			var body provider.CreateProviderDTO
			if err := ctx.BindJSON(&body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			body.AuthorityID = authorityID
			body.Name = name
			body.Version = version

			if err := c.ProviderService.Upload(&body); err != nil {
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
		"/:namespace/:name/remove",
		requireAuthorization(rbac.ActionDelete, slugComposer),
		func(ctx *gin.Context) {
			authorityID, ok := c.resolveAuthorityID(ctx)
			if !ok {
				return
			}

			name := ctx.Param("name")

			if err := c.ProviderService.Delete(authorityID, name); err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
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
		"/:namespace/:name/:version/remove",
		requireAuthorization(rbac.ActionDelete, slugComposer),
		func(ctx *gin.Context) {
			authorityID, ok := c.resolveAuthorityID(ctx)
			if !ok {
				return
			}

			name := ctx.Param("name")
			version := ctx.Param("version")

			if err := c.ProviderService.DeleteVersion(authorityID, name, version); err != nil {
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

// resolveAuthorityID resolves the authority ID from the namespace URL parameter.
func (c *DefaultProviderController) resolveAuthorityID(ctx *gin.Context) (uuid.UUID, bool) {
	namespace := ctx.Param("namespace")

	authority, err := c.AuthorityService.GetByName(namespace)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"errors": []string{fmt.Sprintf("authority %q not found", namespace)},
		})
		return uuid.UUID{}, false
	}

	return authority.ID, true
}
