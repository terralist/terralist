package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valentindeaconu/terralist/internal/server"
)

const (
	ConfigFlag          = "config"
	DatabaseBackendFlag = "database-backend"

	// GitHub OAuth Flags
	GitHubClientIDFlag     = "gh-client-id"
	GitHubClientSecretFlag = "gh-client-secret"
	GitHubOrganizationFlag = "gh-organization"

	LogLevelFlag           = "log-level"
	OAuthProviderFlag      = "oauth-provider"
	PortFlag               = "port"
	TokenSigningSecretFlag = "token-signing-secret"

	// Defaults
	DefaultDatabaseBackend = "sqlite"
	DefaultLogLevel        = "info"
	DefaultPort            = 5758
)

var stringFlags = map[string]stringFlag{
	ConfigFlag: {
		description: "Path to YAML config file where flag values are set.",
	},
	DatabaseBackendFlag: {
		description:  "The database backend.",
		choices:      []string{"sqlite"},
		defaultValue: DefaultDatabaseBackend,
	},
	GitHubClientIDFlag: {
		description:  "The GitHub OAuth Application client ID.",
		defaultValue: "",
	},
	GitHubClientSecretFlag: {
		description:  "The GitHub OAuth Application client secret.",
		defaultValue: "",
	},
	GitHubOrganizationFlag: {
		description:  "The GitHub organization to use for user validation.",
		defaultValue: "",
	},
	LogLevelFlag: {
		description:  "The log level.",
		choices:      []string{"debug", "info", "warn", "error"},
		defaultValue: DefaultLogLevel,
	},
	OAuthProviderFlag: {
		description:  "The OAuth 2.0 provider.",
		choices:      []string{"github"},
		defaultValue: "",
	},
	TokenSigningSecretFlag: {
		description:  "The secret to use when signing authorization tokens.",
		defaultValue: "",
	},
}

var intFlags = map[string]intFlag{
	PortFlag: {
		description:  "The port to bind to.",
		defaultValue: DefaultPort,
	},
}

var boolFlags = map[string]boolFlag{}

type stringFlag struct {
	description string
	// If empty means any value
	choices      []string
	defaultValue string
	hidden       bool
}

type intFlag struct {
	description  string
	defaultValue int
	hidden       bool
}

type boolFlag struct {
	description  string
	defaultValue bool
	hidden       bool
}

// ServerCmd is an abstraction for the server command
type ServerCmd struct {
	ServerCreator ServerCreator
	Viper         *viper.Viper
	Version       string
	Logger        *logrus.Logger
	SilenceOutput bool
}

// ServerCreator creates the server
type ServerCreator interface {
	NewServer(userConfig server.UserConfig, config server.Config) (ServerStarter, error)
}

// DefaultServerCreator is the concrete implementation of ServerCreator
type DefaultServerCreator struct{}

// ServerStarter starts the server
type ServerStarter interface {
	Start() error
}

// NewServer returns the real server object
func (d *DefaultServerCreator) NewServer(userConfig server.UserConfig, config server.Config) (ServerStarter, error) {
	return server.NewServer(userConfig, config)
}

func (s *ServerCmd) Init() *cobra.Command {
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

	c.SetUsageTemplate(usageTmpl(stringFlags, intFlags, boolFlags))
	// In case of invalid flags, print the error
	c.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		s.printErr(err)
		return err
	})

	// Set string flags
	for name, f := range stringFlags {
		usage := f.description

		if len(f.choices) > 0 {
			validOptions := strings.Join(f.choices, ", ")
			usage = fmt.Sprintf("%s Options: [%s].", usage, validOptions)
		}

		if f.defaultValue != "" {
			usage = fmt.Sprintf("%s (default %q)", usage, f.defaultValue)
		}

		c.Flags().String(name, "", usage+"\n")

		if f.hidden {
			c.Flags().MarkHidden(name)
		}

		s.Viper.BindPFlag(name, c.Flags().Lookup(name))
	}

	// Set int flags
	for name, f := range intFlags {
		usage := f.description

		if f.defaultValue != 0 {
			usage = fmt.Sprintf("%s (default %q)", usage, f.defaultValue)
		}

		c.Flags().Int(name, 0, usage+"\n")

		if f.hidden {
			c.Flags().MarkHidden(name)
		}

		s.Viper.BindPFlag(name, c.Flags().Lookup(name))
	}

	// Set bool flags
	for name, f := range boolFlags {
		usage := fmt.Sprintf("%s (default %v)", f.description, f.defaultValue)

		c.Flags().Bool(name, f.defaultValue, usage+"\n")

		if f.hidden {
			c.Flags().MarkHidden(name)
		}

		s.Viper.BindPFlag(name, c.Flags().Lookup(name))
	}

	return c
}

func (s *ServerCmd) preRun() error {
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

func (s *ServerCmd) run() error {
	var userConfig server.UserConfig

	if err := s.Viper.Unmarshal(&userConfig); err != nil {
		return err
	}

	s.setDefaults(&userConfig)
	s.Logger.SetLevel(userConfig.ToLogLevel())

	if err := s.validate(userConfig); err != nil {
		return err
	}

	s.securityWarnings(&userConfig)

	server, err := s.ServerCreator.NewServer(userConfig, server.Config{
		Version: s.Version,
	})

	if err != nil {
		return errors.Wrap(err, "initializing server")
	}

	return server.Start()
}

func (s *ServerCmd) setDefaults(c *server.UserConfig) {
	if c.Port == 0 {
		c.Port = DefaultPort
	}

	if c.LogLevel == "" {
		c.LogLevel = DefaultLogLevel
	}

	if c.DatabaseBackend == "" {
		c.DatabaseBackend = DefaultDatabaseBackend
	}
}

func (s *ServerCmd) validate(userConfig server.UserConfig) error {
	if err := validateChoice("log level", stringFlags["log-level"].choices, userConfig.LogLevel); err != nil {
		return err
	}

	if err := validateChoice("oauth provider", stringFlags["oauth-provider"].choices, userConfig.OAuthProvider); err != nil {
		return err
	}

	// If the github provider is given, assure that its parameters are also given
	if userConfig.OAuthProvider == "github" && (userConfig.GitHubClientID == "" || userConfig.GitHubClientSecret == "") {
		return fmt.Errorf("github oauth provider requires a client id and a client secret")
	}

	return nil
}

func (s *ServerCmd) securityWarnings(userConfig *server.UserConfig) {
	if s.SilenceOutput {
		return
	}

	if userConfig.OAuthProvider == "github" && userConfig.GitHubOrganization == "" {
		s.Logger.Warn("no github organization is set. Every request from a github authenticated user will pass")
	}

	if userConfig.TokenSigningSecret == "" {
		s.Logger.Warn("no token signing secret was provided. Tokens will be signed with an randomly generated secret which will be lost when the process exits.")
	}
}

// withErrPrint prints out any cmd errors to stderr
func (s *ServerCmd) withErrPrint(f func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := f(cmd, args)
		if err != nil && !s.SilenceOutput {
			s.printErr(err)
		}
		return err
	}
}

// printErr prints err to stderr using a red terminal color
func (s *ServerCmd) printErr(err error) {
	fmt.Fprintf(os.Stderr, "%sError: %s%s\n", "\033[31m", err.Error(), "\033[39m")
}

// validateChoice takes an argument name, a choices list and a value and
// validates that the value is one of the available choices, else it will
// return an error containing the argument name and the available options
// The choices array must contain only lowercase strings
func validateChoice(argument string, choices []string, value string) error {
	lv := strings.ToLower(value)

	if !contains(choices, lv) {
		options := strings.Join(choices, ", ")
		return fmt.Errorf("invalid %s, must be one of the values: %s", argument, options)
	}

	return nil
}

// contains checks if a string array contains a value
func contains(arr []string, i string) bool {
	for _, e := range arr {
		if e == i {
			return true
		}
	}

	return false
}
