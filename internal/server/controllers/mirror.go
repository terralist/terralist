package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/repositories"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/rbac"
	"terralist/pkg/registry"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
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

	// AutoCreate lists the upstream hostnames for which an authority is
	// created on the first authenticated request for an unknown namespace.
	AutoCreate []string
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
	tfApi.Use(c.Authentication.AttemptAuthentication())
	tfApi.Use(c.resolveNamespace())
	if !c.AnonymousRead {
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

	// The last path element is either a version document, <version>.json, or
	// a package listed by such a document, which sits next to it.
	tfApi.GET(
		"/:hostname/:namespace/:name/:version",
		func(ctx *gin.Context) {
			namespace := handlers.MustGetFromContext[string](ctx, mirrorNamespaceKey)
			name := ctx.Param("name")
			document := ctx.Param("version")

			if version, ok := strings.CutSuffix(document, ".json"); ok {
				c.listArchives(ctx, *namespace, name, version)
				return
			}

			pkg, ok := provider.ParsePackageFileName(document)
			if !ok || pkg.Name != name {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			c.download(ctx, *namespace, pkg)
		},
	)
}

// listArchives serves the network mirror version document.
func (c *DefaultMirrorController) listArchives(ctx *gin.Context, namespace, name, version string) {
	dto, err := c.ProviderService.ListMirrorArchives(namespace, name, version, c.mayFetch(ctx, namespace, name))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"errors": []string{err.Error()},
		})
		return
	}

	ctx.JSON(http.StatusOK, dto)
}

// download redirects to the storage location of a package, fetching it from
// the upstream registry first when the caller may do so.
func (c *DefaultMirrorController) download(ctx *gin.Context, namespace string, pkg provider.Package) {
	url, err := c.ProviderService.Download(namespace, pkg.Name, pkg.Version, pkg.System, pkg.Architecture, c.mayFetch(ctx, namespace, pkg.Name))

	switch {
	case err == nil:
		ctx.Redirect(http.StatusFound, url)
	case errors.Is(err, services.ErrFetchRequiresCreate):
		ctx.AbortWithStatus(http.StatusForbidden)
	case errors.Is(err, repositories.ErrNotFound), errors.Is(err, registry.ErrNotFound), errors.Is(err, services.ErrUpstreamDenied):
		ctx.JSON(http.StatusNotFound, gin.H{
			"errors": []string{err.Error()},
		})
	default:
		ctx.JSON(http.StatusBadGateway, gin.H{
			"errors": []string{err.Error()},
		})
	}
}

// autoCreate creates the authority standing for an allowlisted upstream
// namespace on behalf of the authenticated caller, and reports whether it did.
func (c *DefaultMirrorController) autoCreate(ctx *gin.Context, hostname, namespace string) bool {
	if !lo.ContainsBy(c.AutoCreate, func(allowed string) bool { return strings.EqualFold(allowed, hostname) }) {
		return false
	}

	user, err := handlers.GetFromContext[auth.User](ctx, "user")
	if err != nil || user.Email == "" {
		return false
	}

	_, err = c.AuthorityService.Create(authority.AuthorityCreateDTO{
		Name:              namespace,
		Owner:             user.Email,
		UpstreamHostname:  strings.ToLower(hostname),
		UpstreamNamespace: namespace,
		UpstreamEnabled:   true,
	})
	if err != nil {
		log.Warn().Err(err).Str("hostname", hostname).Str("namespace", namespace).Msg("Could not create the authority for an upstream namespace.")
		return false
	}

	log.Info().Str("hostname", hostname).Str("namespace", namespace).Str("owner", user.Email).Msg("Created an authority for an upstream namespace.")

	return true
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
				if !c.autoCreate(ctx, hostname, namespace) {
					ctx.AbortWithStatus(http.StatusNotFound)
					return
				}
			} else {
				namespace = a.Name
			}
		}

		ctx.Set(mirrorNamespaceKey, &namespace)
		ctx.Next()
	}
}
