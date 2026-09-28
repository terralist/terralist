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

	// Tokens signs the package links listed in version documents. Terraform
	// downloads packages without credentials, so the link carries the proof.
	Tokens *handlers.DownloadTokens

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

const (
	// mirrorNamespaceKey is the context key holding the name of the authority
	// a mirror request resolved to.
	mirrorNamespaceKey = "mirrorNamespace"

	// packageFetchKey is the context key holding the fetch permission carried
	// by a valid package token, when the request presented one.
	packageFetchKey = "packageFetch"

	packageTokenQuery = "token"
)

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
	tfApi.Use(c.acceptPackageToken())
	if !c.AnonymousRead {
		authorize := requireAuthorization(rbac.ActionGet, slugComposer)
		tfApi.Use(func(ctx *gin.Context) {
			// A valid package token is the proof that the version document
			// listing the package was served to an authorized caller.
			if _, ok := ctx.Get(packageFetchKey); ok {
				ctx.Next()
				return
			}

			authorize(ctx)
		})
	}

	tfApi.GET(
		"/:hostname/:namespace/:name/index.json",
		func(ctx *gin.Context) {
			namespace := handlers.MustGetFromContext[string](ctx, mirrorNamespaceKey)
			name := ctx.Param("name")

			dto, err := c.ProviderService.ListMirrorVersions(*namespace, name, c.mayFetch(ctx, *namespace, name))
			if err != nil {
				ctx.JSON(lookupStatus(err), gin.H{
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
			if !ok || !strings.EqualFold(pkg.Name, name) {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			c.download(ctx, *namespace, pkg)
		},
	)
}

// listArchives serves the network mirror version document. Packages that are
// not stored yet are listed relative to the document, with a token proving
// the caller's permissions to the download that follows.
func (c *DefaultMirrorController) listArchives(ctx *gin.Context, namespace, name, version string) {
	fetch := c.mayFetch(ctx, namespace, name)

	dto, err := c.ProviderService.ListMirrorArchives(namespace, name, version, fetch)
	if err != nil {
		ctx.JSON(lookupStatus(err), gin.H{
			"errors": []string{err.Error()},
		})
		return
	}

	for key, archive := range dto.Archives {
		pkg, ok := provider.ParsePackageFileName(archive.URL)
		if !ok {
			continue
		}

		token, err := c.Tokens.Sign(pkg.Subject(namespace), fetch)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{err.Error()},
			})
			return
		}

		archive.URL = fmt.Sprintf("%s?%s=%s", archive.URL, packageTokenQuery, token)
		dto.Archives[key] = archive
	}

	ctx.JSON(http.StatusOK, dto)
}

// acceptPackageToken verifies the token a package request may carry and
// records the fetch permission it grants.
func (c *DefaultMirrorController) acceptPackageToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Query(packageTokenQuery)
		if token == "" {
			ctx.Next()
			return
		}

		namespace := handlers.MustGetFromContext[string](ctx, mirrorNamespaceKey)
		pkg, ok := provider.ParsePackageFileName(ctx.Param("version"))
		if !ok || !strings.EqualFold(pkg.Name, ctx.Param("name")) {
			ctx.Next()
			return
		}

		if fetch, ok := c.Tokens.Verify(token, pkg.Subject(*namespace)); ok {
			ctx.Set(packageFetchKey, &fetch)
		}

		ctx.Next()
	}
}

// download redirects to the storage location of a package, fetching it from
// the upstream registry first when the caller may do so, by credentials or by
// the token of the link.
func (c *DefaultMirrorController) download(ctx *gin.Context, namespace string, pkg provider.Package) {
	fetch := c.mayFetch(ctx, namespace, pkg.Name)
	if granted, err := handlers.GetFromContext[bool](ctx, packageFetchKey); err == nil {
		fetch = fetch || *granted
	}

	url, err := c.ProviderService.Download(namespace, pkg.Name, pkg.Version, pkg.System, pkg.Architecture, fetch)

	switch {
	case err == nil:
		ctx.Redirect(http.StatusFound, url)
	case errors.Is(err, services.ErrFetchRequiresCreate):
		ctx.AbortWithStatus(http.StatusForbidden)
	case isNotFound(err):
		ctx.JSON(http.StatusNotFound, gin.H{
			"errors": []string{err.Error()},
		})
	default:
		ctx.JSON(http.StatusBadGateway, gin.H{
			"errors": []string{err.Error()},
		})
	}
}

// lookupStatus is the status answering a failed lookup: a bad gateway when
// the upstream is unavailable, not found otherwise.
func lookupStatus(err error) int {
	if errors.Is(err, services.ErrUpstreamUnavailable) {
		return http.StatusBadGateway
	}

	return http.StatusNotFound
}

// isNotFound reports whether err means the artifact is held neither by
// Terralist nor by the upstream, or that the rules keep it from being served.
func isNotFound(err error) bool {
	return errors.Is(err, repositories.ErrNotFound) || errors.Is(err, registry.ErrNotFound) || errors.Is(err, services.ErrUpstreamDenied)
}

// autoCreate creates the authority standing for an allowlisted upstream
// namespace on behalf of the authenticated caller, when the caller may create
// that authority, and reports whether it did.
func (c *DefaultMirrorController) autoCreate(ctx *gin.Context, hostname, namespace string) bool {
	if !lo.ContainsBy(c.AutoCreate, func(allowed string) bool { return strings.EqualFold(allowed, hostname) }) {
		return false
	}

	user, err := handlers.GetFromContext[auth.User](ctx, "user")
	if err != nil || user.Email == "" {
		return false
	}

	if !c.Authorization.CanPerform(*user, rbac.ResourceAuthorities, rbac.ActionCreate, namespace) {
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
