<!-- markdownlint-configure-file {
  "MD013": {
    "code_blocks": false,
    "tables": false
  },
  "MD033": false,
  "MD041": false
} -->

<div align="center">
  <img alt="Terralist Logo" src="./static/terralist.png" width="200" />
  <h1>Terralist</h1>
  <p>A private Terraform/OpenTofu registry for modules and providers</p>
</div>

---

[![Latest Release](https://img.shields.io/github/release/terralist/terralist.svg)](https://github.com/terralist/terralist/releases/latest) [![CI](https://github.com/terralist/terralist/actions/workflows/test.yml/badge.svg)](https://github.com/terralist/terralist/actions/workflows/test.yml) [![Go Version](https://img.shields.io/github/go-mod/go-version/terralist/terralist)](https://go.dev/) [![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

Terralist implements the [Terraform registry protocols](https://developer.hashicorp.com/terraform/internals/module-registry-protocol) and gives you a private, self-hosted registry with a web dashboard, RBAC, and support for multiple storage backends.

## Features

- **Private module and provider registry**: upload, version, and distribute Terraform/OpenTofu modules and providers within your organization
- **Web dashboard**: browse artifacts, view documentation, and manage authorities and API keys
- **RBAC**: fine-grained access control with built-in roles (`admin`, `readonly`, `anonymous`) and custom policies via [Casbin](https://casbin.org/)
- **Multiple OAuth providers**: authenticate via GitHub, GitLab, BitBucket, or any OIDC-compatible provider
- **API keys with scoped policies**: create API keys for CI/CD with per-key RBAC policies and organizational scopes
- **Storage backends**: store artifacts in AWS S3, Azure Blob, Google Cloud Storage, local filesystem, or proxy mode
- **Prometheus metrics**: monitor uploads, downloads, API key usage, storage operations, and HTTP request latency
- **Single binary**: no external dependencies, runs anywhere Go compiles to

## Quick start

```bash
# Download the latest release
curl -sL "https://github.com/terralist/terralist/releases/latest/download/terralist_$(go env GOOS)_$(go env GOARCH).zip" -o terralist.zip
unzip terralist.zip

# Create a minimal config
cat > config.yaml <<EOF
oauth-provider: github
gh-client-id: ${GITHUB_OAUTH_CLIENT_ID}
gh-client-secret: ${GITHUB_OAUTH_CLIENT_SECRET}
token-signing-secret: $(openssl rand -hex 16)
cookie-secret: $(openssl rand -hex 16)
EOF

# Start the server
./terralist server --config config.yaml
```

Then open [http://localhost:5758](http://localhost:5758) in your browser.

See the [getting started guide](https://www.terralist.io/getting-started/) for detailed setup instructions including HTTPS configuration for Terraform CLI integration.

## Usage

```hcl
# Use a module from your private registry
module "vpc" {
  source  = "registry.example.com/my-org/vpc/aws"
  version = "1.0.0"
}

# Use a provider from your private registry
terraform {
  required_providers {
    custom = {
      source  = "registry.example.com/my-org/custom"
      version = "2.0.0"
    }
  }
}
```

## Documentation

Full documentation is available at [www.terralist.io](https://www.terralist.io/), including:

- [Installation](https://www.terralist.io/installation/)
- [Configuration](https://www.terralist.io/configuration/)
- [RBAC Configuration](https://www.terralist.io/user-guide/rbac-configuration/)
- [API Reference](https://www.terralist.io/dev-guide/api-reference/)
- [Local Development](https://www.terralist.io/dev-guide/local-development/)

## Contributing

Contributions are welcome. All input is appreciated, whether it's a bug report, feature request, or pull request.

- **Issues**: [github.com/terralist/terralist/issues](https://github.com/terralist/terralist/issues)
- **Discussions**: [github.com/terralist/terralist/discussions](https://github.com/terralist/terralist/discussions)

## License

Terralist is licensed under the [Mozilla Public License 2.0](./LICENSE).
