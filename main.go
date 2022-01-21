package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valentindeaconu/terralist/internal/cmd"
)

const terralistVersion = "0.1.0"

func main() {
	v := viper.New()

	logger := logrus.New()

	server := &cmd.ServerCmd{
		ServerCreator: &cmd.DefaultServerCreator{},
		Viper:         v,
		Version:       terralistVersion,
		Logger:        logger,
	}

	version := &cmd.VersionCmd{
		Version: terralistVersion,
	}

	cmd.RootCmd.AddCommand(server.Init())
	cmd.RootCmd.AddCommand(version.Init())
	cmd.Execute()
}
