# Terralist

Terralist is a private Terraform registry for providers and modules that follows the published HashiCorp protocols.

## Features

### Fully private

Terralist's only way to operate is private<sup>*</sup>. It requires authentication for any module/provider operation, including fetching them.
It is integrated with the terraform-cli and allows its users to authenticate with a simple `terraform login` command, via an Oauth provider.
You can also generate API keys for programmatic access.

<div style="font-size: 12px;"><sup>*</sup>If you plan to deploy Terralist in an isolated environment, there is also the option of allowing anonymous (unauthenticated) downloads.</div>

### Securely distributing your data

Terralist can host your modules code and providers binaries either locally or remotely in a private storage environment (e.g. a cloud bucket).
If you opt for a remote storage environment, every time Terralist is asked for a download request, it will ask the cloud environment to generate a temporarily presigned URL, then forward that particular URL to the requester.

### Proxy mode

Terralist can also operate in proxy mode, where Terralist will simply forward any URL that it receives from the creators. This means you can take advantage of the Terraform `version` attribute while storing your modules in a git mono-repository.

### Web Dashboard

Terralist exposes an SPA dashboard to the web, that you can use to control your modules, providers and authorities.
