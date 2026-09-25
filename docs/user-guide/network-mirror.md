# Provider Network Mirror

Terralist serves its providers through the [Provider Network Mirror Protocol](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol) in addition to the Provider Registry Protocol. Both protocols expose the same providers: everything uploaded to a Terralist authority is available through both.

The mirror protocol lets a Terraform CLI configuration route provider installation through Terralist with a single `network_mirror` block, without service discovery and without changing anything in the Terraform configurations themselves.

!!! note "Providers only"
    The network mirror protocol is defined by Terraform for providers only. Modules keep using the [Module Registry Protocol](../getting-started.md#upload-a-new-module), through which they can also be [pulled through from an upstream](#pulling-modules-through-from-an-upstream).

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

## Pulling providers through from an upstream

An authority standing for an upstream namespace can also fetch the providers it does not hold from that upstream registry, on first request, and keep them. Enable the upstream on the authority. An update replaces the whole authority, its signing keys included, so send it back as read with the upstream enabled:

```shell
curl -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  https://terralist.example.com/v1/api/authorities/$AUTHORITY_ID \
  | jq '.upstream_enabled = true' \
  | curl -X PATCH \
      -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
      -d @- \
      https://terralist.example.com/v1/api/authorities/$AUTHORITY_ID
```

Pulling through needs a storage backend for providers, since fetched packages are stored and then served from there like uploaded ones. With the `proxy` [`providers-storage-resolver`](../configuration.md#providers-storage-resolver) the upstream is never consulted. Terralist never streams a package to a client: a stored package is answered with a redirect to storage, and a package not stored yet is downloaded from the upstream, verified and stored first, then answered with the same redirect.

The upstream is reached at `https://<upstream_hostname>` unless `upstream_url` names another location, such as a private mirror of the public registry. A private upstream can be given an `upstream_token`, which Terralist sends as a bearer token and stores sealed with the [`upstream-secret`](../configuration.md#upstream-secret). The token is never returned by the API; `upstream_has_token` tells whether one is stored, and an update without a token keeps the stored one. Private upstreams on internal networks also need [`fetch-allow-private-addresses`](../configuration.md#fetch-allow-private-addresses).

### What happens on a request

- The version list, in both protocols, merges the versions the upstream offers with the versions Terralist holds. A version uploaded to Terralist always wins over the upstream one and is served exactly as uploaded: the platforms it lacks are not completed from the upstream.
- The network mirror version document lists the packages Terralist holds with their storage location, and the packages it does not hold yet with their `zh:` hash from the upstream `SHA256SUMS` file and a link back to the mirror. The registry protocol download metadata does the same for a single platform.
- Following such a link downloads the package from the upstream with its digest enforced, stores it next to the `SHA256SUMS` file and its signature, records the platform with the `upstream` origin, and redirects to storage. Concurrent requests for the same package wait for one download. The next request is served from storage without touching the upstream.
- Upstream metadata is cached for [`upstream-cache-ttl`](../configuration.md#upstream-cache-ttl) and kept for [`upstream-cache-retention`](../configuration.md#upstream-cache-retention). When the upstream is unreachable, the last known answer is served, and everything already stored stays available regardless.

Terraform downloads packages without credentials, so the links Terralist lists carry a short-lived token that proves the document was served to a caller allowed to read the package, and whether that caller may fetch it. The links expire after fifteen minutes.

### Who may pull through

Reading a version list or a document only needs the `get` action on the provider. Merging upstream versions and fetching packages writes to storage and to the database, so it needs the `create` action on the `<authority>/<type>` object of the `providers` resource. A caller with `get` only sees and downloads what Terralist already holds. The API key Terraform uses must therefore be allowed to `create` the providers it should pull through.

### Rules

The authority's `upstream_default_policy`, `allow` or `deny`, decides which upstream versions may be served when no rule says otherwise. Rules refine it per artifact:

```shell
curl -X POST \
  -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  -d '{"kind": "provider", "name": "aws", "version": "5.*", "effect": "deny"}' \
  https://terralist.example.com/v1/api/authorities/$AUTHORITY_ID/rules
```

`name` and `version` are globs; `kind` is `provider` or `module`. A version is served when the upstream is enabled, no deny rule matches, and either the default policy is `allow` or an allow rule matches. Deny always wins. Rules filter the version list itself, so Terraform never selects a version it cannot download. They apply to upstream versions only; versions uploaded to Terralist are always served. Rules are removed with `DELETE /v1/api/authorities/<id>/rules/<rule id>`.

A deny rule is the way to stop serving a version that was pulled through: deleting the stored version alone would only make the next request fetch it again.

### Pre-warming

A cold fetch of a large provider happens inside Terraform's download request. Terraform itself puts no timeout on it, but a reverse proxy or load balancer in front of Terralist may. Packages can be fetched ahead of time instead:

```shell
curl -X POST \
  -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  -d '{"platforms": ["linux_amd64", "darwin_arm64"]}' \
  https://terralist.example.com/v1/api/providers/hashicorp/aws/5.0.0/fetch
```

The response reports the outcome per platform.

### Trust

Before anything from an upstream is trusted, Terralist fetches the version's `SHA256SUMS` file and its signature and verifies the signature against the keys the upstream advertises for that version, the way Terraform does. Only digests from a verified file are listed, every package download enforces its digest, and the signing keys are stored with the version so that Terraform installing through the registry protocol verifies the same chain.

Registries keep advertising the key that signed a release after that key expired and do not re-sign old releases. Terralist accepts such signatures once every other check passed, as Terraform does, and logs a warning naming the key. Set [`upstream-reject-expired-signing-keys`](../configuration.md#upstream-reject-expired-signing-keys) to refuse them instead.

### Pulling modules through from an upstream

Modules have no network mirror protocol and no identity problem: a module is fetched from whatever registry its source names, so consumers point the source at Terralist, `terralist.example.com/hashicorp/dir/template`, and Terralist fills the gaps from the upstream namespace the authority stands for, under the same rules and the same `create` requirement.

The version list of the module registry protocol merges the upstream versions the rules allow. The download endpoint of a version Terralist does not hold answers with an `X-Terraform-Get` location pointing at Terralist's own archive route, carrying the same kind of short-lived token as provider packages, because Terraform hands module locations to go-getter, which downloads without credentials. Following that location fetches the module from its upstream source, a git repository or an archive over HTTP(S), stores it exactly as an uploaded module is stored, documentation included, and hands go-getter the storage location. The next download of that version goes to storage directly. Other upstream sources, such as `s3::` or `gcs::`, are refused, and git hosts resolving to private addresses are refused unless [`fetch-allow-private-addresses`](../configuration.md#fetch-allow-private-addresses) is set.

Module rules use `kind` `module` and name the module as `<name>/<system>`:

```shell
curl -X POST \
  -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  -d '{"kind": "module", "name": "dir/template", "version": "*", "effect": "allow"}' \
  https://terralist.example.com/v1/api/authorities/$AUTHORITY_ID/rules
```

A module version can be pre-warmed as well:

```shell
curl -X POST \
  -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
  https://terralist.example.com/v1/api/modules/hashicorp/dir/template/1.0.2/fetch
```

### Creating authorities on demand

With [`upstream-auto-create`](../configuration.md#upstream-auto-create) listing an upstream hostname, the first network mirror request for an unknown namespace of that hostname, from a caller allowed to `create` that authority, creates it: named after the namespace, standing for it, enabled, with the `allow` policy, owned by the caller. Other callers, anonymous ones included, never create authorities.

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
