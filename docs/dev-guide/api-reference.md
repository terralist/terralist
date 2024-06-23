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
POST /v1/api/providers/:name/:version/upload
```

Upload a new provider version.

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/v1/api/providers/NAME/VERSION/upload
```

!!! note "There is no need for you to specify the namespace, as Terralist will resolve it based on your API key."

### Example Response

=== "Status 200"

    TODO

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

## Remove a provider

```
DELETE /v1/api/providers/:name/remove
```

Remove a provider together with all its uploaded versions.

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/v1/api/providers/NAME/remove
```

!!! note "There is no need for you to specify the namespace, as Terralist will resolve it based on your API key."

### Example Response

=== "Status 200"

    TODO

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

## Remove a provider version

```
DELETE /v1/api/providers/:name/:version/remove
```

Remove a specific provider version.

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/v1/api/providers/NAME/VERSION/remove
```

!!! note "There is no need for you to specify the namespace, as Terralist will resolve it based on your API key."

### Example Response

=== "Status 200"

    TODO

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
POST /v1/api/modules/:name/:provider/:version/upload
```

Upload a new module version.

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/v1/api/modules/NAME/PROVIDER/VERSION/upload
```

!!! note "There is no need for you to specify the namespace, as Terralist will resolve it based on your API key."

### Example Response

=== "Status 200"

    TODO

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
## Remove a module

```
DELETE /v1/api/modules/:name/:provider/remove
```

Remove a module together with all its uploaded versions.

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/v1/api/modules/NAME/PROVIDER/remove
```

!!! note "There is no need for you to specify the namespace, as Terralist will resolve it based on your API key."

### Example Response

=== "Status 200"

    TODO

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

## Remove a module version

```
DELETE /v1/api/modules/:name/:provider/:version/remove
```

Remove a specific module version.

### Example Request

``` shell
curl -L -X POST \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/v1/api/modules/NAME/PROVIDER/VERSION/remove
```

!!! note "There is no need for you to specify the namespace, as Terralist will resolve it based on your API key."

### Example Response

=== "Status 200"

    TODO

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
