package server

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"terralist/internal/server/controllers"
	"terralist/internal/server/handlers"
	"terralist/internal/server/repositories"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/database"
	"terralist/pkg/file"
	"terralist/pkg/session"
	"terralist/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/mazen160/go-random"
	"github.com/rs/zerolog/log"
)

// Server represents the Terralist server
type Server struct {
	Port int

	Router *gin.Engine

	JWT      jwt.JWT
	Provider auth.Provider
	Database database.Engine
	Resolver storage.Resolver

	Routers []api.Router

	Readiness *atomic.Bool
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
	// manager, err := webui.NewManager(web.FS, map[string]string{
	// 	"TERRALIST_HOST": 						"localhost",
	// 	"TERRALIST_COMPANY_NAME":     "",
	// 	"TERRALIST_OAUTH_PROVIDERS":  `["github"]`,
	// })

	// if err != nil {
	// 	return nil, fmt.Errorf("could not create a new view manager: %v", err)
	// }

	probeRouter := api.NewRouter(api.RouterOptions{
		Prefix:   "/check",
		Priority: api.RouterMaxPriority,
	})

	readiness := &atomic.Bool{}

	probeRouter.Register(&controllers.DefaultProbeController{
		Ready: readiness,
	})

	apiV1Router := api.NewRouter(api.RouterOptions{
		Prefix:   "/v1",
		Priority: api.RouterMinPriority - 1,
	})

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

	apiV1Router.Register(loginController)

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

	authorization := &handlers.Authorization{
		JWT:           jwtManager,
		ApiKeyService: apiKeyService,
		Store:         config.Store,
	}

	moduleRepository := &repositories.DefaultModuleRepository{
		Database: config.Database,
	}

	moduleService := &services.DefaultModuleService{
		ModuleRepository: moduleRepository,
		AuthorityService: authorityService,
		Resolver:         config.ModulesResolver,
		Fetcher:          file.NewFetcher(),
	}

	moduleController := &controllers.DefaultModuleController{
		ModuleService: moduleService,

		Authorization: authorization,
	}

	apiV1Router.Register(moduleController)

	providerRepository := &repositories.DefaultProviderRepository{
		Database: config.Database,
	}

	providerService := &services.DefaultProviderService{
		ProviderRepository: providerRepository,
		AuthorityService:   authorityService,
		Resolver:           config.ProvidersResolver,
		Fetcher:            file.NewFetcher(),
	}

	providerController := &controllers.DefaultProviderController{
		ProviderService: providerService,

		Authorization: authorization,
	}

	apiV1Router.Register(providerController)

	authorityController := &controllers.DefaultAuthorityController{
		AuthorityService: authorityService,

		Authorization: authorization,
	}

	apiV1Router.Register(authorityController)

	artifactController := &controllers.DefaultArtifactController{
		AuthorityService: authorityService,
		ModuleService:    moduleService,
		ProviderService:  providerService,

		Authorization: authorization,
	}

	apiV1Router.Register(artifactController)

	staticRouter := api.NewRouter(api.RouterOptions{
		Prefix:   "/",
		Priority: api.RouterMinPriority,
	})

	staticRouter.Register(&controllers.DefaultWebController{})

	wellKnownRouter := api.NewRouter(api.RouterOptions{
		Prefix:   "/.well-known",
		Priority: api.RouterMaxPriority,
	})

	wellKnownRouter.Register(&controllers.DefaultServiceDiscoveryController{
		AuthorizationEndpoint: apiV1Router.Prefix() + loginController.AuthorizationRoute(),
		TokenEndpoint:         apiV1Router.Prefix() + loginController.TokenRoute(),
		ModuleEndpoint:        apiV1Router.Prefix() + moduleController.TerraformApi(),
		ProviderEndpoint:      apiV1Router.Prefix() + providerController.TerraformApi(),
	})

	internalRouter := api.NewRouter(api.RouterOptions{
		Prefix:   "/internal",
		Priority: api.RouterMaxPriority + 1,
	})

	internalRouter.Register(&controllers.DefaultInternalController{
		HostURL:               hostURL.String(),
		CanonicalDomain:       hostURL.Host,
		CustomCompanyName:     userConfig.CustomCompanyName,
		OauthProviders:        []string{userConfig.OauthProvider},
		AuthorizationEndpoint: apiV1Router.Prefix() + loginController.AuthorizationRoute(),
		SessionDetailsRoute:   apiV1Router.Prefix() + loginController.SessionDetailsRoute(),
		ClearSessionRoute:     apiV1Router.Prefix() + loginController.ClearSessionRoute(),
	})

	routers := []api.Router{
		probeRouter,
		apiV1Router,
		staticRouter,
		wellKnownRouter,
		internalRouter,
	}

	sort.SliceStable(routers, func(i, j int) bool {
		return routers[i].Priority() < routers[j].Priority()
	})

	return &Server{
		Port: userConfig.Port,

		Router: router,

		JWT:      jwtManager,
		Provider: config.Provider,
		Database: config.Database,

		Routers: routers,

		Readiness: readiness,
	}, nil
}

// Start initializes the routes and starts serving
func (s *Server) Start() error {
	s.Router.Any("/*path", func(c *gin.Context) {
		path := c.Param("path")

		for _, r := range s.Routers {
			if strings.HasPrefix(path, r.Prefix()) {
				r.Router().HandleContext(c)
				return
			}
		}

		if c.Writer.Status() >= 400 {
			c.String(
				c.Writer.Status(),
				"%d: %s",
				c.Writer.Status(),
				http.StatusText(c.Writer.Status()),
			)
		}
	})

	// Ensure server gracefully drains connections when stopped
	stop := make(chan os.Signal, 1)

	go func() {
		// Mark the service as available
		s.Readiness.Store(true)

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
	// Mark the service as unavailable
	s.Readiness.Store(true)

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
