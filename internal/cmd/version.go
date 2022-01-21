package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type VersionCmd struct {
	Version string
}

func (v *VersionCmd) Init() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the current Terralist version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("terralist v%s\n", v.Version)
		},
	}
}
