package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"terralist/pkg/api"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/storage"
	"terralist/pkg/storage/local"

	"github.com/gin-gonic/gin"
)

// FileServer registers the routes that handles the files
type FileServer interface {
	api.RestController
}

// DefaultFileServer is a concrete implementation of FileServer
type DefaultFileServer struct {
	ModulesResolver   storage.Resolver
	ProvidersResolver storage.Resolver

	JWT jwt.JWT
}

func (c *DefaultFileServer) Paths() []string {
	return []string{"/files"}
}

func (c *DefaultFileServer) Subscribe(apis ...*gin.RouterGroup) {
	api := apis[0]

	localModulesResolver, ok := c.ModulesResolver.(*local.Resolver)
	if !ok {
		localModulesResolver = nil
	}

	localProvidersResolver, ok := c.ProvidersResolver.(*local.Resolver)
	if !ok {
		localProvidersResolver = nil
	}

	api.GET("/modules/*filepath", func(ctx *gin.Context) {
		if localModulesResolver == nil {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		fileKey := filepath.Join("modules", ctx.Param("filepath"))
		token := ctx.Query("token")

		if _, err := c.JWT.Extract(token); err != nil {
			ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		file, err := localModulesResolver.GetObject(fileKey)
		if err != nil {
			if !errors.Is(err, local.ErrFileNotFound) {
				ctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
		ctx.Header("Content-Type", "application/octet-stream")

		ctx.Writer.Write(file.Content)

		ctx.Status(http.StatusOK)
	})

	api.GET("/providers/*filepath", func(ctx *gin.Context) {
		if localProvidersResolver == nil {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		fileKey := filepath.Join("modules", ctx.Param("filepath"))
		token := ctx.Query("token")

		if _, err := c.JWT.Extract(token); err != nil {
			ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		file, err := localProvidersResolver.GetObject(fileKey)
		if err != nil {
			if !errors.Is(err, local.ErrFileNotFound) {
				ctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
		ctx.Header("Content-Type", "application/octet-stream")

		ctx.Writer.Write(file.Content)

		ctx.Status(http.StatusOK)
	})
}
