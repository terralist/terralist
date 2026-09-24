package e2e

import (
	"net/http"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderListVersions(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/versions"))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/nonexistent/versions"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/versions"), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, body, "versions")

		versions, ok := body["versions"].([]any)
		require.True(t, ok)

		byVersion := map[string]map[string]any{}
		for _, v := range versions {
			version, ok := v.(map[string]any)
			require.True(t, ok)

			number, ok := version["version"].(string)
			require.True(t, ok)
			byVersion[number] = version
		}

		assert.Contains(t, byVersion, nullSignedVersion, "a version uploaded with its signature material is served by the registry")
		assert.NotContains(t, byVersion, nullMirrorOnlyVersion, "a version uploaded without signature material is served by the mirror only")
		require.Contains(t, byVersion, "3.2.4")

		version := byVersion["3.2.4"]

		platforms, ok := version["platforms"].([]any)
		require.True(t, ok)
		require.Len(t, platforms, 1)

		platform, ok := platforms[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, runtime.GOOS, platform["os"])
		assert.Equal(t, runtime.GOARCH, platform["arch"])
	})
}

func TestProviderDownload(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/3.2.4/download/%s/%s", runtime.GOOS, runtime.GOARCH))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("not found provider", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/nonexistent/3.2.4/download/%s/%s", runtime.GOOS, runtime.GOARCH), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("not found version", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/9.9.9/download/%s/%s", runtime.GOOS, runtime.GOARCH), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("mirror-only version", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/%s/download/%s/%s", nullMirrorOnlyVersion, runtime.GOOS, runtime.GOARCH), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("signed packages upload", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/%s/download/%s/%s", nullSignedVersion, runtime.GOOS, runtime.GOARCH), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, []any{"5.0"}, body["protocols"])
		assert.Contains(t, body, "download_url")
		assert.Contains(t, body, "shasums_url")
		assert.Contains(t, body, "shasums_signature_url")
	})

	t.Run("success", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/3.2.4/download/%s/%s", runtime.GOOS, runtime.GOARCH), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, runtime.GOOS, body["os"])
		assert.Equal(t, runtime.GOARCH, body["arch"])
		assert.Contains(t, body, "download_url")
		assert.Contains(t, body, "shasum")
		assert.Contains(t, body, "shasums_url")
		assert.Contains(t, body, "shasums_signature_url")
		assert.Contains(t, body, "signing_keys")
	})
}

func TestProviderUpload(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodPost, apiURL("/v1/api/providers/hashicorp/null/4.0.0/upload"))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("no body", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodPost, apiURL("/v1/api/providers/hashicorp/null/4.0.0/upload"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("duplicate version", func(t *testing.T) {
		body := map[string]any{
			"protocols": []string{"6.0"},
			"shasums": map[string]string{
				"url":           "https://example.com/shasums",
				"signature_url": "https://example.com/shasums.sig",
			},
			"platforms": []map[string]string{
				{
					"os":           runtime.GOOS,
					"arch":         runtime.GOARCH,
					"download_url": "https://example.com/provider.zip",
					"shasum":       "abc123",
				},
			},
		}
		resp := doAuthRequest(t, http.MethodPost, apiURL("/v1/api/providers/hashicorp/null/3.2.4/upload"), body)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})
}

func TestProviderUploadPackages(t *testing.T) {
	url := apiURL("/v1/api/providers/hashicorp/null/4.0.0/upload-files")
	files := func() []multipartFile {
		return packagesUploadFiles("4.0.0", bootstrap.NullMirrorOnlyH1, bootstrap.NullMirrorOnlyArchive)
	}

	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodPost, url)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("not multipart", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodPost, url, map[string]any{})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing metadata", func(t *testing.T) {
		resp := doAuthMultipartRequest(t, url, files()[1:])
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("duplicate version", func(t *testing.T) {
		resp := doAuthMultipartRequest(t, apiURL("/v1/api/providers/hashicorp/null/%s/upload-files", nullMirrorOnlyVersion), files())
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("package not listed in the version document", func(t *testing.T) {
		mismatched := files()
		mismatched[1].FileName = "other.zip"

		resp := doAuthMultipartRequest(t, url, mismatched)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("shasums without signature", func(t *testing.T) {
		withSums := append(files(), multipartFile{Field: "shasums", FileName: "SHA256SUMS", Content: []byte("")})

		resp := doAuthMultipartRequestWithValues(t, url, withSums, map[string]string{"protocols": "5.0"})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("shasums mismatch", func(t *testing.T) {
		sums := "0000000000000000000000000000000000000000000000000000000000000000  " + nullProviderArchiveName("4.0.0") + "\n"
		withSums := append(files(),
			multipartFile{Field: "shasums", FileName: "SHA256SUMS", Content: []byte(sums)},
			multipartFile{Field: "shasums_signature", FileName: "SHA256SUMS.sig", Content: []byte("signature")},
		)

		resp := doAuthMultipartRequestWithValues(t, url, withSums, map[string]string{"protocols": "5.0"})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		index := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/null/index.json", mirrorHostname(t)), nil)
		body := readJSON(t, index)
		versions, _ := body["versions"].(map[string]any)
		assert.NotContains(t, versions, "4.0.0")
	})
}

func TestProviderDeleteUnauthenticated(t *testing.T) {
	t.Run("delete version", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodDelete, apiURL("/v1/api/providers/hashicorp/null/3.2.4/remove"))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("delete provider", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodDelete, apiURL("/v1/api/providers/hashicorp/null/remove"))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
