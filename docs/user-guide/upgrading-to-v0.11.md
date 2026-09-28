# Upgrading to v0.11

v0.11 removes the features deprecated in v0.10 and changes how names, groups and fetch sources are handled. Most deployments need one configuration change (`oauth-state-secret`) and a look at their RBAC policy. This guide walks through everything to check, in the order to do it.

Upgrade from v0.10.10. Earlier v0.10 releases should be upgraded to v0.10.10 first.

## Before upgrading

Do these steps while v0.10 is still running. Some of them cannot be done once v0.11 refuses to start.

### Back up the database

v0.11 changes the schema on startup: it drops the `authority_api_keys` table and replaces the unique indexes of authorities, providers and modules. On MySQL, schema changes are not transactional, so a migration that fails halfway cannot be rolled back. Take a backup you can restore.

### Remove names that differ only in case

v0.11 holds authorities, providers and modules unique regardless of case, and refuses to start on a database holding rows that differ only in case. It names them in the error and changes nothing.

These queries list such rows:

```sql
-- Authorities
SELECT LOWER(name), COUNT(*) FROM authorities
GROUP BY LOWER(name) HAVING COUNT(*) > 1;

-- Authorities standing for the same upstream namespace
SELECT LOWER(upstream_hostname), LOWER(upstream_namespace), COUNT(*) FROM authorities
WHERE upstream_hostname IS NOT NULL AND upstream_namespace IS NOT NULL
GROUP BY LOWER(upstream_hostname), LOWER(upstream_namespace) HAVING COUNT(*) > 1;

-- Providers of one authority
SELECT authority_id, LOWER(name), COUNT(*) FROM providers
GROUP BY authority_id, LOWER(name) HAVING COUNT(*) > 1;

-- Modules of one authority
SELECT authority_id, LOWER(name), LOWER(provider), COUNT(*) FROM modules
GROUP BY authority_id, LOWER(name), LOWER(provider) HAVING COUNT(*) > 1;
```

Delete the extra ones from the web UI or the API. Deleting an authority also deletes every artifact uploaded to it, so move the artifacts you want to keep to the remaining authority first.

Since v0.10 already looked providers and modules up regardless of case, such duplicates were answering requests ambiguously. Terraform lowercases provider source addresses, so an authority created as `HashiCorp` never matched the requests made for it.

### Replace authority-linked API keys

Authority-linked API keys stop authenticating in v0.11, and their table is dropped on startup. v0.10 logged this warning on every use of one:

```
Authority-linked API keys are deprecated and will be removed in the next release. Please migrate to standalone API keys with RBAC policies.
```

Search your logs for it to find the clients still using one. For each of them, create an API key from the Settings page with the policies it needs (see [API Key Policies](rbac-configuration.md#api-key-policies)), and switch the client to it.

Clients calling `POST /v1/api/authorities/:id/api-keys` or `DELETE /v1/api/authorities/:id/api-keys/:apiKey` must stop, as both routes are gone. Authority responses no longer include `api_keys`.

API keys created from the Settings page keep working unchanged.

## Configuration changes

### Set `oauth-state-secret`

[`oauth-state-secret`](../configuration.md#oauth-state-secret) is required. v0.10 derived it from `token-signing-secret` when it was not set, and logged a warning about it. v0.11 refuses to start without it.

Any random string works:

```shell
export TERRALIST_OAUTH_STATE_SECRET="$(openssl rand -hex 16)"
```

The Docker entrypoint also reads it from the file named by `TERRALIST_OAUTH_STATE_SECRET_FILE`. Logins in progress during the upgrade fail and must be started again. Existing sessions are not affected.

### Move fetch sources to HTTP(S) or git

Terralist now fetches artifacts only from HTTP(S) URLs and git repositories. This applies to upload URLs, webhook release assets and upstream registries. Uploads from `s3::`, `gcs::` or `hg::` sources are refused.

If you upload from such sources, download the artifacts first, then upload them over HTTP(S), from a git repository, or with the local files upload.

Git hosts are resolved before cloning and refused when they resolve to a private address, as HTTP(S) hosts already were. To upload from a git server in your own network, set [`fetch-allow-private-addresses`](../configuration.md#fetch-allow-private-addresses).

## RBAC policy changes

### Reference groups with `group:`

SAML and OAuth groups are referenced with the `group:` prefix instead of `role:`, so a group can no longer collide with a username or a role. Update every `g` and `p` line of the server-side policy that names a group:

```diff
- g, role:ENGINEERING_ADMINS, role:admin
+ g, group:ENGINEERING_ADMINS, role:admin

- p, role:platform-team, providers, create, platform/*, allow
+ p, group:platform-team, providers, create, platform/*, allow
```

Lines naming roles, usernames or emails do not change. A group left with the `role:` prefix silently matches nobody, so its members lose the access it granted. After upgrading, log in as a member of each group and check what they can see.

### Check policies that depend on case

Policy objects and upstream rule names now match regardless of case, as artifacts are looked up. A deny policy on `Internal/*` also applies to `internal/...`, an allow policy on `team/*` also grants `Team/...`, and an upstream rule on `AWS` applies to `aws`. Rule versions are still matched exactly.

This only changes access for policies that used case to tell names apart. Since names are looked up regardless of case, those names were already the same artifacts.

## After upgrading

Start v0.11 and check its log. On success, the migration finishes before the server listens. On failure, the error names what blocked it, usually rows that differ only in case, or a missing `oauth-state-secret`.

Then check that:

- Terraform can still install providers and modules through Terralist (`terraform init -upgrade` on a configuration that uses it);
- each identity provider group still has the access its policy lines grant;
- CI jobs and other clients authenticate with their API keys.

Authorities created with uppercase letters now serve the requests addressed to them in lowercase. If such an authority stands for an upstream namespace, pull-through starts working for it, which may fetch versions that v0.10 never pulled.
