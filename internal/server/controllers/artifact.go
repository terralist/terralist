package controllers

import (
	"net/http"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/slug"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ssoroka/slice"
)

const (
	artifactApiBase = "/api/artifact"
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
		"/:slug/version",
		func(ctx *gin.Context) {
			s, err := slug.Parse(ctx.Param("slug"))

			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			var versions []string
			if s.Provider != "" {
				dto, err := c.ModuleService.Get(s.Namespace, s.Name, s.Provider)
				if err != nil {
					ctx.JSON(http.StatusNotFound, gin.H{
						"errors": []string{err.Error()},
					})
					return
				}

				versions = slice.Map[module.VersionListDTO, string](
					dto.Modules[0].Versions,
					func(v module.VersionListDTO) string {
						return v.Version
					},
				)
			} else {
				dto, err := c.ProviderService.Get(s.Namespace, s.Name)
				if err != nil {
					ctx.JSON(http.StatusNotFound, gin.H{
						"errors": []string{err.Error()},
					})
					return
				}

				versions = slice.Map[provider.VersionListVersionDTO, string](
					dto.Versions,
					func(v provider.VersionListVersionDTO) string {
						return v.Version
					},
				)
			}

			ctx.JSON(http.StatusOK, versions)
		},
	)

	api.DELETE(
		"/:slug/version/:version",
		func(ctx *gin.Context) {
			s, err := slug.Parse(ctx.Param("slug"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			version := ctx.Param("version")

			if s.Provider != "" {
				err := c.ModuleService.DeleteVersion(
					uuid.Must(uuid.NewRandom()),
					s.Name,
					s.Provider,
					version,
				)

				if err != nil {
					ctx.JSON(http.StatusNotFound, gin.H{
						"errors": []string{err.Error()},
					})
					return
				}

			} else {
				err := c.ProviderService.DeleteVersion(
					uuid.Must(uuid.NewRandom()),
					s.Name,
					version,
				)

				if err != nil {
					ctx.JSON(http.StatusNotFound, gin.H{
						"errors": []string{err.Error()},
					})
					return
				}
			}

			ctx.JSON(http.StatusOK, true)
		},
	)
}
