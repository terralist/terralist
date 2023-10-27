package server

import "terralist/pkg/cli"

const (
	ConfigFlag = "config"

	PortFlag = "port"

	LogLevelFlag = "log-level"

	URLFlag = "url"

	CertFileFlag = "cert-file"
	KeyFileFlag  = "key-file"

	DatabaseBackendFlag = "database-backend"

	SQLitePathFlag = "sqlite-path"

	PostgreSQLURLFlag      = "postgres-url"
	PostgreSQLUsernameFlag = "postgres-username"
	PostgreSQLPasswordFlag = "postgres-password"
	PostgreSQLHostFlag     = "postgres-host"
	PostgreSQLPortFlag     = "postgres-port"
	PostgreSQLDatabaseFlag = "postgres-database"

	MySQLURLFlag      = "mysql-url"
	MySQLUsernameFlag = "mysql-username"
	MySQLPasswordFlag = "mysql-password"
	MySQLHostFlag     = "mysql-host"
	MySQLPortFlag     = "mysql-port"
	MySQLDatabaseFlag = "mysql-database"

	OAuthProviderFlag = "oauth-provider"

	// GitHub OAuth Flags
	GitHubClientIDFlag     = "gh-client-id"
	GitHubClientSecretFlag = "gh-client-secret"
	GitHubOrganizationFlag = "gh-organization"

	// BitBucket OAuth Flags
	BitBucketClientIDFlag     = "bb-client-id"
	BitBucketClientSecretFlag = "bb-client-secret"
	BitBucketWorkspaceFlag    = "bb-workspace"

	// GitLab OAuth Flags
	GitLabClientIDFlag     = "gl-client-id"
	GitLabClientSecretFlag = "gl-client-secret"
	GitLabHostFlag         = "gl-host"

	OidcClientIDFlag     = "oi-client-id"
	OidcClientSecretFlag = "oi-client-secret"
	OidcAuthorizeUrlFlag = "oi-authorize-url"
	OidcTokenUrlFlag     = "oi-token-url"
	OidcUserInfoUrlFlag  = "oi-userinfo-url"
	OidcScopeFlag        = "oi-scope"

	TokenSigningSecretFlag = "token-signing-secret"

	ModulesStorageResolverFlag   = "modules-storage-resolver"
	ProvidersStorageResolverFlag = "providers-storage-resolver"

	ModulesAnonymousReadFlag   = "modules-anonymous-read"
	ProvidersAnonymousReadFlag = "providers-anonymous-read"

	S3BucketNameFlag      = "s3-bucket-name"
	S3BucketRegionFlag    = "s3-bucket-region"
	S3BucketPrefixFlag    = "s3-bucket-prefix"
	S3AccessKeyIDFlag     = "s3-access-key-id"
	S3SecretAccessKeyFlag = "s3-secret-access-key"
	S3PresignExpireFlag   = "s3-presign-expire"

	LocalStoreFlag = "local-store"

	SessionStoreFlag = "session-store"

	CookieSecretFlag = "cookie-secret"

	CustomCompanyNameFlag = "custom-company-name"
)

var flags = map[string]cli.Flag{
	ConfigFlag: &cli.StringFlag{
		Description: "Path to YAML config file where flag values are set.",
	},

	PortFlag: &cli.IntFlag{
		Description:  "The port to bind to.",
		DefaultValue: 5758,
	},

	LogLevelFlag: &cli.StringFlag{
		Description:  "The log level.",
		Choices:      []string{"trace", "debug", "info", "warn", "error"},
		DefaultValue: "info",
	},

	URLFlag: &cli.StringFlag{
		Description:  "The URL that Terralist is accessible from.",
		DefaultValue: "http://localhost:5758",
	},

	CertFileFlag: &cli.StringFlag{
		Description: "The path to the certificate file (pem format).",
	},
	KeyFileFlag: &cli.StringFlag{
		Description: "The path to the certificate key file (pem format).",
	},

	DatabaseBackendFlag: &cli.StringFlag{
		Description:  "The database backend.",
		Choices:      []string{"sqlite", "postgresql", "mysql"},
		DefaultValue: "sqlite",
	},

	SQLitePathFlag: &cli.StringFlag{
		Description: "The path to the SQLite database.",
	},

	PostgreSQLURLFlag: &cli.StringFlag{
		Description: "The URL that can be used to connect to PostgreSQL database.",
	},
	PostgreSQLUsernameFlag: &cli.StringFlag{
		Description: "The username that can be used to authenticate to PostgreSQL database.",
	},
	PostgreSQLPasswordFlag: &cli.StringFlag{
		Description: "The password that can be used to authenticate to PostgreSQL database.",
	},
	PostgreSQLHostFlag: &cli.StringFlag{
		Description: "The host where the PostgreSQL database can be found.",
	},
	PostgreSQLPortFlag: &cli.IntFlag{
		Description: "The port on which the PostgreSQL database listens.",
	},
	PostgreSQLDatabaseFlag: &cli.StringFlag{
		Description: "The schema name on which application data should be stored.",
	},

	MySQLURLFlag: &cli.StringFlag{
		Description: "The URL that can be used to connect to MySQL database.",
	},
	MySQLUsernameFlag: &cli.StringFlag{
		Description: "The username that can be used to authenticate to MySQL database.",
	},
	MySQLPasswordFlag: &cli.StringFlag{
		Description: "The password that can be used to authenticate to MySQL database.",
	},
	MySQLHostFlag: &cli.StringFlag{
		Description: "The host where the MySQL database can be found.",
	},
	MySQLPortFlag: &cli.IntFlag{
		Description: "The port on which the MySQL database listens.",
	},
	MySQLDatabaseFlag: &cli.StringFlag{
		Description: "The schema name on which application data should be stored.",
	},

	OAuthProviderFlag: &cli.StringFlag{
		Description: "The OAuth 2.0 provider.",
		Choices:     []string{"github", "bitbucket", "gitlab", "oidc"},
		Required:    true,
	},
	GitHubClientIDFlag: &cli.StringFlag{
		Description: "The GitHub OAuth Application client ID.",
	},
	GitHubClientSecretFlag: &cli.StringFlag{
		Description: "The GitHub OAuth Application client secret.",
	},
	GitHubOrganizationFlag: &cli.StringFlag{
		Description: "The GitHub organization to use for user validation.",
	},
	BitBucketClientIDFlag: &cli.StringFlag{
		Description: "The BitBucket OAuth Application client ID.",
	},
	BitBucketClientSecretFlag: &cli.StringFlag{
		Description: "The BitBucket OAuth Application client secret.",
	},
	BitBucketWorkspaceFlag: &cli.StringFlag{
		Description: "The BitBucket workspace to use for user validation.",
	},
	GitLabClientIDFlag: &cli.StringFlag{
		Description: "The GitLab OAuth Application client ID.",
	},
	GitLabClientSecretFlag: &cli.StringFlag{
		Description: "The GitLab OAuth Application client secret.",
	},
	GitLabHostFlag: &cli.StringFlag{
		Description:  "The GitLab host to use.",
		DefaultValue: "gitlab.com",
	},
	OidcClientIDFlag: &cli.StringFlag{
		Description: "The OIDC Application client ID.",
	},
	OidcClientSecretFlag: &cli.StringFlag{
		Description: "The OIDC Application client secret.",
	},
	OidcAuthorizeUrlFlag: &cli.StringFlag{
		Description: "The OIDC Provider authorization endpoint url.",
	},
	OidcTokenUrlFlag: &cli.StringFlag{
		Description: "The OIDC Provider token endpoint url.",
	},
	OidcUserInfoUrlFlag: &cli.StringFlag{
		Description: "The OIDC Provider userinfo endpoint url.",
	},
	OidcScopeFlag: &cli.StringFlag{
		Description:  "The scopes requested during OIDC authorization.",
		DefaultValue: "openid email",
	},

	TokenSigningSecretFlag: &cli.StringFlag{
		Description: "The secret to use when signing authorization tokens.",
		Required:    true,
	},

	LocalStoreFlag: &cli.StringFlag{
		Description:  "The path to a directory in which Terralist can store files.",
		DefaultValue: "~/.terralist.d",
	},

	ModulesStorageResolverFlag: &cli.StringFlag{
		Description:  "The modules storage resolver.",
		Choices:      []string{"proxy", "local", "s3"},
		DefaultValue: "proxy",
	},

	ProvidersStorageResolverFlag: &cli.StringFlag{
		Description:  "The providers storage resolver.",
		Choices:      []string{"proxy", "local", "s3"},
		DefaultValue: "proxy",
	},

	ModulesAnonymousReadFlag: &cli.BoolFlag{
		Description:  "Allow anonymous read to modules.",
		DefaultValue: false,
	},

	ProvidersAnonymousReadFlag: &cli.BoolFlag{
		Description:  "Allow anonymous read to providers.",
		DefaultValue: false,
	},

	S3BucketNameFlag: &cli.StringFlag{
		Description: "The S3 bucket name.",
	},
	S3BucketRegionFlag: &cli.StringFlag{
		Description: "The S3 bucket region.",
	},
	S3BucketPrefixFlag: &cli.StringFlag{
		Description: "A prefix to be added to the S3 bucket keys.",
	},
	S3AccessKeyIDFlag: &cli.StringFlag{
		Description: "The AWS access key ID to access the S3 bucket.",
	},
	S3SecretAccessKeyFlag: &cli.StringFlag{
		Description: "The AWS secret access key to access the S3 bucket.",
	},
	S3PresignExpireFlag: &cli.IntFlag{
		Description:  "The number of minutes after which the presigned URLs should expire.",
		DefaultValue: 15,
	},

	SessionStoreFlag: &cli.StringFlag{
		Description:  "The session store backend.",
		Choices:      []string{"cookie"},
		DefaultValue: "cookie",
	},

	CookieSecretFlag: &cli.StringFlag{
		Description: "The secret to use for cookie encryption.",
	},

	CustomCompanyNameFlag: &cli.StringFlag{
		Description: "The name of the company hosting the Terralist instance.",
	},
}
