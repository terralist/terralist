package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// config holds the e2e test configuration, populated from environment variables.
var config struct {
	URL          string
	MasterAPIKey string
}

func initConfig() {
	config.URL = envOrDefault("TERRALIST_URL", "http://localhost:5758")
	config.MasterAPIKey = envOrDefault("TERRALIST_MASTER_API_KEY", "e2e-master-api-key-00000000-0000-0000-0000-000000000000")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// httpClient returns a shared HTTP client with reasonable timeouts.
func httpClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
	}
}

// apiURL constructs a full URL from a path.
func apiURL(path string, args ...any) string {
	if len(args) > 0 {
		path = fmt.Sprintf(path, args...)
	}
	return config.URL + path
}

// doRequest executes an HTTP request and returns the response.
func doRequest(t *testing.T, method, url string, body any, headers map[string]string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient().Do(req)
	require.NoError(t, err)

	return resp
}

// doAuthRequest executes an authenticated HTTP request using the master API key.
func doAuthRequest(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()
	return doRequest(t, method, url, body, map[string]string{
		"Authorization": "Bearer x-api-key:" + config.MasterAPIKey,
	})
}

// doAPIKeyRequest executes an HTTP request authenticated with a specific API key.
func doAPIKeyRequest(t *testing.T, method, url string, body any, apiKey string) *http.Response {
	t.Helper()
	return doRequest(t, method, url, body, map[string]string{
		"Authorization": "Bearer x-api-key:" + apiKey,
	})
}

// doUnauthRequest executes an unauthenticated HTTP request.
func doUnauthRequest(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()
	return doRequest(t, method, url, body, nil)
}

// readJSON reads the response body and unmarshals it into a map.
func readJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	if len(data) == 0 {
		return nil
	}

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result), "response body: %s", string(data))
	return result
}

// readBody reads the response body as a string.
func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(data)
}

// uploadFile sends a multipart form POST with a file attachment.
func uploadFile(t *testing.T, url, fieldName, filePath string, headers map[string]string) *http.Response {
	t.Helper()

	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	require.NoError(t, err)

	_, err = io.Copy(part, file)
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	require.NoError(t, err)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient().Do(req)
	require.NoError(t, err)
	return resp
}

// uploadAuthFile sends an authenticated multipart form POST with a file attachment.
func uploadAuthFile(t *testing.T, url, fieldName, filePath string) *http.Response {
	t.Helper()
	return uploadFile(t, url, fieldName, filePath, map[string]string{
		"Authorization": "Bearer x-api-key:" + config.MasterAPIKey,
	})
}
