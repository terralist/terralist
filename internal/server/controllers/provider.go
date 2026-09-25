package controllers

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/file"
	"terralist/pkg/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	providersTerraformApiBase = "/providers"
	providersDefaultApiBase   = "/api/providers"

	// Multipart fields of a provider packages upload.
	packagesMetadataField         = "metadata"
	packagesArchivesField         = "archives"
	packagesShaSumsField          = "shasums"
	packagesShaSumsSignatureField = "shasums_signature"
	packagesProtocolsField        = "protocols"
)

// ProviderController registers the routes that handles the modules.
type ProviderController interface {
	api.RestController

	// TerraformApi returns the endpoint where Terraform can query
	// providers.
	TerraformApi() string
}

// DefaultProviderController is a concrete implementation of ProviderController.
type DefaultProviderController struct {
	ProviderService  services.ProviderService
	AuthorityService services.AuthorityService
	VcsService       services.VcsService
	Authentication   *handlers.Authentication
	Authorization    *handlers.Authorization
	AnonymousRead    bool

	// Tokens signs the download links that point at the network mirror, where
	// packages not stored yet are fetched from. MirrorBaseURL is where those
	// links start.
	Tokens        *handlers.DownloadTokens
	MirrorBaseURL string
}

func (c *DefaultProviderController) Paths() []string {
	return []string{
		providersTerraformApiBase,
		providersDefaultApiBase,
	}
}

func (c *DefaultProviderController) TerraformApi() string {
	return providersTerraformApiBase + "/"
}

func (c *DefaultProviderController) Subscribe(apis ...*gin.RouterGroup) {
	requireAuthorization := c.Authorization.RequireAuthorization(rbac.ResourceProviders)

	slugComposer := func(ctx *gin.Context) string {
		namespace := ctx.Param("namespace")
		name := ctx.Param("name")

		return fmt.Sprintf("%s/%s", namespace, name)
	}

	// tfApi should be compliant with the Terraform Registry Protocol for
	// providers
	// Docs: https://www.terraform.io/docs/internals/provider-registry-protocol.html#find-a-provider-package
	tfApi := apis[0]
	tfApi.Use(c.Authentication.AttemptAuthentication())
	if !c.AnonymousRead {
		tfApi.Use(requireAuthorization(rbac.ActionGet, slugComposer))
	}

	tfApi.GET(
		"/:namespace/:name/versions",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")

			d, err := c.ProviderService.Get(namespace, name, c.mayFetch(ctx, namespace, name))
			if err != nil {
				ctx.JSON(lookupStatus(err), gin.H{
					"errors": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusOK, d)
		},
	)

	// Docs: https://www.terraform.io/docs/internals/provider-registry-protocol.html#find-a-provider-package
	tfApi.GET(
		"/:namespace/:name/:version/download/:os/:arch",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			version := ctx.Param("version")
			os := ctx.Param("os")
			arch := ctx.Param("arch")

			fetch := c.mayFetch(ctx, namespace, name)

			dto, err := c.ProviderService.GetVersion(namespace, name, version, os, arch, fetch)
			if err != nil {
				ctx.JSON(lookupStatus(err), gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			if err := c.signMirrorDownload(dto, namespace, fetch); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, dto)
		},
	)

	// api holds the routes that are not described by the Terraform protocol
	api := apis[1]
	api.POST(
		"/:namespace/:name/webhook/:vcs",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			vcs := ctx.Param("vcs")

			ev, err := c.VcsService.ParseProviderReleaseWebhook(ctx, vcs, namespace, name)
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
			dto, err := c.VcsService.BuildProviderCreateDTO(authority.ID, namespace, name, ev)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"errors": []string{err.Error()}})
				return
			}
			if err := c.ProviderService.Upload(dto); err != nil {
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

	// Upload a new provider version
	api.POST(
		"/:namespace/:name/:version/upload",
		requireAuthorization(rbac.ActionCreate, slugComposer),
		func(ctx *gin.Context) {
			authorityID, ok := c.resolveAuthorityID(ctx)
			if !ok {
				return
			}

			name := ctx.Param("name")
			version := ctx.Param("version")

			var body provider.CreateProviderDTO
			if err := ctx.BindJSON(&body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			body.AuthorityID = authorityID
			body.Name = name
			body.Version = version

			if err := c.ProviderService.Upload(&body); err != nil {
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

	// Upload the packages of a new provider version
	api.POST(
		"/:namespace/:name/:version/upload-files",
		requireAuthorization(rbac.ActionCreate, slugComposer),
		func(ctx *gin.Context) {
			authorityID, ok := c.resolveAuthorityID(ctx)
			if !ok {
				return
			}

			form, err := ctx.MultipartForm()
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			dto := provider.PackagesUploadDTO{
				AuthorityID: authorityID,
				Name:        ctx.Param("name"),
				Version:     ctx.Param("version"),
				Protocols:   splitProtocols(ctx.PostForm(packagesProtocolsField)),
			}

			if err := readPackagesUpload(form, &dto); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}
			defer closeFiles(dto)

			if err := c.ProviderService.UploadPackages(&dto); err != nil {
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

	// Fetch the packages of a version from the upstream registry
	api.POST(
		"/:namespace/:name/:version/fetch",
		requireAuthorization(rbac.ActionCreate, slugComposer),
		func(ctx *gin.Context) {
			var body struct {
				Platforms []string `json:"platforms"`
			}
			if err := ctx.BindJSON(&body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			if len(body.Platforms) == 0 {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{"expecting at least one os_arch platform to fetch"},
				})
				return
			}

			results := c.ProviderService.Fetch(ctx.Param("namespace"), ctx.Param("name"), ctx.Param("version"), body.Platforms)

			ctx.JSON(http.StatusOK, gin.H{
				"results": results,
			})
		},
	)

	// Delete a provider
	api.DELETE(
		"/:namespace/:name/remove",
		requireAuthorization(rbac.ActionDelete, slugComposer),
		func(ctx *gin.Context) {
			authorityID, ok := c.resolveAuthorityID(ctx)
			if !ok {
				return
			}

			name := ctx.Param("name")

			if err := c.ProviderService.Delete(authorityID, name); err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
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
		"/:namespace/:name/:version/remove",
		requireAuthorization(rbac.ActionDelete, slugComposer),
		func(ctx *gin.Context) {
			authorityID, ok := c.resolveAuthorityID(ctx)
			if !ok {
				return
			}

			name := ctx.Param("name")
			version := ctx.Param("version")

			if err := c.ProviderService.DeleteVersion(authorityID, name, version); err != nil {
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

// signMirrorDownload appends a package token to a download URL pointing at the
// network mirror, since Terraform downloads packages without credentials.
func (c *DefaultProviderController) signMirrorDownload(dto *provider.DownloadPlatformDTO, namespace string, fetch bool) error {
	if c.MirrorBaseURL == "" || !strings.HasPrefix(dto.DownloadUrl, c.MirrorBaseURL+"/") {
		return nil
	}

	pkg, ok := provider.ParsePackageFileName(dto.FileName)
	if !ok {
		return nil
	}

	token, err := c.Tokens.Sign(pkg.Subject(namespace), fetch)
	if err != nil {
		return err
	}

	dto.DownloadUrl = fmt.Sprintf("%s?token=%s", dto.DownloadUrl, token)

	return nil
}

// mayFetch reports whether the caller may create packages of the provider,
// which is what fetching them from the upstream registry amounts to.
func (c *DefaultProviderController) mayFetch(ctx *gin.Context, namespace, name string) bool {
	user, err := handlers.GetFromContext[auth.User](ctx, "user")
	if err != nil {
		return false
	}

	return c.Authorization.CanPerform(*user, rbac.ResourceProviders, rbac.ActionCreate, fmt.Sprintf("%s/%s", namespace, name))
}

// readPackagesUpload fills the upload with the files of the multipart form: the
// version document, the package archives and, when present, the SHA256SUMS
// file with its signature. The caller owns the opened files.
func readPackagesUpload(form *multipart.Form, dto *provider.PackagesUploadDTO) error {
	metadata := form.File[packagesMetadataField]
	if len(metadata) != 1 {
		return fmt.Errorf("expecting exactly one version document in the %q field", packagesMetadataField)
	}

	document, err := metadata[0].Open()
	if err != nil {
		return fmt.Errorf("cannot read the version document: %v", err)
	}
	defer document.Close()

	if err := json.NewDecoder(document).Decode(&dto.Metadata); err != nil {
		return fmt.Errorf("cannot decode the version document: %v", err)
	}

	if len(form.File[packagesArchivesField]) == 0 {
		return fmt.Errorf("expecting at least one provider package in the %q field", packagesArchivesField)
	}

	for _, header := range form.File[packagesArchivesField] {
		archive, err := openUploadedFile(header)
		if err != nil {
			return err
		}

		dto.Archives = append(dto.Archives, archive)
	}

	if dto.ShaSums, err = openOptionalFile(form, packagesShaSumsField); err != nil {
		return err
	}

	if dto.ShaSumsSignature, err = openOptionalFile(form, packagesShaSumsSignatureField); err != nil {
		return err
	}

	return nil
}

// openOptionalFile opens the single file of a multipart field, or returns nil
// when the field is absent.
func openOptionalFile(form *multipart.Form, field string) (file.File, error) {
	headers := form.File[field]
	if len(headers) == 0 {
		return nil, nil
	}

	if len(headers) > 1 {
		return nil, fmt.Errorf("expecting at most one file in the %q field", field)
	}

	return openUploadedFile(headers[0])
}

func openUploadedFile(header *multipart.FileHeader) (file.File, error) {
	uploaded, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("cannot read the uploaded file %s: %v", header.Filename, err)
	}

	return file.NewStreamingFile(header.Filename, uploaded, header.Size), nil
}

// closeFiles closes every file opened for a packages upload.
func closeFiles(dto provider.PackagesUploadDTO) {
	for _, f := range dto.Archives {
		_ = f.Close()
	}

	for _, f := range []file.File{dto.ShaSums, dto.ShaSumsSignature} {
		if f != nil {
			_ = f.Close()
		}
	}
}

// splitProtocols parses the comma separated provider protocols form value.
func splitProtocols(value string) []string {
	var protocols []string
	for _, protocol := range strings.Split(value, ",") {
		if protocol = strings.TrimSpace(protocol); protocol != "" {
			protocols = append(protocols, protocol)
		}
	}

	return protocols
}

// resolveAuthorityID resolves the authority ID from the namespace URL parameter.
func (c *DefaultProviderController) resolveAuthorityID(ctx *gin.Context) (uuid.UUID, bool) {
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
