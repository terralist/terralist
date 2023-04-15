package controllers

import (
	"fmt"
	"net/http"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/artifact"
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/services"
	"terralist/pkg/api"

	"github.com/gin-gonic/gin"
	"github.com/ssoroka/slice"
)

const (
	artifactApiBase = "/api/artifacts"
)

// ArtifactController registers the endpoints to control authorities
type ArtifactController interface {
	api.RestController
}

// DefaultArtifactController is a concrete implementation of
// ArtifactController
type DefaultArtifactController struct {
	AuthorityService services.AuthorityService
	ModuleService    services.ModuleService
	ProviderService  services.ProviderService

	Authorization *handlers.Authorization
}

func (c *DefaultArtifactController) Paths() []string {
	return []string{artifactApiBase}
}

func (c *DefaultArtifactController) Subscribe(apis ...*gin.RouterGroup) {
	api := apis[0]
	api.Use(c.Authorization.AnyAuthentication())

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

			artifactsCount := 0
			for _, a := range authorities {
				artifactsCount += len(a.Modules) + len(a.Providers)
			}

			artifacts := make([]artifact.Artifact, 0, artifactsCount)
			for _, a := range authorities {
				for _, m := range a.Modules {
					artifact := m.ToArtifact()
					artifact.Namespace = a.Name
					artifact.FullName = fmt.Sprintf("%s/%s/%s", a.Name, m.Name, m.Provider)

					artifacts = append(artifacts, artifact)
				}

				for _, p := range a.Providers {
					artifact := p.ToArtifact()
					artifact.Namespace = a.Name
					artifact.FullName = fmt.Sprintf("%s/%s", a.Name, p.Name)

					artifacts = append(artifacts, artifact)
				}
			}

			ctx.JSON(http.StatusOK, artifacts)
		},
	)

	api.GET(
		"/:namespace/:name/:provider/version",
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

			versions := slice.Map[module.VersionListDTO, string](
				dto.Modules[0].Versions,
				func(v module.VersionListDTO) string {
					return v.Version
				},
			)

			ctx.JSON(http.StatusOK, versions)
		},
	)

	api.GET(
		"/:namespace/:name/version",
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

			versions := slice.Map[provider.VersionListVersionDTO, string](
				dto.Versions,
				func(v provider.VersionListVersionDTO) string {
					return v.Version
				},
			)

			ctx.JSON(http.StatusOK, versions)
		},
	)

	api.DELETE(
		"/:namespace/:name/:provider/version/:version",
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
