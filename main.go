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

/* func main() {
	var isDebugMode bool
	if level, isSet := os.LookupEnv("TERRALIST_LEVEL"); !isSet {
		isDebugMode = true
	} else {
		if level == "debug" {
			isDebugMode = true
		} else {
			isDebugMode = false
		}
	}

	log := logrus.New()

	if isDebugMode {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.ErrorLevel)
	}

	if err := database.Connect(); err != nil {
		log.Fatal(err.Error())
	}

	if err := provider.InitProvider("github"); err != nil {
		log.Fatal(err.Error())
	}

	if !isDebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(ginlogrus.Logger(log), gin.Recovery())

	// Entry
	routes.InitEntryRoutes(r)

	// Login
	// https://www.terraform.io/docs/internals/login-protocol.html
	routes.InitLoginRoutes(r)

	// Modules
	// https://www.terraform.io/docs/internals/module-registry-protocol.html
	routes.InitModuleRoutes(r)

	// Providers
	// https://www.terraform.io/docs/internals/provider-registry-protocol.html
	routes.InitProviderRoutes(r)

	port, isSet := os.LookupEnv("TERRALIST_PORT")
	if !isSet {
		port = "8080"
	}

	r.Run(fmt.Sprintf(":%s", port))
} */
