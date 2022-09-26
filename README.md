<!-- markdownlint-configure-file {
  "MD013": {
    "code_blocks": false,
    "tables": false
  },
  "MD033": false,
  "MD041": false
} -->

<div align="center" markdown="1">

# Terralist

A _truly_ private Terraform registry
<br />

</div>

## About

Terralist is a private Terraform registry for providers and modules following the published HashiCorp protocols. It provides:
* A secure way to distribute your confidential modules and providers;
* [_Soon_] A management interface to visualize documentation;

## Highlights
* **Login Functionality** ([docs](https://www.terraform.io/docs/internals/login-protocol.html)): Require a token to access the data. It is integrated with Terraform, so you can authenticate to the registry directly through Terraform:
  ```
  $ terraform login registry.example.com
  $ terraform logout registry.example.com
  ```
  It can also generate custom API keys for an authenticated user, which can be used in pipelines: to upload and delete modules and providers; to fetch data - useful if you are using a Terraform Pull Request Automation tool.

* **Modules Registry**: ([docs](https://www.terraform.io/docs/internals/module-registry-protocol.html)) Stores modules data in a *private* storage (for example, an S3 bucket). When download request is received, calls the remote storage to generate a temporary public download URL and forwards the URL to the requester.
  Currently supported private storage:
  * AWS S3: uses a private S3 bucket
  * Proxy: forwards the URL received at creation

* **Provider Registry**: ([docs](https://www.terraform.io/docs/internals/provider-registry-protocol.html)) Similar with modules registry.
  Currently supported private storage:
  * Proxy: forwards the URL received at creation

_Note_: For _Proxy_ storage mode, the URL management is up to you. If, for example, you are providing a git URL, then the same URL will be forwarded to the requester.

## Disclaimer

This project is not meant to replace the public Terraform Registry. Its purpose is to mimic the public registry in a private environment.

## Build

### Release Mode
```
task build -- release
```

### Debug Mode
Debug mode provides additional logging, but decrease the overall performance.
```
task build -- debug
```

A `terralist[.exe]` file should be generated in the repository root directory.

If you cannot use the build script, you can either run the `go build` command manually. Build-time variables are, for now, optionals; they provide a default value.

## Examples

### Authenticate
```
$ terraform login registry.example.com
```

### Create an authority

Use the application web interface to authenticate using your third-party OAUTH 2.0 provider, then create an authority and assign a GPG key to it.

### Generate an API Key

Use the application web interface to generate an API key. You can allocate more than one API key for an authority and use them to upload modules and providers under that specific authority (namespace).

### Upload a new module
```
$ curl -X POST registry.example.com/v1/api/modules/my-module/provider/1.0.0/upload \
       -H "Authorization: Bearer $TERRALIST_API_KEY" \
       -d '{ "download_url": "/home/bob/terraform-modules/example-module" }'
```

### Use the module
```
module "example-module" {
  source  = "registry.example.com/example/my-module/provider"
  version = "1.0.0"

  // ...
}
```

### Upload a new provider

1. Create a new file with the API JSON body (`~/random-2.0.0.json`):
```json
{
  "protocols": [
    "4.0",
    "5.1"
  ],
  "shasums": {
    "url": "https://releases.hashicorp.com/terraform-provider-random/2.0.0/terraform-provider-random_2.0.0_SHA256SUMS",
    "signature_url": "https://releases.hashicorp.com/terraform-provider-random/2.0.0/terraform-provider-random_2.0.0_SHA256SUMS.sig"
  },
  "platforms": [
    {
      "os": "darwin",
      "arch": "amd64",
      "download_url": "https://releases.hashicorp.com/terraform-provider-random/2.0.0/terraform-provider-random_2.0.0_darwin_amd64.zip",
      "shasum": "55ced41e5f68730ef36272d4953f336a50f318c1d1d174665f5fa76cb5df08ae"
    },
    {
      "os": "linux",
      "arch": "amd64",
      "download_url": "https://releases.hashicorp.com/terraform-provider-random/2.0.0/terraform-provider-random_2.0.0_linux_amd64.zip",
      "shasum": "5f9c7aa76b7c34d722fc9123208e26b22d60440cb47150dd04733b9b94f4541a"
    }
  ]
}
```

2. Upload the provider
```
$ curl -X POST registry.example.com/v1/providers/random/2.0.0/upload \
       -H "Authorization: Bearer $TERRALIST_API_KEY" \
       -d "$(cat ~/random-2.0.0.json)"
```

### Use the provider
```
terraform {
  required_providers {
    aws = {
      source  = "registry.example.com/hashicorp/random"
      version = "2.0.0"
    }
  }
}
```

## Endpoints

* `GET /health`: Health Endpoint
* `GET /.well-known/terraform.json`: Terraform Service Discovery endpoint

* `GET /v1/providers/:namespace/:name/versions`: List all versions for a provider
* `GET /v1/providers/:namespace/:name/:version/download/:system/:arch`: Download a specific provider version
* `POST /v1/api/providers/:name/:version/upload`: Upload a new provider version
* `DELETE /v1/api/providers/:name/remove`: Remove a provider
* `DELETE /v1/api/providers/:name/:version/remove`: Remove a provider version


* `GET /v1/modules/:namespace/:name/:provider/versions`: List all versions for a module
* `GET /v1/modules/:namespace/:name/:provider/:version/download`: Download a specific module version
* `POST /v1/api/modules/:name/:provider/:version/upload`: Upload a new modules version
* `DELETE /v1/api/modules/:name/:provider/remove`: Remove a modules
* `DELETE /v1/api/modules/:name/:provider/remove`: Remove a modules version

## Work In Progress

This project is still work-in-progress and I am planning to release it soon.

## Planned Features
* Ability to create an API key to use instead of a Bearer Token;
* A containerized version;
* Web interface to manage the resources;
* Web interface to visualize modules and providers documentation;
* Replace PostgreSQL with a lighter database;
* Multiple authorities support;
* S3 support for providers;
* Google OAUTH 2.0 provider;
* Web documentation for the entire project;

## Contributions

Each contribution is welcomed, if you want to contribute, open an issue or fork the repository and open a PR.