package controllers

import (
	"net/http"
	"strings"

	"terralist/internal/server/handlers"
	"terralist/internal/server/services"
	"terralist/pkg/api"

	"github.com/gin-gonic/gin"
)

type NetworkMirrorController interface {
	api.RestController
}

type DefaultNetworkMirrorController struct {
	ProviderService services.ProviderService
	Authorization   *handlers.Authorization
	AnonymousRead   bool
}

func (c *DefaultNetworkMirrorController) Paths() []string {
	return []string{""}
}

func (c *DefaultNetworkMirrorController) Subscribe(apis ...*gin.RouterGroup) {
	rootApi := apis[0]
	if !c.AnonymousRead {
		rootApi.Use(c.Authorization.ApiAuthentication())
	}

	rootApi.GET(
		"/:hostname/:namespace/:name/index.json",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")

			provider, err := c.ProviderService.Get(namespace, name)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": err.Error(),
				})
				return
			}

			versions := make(map[string]interface{})
			for _, v := range provider.Versions {
				versions[v.Version] = struct{}{}
			}

			ctx.JSON(http.StatusOK, gin.H{
				"versions": versions,
			})
		},
	)

	rootApi.GET(
		"/:hostname/:namespace/:name/:version.json",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			version := ctx.Param("version")

			versionData, err := c.ProviderService.GetVersionAllPlatforms(namespace, name, version)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": err.Error(),
				})
				return
			}

			archives := make(map[string]interface{})
			for _, platform := range versionData.Platforms {
				key := platform.OS + "_" + platform.Arch
				shasum := platform.Shasum
				if !strings.HasPrefix(shasum, "h1:") {
					shasum = "h1:" + shasum
				}
				archives[key] = gin.H{
					"url": platform.DownloadURL,
					"hashes": []string{shasum},
				}
			}

			ctx.JSON(http.StatusOK, gin.H{
				"archives": archives,
			})
		},
	)
}