# Provider Network Mirror

Terralist serves its providers through the [Provider Network Mirror Protocol](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol) in addition to the Provider Registry Protocol. Both protocols expose the same providers: everything uploaded to a Terralist authority is available through both.

The mirror protocol lets a Terraform CLI configuration route provider installation through Terralist with a single `network_mirror` block, without service discovery and without changing anything in the Terraform configurations themselves.

!!! note "Providers only"
    The network mirror protocol is defined by Terraform for providers only. Modules keep using the [Module Registry Protocol](../getting-started.md#upload-a-new-module).

## How it works

The protocol does not use service discovery, so its endpoints live at a fixed base URL:

```
https://terralist.example.com/providers/
```

Terraform queries two documents under that base URL:

| Document | Endpoint |
| --- | --- |
| Available versions | `GET /providers/<hostname>/<namespace>/<type>/index.json` |
| Installation packages of a version | `GET /providers/<hostname>/<namespace>/<type>/<version>.json` |

The `<hostname>` is the hostname of the provider's source address. Terralist answers only for its own hostname, the one configured with the [`url`](../configuration.md#url) option; providers addressed with any other hostname are not found. The `<namespace>` is the authority name and the `<type>` is the provider name, exactly as in the registry protocol.

Both documents are generated from the Terralist database. Package download URLs point to the configured providers storage, or to the location given at upload time when the [`providers-storage-resolver`](../configuration.md#providers-storage-resolver) is `proxy`. Each package is listed with its `zh:` hash, the sha256 published in the provider's `SHA256SUMS` file, which Terraform verifies before installing the package.

## Configuring Terraform

Point Terraform at the mirror in the [CLI configuration file](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-installation). The trailing slash in the URL is required.

```hcl title=".terraformrc"
provider_installation {
  network_mirror {
    url     = "https://terralist.example.com/providers/"
    include = ["terralist.example.com/*/*"]
  }

  direct {
    exclude = ["terralist.example.com/*/*"]
  }
}

credentials "terralist.example.com" {
  token = "x-api-key:YOUR_API_KEY"
}
```

With this configuration, providers published on Terralist are installed through the mirror and every other provider is installed directly from its origin registry. Provider addresses in the configuration stay unchanged:

```hcl
terraform {
  required_providers {
    null = {
      source  = "terralist.example.com/hashicorp/null"
      version = "3.2.4"
    }
  }
}
```

Terraform requires HTTPS for network mirrors. The `credentials` block is not needed when [`providers-anonymous-read`](../configuration.md#providers-anonymous-read) is enabled or the authority is public.

!!! warning "Hostnames with a port"
    Terraform builds the mirror request URL from the provider hostname and mistakes a hostname that carries a port, such as `terralist.example.com:8443`, for a URL scheme. Providers served under such a hostname cannot be installed through the mirror. Serve Terralist on the default HTTPS port to use the mirror protocol.

## Access control

The mirror serves the same providers as the registry protocol and is governed by the same `providers` [RBAC resource](rbac-configuration.md), with the `<authority>/<provider-name>` object syntax. A policy that grants `get` on a provider grants it through both protocols.

## Reference

- [Terraform Provider Network Mirror Protocol](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol)
- [API reference](../dev-guide/api-reference.md#list-provider-versions-network-mirror)
