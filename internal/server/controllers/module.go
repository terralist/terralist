package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/module"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/file"
	"terralist/pkg/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	ModuleService    services.ModuleService
	AuthorityService services.AuthorityService
	VcsService       services.VcsService
	Authentication   *handlers.Authentication
	Authorization    *handlers.Authorization

	AnonymousRead bool

	// Tokens signs the archive locations pointing back at Terralist, where
	// versions not stored yet are fetched from the upstream, since go-getter
	// downloads them without credentials. ArchiveBaseURL is where those
	// locations start.
	Tokens         *handlers.DownloadTokens
	ArchiveBaseURL string
}

const (
	// moduleFetchKey is the context key holding the fetch permission carried
	// by a valid download token, when the request presented one.
	moduleFetchKey = "moduleFetch"

	// downloadTokenQuery is the query parameter carrying a download token.
	downloadTokenQuery = "token"
)

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

	slugComposer := func(ctx *gin.Context) string {
		namespace := ctx.Param("namespace")
		name := ctx.Param("name")
		provider := ctx.Param("provider")

		return fmt.Sprintf("%s/%s/%s", namespace, name, provider)
	}

	// tfApi should be compliant with the Terraform Registry Protocol for
	// modules
	// Docs: https://www.terraform.io/docs/internals/module-registry-protocol.html#list-available-versions-for-a-specific-module
	tfApi := apis[0]
	tfApi.Use(c.Authentication.AttemptAuthentication())
	tfApi.Use(c.acceptDownloadToken())
	if !c.AnonymousRead {
		authorize := requireAuthorization(rbac.ActionGet, slugComposer)
		tfApi.Use(func(ctx *gin.Context) {
			// A valid download token is the proof that the download location
			// was served to an authorized caller.
			if _, ok := ctx.Get(moduleFetchKey); ok {
				ctx.Next()
				return
			}

			authorize(ctx)
		})
	}

	tfApi.GET(
		"/:namespace/:name/:provider/versions",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			provider := ctx.Param("provider")

			d, err := c.ModuleService.Get(namespace, name, provider, c.mayFetch(ctx, namespace, name, provider))
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

			fetch := c.mayFetch(ctx, namespace, name, provider)

			location, err := c.ModuleService.GetVersionURL(namespace, name, provider, version, fetch)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			signed, err := c.signArchiveLocation(*location, namespace, name, provider, version, fetch)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.Header("X-Terraform-Get", signed)
			ctx.JSON(http.StatusNoContent, gin.H{
				"errors": []string{},
			})
		},
	)

	// Serve the archive of a version, fetching it from the upstream registry
	// first when the caller may do so, by credentials or by the token of the
	// location. go-getter follows the X-Terraform-Get header to storage.
	tfApi.GET(
		"/:namespace/:name/:provider/:version/archive",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

			fetch := c.mayFetch(ctx, namespace, name, provider)
			if granted, err := handlers.GetFromContext[bool](ctx, moduleFetchKey); err == nil {
				fetch = fetch || *granted
			}

			location, err := c.ModuleService.Download(namespace, name, provider, version, fetch)

			switch {
			case err == nil:
				ctx.Header("X-Terraform-Get", location)
				ctx.Status(http.StatusNoContent)
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
		},
	)

	// api holds the routes that are not described by the Terraform protocol
	api := apis[1]
	api.POST(
		"/:namespace/:name/:provider/webhook/:vcs",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			vcs := ctx.Param("vcs")

			ev, err := c.VcsService.ParseModuleReleaseWebhook(ctx, vcs, namespace, name, provider)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"errors": []string{err.Error()}})
				return
			}
			if ev == nil {
				ctx.JSON(http.StatusOK, gin.H{"errors": []string{}})
				return
			}

			authority, err := c.AuthorityService.GetByName(namespace)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{"errors": []string{fmt.Sprintf("authority %q not found", namespace)}})
				return
			}

			dto := module.CreateDTO{
				AuthorityID: authority.ID,
				Name:        name,
				Provider:    provider,
				VersionCreateDTO: module.VersionCreateDTO{
					Version: ev.SemVer,
				},
			}

			headers := c.VcsService.GetHeaders()
			header := file.CreateHeader(headers)
			if err := c.ModuleService.Upload(&dto, file.NewRemoteFile(ev.ModuleArchiveURL, header)); err != nil {
				ctx.JSON(http.StatusConflict, gin.H{"errors": []string{err.Error()}})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"errors": []string{},
			})
		},
	)

	api.Use(c.Authentication.AttemptAuthentication())

	// This is a protected endpoint, every request should be authenticated.
	api.Use(c.Authentication.RequireAuthentication())

	// Fetch a version from the upstream registry
	api.POST(
		"/:namespace/:name/:provider/:version/fetch",
		requireAuthorization(rbac.ActionCreate, slugComposer),
		func(ctx *gin.Context) {
			err := c.ModuleService.Fetch(ctx.Param("namespace"), ctx.Param("name"), ctx.Param("provider"), ctx.Param("version"))
			if err != nil {
				status := http.StatusBadGateway
				if isNotFound(err) {
					status = http.StatusNotFound
				}

				ctx.JSON(status, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"errors": []string{},
			})
		},
	)

	// Upload a new module version
	api.POST(
		"/:namespace/:name/:provider/:version/upload",
		requireAuthorization(rbac.ActionCreate, slugComposer),
		func(ctx *gin.Context) {
			authorityID, ok := c.resolveAuthorityID(ctx)
			if !ok {
				return
			}

			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

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

			if err := c.ModuleService.Upload(&dto, file.NewRemoteFile(body.DownloadUrl, header)); err != nil {
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
		"/:namespace/:name/:provider/:version/upload-files",
		requireAuthorization(rbac.ActionCreate, slugComposer),
		func(ctx *gin.Context) {
			authorityID, ok := c.resolveAuthorityID(ctx)
			if !ok {
				return
			}

			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

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

			// Open the uploaded archive for streaming to the service
			uploaded, err := moduleFiles[0].Open()
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"errors": []string{"cannot read the uploaded file", err.Error()},
				})
				return
			}
			defer uploaded.Close()

			dto := module.CreateDTO{
				AuthorityID: authorityID,
				Name:        name,
				Provider:    provider,
				VersionCreateDTO: module.VersionCreateDTO{
					Version: version,
				},
			}

			archive := file.NewStreamingFile(moduleFiles[0].Filename, uploaded, moduleFiles[0].Size)

			if err := c.ModuleService.Upload(&dto, archive); err != nil {
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

	// Get submodule documentation for a specific module version
	api.GET(
		"/:namespace/:name/:provider/:version/submodules/*submodulePath",
		requireAuthorization(rbac.ActionGet, slugComposer),
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")
			submodulePath := strings.TrimPrefix(ctx.Param("submodulePath"), "/")

			doc, err := c.ModuleService.GetSubmoduleDocumentation(namespace, name, provider, version, submodulePath)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"documentation": doc,
			})
		},
	)

	// Delete a module
	api.DELETE(
		"/:namespace/:name/:provider/remove",
		requireAuthorization(rbac.ActionDelete, slugComposer),
		func(ctx *gin.Context) {
			authorityID, ok := c.resolveAuthorityID(ctx)
			if !ok {
				return
			}

			name := ctx.Param("name")
			provider := ctx.Param("provider")

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

	// Delete a module version
	api.DELETE(
		"/:namespace/:name/:provider/:version/remove",
		requireAuthorization(rbac.ActionDelete, slugComposer),
		func(ctx *gin.Context) {
			authorityID, ok := c.resolveAuthorityID(ctx)
			if !ok {
				return
			}

			name := ctx.Param("name")
			provider := ctx.Param("provider")
			version := ctx.Param("version")

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

// resolveAuthorityID resolves the authority ID from the namespace URL parameter.
func (c *DefaultModuleController) resolveAuthorityID(ctx *gin.Context) (uuid.UUID, bool) {
	namespace := ctx.Param("namespace")

	authority, err := c.AuthorityService.GetByName(namespace)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"errors": []string{fmt.Sprintf("authority %q not found", namespace)},
		})
		return uuid.UUID{}, false
	}

	return authority.ID, true
}

// signArchiveLocation appends a download token to a location pointing at the
// archive route, since go-getter downloads without credentials.
func (c *DefaultModuleController) signArchiveLocation(location, namespace, name, provider, version string, fetch bool) (string, error) {
	if c.ArchiveBaseURL == "" || !strings.HasPrefix(location, c.ArchiveBaseURL+"/") {
		return location, nil
	}

	token, err := c.Tokens.Sign(module.ArchiveSubject(namespace, name, provider, version), fetch)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s?%s=%s", location, downloadTokenQuery, token), nil
}

// acceptDownloadToken verifies the token an archive request may carry and
// records the fetch permission it grants.
func (c *DefaultModuleController) acceptDownloadToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Query(downloadTokenQuery)
		if token == "" || !strings.HasSuffix(ctx.Request.URL.Path, "/archive") {
			ctx.Next()
			return
		}

		subject := module.ArchiveSubject(ctx.Param("namespace"), ctx.Param("name"), ctx.Param("provider"), ctx.Param("version"))
		if fetch, ok := c.Tokens.Verify(token, subject); ok {
			ctx.Set(moduleFetchKey, &fetch)
		}

		ctx.Next()
	}
}

// mayFetch reports whether the caller may create versions of the module, which
// is what fetching them from the upstream registry amounts to.
func (c *DefaultModuleController) mayFetch(ctx *gin.Context, namespace, name, provider string) bool {
	user, err := handlers.GetFromContext[auth.User](ctx, "user")
	if err != nil {
		return false
	}

	return c.Authorization.CanPerform(*user, rbac.ResourceModules, rbac.ActionCreate, fmt.Sprintf("%s/%s/%s", namespace, name, provider))
}
