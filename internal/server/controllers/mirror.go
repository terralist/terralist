package controllers

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/mirror"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/file"
	"terralist/pkg/rbac"

	"github.com/gin-gonic/gin"
)

const (
	// mirrorProtocolBase is the base path of the Terraform Provider Network
	// Mirror Protocol. The protocol does not use service discovery, so the
	// path is fixed and configured by users in their Terraform CLI config.
	mirrorProtocolBase = "/providers"

	// mirrorApiBase is the base path of the endpoints managing the mirrored
	// providers.
	mirrorApiBase = "/v1/api/mirror"

	// mirrorMetadataField is the multipart field holding the version document
	// produced by `terraform providers mirror`.
	mirrorMetadataField = "metadata"

	// mirrorArchivesField is the multipart field holding the provider packages.
	mirrorArchivesField = "archives"
)

// MirrorController registers the routes that handle the mirrored providers.
type MirrorController interface {
	api.RestController
}

// DefaultMirrorController is a concrete implementation of MirrorController.
type DefaultMirrorController struct {
	MirrorService  services.MirrorService
	Authentication *handlers.Authentication
	Authorization  *handlers.Authorization
	AnonymousRead  bool
}

func (c *DefaultMirrorController) Paths() []string {
	return []string{
		mirrorProtocolBase,
		mirrorApiBase,
	}
}

func (c *DefaultMirrorController) Subscribe(apis ...*gin.RouterGroup) {
	requireAuthorization := c.Authorization.RequireAuthorization(rbac.ResourceMirror)

	// slugComposer builds the RBAC object from all the path parameters that
	// identify the target: hostname, then namespace, then name, if present.
	slugComposer := func(ctx *gin.Context) string {
		parts := []string{ctx.Param("hostname")}
		for _, param := range []string{"namespace", "name"} {
			if value := ctx.Param(param); value != "" {
				parts = append(parts, value)
			}
		}

		return strings.Join(parts, "/")
	}

	// tfApi should be compliant with the Terraform Provider Network Mirror
	// Protocol.
	// Docs: https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol
	tfApi := apis[0]
	if !c.AnonymousRead {
		tfApi.Use(c.Authentication.AttemptAuthentication())
		tfApi.Use(requireAuthorization(rbac.ActionGet, slugComposer))
	}

	tfApi.GET(
		"/:hostname/:namespace/:name/index.json",
		func(ctx *gin.Context) {
			hostname := ctx.Param("hostname")
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")

			dto, err := c.MirrorService.ListVersions(hostname, namespace, name)
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
			hostname := ctx.Param("hostname")
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")

			version, ok := strings.CutSuffix(ctx.Param("version"), ".json")
			if !ok {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			dto, err := c.MirrorService.GetVersion(hostname, namespace, name, version)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, dto)
		},
	)

	// api holds the routes that manage the mirrored providers. Every request
	// must be authenticated.
	api := apis[1]
	api.Use(c.Authentication.AttemptAuthentication())
	api.Use(c.Authentication.RequireAuthentication())

	// Upload the packages of a mirrored provider version
	api.POST(
		"/:hostname/:namespace/:name/:version/upload",
		requireAuthorization(rbac.ActionCreate, slugComposer),
		func(ctx *gin.Context) {
			hostname := ctx.Param("hostname")
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			version := ctx.Param("version")

			form, err := ctx.MultipartForm()
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			metadata, err := parseMirrorMetadata(form.File[mirrorMetadataField])
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			headers := form.File[mirrorArchivesField]
			if len(headers) == 0 {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{fmt.Sprintf("expecting at least one provider package in the %q field", mirrorArchivesField)},
				})
				return
			}

			archives := make([]file.File, 0, len(headers))
			for _, header := range headers {
				uploaded, err := header.Open()
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{
						"errors": []string{"cannot read the uploaded file", err.Error()},
					})
					return
				}
				defer uploaded.Close()

				archives = append(archives, file.NewStreamingFile(header.Filename, uploaded, header.Size))
			}

			if err := c.MirrorService.Upload(hostname, namespace, name, version, *metadata, archives); err != nil {
				ctx.JSON(http.StatusConflict, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"errors": []string{},
			})
		},
	)

	// Delete a mirrored provider version
	api.DELETE(
		"/:hostname/:namespace/:name/:version",
		requireAuthorization(rbac.ActionDelete, slugComposer),
		func(ctx *gin.Context) {
			err := c.MirrorService.DeleteVersion(
				ctx.Param("hostname"),
				ctx.Param("namespace"),
				ctx.Param("name"),
				ctx.Param("version"),
			)

			respondDeletion(ctx, err)
		},
	)

	// Delete a mirrored provider
	api.DELETE(
		"/:hostname/:namespace/:name",
		requireAuthorization(rbac.ActionDelete, slugComposer),
		func(ctx *gin.Context) {
			err := c.MirrorService.Delete(
				ctx.Param("hostname"),
				ctx.Param("namespace"),
				ctx.Param("name"),
			)

			respondDeletion(ctx, err)
		},
	)

	// Delete all mirrored providers under a namespace
	api.DELETE(
		"/:hostname/:namespace",
		requireAuthorization(rbac.ActionDelete, slugComposer),
		func(ctx *gin.Context) {
			err := c.MirrorService.DeleteNamespace(
				ctx.Param("hostname"),
				ctx.Param("namespace"),
			)

			respondDeletion(ctx, err)
		},
	)

	// Delete all mirrored providers under a hostname
	api.DELETE(
		"/:hostname",
		requireAuthorization(rbac.ActionDelete, slugComposer),
		func(ctx *gin.Context) {
			err := c.MirrorService.DeleteHostname(ctx.Param("hostname"))

			respondDeletion(ctx, err)
		},
	)
}

// parseMirrorMetadata decodes the version document from the multipart field.
func parseMirrorMetadata(headers []*multipart.FileHeader) (*mirror.ArchivesDTO, error) {
	if len(headers) != 1 {
		return nil, fmt.Errorf("expecting exactly one version document in the %q field", mirrorMetadataField)
	}

	uploaded, err := headers[0].Open()
	if err != nil {
		return nil, fmt.Errorf("cannot read the version document: %v", err)
	}
	defer uploaded.Close()

	var metadata mirror.ArchivesDTO
	if err := json.NewDecoder(uploaded).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("cannot decode the version document: %v", err)
	}

	return &metadata, nil
}

// respondDeletion writes the response of a deletion request.
func respondDeletion(ctx *gin.Context, err error) {
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"errors": []string{err.Error()},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"errors": []string{},
	})
}
