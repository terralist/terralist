package server

import "terralist/pkg/cli"

const (
	ConfigFlag = "config"

	PortFlag = "port"

	LogLevelFlag = "log-level"

	URLFlag = "url"

	HomeFlag = "home"

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

	GitHubClientIDFlag     = "gh-client-id"
	GitHubClientSecretFlag = "gh-client-secret"
	GitHubOrganizationFlag = "gh-organization"
	GitHubTeamsFlag        = "gh-teams"
	GitHubDomainFlag       = "gh-domain"

	BitBucketClientIDFlag     = "bb-client-id"
	BitBucketClientSecretFlag = "bb-client-secret"
	BitBucketWorkspaceFlag    = "bb-workspace"

	GitLabClientIDFlag     = "gl-client-id"
	GitLabClientSecretFlag = "gl-client-secret"
	GitLabHostFlag         = "gl-host"
	GitLabGroupsFlag       = "gl-groups"

	OidcClientIDFlag     = "oi-client-id"
	OidcClientSecretFlag = "oi-client-secret"
	OidcHostFlag         = "oi-host"
	OidcAuthorizeUrlFlag = "oi-authorize-url"
	OidcTokenUrlFlag     = "oi-token-url"
	OidcUserInfoUrlFlag  = "oi-userinfo-url"

	SamlDisplayNameFlag                  = "saml-display-name"
	SamlIdPMetadataURLFlag               = "saml-idp-metadata-url"
	SamlIdPMetadataFileFlag              = "saml-idp-metadata-file"
	SamlIdPEntityIDFlag                  = "saml-idp-entity-id"
	SamlIdPSSOURLFlag                    = "saml-idp-sso-url"
	SamlIdPSSOCertificateFlag            = "saml-idp-sso-certificate"
	SamlNameAttributeFlag                = "saml-name-attribute"
	SamlEmailAttributeFlag               = "saml-email-attribute"
	SamlGroupsAttributeFlag              = "saml-groups-attribute"
	SamlCertFileFlag                     = "saml-cert-file"
	SamlKeyFileFlag                      = "saml-key-file"
	SamlPrivateKeySecretFlag             = "saml-private-key-secret"
	SamlHTTPClientTimeoutFlag            = "saml-http-client-timeout"
	SamlAssertionClockSkewFlag           = "saml-assertion-clock-skew"
	SamlRequestIDExpirationFlag          = "saml-request-id-expiration"
	SamlRequestIDCleanupIntervalFlag     = "saml-request-id-cleanup-interval"
	SamlMetadataRefreshIntervalFlag      = "saml-metadata-refresh-interval"
	SamlMetadataRefreshCheckIntervalFlag = "saml-metadata-refresh-check-interval"
	SamlMaxAssertionAgeFlag              = "saml-max-assertion-age"
	SamlAllowIdPInitiatedFlag            = "saml-allow-idp-initiated"
	SamlDisableRequestIDValidationFlag   = "saml-disable-request-id-validation"

	TokenSigningSecretFlag = "token-signing-secret"

	ModulesStorageResolverFlag   = "modules-storage-resolver"
	ProvidersStorageResolverFlag = "providers-storage-resolver"

	ModulesAnonymousReadFlag   = "modules-anonymous-read"
	ProvidersAnonymousReadFlag = "providers-anonymous-read"

	S3EndpointFlag             = "s3-endpoint"
	S3BucketNameFlag           = "s3-bucket-name"
	S3BucketRegionFlag         = "s3-bucket-region"
	S3BucketPrefixFlag         = "s3-bucket-prefix"
	S3AccessKeyIDFlag          = "s3-access-key-id"
	S3SecretAccessKeyFlag      = "s3-secret-access-key"
	S3PresignExpireFlag        = "s3-presign-expire"
	S3ServerSideEncryptionFlag = "s3-server-side-encryption"
	S3UsePathStyleFlag         = "s3-use-path-style"
	S3UseACLsFlag              = "s3-use-acls"

	AzureAccountNameFlag   = "azure-account-name"
	AzureAccountKeyFlag    = "azure-account-key"
	AzureContainerNameFlag = "azure-container-name"
	AzureSASExpireFlag     = "azure-sas-expire"

	GcsBucketNameFlag                 = "gcs-bucket-name"
	GcsBucketPrefixFlag               = "gcs-bucket-prefix"
	GcsSignExpireFlag                 = "gcs-sign-expire"
	GcsServiceAccountCredFilePathFlag = "gcs-service-account-cred-file-path"

	LocalStoreFlag              = "local-store"
	LocalRegistryFlag           = "local-registry"
	LocalTokenSigningSecretFlag = "local-token-signing-secret"
	LocalPresignExpireFlag      = "local-presign-expire"

	SessionStoreFlag = "session-store"

	CookieSecretFlag = "cookie-secret"

	CustomCompanyNameFlag = "custom-company-name"

	RbacPolicyPathFlag  = "rbac-policy-path"
	RbacDefaultRoleFlag = "rbac-default-role"

	MasterApiKeyFlag = "master-api-key"

	AuthTokenExpirationFlag = "auth-token-expiration"
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

	HomeFlag: &cli.PathFlag{
		Description:  "The path to the directory where Terralist can store files.",
		DefaultValue: "~/.terralist.d",
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
		Choices:     []string{"github", "bitbucket", "gitlab", "oidc", "saml"},
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
	GitHubTeamsFlag: &cli.StringFlag{
		Description: "The GitHub team slugs in CSV format to use for user validation.",
	},
	GitHubDomainFlag: &cli.StringFlag{
		Description:  "The GitHub base domain if you are using GitHub Enterprise. (default: 'github.com')",
		DefaultValue: "github.com",
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
	GitLabGroupsFlag: &cli.StringFlag{
		Description:  "The GitLab groups the user must be member. Comma separated.",
		DefaultValue: "",
	},
	OidcClientIDFlag: &cli.StringFlag{
		Description: "The OIDC Application client ID.",
	},
	OidcClientSecretFlag: &cli.StringFlag{
		Description: "The OIDC Application client secret.",
	},
	OidcHostFlag: &cli.StringFlag{
		Description: "The OIDC issuer URL used for discovery. Preferred over the manual OIDC endpoint flags.",
	},
	OidcAuthorizeUrlFlag: &cli.StringFlag{
		Description: "Deprecated manual OIDC authorization endpoint URL. Use oi-host instead.",
	},
	OidcTokenUrlFlag: &cli.StringFlag{
		Description: "Deprecated manual OIDC token endpoint URL. Use oi-host instead.",
	},
	OidcUserInfoUrlFlag: &cli.StringFlag{
		Description: "Deprecated manual OIDC userinfo endpoint URL. Use oi-host instead.",
	},
	SamlDisplayNameFlag: &cli.StringFlag{
		Description:  "The display name for SAML authentication in the UI.",
		DefaultValue: "SSO",
	},
	SamlIdPMetadataURLFlag: &cli.StringFlag{
		Description: "The URL where the IdP metadata can be fetched from. Either this, saml-idp-metadata-file, or both saml-idp-entity-id and saml-idp-sso-url must be provided.",
	},
	SamlIdPMetadataFileFlag: &cli.StringFlag{
		Description: "The local file path to the IdP metadata XML file. Either this, saml-idp-metadata-url, or both saml-idp-entity-id and saml-idp-sso-url must be provided.",
	},
	SamlIdPEntityIDFlag: &cli.StringFlag{
		Description: "The Identity Provider entity ID. Can be used instead of saml-idp-metadata-url/saml-idp-metadata-file if saml-idp-sso-url is also provided.",
	},
	SamlIdPSSOURLFlag: &cli.StringFlag{
		Description: "The Identity Provider Single Sign-On URL. Can be used instead of saml-idp-metadata-url/saml-idp-metadata-file if saml-idp-entity-id is also provided.",
	},
	SamlIdPSSOCertificateFlag: &cli.StringFlag{
		Description: "The Identity Provider SSO certificate (PEM format). Required if certificate cannot be extracted from IdP metadata.",
	},
	SamlNameAttributeFlag: &cli.StringFlag{
		Description:  "The SAML attribute name that contains the user's name.",
		DefaultValue: "displayName",
	},
	SamlEmailAttributeFlag: &cli.StringFlag{
		Description:  "The SAML attribute name that contains the user's email.",
		DefaultValue: "email",
	},
	SamlGroupsAttributeFlag: &cli.StringFlag{
		Description: "The SAML attribute name that contains the user's groups. Used for RBAC group mapping.",
	},
	SamlCertFileFlag: &cli.StringFlag{
		Description: "The path to the certificate file (PEM format) used for signing SAML requests. Optional but recommended for production.",
	},
	SamlKeyFileFlag: &cli.StringFlag{
		Description: "The path to the private key file (PEM format) used for signing SAML requests. Optional but recommended for production.",
	},
	SamlPrivateKeySecretFlag: &cli.StringFlag{
		Description: "The passphrase for the SAML private key if it is encrypted.",
	},
	SamlHTTPClientTimeoutFlag: &cli.StringFlag{
		Description:  "The timeout for HTTP requests to fetch IdP metadata.",
		DefaultValue: "30s",
	},
	SamlAssertionClockSkewFlag: &cli.StringFlag{
		Description:  "The allowed time difference between SP and IdP clocks (SAML 2.0 spec recommends clock skew tolerance).",
		DefaultValue: "5m",
	},
	SamlRequestIDExpirationFlag: &cli.StringFlag{
		Description:  "How long SAML request IDs are kept to prevent replay attacks.",
		DefaultValue: "1h",
	},
	SamlRequestIDCleanupIntervalFlag: &cli.StringFlag{
		Description:  "How often expired SAML request IDs are cleaned up.",
		DefaultValue: "15m",
	},
	SamlMetadataRefreshIntervalFlag: &cli.StringFlag{
		Description:  "The interval for refreshing IdP metadata from URL.",
		DefaultValue: "24h",
	},
	SamlMetadataRefreshCheckIntervalFlag: &cli.StringFlag{
		Description:  "How often to check if IdP metadata needs refresh.",
		DefaultValue: "1h",
	},
	SamlMaxAssertionAgeFlag: &cli.StringFlag{
		Description:  "The maximum age of SAML assertions from IssueInstant.",
		DefaultValue: "1h",
	},
	SamlAllowIdPInitiatedFlag: &cli.BoolFlag{
		Description:  "Whether to allow IdP-initiated SSO (security best practice is to disable).",
		DefaultValue: false,
	},
	SamlDisableRequestIDValidationFlag: &cli.BoolFlag{
		Description:  "Disable SAML request ID validation (useful in Kubernetes environments with multiple pods).",
		DefaultValue: false,
	},

	TokenSigningSecretFlag: &cli.StringFlag{
		Description: "The secret to use when signing authorization tokens.",
		Required:    true,
	},

	LocalStoreFlag: &cli.StringFlag{
		Description:  "The path to a directory in which Terralist can store files.",
		DefaultValue: "~/.terralist.d",
	},
	LocalRegistryFlag: &cli.PathFlag{
		Description:  "Directory where locally stored module/provider artifacts are written.",
		DefaultValue: "~/.terralist.d/registry",
	},
	LocalTokenSigningSecretFlag: &cli.StringFlag{
		Description: "Secret used by the local resolver to sign file download JWTs.",
	},
	LocalPresignExpireFlag: &cli.IntFlag{
		Description:  "The number of minutes after which local presigned URLs should expire.",
		DefaultValue: 15,
	},

	ModulesStorageResolverFlag: &cli.StringFlag{
		Description:  "The modules storage resolver.",
		Choices:      []string{"proxy", "local", "s3", "azure", "gcs"},
		DefaultValue: "proxy",
	},

	ProvidersStorageResolverFlag: &cli.StringFlag{
		Description:  "The providers storage resolver.",
		Choices:      []string{"proxy", "local", "s3", "azure", "gcs"},
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

	S3EndpointFlag: &cli.StringFlag{
		Description: "The endpoint where the S3 SDK should connect.",
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
	S3UsePathStyleFlag: &cli.BoolFlag{
		Description:  "Set this to `true` to force the request to use path-style addressing.",
		DefaultValue: false,
	},
	S3ServerSideEncryptionFlag: &cli.StringFlag{
		Description:  "The server-side encryption algorithm that was used when you store this object in Amazon S3.",
		Choices:      []string{"none", "AES256", "aws:kms", "aws:kms:dsse"},
		DefaultValue: "AES256",
	},
	S3UseACLsFlag: &cli.BoolFlag{
		Description:  "Use S3 ACLs for access control.",
		DefaultValue: false,
	},

	AzureAccountNameFlag: &cli.StringFlag{
		Description: "The Azure account name.",
	},
	AzureAccountKeyFlag: &cli.StringFlag{
		Description: "The Azure account key.",
	},
	AzureContainerNameFlag: &cli.StringFlag{
		Description: "The Azure container name.",
	},
	AzureSASExpireFlag: &cli.IntFlag{
		Description:  "The number of minutes after which the Azure Shared Access Signature(SAS) should expire.",
		DefaultValue: 15,
	},
	GcsBucketNameFlag: &cli.StringFlag{
		Description: "The GCS bucket name.",
	},
	GcsBucketPrefixFlag: &cli.StringFlag{
		Description: "The GCS bucket folder.",
	},
	GcsSignExpireFlag: &cli.IntFlag{
		Description:  "The number of minutes after which the GCS Sign should expire.",
		DefaultValue: 15,
	},
	GcsServiceAccountCredFilePathFlag: &cli.StringFlag{
		Description: "The GCP Service Account key file path.",
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

	RbacPolicyPathFlag: &cli.StringFlag{
		Description: "Path to the RBAC server-side policy.",
	},
	RbacDefaultRoleFlag: &cli.StringFlag{
		Description:  "The name of the RBAC role that should be assigned by default to all users.",
		DefaultValue: "readonly",
	},

	MasterApiKeyFlag: &cli.StringFlag{
		Description: "A pre-shared API key with full access for bootstrapping. Use via X-API-Key header.",
	},

	AuthTokenExpirationFlag: &cli.StringFlag{
		Description:  "The duration for which auth tokens remain valid.",
		Choices:      []string{"1d", "1w", "1m", "1y", "never"},
		DefaultValue: "1d",
	},
}
