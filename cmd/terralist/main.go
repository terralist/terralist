package main

import (
	"fmt"
	"os"

	"terralist/cmd/terralist/server"
	"terralist/cmd/terralist/version"
	"terralist/pkg/cli"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version        = "dev"
	CommitHash     = "n/a"
	BuildTimestamp = "n/a"

	Mode = "debug"
)

func main() {
	if err := cli.LoadEnvFiles(
		"TERRALIST_TOKEN_SIGNING_SECRET",
		"TERRALIST_COOKIE_SECRET",
		"TERRALIST_GH_CLIENT_ID",
		"TERRALIST_GH_CLIENT_SECRET",
		"TERRALIST_BB_CLIENT_ID",
		"TERRALIST_BB_CLIENT_SECRET",
		"TERRALIST_GL_CLIENT_ID",
		"TERRALIST_GL_CLIENT_SECRET",
		"TERRALIST_OI_CLIENT_ID",
		"TERRALIST_OI_CLIENT_SECRET",
		"TERRALIST_MYSQL_URL",
		"TERRALIST_MYSQL_PASSWORD",
		"TERRALIST_POSTGRES_URL",
		"TERRALIST_POSTGRES_PASSWORD",
	); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "load env files: %v\n", err)
		os.Exit(1)
	}

	// rootCmd is the base command onto which other commands are added
	rootCmd := &cobra.Command{
		Use:   "terralist",
		Short: "Private Terraform Registry",
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	v := viper.New()

	serverCmd := &server.Command{
		RunningMode:    Mode,
		Version:        Version,
		CommitHash:     CommitHash,
		BuildTimestamp: BuildTimestamp,
		ServerCreator:  &server.DefaultCreator{},
		Viper:          v,
	}

	versionCmd := &version.Command{
		Version:        Version,
		CommitHash:     CommitHash,
		BuildTimestamp: BuildTimestamp,
	}

	rootCmd.AddCommand(serverCmd.Init())
	rootCmd.AddCommand(versionCmd.Init())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
