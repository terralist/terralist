# API Endpoints

## General Purpose

* `GET /health`: Health Endpoint

## Service Discovery

* `GET /.well-known/terraform.json`: Terraform Service Discovery endpoint

## Providers

* `GET /v1/providers/:namespace/:name/versions`: List all versions for a provider
* `GET /v1/providers/:namespace/:name/:version/download/:system/:arch`: Download a specific provider version
* `POST /v1/api/providers/:name/:version/upload`: Upload a new provider version
* `DELETE /v1/api/providers/:name/remove`: Remove a provider
* `DELETE /v1/api/providers/:name/:version/remove`: Remove a provider version

## Modules

* `GET /v1/modules/:namespace/:name/:provider/versions`: List all versions for a module
* `GET /v1/modules/:namespace/:name/:provider/:version/download`: Download a specific module version
* `POST /v1/api/modules/:name/:provider/:version/upload`: Upload a new modules version
* `DELETE /v1/api/modules/:name/:provider/remove`: Remove a modules
* `DELETE /v1/api/modules/:name/:provider/remove`: Remove a modules version
