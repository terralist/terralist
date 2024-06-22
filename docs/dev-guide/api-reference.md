# API Reference

## Liveness Probe

``` 
GET /check/healthz
```

Responds with status 200 OK if the Terralist instance is healthy.

### Example Request

``` shell
curl -L \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/check/healthz
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
curl -L \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/check/readyz
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
curl -L \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  http://localhost:5758/.well-known/terraform.json
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
      
    }
    ```

=== "Status 404"

    ``` json
    
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
      
    }
    ```

=== "Status 404"

    ``` json

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

    ``` json
    {
      
    }
    ```

=== "Status 404"

    ``` json

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

    ``` json
    {
      
    }
    ```

=== "Status 404"

    ``` json

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

    ``` json
    {
      
    }
    ```

=== "Status 404"

    ``` json

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
      
    }
    ```

=== "Status 404"

    ``` json
    
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

=== "Status 200"

    ``` json
    {
      
    }
    ```

=== "Status 404"

    ``` json

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

    ``` json
    {
      
    }
    ```

=== "Status 404"

    ``` json

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

    ``` json
    {
      
    }
    ```

=== "Status 404"

    ``` json

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

    ``` json
    {
      
    }
    ```

=== "Status 404"

    ``` json

    ```
