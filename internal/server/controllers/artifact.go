package controllers

import (
	"net/http"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/services"
	"terralist/pkg/api"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

			ctx.JSON(http.StatusOK, gin.H{})
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

			_ = namespace

			err := c.ModuleService.DeleteVersion(
				uuid.Must(uuid.NewRandom()), // TODO: Find authority based on namespace
				name,
				provider,
				version,
			)

			if err != nil {
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

			_ = namespace

			err := c.ProviderService.DeleteVersion(
				uuid.Must(uuid.NewRandom()), // TODO: Find authority based on namespace
				name,
				version,
			)

			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, true)
		},
	)
}
