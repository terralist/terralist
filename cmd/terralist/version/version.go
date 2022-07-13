package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Command struct {
	Version        string
	CommitHash     string
	BuildTimestamp string
}

func (v *Command) Init() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the current Terralist version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(
				"terralist %s\nInfo:\n Commit Hash: %s\n Build Timestamp: %s\n",
				v.Version,
				v.CommitHash,
				v.BuildTimestamp,
			)
		},
	}
}
