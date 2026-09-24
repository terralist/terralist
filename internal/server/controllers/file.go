package controllers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"terralist/pkg/api"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/storage"
	"terralist/pkg/storage/local"

	"github.com/gin-gonic/gin"
)

// FileServer registers the routes that handle the files.
type FileServer interface {
	api.RestController
}

// DefaultFileServer is a concrete implementation of FileServer.
type DefaultFileServer struct {
	ModulesResolver   storage.Resolver
	ProvidersResolver storage.Resolver

	JWT jwt.JWT
}

type downloadTokenPayload struct {
	Key string `json:"key"`
}

func extractTokenKey(data jwt.Serializer) (string, bool) {
	if payload, ok := data.(*downloadTokenPayload); ok && payload != nil && payload.Key != "" {
		return payload.Key, true
	}

	if payload, ok := data.(map[string]any); ok {
		if key, ok := payload["key"].(string); ok && key != "" {
			return key, true
		}
	}

	return "", false
}

func (c *DefaultFileServer) Paths() []string {
	return []string{"/files"}
}

func (c *DefaultFileServer) Subscribe(apis ...*gin.RouterGroup) {
	api := apis[0]

	localModulesResolver := local.UnwrapResolver(c.ModulesResolver)
	localProvidersResolver := local.UnwrapResolver(c.ProvidersResolver)

	api.GET("/*filepath", func(ctx *gin.Context) {
		fileKey := strings.TrimPrefix(ctx.Param("filepath"), "/")
		if fileKey == "" {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		token := ctx.Query("token")
		data, err := c.JWT.Extract(token)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		tokenKey, ok := extractTokenKey(data)
		if !ok || tokenKey != fileKey {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var resolver *local.Resolver
		switch {
		case strings.HasPrefix(fileKey, "modules/"):
			resolver = localModulesResolver
		case strings.HasPrefix(fileKey, "providers/"):
			resolver = localProvidersResolver
		default:
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		if resolver == nil {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		file, err := resolver.GetObject(fileKey)
		if err != nil {
			if !errors.Is(err, local.ErrFileNotFound) {
				_ = ctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name()))
		ctx.Header("Content-Type", "application/octet-stream")

		if _, err := io.Copy(ctx.Writer, file); err != nil {
			_ = ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.Status(http.StatusOK)
	})
}
