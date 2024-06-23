# Local Development

Terralist is built in go, which makes it easy to distribute and especially, to run locally, but unfortunately, Terraform/OpenTofu expects the registry to run over HTTPS. If you are developing Terralist and don't plan to use it via a Terraform/OpenTofu interaction, you don't need to follow this document, but otherwise, let's start.

There are multiple options to expose Terralist over HTTPS:

1. Use the built-in TLS support (with `cert-file` and `key-file` configuration options).
  <br />In this case, you will have to bring your own certificate. You can also generate a self-signed certificate.
2. Use a reverse-proxy in front of Terralist that can expose Terralist over HTTPS.
3. Use a managed reverse-proxy in front of Terralist, such as [ngrok](https://ngrok.com/docs/).

It's up to you which method you want to choose. Let's break them down.

## Self-Signed Certificate

For this method you need a self-signed certificate. There are plenty of resources over the internet that can teach you how to do it, for example, you can check [this blog post](https://devopscube.com/create-self-signed-certificates-openssl/).

Once you have your certificate and you configured your computer to trust it, you can configure Terralist with the following two options:

``` yaml title="config.yaml"
cert-file: /path/to/your/cert-in-pem-format
key-file: /path/to/your/cert-key-in-pem-format
```

## Reverse proxy

For this method you need a reverse-proxy in front of Terralist. You can either use NGINX, traefik or others, but keep in mind they might need the same self-signed certificate setup as above. This method is recommended if you already have a reverse-proxy configured and you don't want to redo the certificate setup.

## Managed reverse proxy

There are multiple managed reverse proxy software tools out there, but in this document we will present [ngrok](https://ngrok.com/docs/). Start by following the ngrok's [Quickstart](https://ngrok.com/docs/getting-started/) guide.

Once you have your ngrok setup, you can make it point to your Terralist instance by using:

``` bash
ngrok http 5758
```

!!! note "By default, Terralist is listening on the `5758` port. If you changed it, make sure to update the ngrok command accordingly."

Watch the ngrok output and look for the following line:

```
Forwarding  https://<some-uuid>.ngrok.io -> http://localhost:5758
```

The `<some-uuid>.ngrok.io` value is the one you need. Copy it and leave the ngrok instance running in the background.

Update your Oauth application to use this URL instead (for the callback URL attribute).

!!! warning "If you are using a free ngrok installation, this URL may rotate every time you reboot your ngrok process. Make sure to update the Oauth application every time."

Now, your setup is complete. You may proceed to authenticate from the Terraform/OpenTofu CLI.

=== "Terraform"

    ``` bash
    terraform login <some-uuid>.ngrok.io
    ```

=== "OpenTofu"

    ``` bash
    tofu login <some-uuid>.ngrok.io
    ```

!!! note "Terraform stores the token from the login command in `$HOME/.terraform.d/credentials.tfrc.json`. As you may start and stop your server multiple times and the URL changes everytime, this file will accumulate a lot of garbage entries fast. You can clean it up every once in a while as you please."

Now that you are authenticated, you can start using Terralist in your TF code:

``` hcl
module "example" {
  source  = "<some-uuid>.ngrok.io/NAMESPACE/NAME/PROVIDER"
  version = "1.0.0"
}
```
