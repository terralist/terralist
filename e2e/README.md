# End-to-end Testing

End-to-end tests for Terralist. The suite covers three areas:

- **API tests** (Go) — registry protocol, module/provider CRUD, probes, service discovery
- **Frontend tests** (Playwright) — authentication, dashboard, settings
- **Terraform integration tests** (Go) — real `terraform init`, `plan`, and `apply` against Terralist

## Prerequisites

- Go 1.25+
- Node.js (for Playwright)
- Docker (for RustFS and mock OAuth2 server)
- [mkcert](https://github.com/FiloSottile/mkcert) (for Terraform tests, which require HTTPS)
- Terraform CLI (for Terraform integration tests)

## Running locally

### 1. Start the test server

The `test:server` task builds Terralist with coverage instrumentation and starts all dependencies (RustFS for S3 storage, mock OAuth2 server for OIDC authentication):

```bash
task test:server
```

This starts the server on `http://localhost:5758` and blocks until you press a key to stop.

### 2. Run the tests

In a separate terminal:

```bash
# API and Terraform integration tests
task test:e2e

# Playwright frontend tests
cd e2e/frontend && npm ci && npx playwright test
```

### Terraform integration tests

Terraform requires HTTPS for registry access. To run the Terraform tests locally:

```bash
mkcert -install
mkcert -cert-file /tmp/terralist.pem -key-file /tmp/terralist-key.pem localhost 127.0.0.1 localhost.direct
```

Then start the server with TLS:

```bash
./terralist server --config=config.yaml --cert-file /tmp/terralist.pem --key-file /tmp/terralist-key.pem
```

And run with the HTTPS URL:

```bash
TERRALIST_URL=https://localhost:5758 go test -v -count=1 ./e2e/...
```

The Terraform tests use `localhost.direct` as the registry hostname (a public domain that resolves to `127.0.0.1`) because Terraform requires registry hostnames to contain at least one dot. Tests skip gracefully when HTTPS is not configured or the `terraform` binary is not in PATH.

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `TERRALIST_URL` | `http://localhost:5758` | Terralist server URL |
| `TERRALIST_MASTER_API_KEY` | `e2e-master-api-key-00000000-...` | Master API key for bootstrapping |
| `TERRALIST_S3_ENDPOINT` | `http://localhost:9000` | S3-compatible storage endpoint |
| `TERRALIST_S3_BUCKET_NAME` | `terralist` | S3 bucket name |
| `TERRALIST_S3_BUCKET_REGION` | `us-east-1` | S3 bucket region |
| `TERRALIST_S3_ACCESS_KEY_ID` | `AKIAIOSFODNN7EXAMPLE` | S3 access key |
| `TERRALIST_S3_SECRET_ACCESS_KEY` | `wJalrXUtnFEMI/K7MDENG/...` | S3 secret key |

## Test data

Tests bootstrap their own data at startup — no database snapshots or fixture files needed. The Go `TestMain` function:

1. Creates an S3 bucket via the AWS SDK
2. Creates a `hashicorp` authority via the master API key
3. Fetches the null provider (v3.2.4) metadata from the Terraform registry and uploads it
4. Uploads the `hashicorp/subnets/cidr` module (v1.0.0) from GitHub

## Coverage

To generate a merged coverage report from unit and e2e tests:

```bash
task test:coverage
```

This requires both `coverage/unit.txt` (from `task test:unit -- --cover`) and `coverage/e2e/` (from running the server with `GOCOVERDIR` set) to be present.
