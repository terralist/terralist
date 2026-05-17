# VCS release webhooks

Terralist can create a new **module** or **provider** version when GitHub sends a **release** webhook. Only **GitHub** is supported today; the URL includes a `:vcs` segment for future providers, but the server must be configured with [`vcs-provider: github`](../configuration.md#vcs-provider).

The registry target (authority namespace and module or provider identity) is taken from the URL path; the release payload supplies the tag, assets, and archive URLs.

Webhook routes are **not** protected by normal API authentication. Optionally verify inbound requests with [`gh-webhook-secret`](../configuration.md#gh-webhook-secret) (`X-Hub-Signature-256`). Outbound downloads from private repositories use [`gh-access-token`](../configuration.md#gh-access-token) or a [GitHub App](#github-app-outbound-auth).

## Prerequisites

Set [`vcs-provider`](../configuration.md#vcs-provider) to `github` and configure outbound credentials as below. If `vcs-provider` is left empty, no VCS provider is initialized and webhook handlers are not usable.

## Endpoints

Base URL: `{terralist-url}/v1` (same host as `url` in [configuration](../configuration.md#url)).

| Kind | Method and path |
| --- | --- |
| Module | `POST /api/modules/:namespace/:name/:provider/webhook/:vcs` |
| Provider | `POST /api/providers/:namespace/:name/webhook/:vcs` |

Use `:vcs` = `github` (case-insensitive). Other values are not implemented.

- **`namespace`** is the **authority** name (it must already exist in Terralist).
- **`name`** is the module or provider short name in the registry.
- **`provider`** (modules only) is the Terraform provider suffix for the module (for example `aws`).

**Example** (module): `https://registry.example.com/v1/api/modules/my-corp/network/aws/webhook/github`

**Example** (provider): `https://registry.example.com/v1/api/providers/my-corp/null/webhook/github`

## Behavior summary

- Only `release` events with `action: published` are handled; draft releases are ignored.
- The module archive is taken from `zipball_url`, falling back to `tarball_url`.
- Provider binaries come from release **assets** plus a fetched checksums file (see [Provider release artifacts](#provider-release-artifacts)).
- The Git tag must normalize to a valid semantic version (a leading `v` is stripped), matching manual uploads.

If the module or provider does not exist yet, it is created on first successful upload, same as the authenticated upload API.

## Outbound credentials (private repositories)

Terralist uses these settings when **fetching** archives, assets, or checksum files from GitHub—not for verifying the inbound webhook.

Startup validation requires **either** [`gh-access-token`](../configuration.md#gh-access-token) **or** all three: [`gh-app-id`](../configuration.md#gh-app-id), [`gh-app-installation-id`](../configuration.md#gh-app-installation-id), and [`gh-app-private-key-path`](../configuration.md#gh-app-private-key-path).

### Personal access token

Set [`gh-access-token`](../configuration.md#gh-access-token). Terralist sends `Authorization: Bearer <token>` on outbound fetches.

The token must be allowed to **read** the repository contents needed for downloads: for private repos, a classic PAT needs at least **`repo`**, or fine-grained access with **Contents: Read** on that repository. Public repositories do not require a token.

### GitHub App outbound auth

Set [`gh-app-id`](../configuration.md#gh-app-id), [`gh-app-installation-id`](../configuration.md#gh-app-installation-id), and [`gh-app-private-key-path`](../configuration.md#gh-app-private-key-path).

The App must be **installed** on the organization or user account that owns the repository, with permission to read repository contents for private repos. Use the numeric **App ID** from the App settings page and the **Installation ID** for that installation.

The private key file must be PEM for the App’s RSA key (PKCS#1 or PKCS#8).

Installation access tokens are requested from GitHub’s REST API at `https://api.github.com/app/installations/{installation_id}/access_tokens`. GitHub Enterprise Server may use a different API host for App tokens than for repository webhooks; confirm compatibility with your deployment.

## Provider release artifacts

Provider publishing expects release assets consistent with manual provider uploads:

- One zip per platform: `terraform-provider-<name>_<version>_<os>_<arch>.zip` (the path `:name` is the short provider name, e.g. `null` for the `null` provider).
- A checksums file named **`terraform-provider-<name>_<version>SHA256SUMS`** (no `_` before `SHA256SUMS`) listing those zips and their SHA-256 hashes. Terralist fetches it by URL from the matching release asset.
- Optionally **`terraform-provider-<name>_<version>SHA256SUMS.sig`** for signing metadata.

Unknown `os` / `arch` tokens in filenames are skipped.

## Configuring GitHub

1. In the repository, open **Settings → Webhooks → Add webhook** (or configure an organization webhook scoped to the repositories that should publish).
2. **Payload URL**: one of the [endpoints](#endpoints) above, with `github` as `:vcs` and path segments set to your authority name, module or provider name, and module provider suffix as needed.
3. **Content type**: `application/json`.
4. **Secret**: optional; if set, it must match [`gh-webhook-secret`](../configuration.md#gh-webhook-secret). GitHub sends `X-Hub-Signature-256`.
5. **Which events**: choose **Let me select individual events** and enable **Releases** (Terralist only acts on **published** releases).

Save the webhook. Creating or publishing a release whose tag is a valid semver should trigger a module or provider version upload.

## Security notes

- Leaving [`gh-webhook-secret`](../configuration.md#gh-webhook-secret) empty accepts any caller who can reach the endpoint; use a secret in production when the URL is exposed.
