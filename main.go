package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/valentindeaconu/terralist/api"
	"github.com/valentindeaconu/terralist/database"
	"github.com/valentindeaconu/terralist/service"
)

func main() {
	if err := database.Init(); err != nil {
		log.Fatal(err.Error())
	}

	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Entry
	router.GET("/health", api.Health)
	router.GET("/.well-known/terraform.json", api.ServiceDiscovery)

	// Login
	// https://www.terraform.io/docs/internals/login-protocol.html
	// TODO

	// Modules
	// https://www.terraform.io/docs/internals/module-registry-protocol.html
	moduleController := api.CreateModuleController(router, &service.ModuleService{})
	moduleController.PrepareRoutes()

	// Providers
	// https://www.terraform.io/docs/internals/provider-registry-protocol.html
	providerController := api.CreateProviderController(router, &service.ProviderService{})
	providerController.PrepareRoutes()

	router.Run(":8080")
}
