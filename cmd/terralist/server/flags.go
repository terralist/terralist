package server

import "terralist/pkg/cli"

const (
	ConfigFlag = "config"

	PortFlag = "port"

	LogLevelFlag = "log-level"

	URLFlag = "url"

	DatabaseBackendFlag = "database-backend"

	SQLitePathFlag = "sqlite-path"

	PostgreSQLURLFlag      = "postgres-url"
	PostgreSQLUsernameFlag = "postgres-username"
	PostgreSQLPasswordFlag = "postgres-password"
	PostgreSQLHostFlag     = "postgres-host"
	PostgreSQLPortFlag     = "postgres-port"
	PostgreSQLDatabaseFlag = "postgres-database"

	OAuthProviderFlag = "oauth-provider"

	// GitHub OAuth Flags
	GitHubClientIDFlag     = "gh-client-id"
	GitHubClientSecretFlag = "gh-client-secret"
	GitHubOrganizationFlag = "gh-organization"

	// BitBucket OAuth Flags
	BitBucketClientIDFlag     = "bb-client-id"
	BitBucketClientSecretFlag = "bb-client-secret"
	BitBucketWorkspaceFlag    = "bb-workspace"

	TokenSigningSecretFlag = "token-signing-secret"

	ModulesStorageResolverFlag   = "modules-storage-resolver"
	ProvidersStorageResolverFlag = "providers-storage-resolver"

	S3BucketNameFlag      = "s3-bucket-name"
	S3BucketRegionFlag    = "s3-bucket-region"
	S3BucketPrefixFlag    = "s3-bucket-prefix"
	S3AccessKeyIDFlag     = "s3-access-key-id"
	S3SecretAccessKeyFlag = "s3-secret-access-key"
	S3PresignExpireFlag   = "s3-presign-expire"

	LocalStoreFlag = "local-store"

	SessionStoreFlag = "session-store"

	CookieSecretFlag = "cookie-secret"
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

	DatabaseBackendFlag: &cli.StringFlag{
		Description:  "The database backend.",
		Choices:      []string{"sqlite", "postgresql"},
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

	OAuthProviderFlag: &cli.StringFlag{
		Description: "The OAuth 2.0 provider.",
		Choices:     []string{"github", "bitbucket"},
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
}
