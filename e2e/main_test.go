package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"testing"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/mod/sumdb/dirhash"
)

const (
	// nullMirrorOnlyVersion is uploaded from packages without signature
	// material, so it is served through the network mirror only.
	nullMirrorOnlyVersion = "3.2.1"

	// nullSignedVersion is uploaded from packages together with its
	// SHA256SUMS file and signature, so it is served through both protocols.
	nullSignedVersion = "3.2.0"

	// upstreamProvider is never uploaded and is served by pulling it through
	// from registry.terraform.io, except upstreamDeniedVersion.
	upstreamProvider      = "random"
	upstreamDeniedVersion = "3.5.0"

	// upstreamModule is never uploaded and is served by pulling it through
	// from registry.terraform.io; the name carries the system, as module
	// rules do.
	upstreamModule = "dir/template"
)

// bootstrapState holds data created during bootstrap that tests can reference.
var bootstrap struct {
	HashicorpAuthorityID string

	// NullProviderShaSum is the sha256 of the null provider package for the
	// current platform, as published in its SHA256SUMS file.
	NullProviderShaSum string

	// NullMirrorOnlyArchive is the package of nullMirrorOnlyVersion for the
	// current platform, NullMirrorOnlyH1 its h1 hash and NullMirrorOnlyShaSum
	// its sha256.
	NullMirrorOnlyArchive []byte
	NullMirrorOnlyH1      string
	NullMirrorOnlyShaSum  string
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
	if err := createS3Bucket(); err != nil {
		return fmt.Errorf("creating S3 bucket: %w", err)
	}

	if err := createAuthorities(); err != nil {
		return fmt.Errorf("creating authorities: %w", err)
	}

	if err := uploadNullProvider(); err != nil {
		return fmt.Errorf("uploading null provider: %w", err)
	}

	if err := uploadNullProviderPackages(); err != nil {
		return fmt.Errorf("uploading null provider packages: %w", err)
	}

	if err := uploadModule(); err != nil {
		return fmt.Errorf("uploading module: %w", err)
	}

	return nil
}

func createS3Bucket() error {
	endpoint := envOrDefault("TERRALIST_S3_ENDPOINT", "http://localhost:9000")
	bucket := envOrDefault("TERRALIST_S3_BUCKET_NAME", "terralist")
	accessKey := envOrDefault("TERRALIST_S3_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
	secretKey := envOrDefault("TERRALIST_S3_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	region := envOrDefault("TERRALIST_S3_BUCKET_REGION", "us-east-1")

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return fmt.Errorf("loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = &endpoint
		o.UsePathStyle = true
	})

	_, err = client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: &bucket,
	})
	if err != nil {
		// Ignore "bucket already exists" errors.
		var bae *types.BucketAlreadyOwnedByYou
		var bex *types.BucketAlreadyExists
		if !errors.As(err, &bae) && !errors.As(err, &bex) {
			return fmt.Errorf("creating bucket %q: %w", bucket, err)
		}
	}

	return nil
}

func createAuthorities() error {
	// The hashicorp authority stands for the hashicorp namespace of the public
	// registry, so its providers are also served through the network mirror
	// under their registry.terraform.io address. Its upstream is enabled with
	// the deny policy, so only the providers the rules allow are pulled
	// through: the random provider, except one denied version.
	id, err := createAuthority(map[string]any{
		"name":                    "hashicorp",
		"upstream_hostname":       "registry.terraform.io",
		"upstream_enabled":        true,
		"upstream_default_policy": "deny",
	})
	if err != nil {
		return err
	}
	bootstrap.HashicorpAuthorityID = id

	for _, rule := range []map[string]string{
		{"kind": "provider", "name": upstreamProvider, "version": "*", "effect": "allow"},
		{"kind": "provider", "name": upstreamProvider, "version": upstreamDeniedVersion, "effect": "deny"},
		{"kind": "module", "name": upstreamModule, "version": "*", "effect": "allow"},
	} {
		if _, err := addRule(id, rule); err != nil {
			return err
		}
	}

	return nil
}

// addRule adds an upstream rule to an authority and returns the rule id.
func addRule(authorityID string, rule map[string]string) (string, error) {
	resp, err := doBootstrapRequest(apiURL("/v1/api/authorities/%s/rules", authorityID), rule)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected 200, got %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	id, ok := result["id"].(string)
	if !ok {
		return "", fmt.Errorf("rule response missing 'id' field")
	}

	return id, nil
}

func createAuthority(body map[string]any) (string, error) {
	resp, err := doBootstrapRequest(apiURL("/v1/api/authorities"), body)
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

// fetchNullProviderMetadata fetches the download metadata of a null provider
// version from the Terraform registry for the current platform.
func fetchNullProviderMetadata(version string) (map[string]any, error) {
	registryURL := fmt.Sprintf("https://registry.terraform.io/v1/providers/hashicorp/null/%s/download/%s/%s", version, runtime.GOOS, runtime.GOARCH)
	resp, err := http.Get(registryURL)
	if err != nil {
		return nil, fmt.Errorf("fetching provider metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned %d", resp.StatusCode)
	}

	var metadata map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("decoding registry response: %w", err)
	}

	return metadata, nil
}

func uploadNullProvider() error {
	metadata, err := fetchNullProviderMetadata("3.2.4")
	if err != nil {
		return err
	}

	// Upload the GPG signing key to the authority.
	signingKeys, ok := metadata["signing_keys"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing signing_keys in registry response")
	}
	gpgKeys, ok := signingKeys["gpg_public_keys"].([]any)
	if !ok || len(gpgKeys) == 0 {
		return fmt.Errorf("missing gpg_public_keys in registry response")
	}
	gpgKey, ok := gpgKeys[0].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid gpg key format in registry response")
	}

	keyID, _ := gpgKey["key_id"].(string)
	asciiArmor, _ := gpgKey["ascii_armor"].(string)

	gpgBody := map[string]string{
		"key_id":          keyID,
		"ascii_armor":     asciiArmor,
		"trust_signature": "",
	}

	gpgResp, err := doBootstrapRequest(
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
	shaSumsURL, _ := metadata["shasums_url"].(string)
	shaSumsSignatureURL, _ := metadata["shasums_signature_url"].(string)
	downloadURL, _ := metadata["download_url"].(string)
	shasum, _ := metadata["shasum"].(string)

	bootstrap.NullProviderShaSum = shasum

	providerBody := map[string]any{
		"protocols": []string{"6.0"},
		"shasums": map[string]string{
			"url":           shaSumsURL,
			"signature_url": shaSumsSignatureURL,
		},
		"platforms": []map[string]string{
			{
				"os":           runtime.GOOS,
				"arch":         runtime.GOARCH,
				"download_url": downloadURL,
				"shasum":       shasum,
			},
		},
	}

	provResp, err := doBootstrapRequest(
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

// uploadNullProviderPackages uploads two null provider versions from their
// packages, the way `terraform providers mirror` output is uploaded by an
// operator: one without signature material and one with it.
func uploadNullProviderPackages() error {
	metadata, archive, err := downloadNullPackage(nullMirrorOnlyVersion)
	if err != nil {
		return err
	}

	h1, err := hashArchive(archive)
	if err != nil {
		return fmt.Errorf("hashing provider package: %w", err)
	}

	bootstrap.NullMirrorOnlyArchive = archive
	bootstrap.NullMirrorOnlyH1 = h1
	bootstrap.NullMirrorOnlyShaSum, _ = metadata["shasum"].(string)

	files := packagesUploadFiles(nullMirrorOnlyVersion, h1, archive)
	if err := uploadPackages(nullMirrorOnlyVersion, files, nil); err != nil {
		return err
	}

	metadata, archive, err = downloadNullPackage(nullSignedVersion)
	if err != nil {
		return err
	}

	h1, err = hashArchive(archive)
	if err != nil {
		return fmt.Errorf("hashing provider package: %w", err)
	}

	shaSumsURL, _ := metadata["shasums_url"].(string)
	shaSums, err := download(shaSumsURL)
	if err != nil {
		return err
	}

	shaSumsSignatureURL, _ := metadata["shasums_signature_url"].(string)
	shaSumsSignature, err := download(shaSumsSignatureURL)
	if err != nil {
		return err
	}

	files = append(packagesUploadFiles(nullSignedVersion, h1, archive),
		multipartFile{Field: "shasums", FileName: "SHA256SUMS", Content: shaSums},
		multipartFile{Field: "shasums_signature", FileName: "SHA256SUMS.sig", Content: shaSumsSignature},
	)

	return uploadPackages(nullSignedVersion, files, map[string]string{"protocols": "5.0"})
}

// downloadNullPackage fetches the registry metadata of a null provider version
// and downloads its package for the current platform.
func downloadNullPackage(version string) (map[string]any, []byte, error) {
	metadata, err := fetchNullProviderMetadata(version)
	if err != nil {
		return nil, nil, err
	}

	downloadURL, _ := metadata["download_url"].(string)
	archive, err := download(downloadURL)
	if err != nil {
		return nil, nil, err
	}

	return metadata, archive, nil
}

// download fetches the content of a public URL.
func download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download of %s returned %d", url, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", url, err)
	}

	return content, nil
}

// uploadPackages posts the files of a null provider version to the packages
// upload endpoint with the master API key.
func uploadPackages(version string, files []multipartFile, values map[string]string) error {
	body, contentType, err := multipartBody(files, values)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, apiURL("/v1/api/providers/hashicorp/null/%s/upload-files", version), body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer x-api-key:"+config.MasterAPIKey)

	resp, err := httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("uploading provider packages: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("provider packages upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// nullProviderArchiveName returns the package file name of a null provider
// version for the current platform, as produced by `terraform providers mirror`.
func nullProviderArchiveName(version string) string {
	return fmt.Sprintf("terraform-provider-null_%s_%s_%s.zip", version, runtime.GOOS, runtime.GOARCH)
}

// packagesUploadFiles builds the multipart files of a packages upload: the
// version document listing the archive with the given h1 hash, and the
// archive itself.
func packagesUploadFiles(version, h1 string, archive []byte) []multipartFile {
	archiveName := nullProviderArchiveName(version)

	metadata, _ := json.Marshal(map[string]any{
		"archives": map[string]any{
			fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH): map[string]any{
				"url":    archiveName,
				"hashes": []string{h1},
			},
		},
	})

	return []multipartFile{
		{Field: "metadata", FileName: version + ".json", Content: metadata},
		{Field: "archives", FileName: archiveName, Content: archive},
	}
}

// hashArchive computes the h1 hash of a zip archive, as Terraform does.
func hashArchive(archive []byte) (string, error) {
	tmp, err := os.CreateTemp("", "terralist-e2e-*.zip")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := tmp.Write(archive); err != nil {
		return "", err
	}

	return dirhash.HashZip(tmp.Name(), dirhash.DefaultHash)
}

func uploadModule() error {
	moduleBody := map[string]string{
		"download_url": "https://github.com/hashicorp/terraform-cidr-subnets/archive/refs/tags/v1.0.0.zip",
	}

	resp, err := doBootstrapRequest(
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
func doBootstrapRequest(url string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer x-api-key:"+config.MasterAPIKey)

	return httpClient().Do(req)
}
