package main

import (
	"os"

	"terralist/cmd/terralist/server"
	"terralist/cmd/terralist/version"

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
	// rootCmd is the base command onto which other commands are added
	rootCmd := &cobra.Command{
		Use:   "terralist",
		Short: "Private Terraform Registry",
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	v := viper.New()

	serverCmd := &server.Command{
		RunningMode:   Mode,
		ServerCreator: &server.DefaultCreator{},
		Viper:         v,
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
