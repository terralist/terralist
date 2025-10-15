package controllers

import (
	"fmt"
	"net/http"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/module"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/file"
	"terralist/pkg/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	modulesTerraformApiBase = "/modules"
	modulesDefaultApiBase   = "/api/modules"
)

// ModuleController registers the routes that handles the modules.
type ModuleController interface {
	api.RestController

	// TerraformApi returns the endpoint where Terraform can query
	// modules.
	TerraformApi() string
}

// DefaultModuleController is a concrete implementation of ModuleController.
type DefaultModuleController struct {
	ModuleService  services.ModuleService
	Authentication *handlers.Authentication
	Authorization  *handlers.Authorization
	AnonymousRead  bool
	HomeDir        string
}

func (c *DefaultModuleController) TerraformApi() string {
	return modulesTerraformApiBase + "/"
}

func (c *DefaultModuleController) Paths() []string {
	return []string{
		modulesTerraformApiBase,
		modulesDefaultApiBase,
	}
}

func (c *DefaultModuleController) Subscribe(apis ...*gin.RouterGroup) {
	requireAuthorization := c.Authorization.RequireAuthorization(rbac.ResourceModules)

	fullSlugComposer := func(ctx *gin.Context) string {
		namespace := ctx.Param("namespace")
		name := ctx.Param("name")
		provider := ctx.Param("provider")

		return fmt.Sprintf("%s/%s/%s", namespace, name, provider)
	}

	partialSlugComposer := func(ctx *gin.Context) string {
		name := ctx.Param("name")
		provider := ctx.Param("provider")

		return fmt.Sprintf("%s/%s/%s", ctx.GetString("authorityName"), name, provider)
	}

	// tfApi should be compliant with the Terraform Registry Protocol for
	// modules
	// Docs: https://www.terraform.io/docs/internals/module-registry-protocol.html#list-available-versions-for-a-specific-module
	tfApi := apis[0]
	if !c.AnonymousRead {
		tfApi.Use(c.Authentication.AttemptAuthentication())
		tfApi.Use(requireAuthorization(rbac.ActionGet, fullSlugComposer))
	}

	tfApi.GET(
		"/:namespace/:name/:provider/versions",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			provider := ctx.Param("provider")

			d, err := c.ModuleService.Get(namespace, name, provider)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusOK, d)
		},
	)

	// Docs: https://www.terraform.io/docs/internals/module-registry-protocol.html#download-source-code-for-a-specific-module-version
	tfApi.GET(
		"/:namespace/:name/:provider/:version/download",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

			location, err := c.ModuleService.GetVersionURL(namespace, name, provider, version)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.Header("X-Terraform-Get", *location)
			ctx.JSON(http.StatusNoContent, gin.H{
				"errors": []string{},
			})
		},
	)

	// api holds the routes that are not described by the Terraform protocol
	api := apis[1]
	api.Use(c.Authentication.AttemptAuthentication())

	// This is a protected endpoint, every request should be authenticated.
	api.Use(c.Authentication.RequireAuthentication())

	// Upload a new module version
	api.POST(
		"/:name/:provider/:version/upload",
		handlers.RequireAuthority(),
		requireAuthorization(rbac.ActionCreate, partialSlugComposer),
		func(ctx *gin.Context) {
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

			authorityID, err := uuid.Parse(ctx.GetString("authorityID"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"invalid authority"},
				})
				return
			}

			dto := module.CreateDTO{
				AuthorityID: authorityID,
				Name:        name,
				Provider:    provider,
				VersionCreateDTO: module.VersionCreateDTO{
					Version: version,
				},
			}

			var body module.CreateFromURLDTO
			if err := ctx.BindJSON(&body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			header := file.CreateHeader(body.Headers)

			if err := c.ModuleService.Upload(&dto, body.DownloadUrl, header); err != nil {
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

	// Upload a new module version (with files)
	api.POST(
		"/:name/:provider/:version/upload-files",
		handlers.RequireAuthority(),
		requireAuthorization(rbac.ActionCreate, partialSlugComposer),
		func(ctx *gin.Context) {
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

			authorityID, err := uuid.Parse(ctx.GetString("authorityID"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"invalid authority"},
				})
				return
			}

			form, err := ctx.MultipartForm()
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			moduleFiles := form.File["module"]

			if len(moduleFiles) != 1 {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"expecting exactly one archive file containing the module"},
				})
				return
			}

			// Create a temp file
			onDiskFile, err := file.SaveToDisk(file.NewFromMultipartFileHeader(moduleFiles[0]), c.HomeDir)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"errors": []string{"cannot allocate a new file", err.Error()},
				})
				return
			}

			defer func() {
				if err := onDiskFile.Close(); err != nil {
					log.Error().
						Err(err).
						Str("artifact", "module").
						Str("name", name).
						Str("provider", provider).
						Str("version", version).
						Str("file", onDiskFile.Name()).
						Str("filepath", onDiskFile.Path()).
						Msg("could not close the file")
				}

				if err := onDiskFile.Remove(); err != nil {
					log.Error().
						Err(err).
						Str("artifact", "module").
						Str("name", name).
						Str("provider", provider).
						Str("version", version).
						Str("file", onDiskFile.Name()).
						Str("filepath", onDiskFile.Path()).
						Msg("could not remove module temp disk file")
				}
			}()

			// Write form content to the temp file
			if err := ctx.SaveUploadedFile(moduleFiles[0], onDiskFile.Path()); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"errors": []string{"cannot save content to the local disk", err.Error()},
				})
				return
			}

			dto := module.CreateDTO{
				AuthorityID: authorityID,
				Name:        name,
				Provider:    provider,
				VersionCreateDTO: module.VersionCreateDTO{
					Version: version,
				},
			}

			// Pass-in local-file URI for go-getter
			uri := fmt.Sprintf("file://%v", onDiskFile.Path())
			if err := c.ModuleService.Upload(&dto, uri, nil); err != nil {
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

	// Delete a provider
	api.DELETE(
		"/:name/:provider/remove",
		handlers.RequireAuthority(),
		requireAuthorization(rbac.ActionDelete, partialSlugComposer),
		func(ctx *gin.Context) {
			name := ctx.Param("name")
			provider := ctx.Param("provider")

			authorityID, err := uuid.Parse(ctx.GetString("authorityID"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"invalid authority"},
				})
				return
			}

			if err := c.ModuleService.Delete(authorityID, name, provider); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"errors": []string{},
			})
		},
	)

	// Delete a provider version
	api.DELETE(
		"/:name/:provider/:version/remove",
		handlers.RequireAuthority(),
		requireAuthorization(rbac.ActionDelete, partialSlugComposer),
		func(ctx *gin.Context) {
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

			authorityID, err := uuid.Parse(ctx.GetString("authorityID"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"invalid authority"},
				})
				return
			}

			if err := c.ModuleService.DeleteVersion(authorityID, name, provider, version); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"errors": []string{},
			})
		},
	)
}
