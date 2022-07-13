package server

import (
	"fmt"
	"strings"

	"terralist/internal/server"
	"terralist/pkg/cli"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ConfigFlag = "config"

	PortFlag    = "port"
	DefaultPort = 5758

	LogLevelFlag    = "log-level"
	DefaultLogLevel = "info"

	DatabaseBackendFlag    = "database-backend"
	DefaultDatabaseBackend = "sqlite"

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
		DefaultValue: DefaultPort,
	},

	LogLevelFlag: &cli.StringFlag{
		Description:  "The log level.",
		Choices:      []string{"trace", "debug", "info", "warn", "error"},
		DefaultValue: DefaultLogLevel,
	},

	DatabaseBackendFlag: &cli.StringFlag{
		Description:  "The database backend.",
		Choices:      []string{"sqlite"},
		DefaultValue: DefaultDatabaseBackend,
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

	// Set values from viper
	for k, v := range raw {
		if _, ok := flags[k]; ok {
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

	var userConfig server.UserConfig
	if err := mapstructure.Decode(raw, &userConfig); err != nil {
		return fmt.Errorf("could not unpack to user configuration: %v", err)
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

	if err := s.validate(userConfig); err != nil {
		return err
	}

	s.securityWarnings(&userConfig)

	srv, err := s.ServerCreator.NewServer(userConfig, server.Config{
		RunningMode: s.RunningMode,
	})

	if err != nil {
		return errors.Wrap(err, "initializing server")
	}

	return srv.Start()
}

func (s *Command) validate(userConfig server.UserConfig) error {
	for k, v := range flags {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("could not validate %v: %v", k, err)
		}
	}

	// If the GitHub provider is given, assure that its parameters are also given
	if userConfig.OAuthProvider == "github" && (userConfig.GitHubClientID == "" || userConfig.GitHubClientSecret == "") {
		return fmt.Errorf("github oauth provider requires a client id and a client secret")
	}

	return nil
}

func (s *Command) securityWarnings(userConfig *server.UserConfig) {
	if s.SilenceOutput {
		return
	}

	if userConfig.OAuthProvider == "github" && userConfig.GitHubOrganization == "" {
		log.Warn().
			Msg(
				"No github organization is set. " +
					"Every request from a github authenticated user will pass.",
			)
	}

	if userConfig.TokenSigningSecret == "" {
		log.Warn().
			Msg(
				"No token signing secret was provided. " +
					"Tokens will be signed with an randomly generated secret which will be lost when the process exits.",
			)
	}
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
