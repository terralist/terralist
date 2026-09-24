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

The `<hostname>` and `<namespace>` come from the provider's source address and are mapped to an authority:

- Under Terralist's own hostname, the one configured with the [`url`](../configuration.md#url) option, the namespace is the authority name, exactly as in the registry protocol.
- Under any other hostname, the request is served by the authority that [stands for that upstream namespace](#standing-for-an-upstream-namespace). Without such an authority the provider is not found.

Both documents are generated from the Terralist database. Package download URLs point to the configured providers storage, or to the location given at upload time when the [`providers-storage-resolver`](../configuration.md#providers-storage-resolver) is `proxy`. Each package is listed with its `zh:` hash, the sha256 of the package, and with its `h1:` hash when the uploader provided one. Terraform verifies the package against these hashes before installing it.

## Standing for an upstream namespace

Terraform identifies a provider by its full source address, `registry.terraform.io/hashicorp/aws` for the public `hashicorp/aws`. A module that requires `hashicorp/aws` and a root configuration that requires `terralist.example.com/hashicorp/aws` ask for two different providers, so a mirror that only answered under Terralist's own hostname could not serve public providers to configurations that were not written for it.

An authority can therefore stand for a namespace of an upstream registry. Set its upstream hostname, and optionally the upstream namespace when it differs from the authority name:

```shell
curl -X POST \
  -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  -d '{"name": "hashicorp", "upstream_hostname": "registry.terraform.io"}' \
  https://terralist.example.com/v1/api/authorities
```

Every provider of that authority is then served by the mirror under both addresses, `terralist.example.com/hashicorp/<type>` and `registry.terraform.io/hashicorp/<type>`. Access control does not change: the policy object stays `<authority>/<type>`. An upstream hostname and namespace pair can be claimed by one authority only.

## Uploading provider packages

Besides the [registry upload](../dev-guide/api-reference.md#upload-a-provider-version), which fetches the provider files from URLs, a provider version can be uploaded from its package files directly. This is how providers reach an air-gapped Terralist: an operator downloads them on a machine with internet access, moves them into the isolated network and uploads them.

### 1. Download the providers

On a connected machine, let Terraform download the providers required by a configuration, together with their metadata:

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

### 2. Optionally, download the signature material

The registry protocol requires the provider's `SHA256SUMS` file, its signature and the signing key. `terraform providers mirror` does not download them, so a version uploaded without them is served through the network mirror only. To serve it through both protocols, fetch them from the upstream registry's download endpoint and add the signing key to the authority:

```shell
version=3.2.4
platform=linux_amd64
meta=$(curl -s "https://registry.terraform.io/v1/providers/hashicorp/null/$version/download/${platform%_*}/${platform#*_}")
curl -s -o SHA256SUMS "$(echo "$meta" | jq -r .shasums_url)"
curl -s -o SHA256SUMS.sig "$(echo "$meta" | jq -r .shasums_signature_url)"
echo "$meta" | jq -r '.signing_keys.gpg_public_keys[0] | {key_id, ascii_armor}' > key.json
```

### 3. Create the authority

Create the authority once, standing for the upstream namespace so that Terraform can install the providers under their original address:

```shell
curl -X POST \
  -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  -d '{"name": "hashicorp", "upstream_hostname": "registry.terraform.io"}' \
  https://terralist.example.com/v1/api/authorities
```

When serving through the registry protocol as well, add the signing key from step 2 to the authority through its keys endpoint or the web UI.

### 4. Upload each version

The upload is a multipart request with the version document in the `metadata` field and the packages in the `archives` field. The `shasums` and `shasums_signature` files and the `protocols` value are given together when the version should be served through the registry protocol too:

```shell
curl -X POST \
  -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  -F metadata=@3.2.4.json \
  -F archives=@terraform-provider-null_3.2.4_linux_amd64.zip \
  -F archives=@terraform-provider-null_3.2.4_darwin_arm64.zip \
  -F shasums=@SHA256SUMS \
  -F shasums_signature=@SHA256SUMS.sig \
  -F protocols=5.0,6.0 \
  https://terralist.example.com/v1/api/providers/hashicorp/null/3.2.4/upload-files
```

Terralist matches every archive with the entry of the same file name in the version document, records its sha256 and its `h1` hash, and, when a `SHA256SUMS` file is given, rejects the upload if any archive does not match it. The `index.json` file is not uploaded; the version list is computed by Terralist.

A whole mirror directory can be uploaded as mirror-only versions with a small loop:

```shell
cd mirror
find . -name '*.json' ! -name 'index.json' | while read -r metadata; do
  dir=$(dirname "$metadata")
  version=$(basename "$metadata" .json)
  path=${dir#./*/}

  args=()
  for archive in "$dir"/*_"$version"_*.zip; do
    args+=(-F "archives=@$archive")
  done

  curl -X POST \
    -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
    -F "metadata=@$metadata" "${args[@]}" \
    "https://terralist.example.com/v1/api/providers/$path/$version/upload-files"
done
```

The loop strips the hostname from the mirror directory, so `registry.terraform.io/hashicorp/null` uploads to the `hashicorp` authority, which must exist beforehand.

Mirror-only versions are omitted from the registry protocol version list, since Terraform refuses unsigned packages from a registry. They are removed like any other version, through the [provider deletion endpoints](../dev-guide/api-reference.md#remove-a-provider-version).

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

With this configuration, Terraform installs every provider from the mirror. Provider addresses in the configuration stay unchanged, and providers of authorities standing for an upstream namespace keep their upstream address:

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
    include = ["terralist.example.com/*/*", "registry.terraform.io/hashicorp/*"]
  }

  direct {
    exclude = ["terralist.example.com/*/*", "registry.terraform.io/hashicorp/*"]
  }
}
```

Terraform verifies each package against the `zh:` hash listed by the mirror and reports it as `verified checksum`. The dependency lock file then records only the `h1:` hash Terraform computes for the installed platform, and `terraform init` warns about incomplete lock file information. Run [`terraform providers lock`](https://developer.hashicorp.com/terraform/cli/commands/providers/lock) to record the hashes of every platform your team uses.

Terraform requires HTTPS for network mirrors. The `credentials` block is not needed when [`providers-anonymous-read`](../configuration.md#providers-anonymous-read) is enabled or the authority is public.

!!! warning "Hostnames with a port"
    Terraform builds the mirror request URL from the provider hostname and mistakes a hostname that carries a port, such as `terralist.example.com:8443`, for a URL scheme. Providers addressed with such a hostname cannot be installed through the mirror. Serve Terralist on the default HTTPS port, or address the providers with the upstream hostname their authority stands for.

## Access control

The mirror serves the same providers as the registry protocol and is governed by the same `providers` [RBAC resource](rbac-configuration.md), with the `<authority>/<provider-name>` object syntax. A policy that grants `get` on a provider grants it through both protocols.

## Reference

- [Terraform Provider Network Mirror Protocol](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol)
- [`terraform providers mirror`](https://developer.hashicorp.com/terraform/cli/commands/providers/mirror)
- [API reference](../dev-guide/api-reference.md#list-provider-versions-network-mirror)
