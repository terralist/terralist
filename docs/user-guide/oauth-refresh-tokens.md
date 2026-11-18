# OAuth Refresh Tokens

Terraform and OpenTofu keep only the access token from `terraform login`, so they never refresh it. Other OAuth clients can get refresh tokens, so their users stay logged in without long-lived access tokens.

Refresh tokens are off by default. When they are on, every refresh is checked with the identity provider first: a user who was removed there, or who lost the group, team or workspace that gave them access, can no longer refresh.

## Enable refresh tokens

```yaml
oauth-refresh-enabled: true
oauth-refresh-secret: "a-random-secret-of-at-least-32-characters"

# Optional, shown with their defaults
oauth-refresh-access-token-expiration: "1h"
oauth-refresh-token-expiration: "720h"
oauth-refresh-max-session-age: "2160h"
```

See [Configuration](../configuration.md#oauth-refresh-enabled) for each option.

## Identity providers

| Provider | How Terralist checks the user on refresh |
| --- | --- |
| GitHub | Fetches the user, organization and teams again with the access token from the login. GitHub Apps with expiring user tokens also refresh that token. |
| GitLab | Refreshes the GitLab token, then fetches the user and groups again. |
| BitBucket | Refreshes the BitBucket token, then fetches the user and workspace again. |
| OIDC | Requests the `offline_access` scope, refreshes the token, then fetches the user info again. The identity provider must allow `offline_access` for the client; a login without an upstream refresh token gets no Terralist refresh token. |
| SAML | Not supported. Terralist refuses to start with `oauth-refresh-enabled`. |

## Client flow

The client uses the same authorization code flow with PKCE as Terraform, from the endpoints in `/.well-known/terraform.json`. It must:

- send a `client_id` other than `terraform-cli`, which identifies Terraform and OpenTofu;
- listen on a loopback address on one of the ports in the discovery document (`10000` to `10010`) to receive the code.

The token response then holds a `refresh_token`, and its `expires_in` follows `oauth-refresh-access-token-expiration`.

### Refresh

```shell
curl -X POST https://registry.example.com/v1/auth/token \
  -d grant_type=refresh_token \
  -d client_id=my-client \
  -d refresh_token="$REFRESH_TOKEN"
```

Each refresh returns a new access token and a new refresh token. The previous refresh token stops working; store the new one.

| Response | Meaning |
| --- | --- |
| `200` | New tokens. |
| `400` `invalid_grant` | The refresh token is unknown, expired or revoked, or the identity provider refused the user. Log in again. |
| `503` `temporarily_unavailable` | The identity provider could not be reached. The refresh token still works; retry later. |

Refresh tokens are bound to the `client_id` that logged in. If a refresh token is used again after it was replaced, Terralist treats it as stolen and revokes the session.

### Revoke

On logout, revoke the refresh token ([RFC 7009](https://datatracker.ietf.org/doc/html/rfc7009)):

```shell
curl -X POST https://registry.example.com/v1/auth/revoke \
  -d client_id=my-client \
  -d token="$REFRESH_TOKEN"
```

The endpoint answers `200` whether or not the token was known, and only exists while refresh tokens are enabled.

## Security notes

- Terralist stores only hashes of its refresh tokens. The identity provider grants are encrypted with `oauth-refresh-secret`.
- Access tokens cannot be revoked. Keep `oauth-refresh-access-token-expiration` short: it is how long a removed user keeps access.
- Turning `oauth-refresh-enabled` off removes every refresh session on the next start.
