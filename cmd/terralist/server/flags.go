package server

import "terralist/pkg/cli"

const (
	ConfigFlag = "config"

	PortFlag = "port"

	LogLevelFlag = "log-level"

	DatabaseBackendFlag = "database-backend"

	SQLitePathFlag = "sqlite-path"

	MySQLURLFlag      = "mysql-url"
	MySQLUsernameFlag = "mysql-username"
	MySQLPasswordFlag = "mysql-password"
	MySQLHostFlag     = "mysql-host"
	MySQLPortFlag     = "mysql-port"
	MySQLDatabaseFlag = "mysql-database"

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

	TokenSigningSecretFlag = "token-signing-secret"
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

	DatabaseBackendFlag: &cli.StringFlag{
		Description:  "The database backend.",
		Choices:      []string{"sqlite", "mysql", "postgresql"},
		DefaultValue: "sqlite",
	},

	SQLitePathFlag: &cli.StringFlag{
		Description: "The path to the SQLite database.",
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
		Choices:     []string{"github"},
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

	TokenSigningSecretFlag: &cli.StringFlag{
		Description: "The secret to use when signing authorization tokens.",
		Required:    true,
	},
}
