package e2e

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// config holds the e2e test configuration, populated from environment variables.
var config struct {
	URL           string
	MetricsURL    string
	MasterAPIKey  string
	WebhookSecret string
}

func initConfig() {
	config.URL = envOrDefault("TERRALIST_URL", "http://localhost:5758")
	config.MetricsURL = envOrDefault("TERRALIST_METRICS_URL", "http://localhost:9090")
	config.MasterAPIKey = envOrDefault("TERRALIST_MASTER_API_KEY", "e2e-master-api-key-00000000-0000-0000-0000-000000000000")
	config.WebhookSecret = envOrDefault("TERRALIST_GH_WEBHOOK_SECRET", "e2e-webhook-secret")
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

// doUnauthRequest executes an unauthenticated HTTP request.
func doUnauthRequest(t *testing.T, method, url string) *http.Response {
	t.Helper()
	return doRequest(t, method, url, nil, nil)
}

// githubWebhookSignature returns the X-Hub-Signature-256 header value for a payload.
func githubWebhookSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// doWebhookRequest executes an unsigned GitHub webhook POST.
func doWebhookRequest(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()
	return doWebhookRequestWithSignature(t, url, body, githubWebhookSignature(body, config.WebhookSecret))
}

// doWebhookRequestWithSignature executes a GitHub webhook POST with an explicit signature header.
func doWebhookRequestWithSignature(t *testing.T, url string, body []byte, signature string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if signature != "" {
		req.Header.Set("X-Hub-Signature-256", signature)
	}

	resp, err := httpClient().Do(req)
	require.NoError(t, err)

	return resp
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
