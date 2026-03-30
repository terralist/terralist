package e2e

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// registryHost returns a Terraform-compatible registry hostname.
// Terraform requires registry hostnames to contain at least one dot.
// localhost.direct is a public domain that resolves to 127.0.0.1.
func registryHost(t *testing.T) string {
	t.Helper()

	parsed, err := url.Parse(config.URL)
	require.NoError(t, err)

	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	return fmt.Sprintf("localhost.direct:%s", port)
}

func TestTerraformModuleInit(t *testing.T) {
	requireTerraformCapable(t)
	host := registryHost(t)

	dir := setupTerraformProject(t, host, fmt.Sprintf(`
module "subnets" {
  source  = "%s/hashicorp/subnets/cidr"
  version = "1.0.0"

  base_cidr_block = "10.0.0.0/16"
  networks = [
    { name = "a", new_bits = 8 },
  ]
}
`, host))

	out := runTerraform(t, dir, "init")
	assert.Contains(t, out, "Initializing the backend")
	assert.Contains(t, out, "Terraform has been successfully initialized")
}

func TestTerraformModulePlan(t *testing.T) {
	requireTerraformCapable(t)
	host := registryHost(t)

	dir := setupTerraformProject(t, host, fmt.Sprintf(`
module "subnets" {
  source  = "%s/hashicorp/subnets/cidr"
  version = "1.0.0"

  base_cidr_block = "10.0.0.0/16"
  networks = [
    { name = "a", new_bits = 8 },
  ]
}
`, host))

	runTerraform(t, dir, "init")
	out := runTerraform(t, dir, "plan")
	assert.Contains(t, out, "no changes")
}

func TestTerraformProviderInit(t *testing.T) {
	requireTerraformCapable(t)
	host := registryHost(t)

	dir := setupTerraformProject(t, host, fmt.Sprintf(`
terraform {
  required_providers {
    null = {
      source  = "%s/hashicorp/null"
      version = "3.2.4"
    }
  }
}

resource "null_resource" "test" {}
`, host))

	out := runTerraform(t, dir, "init")
	assert.Contains(t, out, fmt.Sprintf("Installed %s/hashicorp/null v3.2.4", host))
	assert.Contains(t, out, "Terraform has been successfully initialized")
}

func TestTerraformProviderPlanApply(t *testing.T) {
	requireTerraformCapable(t)
	host := registryHost(t)

	dir := setupTerraformProject(t, host, fmt.Sprintf(`
terraform {
  required_providers {
    null = {
      source  = "%s/hashicorp/null"
      version = "3.2.4"
    }
  }
}

resource "null_resource" "test" {}
`, host))

	runTerraform(t, dir, "init")

	planOut := runTerraform(t, dir, "plan")
	assert.Contains(t, planOut, "null_resource.test")
	assert.Contains(t, planOut, "1 to add")

	applyOut := runTerraform(t, dir, "apply", "-auto-approve")
	assert.Contains(t, applyOut, "Apply complete! Resources: 1 added")
}

// requireTerraformCapable skips the test if the environment is not set up
// for Terraform integration tests (requires HTTPS and terraform binary).
func requireTerraformCapable(t *testing.T) {
	t.Helper()

	if !strings.HasPrefix(config.URL, "https://") {
		t.Skip("Terraform tests require HTTPS (set TERRALIST_URL to an https:// URL)")
	}

	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform binary not found in PATH")
	}
}

// setupTerraformProject creates a temporary directory with a Terraform
// configuration and a CLI config that points at the Terralist instance.
func setupTerraformProject(t *testing.T, host, tfConfig string) string {
	t.Helper()

	dir := t.TempDir()

	// Write the Terraform configuration.
	mainTf := filepath.Join(dir, "main.tf")
	require.NoError(t, os.WriteFile(mainTf, []byte(tfConfig), 0644))

	// Write a .terraformrc that authenticates with the master API key.
	rcContent := fmt.Sprintf(`credentials "%s" {
  token = "x-api-key:%s"
}
`, host, config.MasterAPIKey)

	rcPath := filepath.Join(dir, ".terraformrc")
	require.NoError(t, os.WriteFile(rcPath, []byte(rcContent), 0644))

	return dir
}

// runTerraform executes a terraform command in the given directory and returns
// the combined stdout/stderr output.
func runTerraform(t *testing.T, dir, subcmd string, args ...string) string {
	t.Helper()

	cmdArgs := append([]string{subcmd, "-no-color"}, args...)
	cmd := exec.Command("terraform", cmdArgs...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"TF_CLI_CONFIG_FILE="+filepath.Join(dir, ".terraformrc"),
		"TF_IN_AUTOMATION=1",
	)

	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "terraform %s failed:\n%s", subcmd, string(out))
	return string(out)
}
