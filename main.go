package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
	"github.com/valentindeaconu/terralist/database"
	"github.com/valentindeaconu/terralist/oauth/provider"
	"github.com/valentindeaconu/terralist/routes"
)

func main() {
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
}
