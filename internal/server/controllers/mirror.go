package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"terralist/internal/server/handlers"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/rbac"

	"github.com/gin-gonic/gin"
)

const (
	// mirrorProtocolBase is the base path of the Terraform Provider Network
	// Mirror Protocol. The protocol does not use service discovery, so the
	// path is fixed and configured by users in their Terraform CLI config.
	mirrorProtocolBase = "/providers"
)

// MirrorController registers the routes that serve the providers through the
// Terraform Provider Network Mirror Protocol.
type MirrorController interface {
	api.RestController
}

// DefaultMirrorController is a concrete implementation of MirrorController.
type DefaultMirrorController struct {
	ProviderService services.ProviderService
	Authentication  *handlers.Authentication
	Authorization   *handlers.Authorization

	// Hostname is the host under which Terralist serves its providers. The
	// mirror answers only for providers addressed with this hostname.
	Hostname      string
	AnonymousRead bool
}

func (c *DefaultMirrorController) Paths() []string {
	return []string{
		mirrorProtocolBase,
	}
}

func (c *DefaultMirrorController) Subscribe(apis ...*gin.RouterGroup) {
	requireAuthorization := c.Authorization.RequireAuthorization(rbac.ResourceProviders)

	slugComposer := func(ctx *gin.Context) string {
		namespace := ctx.Param("namespace")
		name := ctx.Param("name")

		return fmt.Sprintf("%s/%s", namespace, name)
	}

	// tfApi should be compliant with the Terraform Provider Network Mirror
	// Protocol.
	// Docs: https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol
	tfApi := apis[0]
	tfApi.Use(c.requireHostname())
	if !c.AnonymousRead {
		tfApi.Use(c.Authentication.AttemptAuthentication())
		tfApi.Use(requireAuthorization(rbac.ActionGet, slugComposer))
	}

	tfApi.GET(
		"/:hostname/:namespace/:name/index.json",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")

			dto, err := c.ProviderService.ListMirrorVersions(namespace, name)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, dto)
		},
	)

	tfApi.GET(
		"/:hostname/:namespace/:name/:version",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")

			version, ok := strings.CutSuffix(ctx.Param("version"), ".json")
			if !ok {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			dto, err := c.ProviderService.ListMirrorArchives(namespace, name, version)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, dto)
		},
	)
}

// requireHostname rejects requests for providers addressed with a hostname
// other than the one Terralist serves.
func (c *DefaultMirrorController) requireHostname() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !strings.EqualFold(ctx.Param("hostname"), c.Hostname) {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Next()
	}
}
