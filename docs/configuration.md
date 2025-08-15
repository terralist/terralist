# Configuration

Terralist supports multiple types of configuration:

- CLI arguments <br/>
  Set the option by passing it with the `--` prefix on the CLI command (e.g. `--port`).

- Environment Variable <br/>
  Any option can be set using an environment variable. To do such, replace any dash (`-`) with an underscore (`_`), uppercase everything and add the `TERRALIST_` prefix (e.g. `TERRALIST_PORT`).

- Configuration File <br/>
  Set all options you want to a configuration file, then pass the path to the configuration file using the `config` option (`--config` argument or `TERRALIST_CONFIG` environment variable). Supported file formats: JSON, TOML, YAML, HCL, INI, envfile and Java Properties files.

It is also possible to mix those types.

Terralist also supports reading the environment at run-time. For example, if you only know the port value at run-time (e.g. you are running on Heroku), you can set the `TERRALIST_PORT` environment variable to `${PORT}`; this instruction will inform Terralist to read the value, at run-time, from the environment variable called `PORT`. It is also possible to set a default value, in case the given one is not present, by using a colon (`:`), example: `${PORT:5758}`.

## Options

### `config`

Path to YAML config file where flag values are set.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--config` |
| env | `TERRALIST_CONFIG` |

### `log-level`

The log level.

| Name | Value |
| --- | --- |
| type | select |
| choices | `trace`, `debug`, `info`, `warn`, `error` |
| required | no |
| default | `info` |
| cli | `--log-level` |
| env | `TERRALIST_LOG_LEVEL` |

### `port`

The port to bind to.

| Name | Value |
| --- | --- |
| type | int |
| required | no |
| default | `5758` |
| cli | `--port` |
| env | `TERRALIST_PORT` |

### `url`

The URL that Terralist is accessible from.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `http://localhost:5758` |
| cli | `--url` |
| env | `TERRALIST_URL` |

### `home`

The path to the directory where Terralist can store files.

| Name | Value |
| --- | --- |
| type | string |
| required | yes |
| default | `$HOME/.terralist.d` |
| cli | `--home` |
| env | `TERRALIST_HOME` |


### `cert-file`

The path to the certificate file (pem format).

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--cert-file` |
| env | `TERRALIST_CERT_FILE` |

### `key-file`

The path to the certificate key file (pem format).

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--key-file` |
| env | `TERRALIST_KEY_FILE` |

### `token-signing-secret`

The secret to use when signing authorization tokens.

| Name | Value |
| --- | --- |
| type | string |
| required | yes |
| default | `n/a` |
| cli | `--token-signing-secret` |
| env | `TERRALIST_TOKEN_SIGNING_SECRET` |

### `authorized-users`

Comma separated list of users authorized to access the settings page. If empty, all users are allowed.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--authorized-users` |
| env | `TERRALIST_AUTHORIZED_USERS` |

### `oauth-provider`

The OAuth 2.0 provider.

| Name | Value |
| --- | --- |
| type | select |
| choices | `github`, `bitbucket`, `gitlab`, `oidc` |
| required | yes |
| default | `n/a` |
| cli | `--oauth-provider` |
| env | `TERRALIST_OAUTH_PROVIDER` |

### `gh-client-id`

The GitHub OAuth Application client ID.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--gh-client-id` |
| env | `TERRALIST_GH_CLIENT_ID` |

### `gh-client-secret`

The GitHub OAuth Application client secret.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--gh-client-secret` |
| env | `TERRALIST_GH_CLIENT_SECRET` |

### `gh-organization`

The GitHub organization to use for user validation.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--gh-organization` |
| env | `TERRALIST_GH_ORGANIZATION` |

### `gh-teams`

The GitHub team slugs in CSV format to use for user validation. This requires `gh-organization` to be set.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--gh-teams` |
| env | `TERRALIST_GH_TEAMS` |

### `gh-domain`

The GitHub base domain if you are using GitHub Enterprise.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `github.com` |
| cli | `--gh-domain` |
| env | `TERRALIST_GH_DOMAIN` |

### `bb-client-id`

The BitBucket OAuth Application client ID.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--bb-client-id` |
| env | `TERRALIST_BB_CLIENT_ID` |

### `bb-client-secret`

The BitBucket OAuth Application client secret.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--bb-client-secret` |
| env | `TERRALIST_BB_CLIENT_SECRET` |

### `bb-workspace`

The BitBucket workspace to use for user validation.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--bb-workspace` |
| env | `TERRALIST_BB_WORKSPACE` |

### `gl-client-id`

The GitLab OAuth Application client ID.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--gl-client-id` |
| env | `TERRALIST_GL_CLIENT_ID` |

### `gl-client-secret`

The Gitlab OAuth Application client secret.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--gl-client-secret` |
| env | `TERRALIST_GL_CLIENT_SECRET` |

### `gl-host`

The (self hosted) GitLab host to use. E.g. gitlab.mycompany.com:8443

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `gitlab.com` |
| cli | `--gl-host` |
| env | `TERRALIST_GL_HOST` |

### `gl-groups`

The GitLab groups names the user must be member of. It must be comma separated with no spaces.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--gl-groups` |
| env | `TERRALIST_GL_GROUPS` |


### `oi-client-id`

The OpenID Connect client ID.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--oi-client-id` |
| env | `TERRALIST_OI_CLIENT_ID` |

### `oi-client-secret`

The OpenID Connect client secret.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--oi-client-secret` |
| env | `TERRALIST_OI_CLIENT_SECRET` |

### `oi-authorize-url`

The url to [OpenID Connect authorization endpoint](https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint). E.g. `https://login.mycompany.com/auth/realms/developer/protocol/openid-connect/auth`

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--oi-authorize-url` |
| env | `TERRALIST_OI_AUTHORIZE_URL` |

### `oi-token-url`

The url to [OpenID Connect token endpoint](https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint). E.g. `https://login.mycompany.com/auth/realms/developer/protocol/openid-connect/token`

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--oi-token-url` |
| env | `TERRALIST_OI_TOKEN_URL` |

### `oi-userinfo-url`

The url to [OpenID Connect userinfo endpoint](https://openid.net/specs/openid-connect-core-1_0.html#UserInfo). E.g. `https://login.mycompany.com/auth/realms/developer/protocol/openid-connect/userinfo`

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--oi-userinfo-url` |
| env | `TERRALIST_OI_USERINFO_URL` |

### `oi-scope`

The OpenID Connect scope requested during authorization to ensure to get claims `sub` and `email`.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `openid email` |
| cli | `--oi-scope` |
| env | `TERRALIST_OI_SCOPE` |

### `database-backend`

The database backend.

| Name | Value |
| --- | --- |
| type | select |
| choices | `sqlite`, `postgresql`, `mysql` |
| required | no |
| default | `sqlite` |
| cli | `--database-backend` |
| env | `TERRALIST_DATABASE_BACKEND` |

### `postgres-url`

The URL that can be used to connect to PostgreSQL database.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--postgres-url` |
| env | `TERRALIST_POSTGRES_URL` |

### `postgres-host`

The host where the PostgreSQL database can be found.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--postgres-host` |
| env | `TERRALIST_POSTGRES_HOST` |

### `postgres-port`

The port on which the PostgreSQL database listens.

| Name | Value |
| --- | --- |
| type | int |
| required | no |
| default | `n/a` |
| cli | `--postgres-port` |
| env | `TERRALIST_POSTGRES_PORT` |

### `postgres-username`

The username that can be used to authenticate to PostgreSQL database.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--postgres-username` |
| env | `TERRALIST_POSTGRES_USERNAME` |

### `postgres-password`

The password that can be used to authenticate to PostgreSQL database.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--postgres-password` |
| env | `TERRALIST_POSTGRES_PASSWORD` |

### `postgres-database`

The schema name on which application data should be stored.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--postgres-database` |
| env | `TERRALIST_POSTGRES_DATABASE` |

### `mysql-url`

The URL that can be used to connect to MySQL database.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--mysql-url` |
| env | `TERRALIST_MYSQL_URL` |

### `mysql-host`

The host where the MySQL database can be found.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--mysql-host` |
| env | `TERRALIST_MYSQL_HOST` |

### `mysql-port`

The port on which the MySQL database listens.

| Name | Value |
| --- | --- |
| type | int |
| required | no |
| default | `n/a` |
| cli | `--mysql-port` |
| env | `TERRALIST_MYSQL_PORT` |

### `mysql-username`

The username that can be used to authenticate to MySQL database.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--mysql-username` |
| env | `TERRALIST_MYSQL_USERNAME` |

### `mysql-password`

The password that can be used to authenticate to MySQL database.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--mysql-password` |
| env | `TERRALIST_MYSQL_PASSWORD` |

### `mysql-database`

The schema name on which application data should be stored.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--mysql-database` |
| env | `TERRALIST_MYSQL_DATABASE` |

### `sqlite-path`

The path to the SQLite database.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--sqlite-path` |
| env | `TERRALIST_SQLITE_PATH` |

### `session-store`

The session store backend.

| Name | Value |
| --- | --- |
| type | select |
| choices | `cookie` |
| required | no |
| default | `cookie` |
| cli | `--session-store` |
| env | `TERRALIST_SESSION_STORE` |

### `cookie-secret`

The secret to use for cookie encryption.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--cookie-secret` |
| env | `TERRALIST_COOKIE_SECRET` |

### `modules-storage-resolver`

The modules storage resolver.

| Name | Value |
| --- | --- |
| type | select |
| choices | `proxy`, `local`, `s3`, `azure` |
| required | no |
| default | `proxy` |
| cli | `--modules-storage-resolver` |
| env | `TERRALIST_MODULES_STORAGE_RESOLVER` |

### `providers-storage-resolver`

The providers storage resolver.

| Name | Value |
| --- | --- |
| type | select |
| choices | `proxy`, `local`, `s3`, `azure` |
| required | no |
| default | `proxy` |
| cli | `--providers-storage-resolver` |
| env | `TERRALIST_PROVIDERS_STORAGE_RESOLVER` |

### `modules-anonymous-read`

Allows anonymous read and download of modules.

| Name | Value |
| --- | --- |
| type | bool |
| required | no |
| default | `false` |
| cli | `--modules-anonymous-read` |
| env | `TERRALIST_MODULES_ANONYMOUS_READ` |

### `providers-anonymous-read`

Allows anonymous read and download of providers.

| Name | Value |
| --- | --- |
| type | bool |
| required | no |
| default | `false` |
| cli | `--providers-anonymous-read` |
| env | `TERRALIST_PROVIDERS_ANONYMOUS_READ` |

### `s3-endpoint`

The endpoint where the S3 SDK should connect. By default, Terralist will connect to the AWS S3 endpoint.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--s3-endpoint` |
| env | `TERRALIST_S3_ENDPOINT` |

### `s3-bucket-name`

The S3 bucket name.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--s3-bucket-name` |
| env | `TERRALIST_S3_BUCKET_NAME` |

### `s3-bucket-region`

The S3 bucket region.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--s3-bucket-region` |
| env | `TERRALIST_S3_BUCKET_REGION` |

### `s3-bucket-prefix`

A prefix to be added to the S3 bucket keys.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--s3-bucket-prefix` |
| env | `TERRALIST_S3_BUCKET_PREFIX` |

### `s3-presign-expire`

The number of minutes after which the presigned URLs should expire.

| Name | Value |
| --- | --- |
| type | int |
| required | no |
| default | `15` |
| cli | `--s3-presign-expire` |
| env | `TERRALIST_S3_PRESIGN_EXPIRE` |

### `s3-access-key-id`

The AWS access key ID to access the S3 bucket.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--s3-access-key-id` |
| env | `TERRALIST_S3_ACCESS_KEY_ID` |

### `s3-secret-access-key`

The AWS secret access key to access the S3 bucket.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--s3-secret-access-key` |
| env | `TERRALIST_S3_SECRET_ACCESS_KEY` |

### `s3-use-path-style`

Set this to `true` to force the request to use path-style addressing (i.e. `http://s3.amazonaws.com/BUCKET/KEY`).
By default, the S3 client will use virtual hosted bucket addressing when possible (i.e. `http://BUCKET.s3.amazonaws.com/KEY`).

| Name | Value |
| --- | --- |
| type | bool |
| required | no |
| default | `false` |
| cli | `--s3-use-path-style` |
| env | `TERRALIST_S3_USE_PATH_STYLE` |

## `s3-server-side-encryption`

The server-side encryption algorithm that was used when you store this object in Amazon S3.

| Name | Value |
| --- | --- |
| type | select |
| choices | `none`, `AES256`, `aws:kms`, `aws:kms:dsse` |
| required | no |
| default | `AES256` |
| cli | `--s3-server-side-encryption` |
| env | `TERRALIST_S3_SERVER_SIDE_ENCRYPTION` |

### `s3-disable-acl`

Disable S3 ACL and rely on bucket policy for access control.

| Name | Value |
| --- | --- |
| type | bool |
| required | no |
| default | `false` |
| cli | `--s3-disable-acl` |
| env | `TERRALIST_S3_DISABLE_ACL` |

### `local-store`

The path to a directory in which Terralist can store files.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `~/.terralist.d` |
| cli | `--local-store` |
| env | `TERRALIST_LOCAL_STORE` |

### `azure-account-name`

The Azure account name.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--azure-account-name` |
| env | `TERRALIST_AZURE_ACCOUNT_NAME` |

### `azure-account-key`

The Azure account key.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--azure-account-key` |
| env | `TERRALIST_AZURE_ACCOUNT_KEY` |

### `azure-container-name`

The Azure container name.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--azure-container-name` |
| env | `TERRALIST_AZURE_CONTAINER_NAME` |

### `azure-sas-expire`

The number of minutes after which the Azure Shared Access Signature(SAS) should expire.

| Name | Value |
| --- | --- |
| type | int |
| required | no |
| default | `15` |
| cli | `--azure-sas-expire` |
| env | `TERRALIST_AZURE_SAS_EXPIRE` |


### `gcs-bucket-name`

The GCS bucket name.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--gcs-bucket-name` |
| env | `TERRALIST_GCS_BUCKET_NAME` |

### `gcs-bucket-prefix`

A prefix to be added to the GCS bucket objects.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--gcs-bucket-prefix` |
| env | `TERRALIST_GCS_BUCKET_PREFIX` |

### `gcs-sign-expire`

The number of minutes after which the signed URLs should expire.

| Name | Value |
| --- | --- |
| type | int |
| required | no |
| default | `15` |
| cli | `--gcs-sign-expire` |
| env | `TERRALIST_GCS_SIGN_EXPIRE` |

### `gcs-service-account-cred-file-path`

The GCP Service Account key path access key ID to access the GCS bucket. Leave empty for default credentials

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--gcs-service-account-cred-file-path` |
| env | `TERRALIST_GCS_SERVICE_ACCOUNT_CRED_FILE_PATH` |

### `custom-company-name`

A small NIT branding of Terralist. The name of the company set by this variable will appear on the login page.

| Name | Value |
| --- | --- |
| type | string |
| required | no |
| default | `n/a` |
| cli | `--custom-company-name` |
| env | `TERRALIST_CUSTOM_COMPANY_NAME` |

## Example YAML configuration file

```yaml
# Try to read PORT from the environment variable, and if it's missing,
# fallback to 5758
port: "${PORT:5758}"

log-level: "debug"

oauth-provider: "github"
gh-client-id: "${GITHUB_OAUTH_CLIENT_ID}"
gh-client-secret: "${GITHUB_OAUTH_CLIENT_SECRET}"
# gh-organization is optional, you can set it to restrict access to the registry
# only to members of your GitHub organization
gh-organization: "my-org"
# gh-teams is optional, only users that are part of one of the teams will be able to access the registry
# gh-organization is required for gh-teams to work
# you must use the slug version of the team
gh-teams: "team-a,team-b"
token-signing-secret: "supersecretstring"

database-backend: "sqlite"
sqlite-path: "terralist.db"

# database-backend: "postgresql"
# postgres-url: "${DATABASE_URL:postgres://admin:admin@localhost:5678/public}"

# database-backend: "mysql"
# mysql-url: "admin:admin@tcp(localhost:3306)/terralist"

modules-storage-resolver: "s3" # or "azure"
providers-storage-resolver: "proxy"

s3-bucket-name: "my-bucket"
s3-bucket-region: "us-east-1"
s3-access-key-id: "${AWS_ACCESS_KEY_ID}"
s3-secret-access-key: "${AWS_SECRET_ACCESS_KEY}"

# azure-account-name: "Globally unique name of your storage account"
# azure-container-name: "Name of the container in the storage account"
# azure-account-key: "Access key of the storage account" # If not using DefaultAzureCredentials
# azure-sas-expire: 45 # The number of minutes after which the SAS should expire.

# local-store: "~/.terralist.d"

session-store: "cookie"

cookie-secret: "anothersupersecretstring"
```
