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

## List API keys

```
GET /v1/api/api-keys/
```

List all standalone API keys visible to the authenticated user. Results are filtered based on the caller's RBAC policies — only keys for which the user has `get` permission on `api-keys` are returned.

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

Create a standalone API key with RBAC policies. Requires `create` permission on `api-keys` for the specified scope.

The `scope` field is required and determines who can manage the key via RBAC policies (see [API Key Scopes](/user-guide/rbac-configuration/#api-key-scopes)).

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
      "name": "ci-deploy-key"
    }
    ```

    !!! note "The `id` is the API key value. Store it securely — it cannot be retrieved again."

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

Delete a standalone API key. Requires `delete` permission on `api-keys`.

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
