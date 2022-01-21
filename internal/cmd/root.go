package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// RootCmd is the base command onto which other commands are added
var RootCmd = &cobra.Command{
	Use:   "terralist",
	Short: "Private Terraform Registry",
}

// Execute launches the RootCmd
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
