package server

import (
	"fmt"
	"os"
	"time"

	"terralist/internal/server/controllers"
	"terralist/internal/server/handlers"
	"terralist/internal/server/repositories"
	"terralist/internal/server/services"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/database"
	"terralist/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/mazen160/go-random"
	"github.com/rs/zerolog/log"
)

// Server represents the Terralist server
type Server struct {
	Port int

	JWT      jwt.JWT
	Router   *gin.Engine
	Provider auth.Provider
	Database database.Engine
	Resolver storage.Resolver

	ServiceDiscoveryController *controllers.ServiceDiscoveryController
	LoginController            *controllers.LoginController
	ModuleController           *controllers.ModuleController
	ProviderController         *controllers.ProviderController
}

// Config holds the server configuration that isn't configurable by the user
type Config struct {
	RunningMode string

	Database database.Engine
	Provider auth.Provider
	Resolver storage.Resolver
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

	jwtManager, _ := jwt.New(userConfig.TokenSigningSecret)

	salt, _ := random.String(32)
	exchangeKey, _ := random.String(32)

	loginService := &services.DefaultLoginService{
		Provider: config.Provider,
		JWT:      jwtManager,

		EncryptSalt:     salt,
		CodeExchangeKey: exchangeKey,
	}

	loginController := &controllers.LoginController{
		LoginService: loginService,
		EncryptSalt:  salt,
	}

	moduleRepository := &repositories.DefaultModuleRepository{
		Database: config.Database,
		Resolver: config.Resolver,
	}

	moduleService := &services.DefaultModuleService{
		ModuleRepository: moduleRepository,
	}

	moduleController := &controllers.ModuleController{
		ModuleService: moduleService,
		JWT:           jwtManager,
	}

	providerRepository := &repositories.DefaultProviderRepository{
		Database: config.Database,
	}

	providerService := &services.DefaultProviderService{
		ProviderRepository: providerRepository,
	}

	providerController := &controllers.ProviderController{
		ProviderService: providerService,
		JWT:             jwtManager,
	}

	serviceDiscoveryController := &controllers.ServiceDiscoveryController{
		AuthorizationEndpoint: loginController.AuthorizationRoute(),
		TokenEndpoint:         loginController.TokenRoute(),
		ModuleEndpoint:        moduleController.TerraformApiBase() + "/",
		ProviderEndpoint:      providerController.TerraformApiBase() + "/",
	}

	return &Server{
		Port: userConfig.Port,

		Router:   router,
		Provider: config.Provider,
		Database: config.Database,
		JWT:      jwtManager,

		ServiceDiscoveryController: serviceDiscoveryController,
		LoginController:            loginController,
		ModuleController:           moduleController,
		ProviderController:         providerController,
	}, nil
}

// Start initializes the routes and starts serving
func (s *Server) Start() error {
	// Health Check API
	s.Router.GET("/health", handlers.Health())

	ctrs := []interface {
		Subscribe(*gin.RouterGroup, *gin.RouterGroup)
		TerraformApiBase() string
		ApiBase() string
	}{
		s.ServiceDiscoveryController,
		s.LoginController,
		s.ModuleController,
		s.ProviderController,
	}

	for _, c := range ctrs {
		tfApi := s.Router.Group(c.TerraformApiBase())
		api := s.Router.Group(c.ApiBase())
		c.Subscribe(tfApi, api)
	}

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
