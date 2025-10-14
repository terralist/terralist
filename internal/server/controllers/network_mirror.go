package controllers

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"terralist/internal/server/handlers"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/rbac"

	"github.com/gin-gonic/gin"
	"golang.org/x/mod/sumdb/dirhash"
)

type NetworkMirrorController interface {
	api.RestController
}

type DefaultNetworkMirrorController struct {
	ProviderService services.ProviderService
	Authentication  *handlers.Authentication
	Authorization   *handlers.Authorization
	AnonymousRead   bool
}

func (c *DefaultNetworkMirrorController) Paths() []string {
	return []string{""}
}

func parseSHASUMS(content string, filename string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			file := parts[1]
			if file == filename {
				return hash, nil
			}
		}
	}
	return "", fmt.Errorf("hash not found for %s", filename)
}

func fetchSHASUMS(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch SHASUMS: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func computeH1Hash(zipURL string) (string, error) {
	resp, err := http.Get(zipURL)
	if err != nil {
		return "", fmt.Errorf("failed to download zip: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download zip: status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "provider-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write zip: %v", err)
	}

	h1Hash, err := dirhash.HashZip(tmpFile.Name(), dirhash.DefaultHash)
	if err != nil {
		return "", fmt.Errorf("failed to compute h1 hash: %v", err)
	}

	return h1Hash, nil
}

func (c *DefaultNetworkMirrorController) Subscribe(apis ...*gin.RouterGroup) {
	requireAuthorization := c.Authorization.RequireAuthorization(rbac.ResourceProviders)

	fullSlugComposer := func(ctx *gin.Context) string {
		namespace := ctx.Param("namespace")
		name := ctx.Param("name")

		return fmt.Sprintf("%s/%s", namespace, name)
	}

	rootApi := apis[0]
	if !c.AnonymousRead {
		rootApi.Use(c.Authentication.RequireAuthentication())
		rootApi.Use(requireAuthorization(rbac.ActionGet, fullSlugComposer))
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
		"/:hostname/:namespace/:name/:version",
		func(ctx *gin.Context) {
			namespace := ctx.Param("namespace")
			name := ctx.Param("name")
			versionWithExt := ctx.Param("version")
			version := strings.TrimSuffix(versionWithExt, ".json")

			versionData, err := c.ProviderService.GetVersionAllPlatforms(namespace, name, version)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": err.Error(),
				})
				return
			}

			var shasums string
			var fetchErr error
			if versionData.ShaSumsUrl != "" {
				shasums, fetchErr = fetchSHASUMS(versionData.ShaSumsUrl)
				if fetchErr != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{
						"errors": fmt.Sprintf("failed to fetch SHASUMS: %v", fetchErr),
					})
					return
				}
			}

			archives := make(map[string]interface{})
			for _, platform := range versionData.Platforms {
				key := platform.OS + "_" + platform.Arch

				filename := fmt.Sprintf("terraform-provider-%s_%s_%s_%s.zip", name, version, platform.OS, platform.Arch)

				hashes := []string{}

				if platform.DownloadURL != "" {
					h1Hash, err := computeH1Hash(platform.DownloadURL)
					if err != nil {
						ctx.JSON(http.StatusInternalServerError, gin.H{
							"errors": fmt.Sprintf("failed to compute h1 hash: %v", err),
						})
						return
					}
					hashes = append(hashes, h1Hash)
				}

				if shasums != "" {
					rawHash, fetchErr := parseSHASUMS(shasums, filename)
					if fetchErr == nil {
						hashes = append(hashes, "zh:"+rawHash)
					}
				}

				archives[key] = gin.H{
					"url":    platform.DownloadURL,
					"hashes": hashes,
				}
			}

			ctx.JSON(http.StatusOK, gin.H{
				"archives": archives,
			})
		},
	)
}
