package server

import (
	"fmt"
	"os"
	"terralist/pkg/auth"
	"terralist/pkg/database"
	"terralist/pkg/storage/resolver"
	"time"

	"terralist/internal/server/controllers"
	"terralist/internal/server/handlers"
	"terralist/internal/server/services"
	"terralist/pkg/auth/jwt"

	"github.com/gin-gonic/gin"
	"github.com/mazen160/go-random"
	"github.com/rs/zerolog/log"
)

const (
	OAuthAuthorizationEndpoint = "/oauth/authorization"
	OAuthTokenEndpoint         = "/oauth/token"
	OAuthRedirectEndpoint      = "/oauth/redirect"
	ModuleEndpoint             = "/v1/modules"
	ProviderEndpoint           = "/v1/providers"
)

var (
	TerraformPorts = []int{10000, 10010}
)

// Server represents the Terralist server
type Server struct {
	Port int

	JWT      jwt.JWT
	Router   *gin.Engine
	Provider auth.Provider
	Database database.Engine
	Resolver resolver.Resolver

	EntryController    *controllers.EntryController
	LoginController    *controllers.LoginController
	ModuleController   *controllers.ModuleController
	ProviderController *controllers.ProviderController
}

// Config holds the server configuration that isn't configurable by the user
type Config struct {
	RunningMode string

	Database database.Engine
	Provider auth.Provider
	Resolver resolver.Resolver
}

func NewServer(userConfig UserConfig, config Config) (*Server, error) {
	if config.RunningMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if config.RunningMode == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(handlers.Logger())
	router.Use(gin.Recovery())

	// Apply initial migration
	if err := config.Database.WithMigration(&InitialMigration{}); err != nil {
		return nil, fmt.Errorf("could not apply initial migration: %v", err)
	}

	entryController := &controllers.EntryController{
		AuthorizationEndpoint: OAuthAuthorizationEndpoint,
		TokenEndpoint:         OAuthTokenEndpoint,
		ModuleEndpoint:        ModuleEndpoint + "/",
		ProviderEndpoint:      ProviderEndpoint + "/",
		TerraformPorts:        TerraformPorts,
	}

	jwtManager, _ := jwt.New(userConfig.TokenSigningSecret)

	salt, _ := random.String(32)
	exchangeKey, _ := random.String(32)

	loginController := &controllers.LoginController{
		Provider: config.Provider,
		JWT:      jwtManager,

		EncryptSalt:     salt,
		CodeExchangeKey: exchangeKey,
	}

	moduleService := &services.ModuleService{
		Database: config.Database,
		Resolver: config.Resolver,
	}

	moduleController := &controllers.ModuleController{
		ModuleService: moduleService,
	}

	providerService := &services.ProviderService{
		Database: config.Database,
		Resolver: config.Resolver,
	}

	providerController := &controllers.ProviderController{
		ProviderService: providerService,
	}

	return &Server{
		Port: userConfig.Port,

		Router:   router,
		Provider: config.Provider,
		Database: config.Database,
		JWT:      jwtManager,

		EntryController:    entryController,
		LoginController:    loginController,
		ModuleController:   moduleController,
		ProviderController: providerController,
	}, nil
}

// Start initializes the routes and starts serving
func (s *Server) Start() error {
	// Entry routes (no security checks)
	// Health Check API
	s.Router.GET("/health", s.EntryController.HealthCheck())

	// Terraform Service Discovery API
	s.Router.GET("/.well-known/terraform.json", s.EntryController.ServiceDiscovery())

	// Login routes (no security checks)
	// https://www.terraform.io/docs/internals/login-protocol.html
	s.Router.GET(OAuthAuthorizationEndpoint, s.LoginController.Authorize())
	s.Router.GET(OAuthRedirectEndpoint, s.LoginController.Redirect())
	s.Router.POST(OAuthTokenEndpoint, s.LoginController.TokenValidate())

	// Module routes (with security checks)
	// https://www.terraform.io/docs/internals/module-registry-protocol.html#list-available-versions-for-a-specific-module
	moduleRouter := s.Router.Group(ModuleEndpoint)
	moduleRouter.Use(handlers.Authorize(s.JWT))

	moduleRouter.GET("/:namespace/:name/:provider/versions", s.ModuleController.Get())

	// https://www.terraform.io/docs/internals/module-registry-protocol.html#download-source-code-for-a-specific-module-version
	moduleRouter.GET("/:namespace/:name/:provider/:version/download", s.ModuleController.GetVersion())

	// Upload a new module version
	moduleRouter.POST("/:namespace/:name/:provider/:version/upload", s.ModuleController.Upload())

	// Delete a module
	moduleRouter.DELETE("/:namespace/:name/:provider/remove", s.ModuleController.Delete())

	// Delete a module version
	moduleRouter.DELETE("/:namespace/:name/:provider/:version/remove", s.ModuleController.DeleteVersion())

	// Providers
	// https://www.terraform.io/docs/internals/provider-registry-protocol.html#list-available-versions
	providerRouter := s.Router.Group(ProviderEndpoint)
	providerRouter.Use(handlers.Authorize(s.JWT))

	providerRouter.GET("/:namespace/:name/versions", s.ProviderController.Get())

	// https://www.terraform.io/docs/internals/provider-registry-protocol.html#find-a-provider-package
	providerRouter.GET("/:namespace/:name/:version/download/:os/:arch", s.ProviderController.GetVersion())

	// Upload a new provider version
	providerRouter.POST("/:namespace/:name/:version/upload", s.ProviderController.Upload())

	// Delete a provider
	providerRouter.DELETE("/:namespace/:name/remove", s.ProviderController.Delete())

	// Delete a provider version
	providerRouter.DELETE("/:namespace/:name/:version/remove", s.ProviderController.DeleteVersion())

	// Ensure server gracefully drains connections when stopped
	stop := make(chan os.Signal, 1)

	go func() {
		log.Info().Msgf("Terralist started, listening on port %v", s.Port)

		err := s.Router.Run(fmt.Sprintf(":%d", s.Port))

		if err != nil {
			log.Error().AnErr("error", err).Send()
		}
	}()
	<-stop

	log.Warn().Msg("Received interrupt signal, waiting for in-progress operations to complete")
	s.waitForDrain()

	return nil
}

// waitForDrain blocks the process until draining is complete
func (s *Server) waitForDrain() {
	drainComplete := make(chan bool, 1)

	go func() {
		// TODO: Make Drainer when necessary
		drainComplete <- true
	}()

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-drainComplete:
			log.Info().Msg("All in-progress operations completed, shutting down.")
		case <-ticker.C:
			log.Info().Msg("Waiting for in-progress operations to complete...")
		}
	}
}
