# Testing on your `localhost`

## The Problem

When running the `terraform login` command with `localhost` as the endpoint, the following error occurs:

```bash
terraform login localhost:5758                                                                      
│ Error: Service discovery failed for localhost:5758
│
│ Failed to request discovery document: Get "https://localhost:5758/.well-known/terraform.json": http: server gave HTTP response to HTTPS client.
```

Since the terraform cli client only expects responses on `https` this command will not work.

## Options to solve the problem

To be able to do such on your local development, you either have to:

1. create a trusted certificate and put a proxy in front of the Terralist server, which serves content on HTTPS
2. use some tooling like ngrok.

### Using `ngrok`

- After you spin up your Terralist server (let's say, on port 5758 - the default one), open a new ngrok instance with the command:

```bash
ngrok http 5758
```

- Then copy the subdomain from the ngrok output, which should look like this:

```bash
...
Forwarding  https://<some-uuid>.ngrok.io -> http://localhost:5758
...

```

- The `<some-uuid>.ngrok.io` can be used to authenticate to the local Terralist server:

```bash
terraform login <some-uuid>.ngrok.io
```

- When you close the server, don't forget to remove it from the `$HOME/.terraform.d/credentials.tfrc.json`, since that ngrok UUID is random and you will receive a new one each time you start the server.

- Also, the artifacts can be accessed from Terraform in the same way:

```json
module "example" {
  source  = "<some-uuid>.ngrok.io/namespace/name/provider"
  version = "1.0.0"
}
```

- For the rest of the API (endpoints under /v1/api), you can access them directly from the localhost:5758 address, since HTTPS is not required.
