# Getting Started

## Install Terralist

### Manual Installation

To compile the source code, you will need the golang compiler. Check out the [official documentation](https://go.dev/doc/install) to see how to install it on your machine.

We use Taskfile to automatically set the compile-time variables. You can find the official installation documentation [here](https://taskfile.dev/installation/).

If you don't want to use the build task, you can either run the `go build` command manually. Build-time variables are, for now, optionals; they provide a default value.

#### Release Mode
```
task build -- release
```

#### Debug Mode
Debug mode provides additional logging, but decrease the overall performance.
```
task build -- debug
```

A `terralist[.exe]` file should be generated in the repository root directory.

### Download the binary from the release page

_Soon_.

### Docker

_Soon_.

## Set the OAUTH provider

Terraform authenticates users with [oauth 2.0](https://oauth.net/2/). You will need credentials of an Oauth Application from our supported providers:

+ [GitHub](https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app)
+ [BitBucket](https://developer.atlassian.com/cloud/bitbucket/oauth-2/)
+ [GitLab](https://docs.gitlab.com/ee/integration/oauth_provider.html#create-an-instance-wide-application)
  - the `email` and `openid` scopes must assigned for the gitlab oauth application

For local development, you can set the homepage URL to `http://localhost:5758` and the callback URL to `http://localhost:5758/v1/api/auth/redirect`.

_Note_: The port `5758` is the default. If you decide to change it, you will also need to change it in the Oauth App settings.

## Configuration

Terralist has a lot of configuration options. Take a look over all variables [here](./CONFIGURATION.md).

To get a quick view, you can rely on the default ones and only set the required variables:

+ `oauth-provider`: The [Oauth provider](#set-the-oauth-provider) you configured;
+ The Oauth provider configuration - it depends on what provider you selected (e.g. for GitHub, `gh-client-id` and `gh-client-secret`)
+ `token-signing-secret`: A random string to protect the tokens;
+ `cookie-secret`: A random string to protect the cookies;
  <br/> _Note_: The `cookie` store is the default session storage;

You can set the configuration using:
+ CLI arguments: Add a `--` in front of each variable and pass the value (e.g. `--oauth-provider`);
+ Environment Variables: Replace each dash (`-`) with an underscore (`_`), uppercase everything and add a `TERRALIST_` prefix (e.g. `TERRALIST_OAUTH_PROVIDER`);
+ Configuration file: Create a `.yaml` file with your variables and pass it with a `--config` argument (or set the `TERRALIST_CONFIG` environment variable);
  <br />E.g.:
  ```yaml
  oauth-provider: github

  token-signing-secret: 'MySuperSecret'
  ```

_Note_: You can choose to mix all those options.

## Launch the server

To launch the server in execution, you can run the following command:

+ UNIX & UNIX-like
  ```console
  ./terralist server --config config.yaml
  ```
+ Windows
  ```powershell
  ./terralist.exe server --config config.yaml
  ```

If the server correctly started, you should see the following log line:
```json
{"level":"info","time":"---","message":"Terralist started, listening on port 5758"}
```

## How to use it

### Running on a local development environment

Since the terraform cli expects all responses to be from an HTTPS server, the standard `localhost:5758` will return an error when trying to login. Specifically: 

```console
terraform login localhost:5758                                                                      
│ Error: Service discovery failed for localhost:5758
│
│ Failed to request discovery document: Get "https://localhost:5758/.well-known/terraform.json": http: server gave HTTP response to HTTPS client.
```

To work around this, a proxy with a trusted SSL certificate or a service like `ngrok` should be used. See [local-testing](./LOCAL-TESTING.md) for all the details.

### Authenticate
```console
terraform login registry.example.com
```

### Create an authority

Use the application web interface to authenticate using your third-party oauth provider, then create an authority and assign a GPG key to it.

### Generate an API Key

Use the application web interface to generate an API key. You can allocate more than one API key for an authority and use them to upload modules and providers under that specific authority (namespace).

### Upload a new module
```console
curl -X POST registry.example.com/v1/api/modules/my-module/provider/1.0.0/upload \
     -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
     -d '{ "download_url": "/home/bob/terraform-modules/example-module" }'
```

### Use the module
```hcl
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
```console
curl -X POST registry.example.com/v1/api/providers/random/2.0.0/upload \
     -H "Authorization: Bearer x-api-key:$TERRALIST_API_KEY" \
     -d "$(cat ~/random-2.0.0.json)"
```

### Use the provider
```hcl
terraform {
  required_providers {
    aws = {
      source  = "registry.example.com/hashicorp/random"
      version = "2.0.0"
    }
  }
}
```
