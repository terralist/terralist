# API Reference

## Liveness Probe

``` 
GET /check/healthz
```

Responds with status 200 OK if the Terralist instance is healthy.

### Example Request

``` shell
curl -L http://localhost:5758/check/healthz
```

### Example Response

=== "Status 200"

    ``` json

    ```

## Readiness Probe

``` 
GET /check/readyz
```

Responds with status 200 OK if the Terralist instance is ready.

### Example Request

``` shell
curl -L http://localhost:5758/check/readyz
```

### Example Response

=== "Status 200"

    ``` json

    ```

## Service Discovery

``` 
GET /.well-known/terraform.json
```

Terraform/OpenTofu service discovery endpoint. Instructs the CLI tool where to find resources.

### Example Request

``` shell
curl -L http://localhost:5758/.well-known/terraform.json
```

### Example Response

=== "Status 200"

    ``` json
    {
      "login.v1": {
        "authz": "/v1/auth/authorization",
        "client": "terraform-cli",
        "grant_types": [
          "authz_code"
        ],
        "ports": [10000, 10010],
        "token": "/v1/auth/token"
      },
      "modules.v1": "/v1/modules/",
      "providers.v1": "/v1/providers/"
    }
    ```

## List all versions for a provider

```
GET /v1/providers/:namespace/:name/versions
```

Get all versions for a provider.

### Example Request

``` shell
curl -L \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/v1/providers/NAMESPACE/NAME/versions
```

### Example Response

=== "Status 200"

    ``` json
    {
      "versions": [
        {
          "version": "5.46.0",
          "protocols": [
            "5.0"
          ],
          "platforms": [
            {
              "os": "linux",
              "arch": "amd64"
            },
            {
              "os": "darwin",
              "arch": "amd64"
            },
            {
              "os": "darwin",
              "arch": "arm64"
            },
            {
              "os": "windows",
              "arch": "amd64"
            }
          ]
        }
      ]
    }

    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 404"

    ``` json
    {
      "errors": "requested provider was not found: no provider found with given arguments (provider hashicorp/aws)"
    }
    ```

## Download provider version

```
GET /v1/providers/:namespace/:name/:version/download/:system/:arch
```

Download a specific provider version.

### Example Request

``` shell
curl -L \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/v1/providers/NAMESPACE/NAME/VERSION/download/SYSTEM/ARCH
```

### Example Response

=== "Status 200"

    ``` json
    {
      "protocols": [
        "5.0"
      ],
      "os": "linux",
      "arch": "amd64",
      "filename": "terraform-provider-aws_5.46.0_linux_amd64.zip",
      "download_url": "https://SOME-BUCKET-NAME.s3.SOME-REGION.amazonaws.com/providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_linux_amd64.zip?X-Amz-Algorithm=[REDACTED]&X-Amz-Credential=[REDACTED]&X-Amz-Date=[REDACTED]&X-Amz-Expires=900&X-Amz-SignedHeaders=host&X-Amz-Signature=[REDACTED]",
      "shasums_url": "https://SOME-BUCKET-NAME.s3.SOME-REGION.amazonaws.com/providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_SHA256SUMS?X-Amz-Algorithm=[REDACTED]&X-Amz-Credential=[REDACTED]&X-Amz-Date=[REDACTED]&X-Amz-Expires=900&X-Amz-SignedHeaders=host&X-Amz-Signature=[REDACTED]",
      "shasums_signature_url": "https://SOME-BUCKET-NAME.s3.SOME-REGION.amazonaws.com/providers/hashicorp/aws/5.46.0/terraform-provider-aws_5.46.0_SHA256SUMS.sigX-Amz-Algorithm=[REDACTED]&X-Amz-Credential=[REDACTED]&X-Amz-Date=[REDACTED]&X-Amz-Expires=900&X-Amz-SignedHeaders=host&X-Amz-Signature=[REDACTED]",
      "shasum": "37cdf4292649a10f12858622826925e18ad4eca354c31f61d02c66895eb91274",
      "signing_keys": {
        "gpg_public_keys": [
          {
            "key_id": "34365D9472D7468F",
            "ascii_armor": "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\n[REDACTED FOR SIMPLICITY]\n-----END PGP PUBLIC KEY BLOCK-----",
            "trust_signature": "",
            "string": "hashicorp",
            "source_url": "https://www.hashicorp.com/security.html"
          }
        ]
      }
    }
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 404"

    ``` json
    {
      "errors": [
        "not found"
      ]
    }
    ```

## Upload a provider version

```
POST /v1/api/providers/:namespace/:name/:version/upload
```

Upload a new provider version.

If the URLs from which the provider files should be downloaded are of types `http` or `https`, a dictionary of headers can be additionally passed, depending on your needs. If those headers are passed-in for other URL types, they will be ignored.

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  -d '{
    "protocols": ["5.0"],
    "headers": {
      "Accept": "application/octet-stream",
      "Authorization": "Bearer {TOKEN}",
      "X-GitHub-Api-Version": "2022-11-28"
    },
    "shasums": {
      "url": "https://api.github.com/repos/{OWNER}/{REPO}/releases/assets/{SHA256SUMS-ASSET-ID}",
      "signature_url": "https://api.github.com/repos/{OWNER}/{REPO}/releases/assets/{SHA256SUMS-SIG-ASSET-ID}",
    },
    "platforms": [
      {
        "os": "linux",
        "arch": "amd64",
        "download_url": "https://api.github.com/repos/{OWNER}/{REPO}/releases/assets/{PROVIDER-LINUX-AMD64-ASSET-ID}",
        "shasum": "{SHASUM}"
      }
    ]
  }' \
  http://localhost:5758/v1/api/providers/NAMESPACE/NAME/VERSION/upload
```

=== "Status 200"

    ``` json
    {
      "errors": []
    }
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 4xx/5xx"

    ``` json
    {
      "errors": [
        "...",
      ]
    }
    ```

## Upload provider packages

```
POST /v1/api/providers/:namespace/:name/:version/upload-files
```

Upload a new provider version from its package files, as produced by `terraform providers mirror`. The request is a multipart form:

- `metadata`: the `<version>.json` document listing the packages and their `h1` hashes;
- `archives`: one or more package archives listed in the document;
- `shasums` and `shasums_signature` (optional): the provider's `SHA256SUMS` file and its signature, always together;
- `protocols` (required with `shasums`): comma separated provider protocol versions.

Every archive must be listed in the document under its file name. When a `SHA256SUMS` file is given, every archive must match its entry. A version uploaded without `shasums` is served through the [network mirror](../user-guide/network-mirror.md) only and does not appear in the registry protocol version list. A providers storage resolver must be configured.

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  -F metadata=@3.2.4.json \
  -F archives=@terraform-provider-null_3.2.4_linux_amd64.zip \
  -F archives=@terraform-provider-null_3.2.4_darwin_arm64.zip \
  -F shasums=@SHA256SUMS \
  -F shasums_signature=@SHA256SUMS.sig \
  -F protocols=5.0 \
  http://localhost:5758/v1/api/providers/hashicorp/null/3.2.4/upload-files
```

### Example Response

=== "Status 200"

    ``` json
    {
      "errors": []
    }
    ```

=== "Status 400"

    ``` json
    {
      "errors": [
        "expecting exactly one version document in the \"metadata\" field"
      ]
    }
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 409"

    ``` json
    {
      "errors": [
        "package terraform-provider-null_3.2.4_linux_amd64.zip does not match its SHA256SUMS entry"
      ]
    }
    ```

## Fetch provider packages from the upstream

```
POST /v1/api/providers/:namespace/:name/:version/fetch
```

Download the given `os_arch` platforms of a version from the upstream registry the authority stands for into storage, so that later installs are served without reaching the upstream. Requires the `create` action on the provider. See [pulling providers through](../user-guide/network-mirror.md#pulling-providers-through-from-an-upstream).

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  -d '{"platforms": ["linux_amd64", "darwin_arm64"]}' \
  http://localhost:5758/v1/api/providers/hashicorp/aws/5.0.0/fetch
```

### Example Response

=== "Status 200"

    ``` json
    {
      "results": [
        {"platform": "linux_amd64"},
        {"platform": "darwin_arm64", "error": "version not allowed by the upstream rules"}
      ]
    }
    ```

=== "Status 400"

    ``` json
    {
      "errors": [
        "expecting at least one os_arch platform to fetch"
      ]
    }
    ```

## Manage upstream rules

```
POST   /v1/api/authorities/:id/rules
DELETE /v1/api/authorities/:id/rules/:ruleId
```

Add or remove a rule deciding which versions an authority serves from its upstream registry. `kind` is `provider` or `module`, `name` and `version` are globs, `effect` is `allow` or `deny`. Both require the `update` action on the authority. The rules of an authority are returned with the authority itself.

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  -d '{"kind": "provider", "name": "aws", "version": "5.*", "effect": "deny"}' \
  http://localhost:5758/v1/api/authorities/AUTHORITY_ID/rules
```

### Example Response

=== "Status 200"

    ``` json
    {
      "id": "b3f1c2d4-...",
      "kind": "provider",
      "name": "aws",
      "version": "5.*",
      "effect": "deny"
    }
    ```

=== "Status 409"

    ``` json
    {
      "errors": [
        "invalid rule kind \"bucket\", expected provider or module"
      ]
    }
    ```

## Remove a provider

```
DELETE /v1/api/providers/:namespace/:name/remove
```

Remove a provider together with all its uploaded versions.

### Example Request

``` shell
curl -L -X DELETE \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  http://localhost:5758/v1/api/providers/NAMESPACE/NAME/remove
```

### Example Response

=== "Status 200"

    ``` json
    {
      "errors": []
    }
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 4xx/5xx"

    ``` json
    {
      "errors": [
        "...",
      ]
    }
    ```

## Remove a provider version

```
DELETE /v1/api/providers/:namespace/:name/:version/remove
```

Remove a specific provider version.

### Example Request

``` shell
curl -L -X DELETE \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  http://localhost:5758/v1/api/providers/NAMESPACE/NAME/VERSION/remove
```

### Example Response

=== "Status 200"

    ``` json
    {
      "errors": []
    }
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 4xx/5xx"

    ``` json
    {
      "errors": [
        "...",
      ]
    }
    ```

## List all versions for a module

```
GET /v1/modules/:namespace/:name/:provider/versions
```

Get all versions for a module.

### Example Request

``` shell
curl -L \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/v1/modules/NAMESPACE/NAME/PROVIDER/versions
```

### Example Response

=== "Status 200"

    ``` json
    {
      "modules": [
        {
          "versions": [
            {
              "version": "5.5.3"
            },
            {
              "version": "5.6.0"
            },
            {
              "version": "5.7.0"
            },
            {
              "version": "5.7.1"
            }
          ]
        }
      ]
    }
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 404"

    ``` json
    {
      "errors": "no module found with given arguments (source terraform-aws-modules/vpc/aws)"
    }
    ```

## Download module version

```
GET /v1/modules/:namespace/:name/:provider/:version/download
```

Download a specific provider version.

### Example Request

``` shell
curl -L \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/v1/modules/NAMESPACE/NAME/PROVIDER/VERSION/download
```

### Example Response

=== "Status 204"

    ``` json
    {}
    ```

    !!! note "The `X-Terraform-Get` header should be set to the correct download link for this module."

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 404"

    ``` json
    {
      "errors": [
        "not found"
      ]
    }
    ```

## Upload a module version

```
POST /v1/api/modules/:namespace/:name/:provider/:version/upload
```

Upload a new module version.

If the URL from which the module files should be downloaded is of types `http` or `https`, a dictionary of headers can be additionally passed, depending on your needs. If those headers are passed-in for other URL types, they will be ignored.

### Example Request

=== "GitHub API"

    ``` shell
    curl -L -X POST \
      -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
      -d '{
        "download_url": "https://api.github.com/repos/{OWNER}/{REPO}/releases/assets/{ASSET-ID}?archive=zip",
        "headers": {
            "Accept": "application/octet-stream",
            "Authorization": "Bearer {TOKEN}",
            "X-GitHub-Api-Version": "2022-11-28"
        }
      }' \
      http://localhost:5758/v1/api/modules/NAMESPACE/NAME/PROVIDER/VERSION/upload
    ```

    !!! note "Notice the `archive=zip` query argument. If you want to instruct Terralist to download the asset from the API, you will also need to manually specify that the asset which is being downloaded is a zip archive."

=== "GitHub HTTP"

    ``` shell
    curl -L -X POST \
      -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
      -d '{
        "download_url": "https://github.com/{OWNER}/{REPO}/archive/refs/tags/{RELEASE-TAG-NAME}.zip",
        "headers": {
            "Accept": "application/octet-stream",
            "Authorization": "Basic {YOUR-GITHUB-BASE64ENC-USERNAME-TOKEN}"
        }
      }' \
      http://localhost:5758/v1/api/modules/NAMESPACE/NAME/PROVIDER/VERSION/upload
    ```

    !!! note "To obtain the basic auth token you can base64-encode the following string: `{your-github-username}:{your-github-pat-with-read-access-to-the-repository}`."

### Example Response

=== "Status 200"

    ``` json
    {
      "errors": []
    }
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 4xx/5xx"

    ``` json
    {
      "errors": [
        "...",
      ]
    }
    ```

## Upload a module version (with local files)

```
POST /v1/api/modules/:namespace/:name/:provider/:version/upload-files
```

Upload a new module version (with local files).

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  -F "module=@/path/to/your-module.zip"
  http://localhost:5758/v1/api/modules/NAMESPACE/NAME/PROVIDER/VERSION/upload-files
```

### Example Response

=== "Status 200"

    ``` json
    {
      "errors": []
    }
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 4xx/5xx"

    ``` json
    {
      "errors": [
        "...",
      ]
    }
    ```

## Fetch a module version from the upstream

```
POST /v1/api/modules/:namespace/:name/:provider/:version/fetch
```

Download a module version from the upstream registry the authority stands for into storage, so that later installs are served without reaching the upstream. Requires the `create` action on the module. See [pulling modules through](../user-guide/network-mirror.md#pulling-modules-through-from-an-upstream).

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  http://localhost:5758/v1/api/modules/hashicorp/dir/template/1.0.2/fetch
```

### Example Response

=== "Status 200"

    ``` json
    {
      "errors": []
    }
    ```

=== "Status 502"

    ``` json
    {
      "errors": [
        "version not allowed by the upstream rules"
      ]
    }
    ```

## Remove a module

```
DELETE /v1/api/modules/:namespace/:name/:provider/remove
```

Remove a module together with all its uploaded versions.

### Example Request

``` shell
curl -L -X DELETE \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  http://localhost:5758/v1/api/modules/NAMESPACE/NAME/PROVIDER/remove
```

=== "Status 200"

    ``` json
    {
      "errors": []
    }
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 4xx/5xx"

    ``` json
    {
      "errors": [
        "...",
      ]
    }
    ```

## Remove a module version

```
DELETE /v1/api/modules/:namespace/:name/:provider/:version/remove
```

Remove a specific module version.

### Example Request

``` shell
curl -L -X DELETE \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  http://localhost:5758/v1/api/modules/NAMESPACE/NAME/PROVIDER/VERSION/remove
```

### Example Response

=== "Status 200"

    ``` json
    {
      "errors": []
    }
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 4xx/5xx"

    ``` json
    {
      "errors": [
        "...",
      ]
    }
    ```

## List provider versions (network mirror)

```
GET /providers/:hostname/:namespace/:name/index.json
```

List all versions of a provider, as defined by the [Provider Network Mirror Protocol](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol). Under Terralist's own hostname the namespace is the authority name; under any other hostname the request is served by the authority standing for that upstream hostname and namespace. See the [Provider Network Mirror](../user-guide/network-mirror.md) guide for details.

### Example Request

``` shell
curl -L -X GET \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  http://localhost:5758/providers/localhost:5758/hashicorp/null/index.json
```

### Example Response

=== "Status 200"

    ``` json
    {
      "versions": {
        "3.2.3": {},
        "3.2.4": {}
      }
    }
    ```

=== "Status 403"

    _Empty body._

=== "Status 404"

    ``` json
    {
      "errors": [
        "requested provider was not found: no provider found with given arguments (provider hashicorp/null)"
      ]
    }
    ```

## List provider installation packages (network mirror)

```
GET /providers/:hostname/:namespace/:name/:version.json
```

List the installation packages of a provider version, as defined by the [Provider Network Mirror Protocol](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol). The `url` of each package points to the storage backend and may be temporary. Each package carries its `zh:` hash, the sha256 of the package, preceded by its `h1:` hash when the uploader provided one.

### Example Request

``` shell
curl -L -X GET \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  http://localhost:5758/providers/localhost:5758/hashicorp/null/3.2.4.json
```

### Example Response

=== "Status 200"

    ``` json
    {
      "archives": {
        "darwin_arm64": {
          "url": "https://bucket.s3.amazonaws.com/providers/hashicorp/null/3.2.4/terraform-provider-null_3.2.4_darwin_arm64.zip?...",
          "hashes": [
            "zh:2a3f1f4b7b0d1e2c9a6f0c3e6d5b4a8f7e6d5c4b3a2f1e0d9c8b7a6f5e4d3c2b"
          ]
        },
        "linux_amd64": {
          "url": "https://bucket.s3.amazonaws.com/providers/hashicorp/null/3.2.4/terraform-provider-null_3.2.4_linux_amd64.zip?...",
          "hashes": [
            "zh:9c8b7a6f5e4d3c2b1a0f9e8d7c6b5a4f3e2d1c0b9a8f7e6d5c4b3a2f1e0d9c8b"
          ]
        }
      }
    }
    ```

=== "Status 403"

    _Empty body._

=== "Status 404"

    ``` json
    {
      "errors": [
        "provider hashicorp/null does not contain version 3.2.4"
      ]
    }
    ```

## List API keys

```
GET /v1/api/api-keys/
```

List all API keys visible to the authenticated user. Results are filtered based on the caller's RBAC policies — only keys for which the user has `get` permission on `api-keys` are returned.

### Example Request

``` shell
curl -L \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  http://localhost:5758/v1/api/api-keys/
```

### Example Response

=== "Status 200"

    ``` json
    [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "ci-key",
        "scope": "team-a",
        "created_by": "admin@example.com",
        "expiration": "",
        "policies": [
          {
            "id": "660e8400-e29b-41d4-a716-446655440001",
            "resource": "modules",
            "action": "*",
            "object": "my-authority/*/*",
            "effect": "allow"
          }
        ]
      }
    ]
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

## Create an API key

```
POST /v1/api/api-keys/
```

Create a API key with RBAC policies. Requires `create` permission on `api-keys` for the specified scope.

The `scope` field is required and determines who can manage the key via RBAC policies (see [API Key Scopes](../user-guide/rbac-configuration.md#api-key-scopes)).

The `expire_in` field is optional and specifies the expiration in hours. If omitted or set to `0`, the key does not expire.

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  -d '{
    "name": "ci-deploy-key",
    "scope": "team-a",
    "expire_in": 720,
    "policies": [
      {
        "resource": "modules",
        "action": "create",
        "object": "my-authority/*/*",
        "effect": "allow"
      },
      {
        "resource": "modules",
        "action": "get",
        "object": "my-authority/*/*",
        "effect": "allow"
      }
    ]
  }' \
  http://localhost:5758/v1/api/api-keys/
```

### Example Response

=== "Status 201"

    ``` json
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "ci-deploy-key",
      "key": "tlk_9rJ0mQ3xK8vN2pL5wT7yB1cD4fG6hJ0kM3nP5qR8sU"
    }
    ```

    !!! note "The `key` is the API key value and is returned only in this response. Store it securely, it cannot be retrieved again. Only a hash of it is kept on the server. The `id` identifies the key for listing and deletion."

=== "Status 400"

    ``` json
    {
      "errors": [
        "policy 0: invalid resource \"foo\"; must be one of: modules, providers, authorities, api-keys"
      ]
    }
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

## Delete an API key

```
DELETE /v1/api/api-keys/:id
```

Delete a API key. Requires `delete` permission on `api-keys`.

### Example Request

``` shell
curl -L -X DELETE \
  -H "Authorization: Bearer x-api-key:<YOUR-TOKEN>" \
  http://localhost:5758/v1/api/api-keys/550e8400-e29b-41d4-a716-446655440000
```

### Example Response

=== "Status 200"

    ``` json
    true
    ```

=== "Status 401"

    ``` json
    {
      "errors": [
        "Authorization: missing",
        "X-API-Key: missing"
      ]
    }
    ```

=== "Status 404"

    ``` json
    {
      "errors": [
        "cannot parse api key"
      ]
    }
    ```
