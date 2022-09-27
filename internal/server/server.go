package server

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"terralist/internal/server/controllers"
	"terralist/internal/server/handlers"
	"terralist/internal/server/repositories"
	"terralist/internal/server/services"
	"terralist/internal/server/views"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/database"
	"terralist/pkg/session"
	"terralist/pkg/storage"
	"terralist/pkg/webui"

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

	Controllers []api.RestController
}

// Config holds the server configuration that isn't configurable by the user
type Config struct {
	RunningMode string

	Database          database.Engine
	Provider          auth.Provider
	ModulesResolver   storage.Resolver
	ProvidersResolver storage.Resolver
	Store             session.Store
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

	// Parse host URL
	hostURL, err := url.Parse(userConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("host URL cannot be parsed")
	}

	// Apply initial migration
	if err := config.Database.WithMigration(&InitialMigration{}); err != nil {
		return nil, fmt.Errorf("could not apply initial migration: %v", err)
	}

	// Initialize webUI manager
	manager, err := webui.NewManager(views.FS)
	if err != nil {
		return nil, fmt.Errorf("could not create a new view manager: %v", err)
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

	loginController := &controllers.DefaultLoginController{
		Store:        config.Store,
		LoginService: loginService,

		EncryptSalt: salt,
		HostURL:     hostURL,
	}

	authorityRepository := &repositories.DefaultAuthorityRepository{
		Database: config.Database,
	}

	authorityService := &services.DefaultAuthorityService{
		AuthorityRepository: authorityRepository,
	}

	apiKeyRepository := &repositories.DefaultApiKeyRepository{
		Database: config.Database,
	}

	apiKeyService := &services.DefaultApiKeyService{
		ApiKeyRepository: apiKeyRepository,
		AuthorityService: authorityService,
	}

	moduleRepository := &repositories.DefaultModuleRepository{
		Database: config.Database,
	}

	moduleService := &services.DefaultModuleService{
		ModuleRepository: moduleRepository,
		AuthorityService: authorityService,
		Resolver:         config.ModulesResolver,
	}

	moduleController := &controllers.DefaultModuleController{
		ModuleService: moduleService,
		ApiKeyService: apiKeyService,
		JWT:           jwtManager,
	}

	providerRepository := &repositories.DefaultProviderRepository{
		Database: config.Database,
	}

	providerService := &services.DefaultProviderService{
		ProviderRepository: providerRepository,
		AuthorityService:   authorityService,
		Resolver:           config.ProvidersResolver,
	}

	providerController := &controllers.DefaultProviderController{
		ProviderService: providerService,
		ApiKeyService:   apiKeyService,
		JWT:             jwtManager,
	}

	serviceDiscoveryController := &controllers.DefaultServiceDiscoveryController{
		AuthorizationEndpoint: loginController.AuthorizationRoute(),
		TokenEndpoint:         loginController.TokenRoute(),
		ModuleEndpoint:        moduleController.TerraformApi(),
		ProviderEndpoint:      providerController.TerraformApi(),
	}

	webController := &controllers.DefaultWebController{
		Store:            config.Store,
		UIManager:        manager,
		AuthorityService: authorityService,
		ApiKeyService:    apiKeyService,

		ProviderName:          config.Provider.Name(),
		HostURL:               hostURL,
		AuthorizationEndpoint: loginController.AuthorizationRoute(),
	}

	return &Server{
		Port: userConfig.Port,

		Router:   router,
		Provider: config.Provider,
		Database: config.Database,
		JWT:      jwtManager,

		Controllers: []api.RestController{
			serviceDiscoveryController,
			loginController,
			moduleController,
			providerController,
			webController,
		},
	}, nil
}

// Start initializes the routes and starts serving
func (s *Server) Start() error {
	// Health Check API
	s.Router.GET("/health", handlers.Health())

	for _, c := range s.Controllers {
		var groups []*gin.RouterGroup

		paths := c.Paths()
		for _, p := range paths {
			groups = append(groups, s.Router.Group(p))
		}

		c.Subscribe(groups...)
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
