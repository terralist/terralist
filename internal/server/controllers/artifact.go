package controllers

import (
	"fmt"
	"net/http"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/artifact"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/rbac"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

const (
	artifactApiBase = "/api/artifacts"
)

// ArtifactController registers the endpoints to control authorities.
type ArtifactController interface {
	api.RestController
}

// DefaultArtifactController is a concrete implementation of
// ArtifactController.
type DefaultArtifactController struct {
	AuthorityService services.AuthorityService
	ModuleService    services.ModuleService
	ProviderService  services.ProviderService

	Authentication *handlers.Authentication
	Authorization  *handlers.Authorization
}

func (c *DefaultArtifactController) Paths() []string {
	return []string{artifactApiBase}
}

func (c *DefaultArtifactController) Subscribe(apis ...*gin.RouterGroup) {
	api := apis[0]

	moduleComposer := func(ctx *gin.Context) string {
		namespace := ctx.Param("namespace")
		name := ctx.Param("name")
		provider := ctx.Param("provider")

		return fmt.Sprintf("%s/%s/%s", namespace, name, provider)
	}

	providerComposer := func(ctx *gin.Context) string {
		namespace := ctx.Param("namespace")
		name := ctx.Param("name")

		return fmt.Sprintf("%s/%s", namespace, name)
	}

	// This is a protected endpoint, every request should be authenticated.
	api.Use(c.Authentication.RequireAuthentication())

	api.GET(
		"/",
		func(ctx *gin.Context) {
			authorities, err := c.AuthorityService.GetAll()
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			// The user key should be preset by the RequireAuthentication middleware.
			user := handlers.MustGetFromContext[auth.User](ctx, "user")

			// Filter only authorities where the user has access
			filteredAuthorities := lo.Filter(authorities, func(authority *authority.Authority, idx int) bool {
				return c.Authorization.CanPerform(*user, rbac.ResourceAuthorities, rbac.ActionGet, authority.Name)
			})

			artifactsCount := 0
			for _, a := range filteredAuthorities {
				artifactsCount += len(a.Modules) + len(a.Providers)
			}

			artifacts := make([]artifact.Artifact, 0, artifactsCount)
			for _, a := range authorities {
				for _, m := range a.Modules {
					artifact := m.ToArtifact()
					artifact.Namespace = a.Name
					artifact.FullName = fmt.Sprintf("%s/%s/%s", a.Name, m.Name, m.Provider)

					if !c.Authorization.CanPerform(*user, rbac.ResourceModules, rbac.ActionGet, artifact.FullName) {
						continue
					}

					artifacts = append(artifacts, artifact)
				}

				for _, p := range a.Providers {
					artifact := p.ToArtifact()
					artifact.Namespace = a.Name
					artifact.FullName = fmt.Sprintf("%s/%s", a.Name, p.Name)

					if !c.Authorization.CanPerform(*user, rbac.ResourceProviders, rbac.ActionGet, artifact.FullName) {
						continue
					}

					artifacts = append(artifacts, artifact)
				}
			}

			ctx.JSON(http.StatusOK, artifacts)
		},
	)

	// TODO: I'm pretty sure this is unused
	api.GET(
		"/:namespace/:name/:provider/version",
		c.Authorization.RequireAuthorization(rbac.ResourceModules)(rbac.ActionGet, moduleComposer),
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			provider := ctx.Param("provider")

			dto, err := c.ModuleService.Get(namespace, name, provider)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			versions := lo.Map(
				dto.Modules[0].Versions,
				func(v module.VersionListDTO, _ int) string {
					return v.Version
				},
			)

			ctx.JSON(http.StatusOK, versions)
		},
	)

	api.GET(
		"/:namespace/:name/:provider/version/:version",
		c.Authorization.RequireAuthorization(rbac.ResourceModules)(rbac.ActionGet, moduleComposer),
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

			dto, err := c.ModuleService.GetVersion(namespace, name, provider, version)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, dto)
		},
	)

	api.GET(
		"/:namespace/:name/version",
		c.Authorization.RequireAuthorization(rbac.ResourceProviders)(rbac.ActionGet, providerComposer),
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")

			dto, err := c.ProviderService.Get(namespace, name)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			versions := lo.Map(
				dto.Versions,
				func(v provider.VersionListVersionDTO, _ int) string {
					return v.Version
				},
			)

			ctx.JSON(http.StatusOK, versions)
		},
	)

	api.DELETE(
		"/:namespace/:name/:provider/version/:version",
		c.Authorization.RequireAuthorization(rbac.ResourceModules)(rbac.ActionDelete, moduleComposer),
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

			authority, err := c.AuthorityService.GetByName(namespace)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			if err := c.ModuleService.DeleteVersion(
				authority.ID,
				name,
				provider,
				version,
			); err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, true)
		},
	)

	api.DELETE(
		"/:namespace/:name/version/:version",
		c.Authorization.RequireAuthorization(rbac.ResourceProviders)(rbac.ActionDelete, providerComposer),
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			version := ctx.Param("version")

			authority, err := c.AuthorityService.GetByName(namespace)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			if err := c.ProviderService.DeleteVersion(
				authority.ID,
				name,
				version,
			); err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, true)
		},
	)
}
