package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"terralist/pkg/auth/bitbucket"
	"terralist/pkg/auth/gitlab"
	"terralist/pkg/auth/oidc"

	"terralist/internal/server"
	"terralist/pkg/auth"
	authFactory "terralist/pkg/auth/factory"
	"terralist/pkg/auth/github"
	"terralist/pkg/cli"
	"terralist/pkg/database"
	dbFactory "terralist/pkg/database/factory"
	"terralist/pkg/database/mysql"
	"terralist/pkg/database/postgresql"
	"terralist/pkg/database/sqlite"
	"terralist/pkg/session"
	"terralist/pkg/session/cookie"
	sessionFactory "terralist/pkg/session/factory"
	"terralist/pkg/storage"
	"terralist/pkg/storage/azure"
	storageFactory "terralist/pkg/storage/factory"
	"terralist/pkg/storage/gcs"
	"terralist/pkg/storage/local"
	"terralist/pkg/storage/s3"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ssoroka/slice"
)

// Command is an abstraction for the server command
type Command struct {
	ServerCreator Creator
	Viper         *viper.Viper

	RunningMode string

	SilenceOutput bool
}

// Creator creates the server
type Creator interface {
	NewServer(userConfig server.UserConfig, config server.Config) (Starter, error)
}

// DefaultCreator is the concrete implementation of Creator
type DefaultCreator struct{}

// Starter starts the server
type Starter interface {
	Start() error
}

// NewServer returns the real server object
func (d *DefaultCreator) NewServer(userConfig server.UserConfig, config server.Config) (Starter, error) {
	return server.NewServer(userConfig, config)
}

func (s *Command) Init() *cobra.Command {
	c := &cobra.Command{
		Use:           "server",
		Short:         "Starts the Terralist server",
		Long:          "Starts the Terralist RESTful server.",
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRunE: s.withErrPrint(func(cmd *cobra.Command, args []string) error {
			return s.preRun()
		}),
		RunE: s.withErrPrint(func(cmd *cobra.Command, args []string) error {
			return s.run()
		}),
	}

	// Configure viper to accept env vars with prefix instead of flags
	s.Viper.SetEnvPrefix("TERRALIST")
	s.Viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	s.Viper.AutomaticEnv()
	s.Viper.SetTypeByDefaultValue(true)

	c.SetUsageTemplate(cli.UsageTmpl(flags))
	// In case of invalid flags, print the error
	c.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		s.printErr(err)
		return err
	})

	for name, f := range flags {
		usage := f.Format() + "\n"

		if fg, ok := f.(*cli.StringFlag); ok {
			c.Flags().String(name, fg.DefaultValue, usage)
		} else if fg, ok := f.(*cli.IntFlag); ok {
			c.Flags().Int(name, fg.DefaultValue, usage)
		} else if fg, ok := f.(*cli.BoolFlag); ok {
			c.Flags().Bool(name, fg.DefaultValue, usage)
		}

		if f.IsHidden() {
			_ = c.Flags().MarkHidden(name)
		}

		_ = s.Viper.BindPFlag(name, c.Flags().Lookup(name))
	}

	return c
}

func (s *Command) preRun() error {
	// If passed a config file then try and load it.
	configFile := s.Viper.GetString(ConfigFlag)

	if configFile != "" {
		s.Viper.SetConfigFile(configFile)
		if err := s.Viper.ReadInConfig(); err != nil {
			return errors.Wrapf(err, "invalid config: reading %s", configFile)
		}
	}

	return nil
}

func (s *Command) run() error {
	var raw map[string]any

	if err := s.Viper.Unmarshal(&raw); err != nil {
		return err
	}

	configuredFlags := []string{}

	// Set values from viper
	for k, v := range raw {
		if _, ok := flags[k]; ok {
			configuredFlags = append(configuredFlags, k)

			// If it's not set, set the default value
			if !s.Viper.IsSet(k) {
				_ = flags[k].Set(nil)

				continue
			}

			if err := flags[k].Set(v); err != nil {
				return fmt.Errorf("could not unpack flags: %v", err)
			}
		}
	}

	// Set defaults for other flags
	for k := range flags {
		if !slice.Contains(configuredFlags, k) {
			_ = flags[k].Set(nil)
		}
	}

	// Validate flag values
	for k, v := range flags {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("could not validate %v: %v", k, err)
		}
	}

	userConfig := server.UserConfig{
		LogLevel:               flags[LogLevelFlag].(*cli.StringFlag).Value,
		Port:                   flags[PortFlag].(*cli.IntFlag).Value,
		URL:                    flags[URLFlag].(*cli.StringFlag).Value,
		CertFile:               flags[CertFileFlag].(*cli.StringFlag).Value,
		KeyFile:                flags[KeyFileFlag].(*cli.StringFlag).Value,
		TokenSigningSecret:     flags[TokenSigningSecretFlag].(*cli.StringFlag).Value,
		OauthProvider:          flags[OAuthProviderFlag].(*cli.StringFlag).Value,
		CustomCompanyName:      flags[CustomCompanyNameFlag].(*cli.StringFlag).Value,
		ModulesAnonymousRead:   flags[ModulesAnonymousReadFlag].(*cli.BoolFlag).Value,
		ProvidersAnonymousRead: flags[ProvidersAnonymousReadFlag].(*cli.BoolFlag).Value,
	}

	if s.RunningMode == "debug" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else {
		switch userConfig.LogLevel {
		case "trace":
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		case "debug":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "info":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case "warn":
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case "error":
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		}
	}

	// Initialize database
	var db database.Engine
	var err error
	switch flags[DatabaseBackendFlag].(*cli.StringFlag).Value {
	case "sqlite":
		db, err = dbFactory.NewDatabase(database.SQLITE, &sqlite.Config{
			Path: flags[SQLitePathFlag].(*cli.StringFlag).Value,
		})
	case "postgresql":
		db, err = dbFactory.NewDatabase(database.POSTGRESQL, &postgresql.Config{
			URL:      flags[PostgreSQLURLFlag].(*cli.StringFlag).Value,
			Username: flags[PostgreSQLUsernameFlag].(*cli.StringFlag).Value,
			Password: flags[PostgreSQLPasswordFlag].(*cli.StringFlag).Value,
			Hostname: flags[PostgreSQLHostFlag].(*cli.StringFlag).Value,
			Port:     flags[PostgreSQLPortFlag].(*cli.IntFlag).Value,
			Name:     flags[PostgreSQLDatabaseFlag].(*cli.StringFlag).Value,
		})
	case "mysql":
		db, err = dbFactory.NewDatabase(database.MYSQL, &mysql.Config{
			URL:      flags[MySQLURLFlag].(*cli.StringFlag).Value,
			Username: flags[MySQLUsernameFlag].(*cli.StringFlag).Value,
			Password: flags[MySQLPasswordFlag].(*cli.StringFlag).Value,
			Hostname: flags[MySQLHostFlag].(*cli.StringFlag).Value,
			Port:     flags[MySQLPortFlag].(*cli.IntFlag).Value,
			Name:     flags[MySQLDatabaseFlag].(*cli.StringFlag).Value,
		})
	}
	if err != nil {
		return err
	}

	// Initialize Auth provider
	var provider auth.Provider

	switch flags[OAuthProviderFlag].(*cli.StringFlag).Value {
	case "github":
		provider, err = authFactory.NewProvider(auth.GITHUB, &github.Config{
			ClientID:     flags[GitHubClientIDFlag].(*cli.StringFlag).Value,
			ClientSecret: flags[GitHubClientSecretFlag].(*cli.StringFlag).Value,
			Organization: flags[GitHubOrganizationFlag].(*cli.StringFlag).Value,
			Teams:        flags[GitHubTeamsFlag].(*cli.StringFlag).Value,
		})
	case "bitbucket":
		provider, err = authFactory.NewProvider(auth.BITBUCKET, &bitbucket.Config{
			ClientID:     flags[BitBucketClientIDFlag].(*cli.StringFlag).Value,
			ClientSecret: flags[BitBucketClientSecretFlag].(*cli.StringFlag).Value,
			Workspace:    flags[BitBucketWorkspaceFlag].(*cli.StringFlag).Value,
		})
	case "gitlab":
		provider, err = authFactory.NewProvider(auth.GITLAB, &gitlab.Config{
			ClientID:                   flags[GitLabClientIDFlag].(*cli.StringFlag).Value,
			ClientSecret:               flags[GitLabClientSecretFlag].(*cli.StringFlag).Value,
			GitlabHostWithOptionalPort: flags[GitLabHostFlag].(*cli.StringFlag).Value,
			TerralistSchemeHostAndPort: userConfig.URL,
		})
	case "oidc":
		provider, err = authFactory.NewProvider(auth.OIDC, &oidc.Config{
			ClientID:                   flags[OidcClientIDFlag].(*cli.StringFlag).Value,
			ClientSecret:               flags[OidcClientSecretFlag].(*cli.StringFlag).Value,
			AuthorizeUrl:               flags[OidcAuthorizeUrlFlag].(*cli.StringFlag).Value,
			TokenUrl:                   flags[OidcTokenUrlFlag].(*cli.StringFlag).Value,
			UserInfoUrl:                flags[OidcUserInfoUrlFlag].(*cli.StringFlag).Value,
			Scope:                      flags[OidcScopeFlag].(*cli.StringFlag).Value,
			TerralistSchemeHostAndPort: userConfig.URL,
		})
	}
	if err != nil {
		return err
	}

	// Initialize storage resolver
	resolvers := map[string]storage.Resolver{
		"modules":   nil,
		"providers": nil,
	}
	resolversFlags := map[string]string{
		"modules":   ModulesStorageResolverFlag,
		"providers": ProvidersStorageResolverFlag,
	}

	for name, key := range resolversFlags {
		switch flags[key].(*cli.StringFlag).Value {
		case "proxy":
			resolvers[name], err = nil, nil
		case "local":
			// Initialize home directory
			homeDirClean := filepath.Clean(flags[LocalStoreFlag].(*cli.StringFlag).Value)
			if strings.HasPrefix(homeDirClean, "~") {
				userHomeDir, _ := os.UserHomeDir()
				homeDirClean = fmt.Sprintf("%s%s", userHomeDir, homeDirClean[1:])
			}

			homeDir, erro := filepath.Abs(homeDirClean)
			if erro != nil {
				return fmt.Errorf("invalid value for home directory: %v", err)
			}

			// Make sure Home Directory exists
			if erro := os.MkdirAll(homeDir, os.ModePerm); err != nil {
				return fmt.Errorf("could not create the home directory: %v", erro)
			}

			resolvers[name], err = storageFactory.NewResolver(storage.LOCAL, &local.Config{
				HomeDirectory: homeDir,
			})
		case "s3":
			resolvers[name], err = storageFactory.NewResolver(storage.S3, &s3.Config{
				Endpoint:             flags[S3EndpointFlag].(*cli.StringFlag).Value,
				BucketName:           flags[S3BucketNameFlag].(*cli.StringFlag).Value,
				BucketRegion:         flags[S3BucketRegionFlag].(*cli.StringFlag).Value,
				BucketPrefix:         flags[S3BucketPrefixFlag].(*cli.StringFlag).Value,
				AccessKeyID:          flags[S3AccessKeyIDFlag].(*cli.StringFlag).Value,
				SecretAccessKey:      flags[S3SecretAccessKeyFlag].(*cli.StringFlag).Value,
				LinkExpire:           flags[S3PresignExpireFlag].(*cli.IntFlag).Value,
				UsePathStyle:         flags[S3UsePathStyleFlag].(*cli.BoolFlag).Value,
				ServerSideEncryption: flags[S3ServerSideEncryptionFlag].(*cli.StringFlag).Value,
			})
		case "azure":
			resolvers[name], err = storageFactory.NewResolver(storage.AZURE, &azure.Config{
				AccountName:        flags[AzureAccountNameFlag].(*cli.StringFlag).Value,
				AccountKey:         flags[AzureAccountKeyFlag].(*cli.StringFlag).Value,
				ContainerName:      flags[AzureContainerNameFlag].(*cli.StringFlag).Value,
				SASExpire:          flags[AzureSASExpireFlag].(*cli.IntFlag).Value,
				DefaultCredentials: false,
			})
		case "gcs":
			resolvers[name], err = storageFactory.NewResolver(storage.GCS, &gcs.Config{
				BucketName:                 flags[GcsBucketNameFlag].(*cli.StringFlag).Value,
				BucketPrefix:               flags[GcsBucketPrefixFlag].(*cli.StringFlag).Value,
				ServiceAccountCredFilePath: flags[GcsServiceAccountCredFilePathFlag].(*cli.StringFlag).Value,
				LinkExpire:                 flags[GcsSignExpireFlag].(*cli.IntFlag).Value,
				DefaultCredentials:         false,
			})
		}

		if err != nil {
			return err
		}
	}

	// Initialize session store
	var store session.Store
	switch flags[SessionStoreFlag].(*cli.StringFlag).Value {
	case "cookie":
		store, err = sessionFactory.NewStore(session.COOKIE, &cookie.Config{
			Secret: flags[CookieSecretFlag].(*cli.StringFlag).Value,
		})
	}
	if err != nil {
		return err
	}

	srv, err := s.ServerCreator.NewServer(userConfig, server.Config{
		Database:          db,
		Provider:          provider,
		ModulesResolver:   resolvers["modules"],
		ProvidersResolver: resolvers["providers"],
		Store:             store,
		RunningMode:       s.RunningMode,
	})

	if err != nil {
		return errors.Wrap(err, "initializing server")
	}

	return srv.Start()
}

// withErrPrint prints out any cmd errors to stderr
func (s *Command) withErrPrint(f func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := f(cmd, args)
		if err != nil && !s.SilenceOutput {
			s.printErr(err)
		}
		return err
	}
}

// printErr prints err to stderr using a red terminal color
func (s *Command) printErr(err error) {
	log.Error().AnErr("error", err).Send()
}
