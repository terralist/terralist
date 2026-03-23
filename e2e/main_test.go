package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// bootstrapState holds data created during bootstrap that tests can reference.
var bootstrap struct {
	HashicorpAuthorityID string
}

func TestMain(m *testing.M) {
	initConfig()

	if err := waitForServer(); err != nil {
		fmt.Fprintf(os.Stderr, "server not ready: %v\n", err)
		os.Exit(1)
	}

	if err := bootstrapEnvironment(); err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func waitForServer() error {
	client := &http.Client{Timeout: 2 * time.Second}

	for i := range 30 {
		resp, err := client.Get(apiURL("/check/readyz"))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		if i < 29 {
			time.Sleep(1 * time.Second)
		}
	}

	return fmt.Errorf("server did not become ready within 30 seconds")
}

func bootstrapEnvironment() error {
	if err := createAuthorities(); err != nil {
		return fmt.Errorf("creating authorities: %w", err)
	}

	if err := uploadNullProvider(); err != nil {
		return fmt.Errorf("uploading null provider: %w", err)
	}

	if err := uploadModule(); err != nil {
		return fmt.Errorf("uploading module: %w", err)
	}

	return nil
}

func createAuthorities() error {
	id, err := createAuthority("hashicorp")
	if err != nil {
		return err
	}
	bootstrap.HashicorpAuthorityID = id

	return nil
}

func createAuthority(name string) (string, error) {
	resp, err := doBootstrapRequest(http.MethodPost, apiURL("/v1/api/authorities"), map[string]string{"name": name})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected 201, got %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	id, ok := result["id"].(string)
	if !ok {
		return "", fmt.Errorf("authority response missing 'id' field")
	}

	return id, nil
}

func uploadNullProvider() error {
	// Fetch provider metadata from the Terraform registry.
	registryURL := "https://registry.terraform.io/v1/providers/hashicorp/null/3.2.4/download/linux/amd64"
	resp, err := http.Get(registryURL)
	if err != nil {
		return fmt.Errorf("fetching provider metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registry returned %d", resp.StatusCode)
	}

	var metadata map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return fmt.Errorf("decoding registry response: %w", err)
	}

	// Upload the GPG signing key to the authority.
	signingKeys := metadata["signing_keys"].(map[string]any)
	gpgKeys := signingKeys["gpg_public_keys"].([]any)
	gpgKey := gpgKeys[0].(map[string]any)

	gpgBody := map[string]string{
		"key_id":          gpgKey["key_id"].(string),
		"ascii_armor":     gpgKey["ascii_armor"].(string),
		"trust_signature": "",
	}

	gpgResp, err := doBootstrapRequest(
		http.MethodPost,
		apiURL("/v1/api/authorities/%s/keys", bootstrap.HashicorpAuthorityID),
		gpgBody,
	)
	if err != nil {
		return fmt.Errorf("uploading GPG key: %w", err)
	}
	defer gpgResp.Body.Close()

	if gpgResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(gpgResp.Body)
		return fmt.Errorf("GPG key upload failed (%d): %s", gpgResp.StatusCode, string(body))
	}

	// Upload the provider version.
	providerBody := map[string]any{
		"protocols": []string{"6.0"},
		"shasums": map[string]string{
			"url":           metadata["shasums_url"].(string),
			"signature_url": metadata["shasums_signature_url"].(string),
		},
		"platforms": []map[string]string{
			{
				"os":           "linux",
				"arch":         "amd64",
				"download_url": metadata["download_url"].(string),
				"shasum":       metadata["shasum"].(string),
			},
		},
	}

	provResp, err := doBootstrapRequest(
		http.MethodPost,
		apiURL("/v1/api/providers/hashicorp/null/3.2.4/upload"),
		providerBody,
	)
	if err != nil {
		return fmt.Errorf("uploading provider: %w", err)
	}
	defer provResp.Body.Close()

	if provResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(provResp.Body)
		return fmt.Errorf("provider upload failed (%d): %s", provResp.StatusCode, string(body))
	}

	return nil
}

func uploadModule() error {
	moduleBody := map[string]string{
		"download_url": "https://github.com/hashicorp/terraform-cidr-subnets/archive/refs/tags/v1.0.0.zip",
	}

	resp, err := doBootstrapRequest(
		http.MethodPost,
		apiURL("/v1/api/modules/hashicorp/subnets/cidr/1.0.0/upload"),
		moduleBody,
	)
	if err != nil {
		return fmt.Errorf("uploading module: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("module upload failed (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// doBootstrapRequest is a non-test helper for use in TestMain where *testing.T is unavailable.
func doBootstrapRequest(method, url string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer x-api-key:"+config.MasterAPIKey)

	return httpClient().Do(req)
}
