# Provider Network Mirror

Terralist can act as a [Provider Network Mirror](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol) for Terraform and OpenTofu. A network mirror serves provider packages that were originally published on other registries, such as `registry.terraform.io`, so that Terraform can install them without reaching the upstream registry.

This is designed for air-gapped deployments: an operator downloads the providers on a machine with internet access, moves them into the isolated network, and uploads them to Terralist. Terraform clients inside the isolated network then install every provider from Terralist.

!!! note "Providers only"
    The network mirror protocol is defined by Terraform for providers only. Modules keep using the [Module Registry Protocol](../getting-started.md#upload-a-new-module).

## How it works

Mirrored providers are stored apart from the providers published directly on Terralist. They are identified by the hostname of their upstream registry, their namespace and their type, exactly as Terraform identifies them in a `source` address. A mirrored `registry.terraform.io/hashicorp/aws` and a `hashicorp/aws` provider published on a Terralist authority never interfere with each other.

The protocol does not use service discovery, so its endpoints live at a fixed base URL:

```
https://terralist.example.com/providers/
```

Terraform queries two documents under that base URL:

| Document | Endpoint |
| --- | --- |
| Available versions | `GET /providers/<hostname>/<namespace>/<type>/index.json` |
| Installation packages of a version | `GET /providers/<hostname>/<namespace>/<type>/<version>.json` |

Both documents are generated from the Terralist database. Package download URLs point to the configured storage backend.

## Enabling the mirror

The mirror is disabled by default. Enable it by choosing a storage backend for the mirrored packages with the [`mirror-storage-resolver`](../configuration.md#mirror-storage-resolver) option:

```yaml title="config.yaml"
mirror-storage-resolver: "s3"
```

Any of the `local`, `s3`, `azure` and `gcs` backends can be used. Mirrored packages are stored under the `mirror/<hostname>/<namespace>/<type>/<version>/` prefix, so the mirror can share a bucket or directory with the modules and providers storage.

Reads through the protocol are subject to [access control](#access-control). To let unauthenticated Terraform clients use the mirror, enable [`mirror-anonymous-read`](../configuration.md#mirror-anonymous-read).

## Populating the mirror

### 1. Download the providers

On a machine with internet access, let Terraform download the providers required by a configuration, together with their metadata:

```shell
terraform providers mirror -platform=linux_amd64 -platform=darwin_arm64 ./mirror
```

The command writes one directory per provider, containing a `<version>.json` document and one package archive per platform:

```
mirror/registry.terraform.io/hashicorp/null/
├── 3.2.4.json
├── index.json
├── terraform-provider-null_3.2.4_darwin_arm64.zip
└── terraform-provider-null_3.2.4_linux_amd64.zip
```

Terraform verifies the signatures of the packages it downloads, so the `h1` hashes listed in `<version>.json` can be trusted as long as the files are not altered afterwards.

### 2. Upload a version

Move the directory into the isolated network and upload each version to Terralist. The upload is a multipart request with two fields:

- `metadata`: the `<version>.json` document, as written by Terraform;
- `archives`: one or more package archives listed in the document.

```shell
curl -X POST \
  -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  -F metadata=@3.2.4.json \
  -F archives=@terraform-provider-null_3.2.4_linux_amd64.zip \
  -F archives=@terraform-provider-null_3.2.4_darwin_arm64.zip \
  https://terralist.example.com/v1/api/mirror/registry.terraform.io/hashicorp/null/3.2.4/upload
```

The `index.json` file is not uploaded. Terralist computes the list of available versions itself, so versions uploaded at different times never overwrite each other.

Terralist matches every uploaded archive with the entry of the same file name in the metadata, computes the `h1` hash of the archive and rejects the whole upload if any archive does not match its declared hash, is not listed in the metadata, or has no `h1` hash. Since Terraform trusts a network mirror completely, this check is what guarantees that clients receive the packages Terraform originally verified.

Archives listed in the metadata but not attached to the request are ignored, so platforms can be uploaded in separate requests. Uploading a platform that already exists for the version is rejected; delete the version first to replace it.

A whole mirror directory can be uploaded with a small loop:

```shell
cd mirror
find . -name '*.json' ! -name 'index.json' | while read -r metadata; do
  dir=$(dirname "$metadata")
  version=$(basename "$metadata" .json)
  path=${dir#./}

  args=()
  for archive in "$dir"/*_"$version"_*.zip; do
    args+=(-F "archives=@$archive")
  done

  curl -X POST \
    -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
    -F "metadata=@$metadata" "${args[@]}" \
    "https://terralist.example.com/v1/api/mirror/$path/$version/upload"
done
```

### 3. Remove providers

Mirrored providers can be removed at any granularity:

```shell
# A single version
curl -X DELETE -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  https://terralist.example.com/v1/api/mirror/registry.terraform.io/hashicorp/null/3.2.4

# A provider with all its versions
curl -X DELETE -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  https://terralist.example.com/v1/api/mirror/registry.terraform.io/hashicorp/null

# Every provider of a namespace
curl -X DELETE -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  https://terralist.example.com/v1/api/mirror/registry.terraform.io/hashicorp

# Every provider mirrored from a hostname
curl -X DELETE -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  https://terralist.example.com/v1/api/mirror/registry.terraform.io
```

The stored packages are removed together with their database records.

## Configuring Terraform

Point Terraform at the mirror in the [CLI configuration file](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-installation). The trailing slash in the URL is required.

```hcl title=".terraformrc"
provider_installation {
  network_mirror {
    url = "https://terralist.example.com/providers/"
  }
}

credentials "terralist.example.com" {
  token = "x-api-key:YOUR_API_KEY"
}
```

With this configuration, Terraform installs every provider from the mirror. Provider addresses in the configuration stay unchanged:

```hcl
terraform {
  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "3.2.4"
    }
  }
}
```

To mirror only some providers and install the rest directly, combine the `network_mirror` and `direct` methods with `include` and `exclude`:

```hcl title=".terraformrc"
provider_installation {
  network_mirror {
    url     = "https://terralist.example.com/providers/"
    include = ["registry.terraform.io/hashicorp/*"]
  }

  direct {
    exclude = ["registry.terraform.io/hashicorp/*"]
  }
}
```

Terraform requires HTTPS for network mirrors. The `credentials` block is not needed when `mirror-anonymous-read` is enabled.

!!! tip "Terralist providers and the mirror"
    Providers published directly on Terralist are installed through the Provider Registry Protocol, under Terralist's own hostname. When a configuration uses both, exclude Terralist's hostname from the mirror: `exclude = ["terralist.example.com/*/*"]`.

## Access control

Mirrored providers are governed by the `mirror` [RBAC resource](rbac-configuration.md). Policy objects use the `<hostname>/<namespace>/<type>` syntax:

```csv
# Everyone can install mirrored providers
p, role:readonly, mirror, get, *, allow

# The platform team maintains everything mirrored from the public registry
p, role:platform, mirror, *, registry.terraform.io*, allow
```

The built-in `role:readonly` role already grants `get` on every mirrored provider. Uploading requires the `create` action and removing requires the `delete` action. Deleting a namespace or a hostname is evaluated against the `<hostname>/<namespace>` and `<hostname>` objects respectively.

## Reference

- [Terraform Provider Network Mirror Protocol](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol)
- [`terraform providers mirror`](https://developer.hashicorp.com/terraform/cli/commands/providers/mirror)
- [API reference](../dev-guide/api-reference.md#list-mirrored-provider-versions)
