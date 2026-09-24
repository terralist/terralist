package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"terralist/internal/server/handlers"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
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
	ProviderService  services.ProviderService
	AuthorityService services.AuthorityService
	Authentication   *handlers.Authentication
	Authorization    *handlers.Authorization

	// Hostname is the host under which Terralist serves its providers.
	// Providers addressed with this hostname belong to the authority named
	// by the namespace; providers addressed with any other hostname belong
	// to the authority standing for that upstream hostname and namespace.
	Hostname      string
	AnonymousRead bool
}

// mirrorNamespaceKey is the context key holding the name of the authority a
// mirror request resolved to.
const mirrorNamespaceKey = "mirrorNamespace"

func (c *DefaultMirrorController) Paths() []string {
	return []string{
		mirrorProtocolBase,
	}
}

func (c *DefaultMirrorController) Subscribe(apis ...*gin.RouterGroup) {
	requireAuthorization := c.Authorization.RequireAuthorization(rbac.ResourceProviders)

	slugComposer := func(ctx *gin.Context) string {
		namespace := handlers.MustGetFromContext[string](ctx, mirrorNamespaceKey)
		name := ctx.Param("name")

		return fmt.Sprintf("%s/%s", *namespace, name)
	}

	// tfApi should be compliant with the Terraform Provider Network Mirror
	// Protocol.
	// Docs: https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol
	tfApi := apis[0]
	tfApi.Use(c.resolveNamespace())
	if !c.AnonymousRead {
		tfApi.Use(c.Authentication.AttemptAuthentication())
		tfApi.Use(requireAuthorization(rbac.ActionGet, slugComposer))
	}

	tfApi.GET(
		"/:hostname/:namespace/:name/index.json",
		func(ctx *gin.Context) {
			namespace := handlers.MustGetFromContext[string](ctx, mirrorNamespaceKey)
			name := ctx.Param("name")

			dto, err := c.ProviderService.ListMirrorVersions(*namespace, name, c.mayFetch(ctx, *namespace, name))
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
			namespace := handlers.MustGetFromContext[string](ctx, mirrorNamespaceKey)
			name := ctx.Param("name")

			version, ok := strings.CutSuffix(ctx.Param("version"), ".json")
			if !ok {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			dto, err := c.ProviderService.ListMirrorArchives(*namespace, name, version, c.mayFetch(ctx, *namespace, name))
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

// mayFetch reports whether the caller may create packages of the provider,
// which is what fetching them from the upstream registry amounts to.
func (c *DefaultMirrorController) mayFetch(ctx *gin.Context, namespace, name string) bool {
	user, err := handlers.GetFromContext[auth.User](ctx, "user")
	if err != nil {
		return false
	}

	return c.Authorization.CanPerform(*user, rbac.ResourceProviders, rbac.ActionCreate, fmt.Sprintf("%s/%s", namespace, name))
}

// resolveNamespace maps the hostname and namespace of a mirror request to the
// name of the authority holding the provider, and rejects requests nobody
// stands for.
func (c *DefaultMirrorController) resolveNamespace() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		hostname := ctx.Param("hostname")
		namespace := ctx.Param("namespace")

		if !strings.EqualFold(hostname, c.Hostname) {
			a, err := c.AuthorityService.GetByUpstream(hostname, namespace)
			if err != nil {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			namespace = a.Name
		}

		ctx.Set(mirrorNamespaceKey, &namespace)
		ctx.Next()
	}
}
