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
	ProviderService services.ProviderService
	Authentication  *handlers.Authentication
	Authorization   *handlers.Authorization
	AnonymousRead   bool
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

	fullSlugComposer := func(ctx *gin.Context) string {
		namespace := ctx.Param("namespace")
		name := ctx.Param("name")

		return fmt.Sprintf("%s/%s", namespace, name)
	}

	partialSlugComposer := func(ctx *gin.Context) string {
		name := ctx.Param("name")

		return fmt.Sprintf("%s/%s", ctx.GetString("authorityName"), name)
	}

	// tfApi should be compliant with the Terraform Registry Protocol for
	// providers
	// Docs: https://www.terraform.io/docs/internals/provider-registry-protocol.html#find-a-provider-package
	tfApi := apis[0]
	if !c.AnonymousRead {
		tfApi.Use(c.Authentication.RequireAuthentication())
		tfApi.Use(requireAuthorization(rbac.ActionGet, fullSlugComposer))
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
	api.Use(c.Authentication.RequireAuthentication())

	// Upload a new provider version
	api.POST(
		"/:name/:version/upload",
		handlers.RequireAuthority(),
		requireAuthorization(rbac.ActionCreate, partialSlugComposer),
		func(ctx *gin.Context) {
			name := ctx.Param("name")
			version := ctx.Param("version")

			authorityID, err := uuid.Parse(ctx.GetString("authorityID"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"invalid authority"},
				})
				return
			}

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
		"/:name/remove",
		handlers.RequireAuthority(),
		requireAuthorization(rbac.ActionDelete, partialSlugComposer),
		func(ctx *gin.Context) {
			name := ctx.Param("name")

			authorityID, err := uuid.Parse(ctx.GetString("authorityID"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"invalid authority"},
				})
				return
			}

			if err := c.ProviderService.Delete(authorityID, name); err != nil {
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
		"/:name/:version/remove",
		handlers.RequireAuthority(),
		requireAuthorization(rbac.ActionDelete, partialSlugComposer),
		func(ctx *gin.Context) {
			name := ctx.Param("name")
			version := ctx.Param("version")

			authorityID, err := uuid.Parse(ctx.GetString("authorityID"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"invalid authority"},
				})
				return
			}

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
