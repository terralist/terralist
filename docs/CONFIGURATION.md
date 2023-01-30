# Configuration

Terralist supports multiple types of configuration:
+ CLI arguments
  Set the option by passing it with the `--` prefix on the CLI command (e.g. `--port`).
+ Environment Variable
  Any option can be set using an environment variable. To do such, replace any dash (`-`) with an underscore (`_`), uppercase everything and add the `TERRALIST_` prefix (e.g. `TERRALIST_PORT`).
+ Configuration File
  Set all options you want to a configuration file, then pass the path to the configuration file using the `config` option (`--config` argument or `TERRALIST_CONFIG` environment variable).
  <br /> Supported file formats: JSON, TOML, YAML, HCL, INI, envfile and Java Properties files.
  <br /> E.g. (YAML):
  ```yaml
  port: 5758

  log-level: debug
  ```

It is also possible to mix those types.

Terralist also supports reading the environment at run-time. For example, if you only know the port value at run-time (e.g. you are running on Heroku), you can set the `TERRALIST_PORT` environment variable to `${PORT}`; this instruction will inform Terralist to read the value, at run-time, from the environment variable called `PORT`. It is also possible to set a default value, in case the given one is not present, by using a colon (`:`), example: `${PORT:5758}`.

## Options

| Name                         | Type   | Required | Default                 | Description                                                           |
| ---------------------------- | ------ | -------- | ----------------------- | --------------------------------------------------------------------- |
| `config`                     | string | no       | `n/a`                   | Path to YAML config file where flag values are set.                   |
| `log-level`                  | string | no       | `info`                  | The log level.                                                        |
| `port`                       | int    | no       | `5758`                  | The port to bind to.                                                  |
| `url`                        | string | no       | `http://localhost:5758` | The URL that Terralist is accessible from.                            |
| `token-signing-secret`       | string | yes      | `n/a`                   | The secret to use when signing authorization tokens.                  |
| `oauth-provider`             | string | yes      | `n/a`                   | The OAuth 2.0 provider.                                               |
| `gh-client-id`               | string | no       | `n/a`                   | The GitHub OAuth Application client ID.                               |
| `gh-client-secret`           | string | no       | `n/a`                   | The GitHub OAuth Application client secret.                           |
| `gh-organization`            | string | no       | `n/a`                   | The GitHub organization to use for user validation.                   |
| `bb-client-id`               | string | no       | `n/a`                   | The BitBucket OAuth Application client ID.                            |
| `bb-client-secret`           | string | no       | `n/a`                   | The BitBucket OAuth Application client secret.                        |
| `bb-workspace`               | string | no       | `n/a`                   | The BitBucket workspace to use for user validation.                   |
| `gl-client-id`               | string | no       | `n/a`                   | The GitLab OAuth Application client ID.                               |
| `gl-client-secret`           | string | no       | `n/a`                   | The Gitlab OAuth Application client secret.                           |
| `gl-host`                    | string | no       | `gitlab.com`            | The (self hosted) GitLab host to use. E.g. gitlab.mycompany.com:8443  |
| `database-backend`           | string | no       | `sqlite`                | The database backend.                                                 |
| `postgres-url`               | string | no       | `n/a`                   | The URL that can be used to connect to PostgreSQL database.           |
| `postgres-host`              | string | no       | `n/a`                   | The host where the PostgreSQL database can be found.                  |
| `postgres-port`              | int    | no       | `n/a`                   | The port on which the PostgreSQL database listens.                    |
| `postgres-username`          | string | no       | `n/a`                   | The username that can be used to authenticate to PostgreSQL database. |
| `postgres-password`          | string | no       | `n/a`                   | The password that can be used to authenticate to PostgreSQL database. |
| `postgres-database`          | string | no       | `n/a`                   | The schema name on which application data should be stored.           |
| `mysql-url`                  | string | no       | `n/a`                   | The URL that can be used to connect to MySQL database.                |
| `mysql-host`                 | string | no       | `n/a`                   | The host where the MySQL database can be found.                       |
| `mysql-port`                 | int    | no       | `n/a`                   | The port on which the MySQL database listens.                         |
| `mysql-username`             | string | no       | `n/a`                   | The username that can be used to authenticate to MySQL database.      |
| `mysql-password`             | string | no       | `n/a`                   | The password that can be used to authenticate to MySQL database.      |
| `mysql-database`             | string | no       | `n/a`                   | The schema name on which application data should be stored.           |
| `sqlite-path`                | string | no       | `n/a`                   | The path to the SQLite database.                                      |
| `session-store`              | string | no       | `cookie`                | The session store backend.                                            |
| `cookie-secret`              | string | no       | `n/a`                   | The secret to use for cookie encryption.                              |
| `modules-storage-resolver`   | string | no       | `proxy`                 | The modules storage resolver.                                         |
| `providers-storage-resolver` | string | no       | `proxy`                 | The providers storage resolver.                                       |
| `s3-bucket-name`             | string | no       | `n/a`                   | The S3 bucket name.                                                   |
| `s3-bucket-region`           | string | no       | `n/a`                   | The S3 bucket region.                                                 |
| `s3-bucket-prefix`           | string | no       | `n/a`                   | A prefix to be added to the S3 bucket keys.                           |
| `s3-presign-expire`          | int    | no       | `15`                    | The number of minutes after which the presigned URLs should expire.   |
| `s3-access-key-id`           | string | no       | `n/a`                   | The AWS access key ID to access the S3 bucket.                        |
| `s3-secret-access-key`       | string | no       | `n/a`                   | The AWS secret access key to access the S3 bucket.                    |
| `local-store`                | string | no       | `~/.terralist.d`        | The path to a directory in which Terralist can store files.           |

## Example config file

```yaml
# Try to read PORT from the environment variable, and if it's missing,
# fallback to 5758
port: "${PORT:5758}"

log-level: "debug"

oauth-provider: "github"
gh-client-id: "<< YOUR_CLIENT_ID >>"
gh-client-secret: "<< YOUR_CLIENT_SECRET >>"
# gh-organization is optional, you can set it to restrict access to the registry
# only to members of your GitHub organization
gh-organization: "<< YOUR_GH_ORGANIZATION >>"

token-signing-secret: "<< ANY_RANDOM_STRING_SECRET >>"

database-backend: "sqlite"
sqlite-path: "terralist.db"

# database-backend: "postgresql"
# postgres-url: "${DATABASE_URL:postgres://admin:admin@localhost:5678/public}"

# database-backend: "mysql"
# mysql-url: "admin:admin@tcp(localhost:3306)/terralist"

modules-storage-resolver: "s3"
providers-storage-resolver: "proxy"

s3-bucket-name: "<< YOUR_S3_BUCKET_NAME >>"
s3-bucket-region: "<< S3_BUCKET_REGION >>"
s3-access-key-id: "<< YOUR_ACCESS_KEY_ID >>"
s3-secret-access-key: "<< YOUR_SECRET_ACCESS_KEY >>"

# local-store: "~/.terralist.d"

session-store: "cookie"

cookie-secret: "<< ANY_RANDOM_STRING_SECRET>>"
```
