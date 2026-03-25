# Terralist

Terralist is a private Terraform/OpenTofu registry for modules and providers. It implements the [Terraform registry protocols](https://developer.hashicorp.com/terraform/internals/module-registry-protocol) and gives you full control over how your infrastructure code is distributed.

## Features

### Private by default

Terralist requires authentication for all module and provider operations, including downloads. Users authenticate via `terraform login` through any supported OAuth provider (GitHub, GitLab, BitBucket, OIDC), or through API keys for programmatic access.

Anonymous (unauthenticated) downloads can be optionally enabled for isolated environments.

### Role-based access control

Access to resources is governed by RBAC policies built on [Casbin](https://casbin.org/). Built-in roles (`admin`, `readonly`, `anonymous`) cover common scenarios, and custom policies allow fine-grained control over who can read, create, or delete specific modules and providers.

API keys carry their own inline RBAC policies, so each CI/CD pipeline or integration can have precisely scoped access.

### Secure artifact storage

Modules and providers are stored in private storage backends. When a download is requested, Terralist generates a temporary presigned URL and forwards it to the requester. Supported backends include AWS S3, Azure Blob, Google Cloud Storage, and local filesystem.

A proxy mode is also available, where Terralist forwards the original source URL directly. This lets you use `version` constraints while storing modules in a git mono-repository.

### Artifacts documentation

Terralist analyses uploaded artifacts and generates versioned documentation, including submodule documentation for modules. The documentation is rendered in the web dashboard with syntax highlighting, Mermaid diagram support, and emoji rendering.

### Web dashboard

A built-in web interface lets you browse modules and providers, view documentation, manage authorities and signing keys, and create API keys with scoped RBAC policies.

### Observability

Terralist exposes Prometheus metrics for artifact operations, API key usage, storage backend performance, and HTTP request latency.
