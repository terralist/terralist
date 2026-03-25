# Local Development

Terralist is built in Go, which makes it easy to distribute and run locally. However, Terraform/OpenTofu expects the registry to run over HTTPS. If you are developing Terralist and don't plan to interact with it via Terraform/OpenTofu, you don't need to follow this document.

There are multiple options to expose Terralist over HTTPS:

1. **mkcert** (recommended) — generates locally-trusted certificates with zero configuration.
2. **Self-signed certificate** — use the built-in TLS support with `cert-file` and `key-file` configuration options.
3. **Reverse proxy** — use NGINX, Caddy, traefik, or similar in front of Terralist.
4. **Managed reverse proxy** — use a service like [ngrok](https://ngrok.com/docs/).

## mkcert (recommended)

[mkcert](https://github.com/FiloSottile/mkcert) generates locally-trusted development certificates. It automatically creates a local Certificate Authority (CA) and adds it to your system trust store, so both your browser and Terraform trust the certificate without any manual configuration.

### Install mkcert

=== "macOS"

    ``` bash
    brew install mkcert
    ```

=== "Linux"

    ``` bash
    # Debian/Ubuntu
    sudo apt install libnss3-tools
    curl -JLO "https://dl.filippo.io/mkcert/latest?for=linux/amd64"
    chmod +x mkcert-v*-linux-amd64
    sudo mv mkcert-v*-linux-amd64 /usr/local/bin/mkcert
    ```

### Generate certificates

``` bash
mkcert -install
mkcert -cert-file cert.pem -key-file key.pem localhost 127.0.0.1
```

The first command installs the local CA into your system trust store (you only need to do this once). The second generates a certificate valid for `localhost` and `127.0.0.1`.

### Configure Terralist

``` yaml title="config.yaml"
cert-file: cert.pem
key-file: key.pem
```

Start Terralist and it will serve over HTTPS:

``` bash
./terralist server --config config.yaml
```

### Authenticate with Terraform/OpenTofu

=== "Terraform"

    ``` bash
    terraform login localhost:5758
    ```

=== "OpenTofu"

    ``` bash
    tofu login localhost:5758
    ```

!!! note "Terraform requires registry hostnames to contain at least one dot. If you're accessing Terralist at `localhost:5758`, Terraform will reject it. You can use `localhost.direct` instead — it's a public domain that resolves to `127.0.0.1`. Generate the certificate with: `mkcert -cert-file cert.pem -key-file key.pem localhost 127.0.0.1 localhost.direct`"

Once authenticated, you can reference modules and providers:

``` hcl
module "example" {
  source  = "localhost.direct:5758/NAMESPACE/NAME/PROVIDER"
  version = "1.0.0"
}
```

## Self-signed certificate

If you prefer not to use mkcert, you can generate a self-signed certificate manually. There are plenty of resources on the internet for this — for example, [this guide using OpenSSL](https://devopscube.com/create-self-signed-certificates-openssl/).

Once you have your certificate and have configured your system to trust it, configure Terralist:

``` yaml title="config.yaml"
cert-file: /path/to/your/cert.pem
key-file: /path/to/your/key.pem
```

## Reverse proxy

If you already have a reverse proxy (NGINX, Caddy, traefik, etc.) configured with TLS termination, you can place it in front of Terralist. Point the proxy at `http://localhost:5758` and let it handle HTTPS. No certificate configuration is needed in Terralist itself.

## Managed reverse proxy (ngrok)

[ngrok](https://ngrok.com/docs/) provides a public HTTPS URL that tunnels to your local server. Start by following ngrok's [Quickstart](https://ngrok.com/docs/getting-started/) guide.

Once ngrok is set up, point it at your Terralist instance:

``` bash
ngrok http 5758
```

!!! note "By default, Terralist listens on port `5758`. If you changed it, update the ngrok command accordingly."

Watch the ngrok output for the forwarding URL:

```
Forwarding  https://<some-uuid>.ngrok.io -> http://localhost:5758
```

Copy the `<some-uuid>.ngrok.io` value and update your OAuth application's callback URL to use it.

!!! warning "If you are using a free ngrok plan, this URL rotates every time you restart ngrok. You will need to update the OAuth application callback URL each time."

Authenticate from the CLI:

=== "Terraform"

    ``` bash
    terraform login <some-uuid>.ngrok.io
    ```

=== "OpenTofu"

    ``` bash
    tofu login <some-uuid>.ngrok.io
    ```

!!! note "Terraform stores login tokens in `$HOME/.terraform.d/credentials.tfrc.json`. With rotating ngrok URLs, this file accumulates stale entries. Clean it up periodically."

Then use Terralist in your Terraform code:

``` hcl
module "example" {
  source  = "<some-uuid>.ngrok.io/NAMESPACE/NAME/PROVIDER"
  version = "1.0.0"
}
```

## Quick development setup

If you want a fully configured local environment for development and testing, you can use the `test:server` task:

``` bash
task test:server
```

This builds Terralist, starts RustFS (S3-compatible storage) and a mock OAuth2 server via Docker, and runs Terralist with all dependencies configured. Press any key to stop the server and clean up the containers.

You can then run the test suite against it in a separate terminal. See [E2E Testing](./e2e-testing.md) for details.