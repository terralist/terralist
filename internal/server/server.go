package server

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
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
	"terralist/pkg/rbac"
	"terralist/pkg/session"
	"terralist/pkg/storage"
	"terralist/web"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	random "github.com/mazen160/go-random"
	"github.com/rs/zerolog/log"
)

// Server represents the Terralist server.
type Server struct {
	Port     int
	CertFile string
	KeyFile  string

	Router *gin.Engine

	JWT      jwt.JWT
	Provider auth.Provider
	Database database.Engine
	Resolver storage.Resolver

	Readiness *atomic.Bool

	AuthorizedUsers string
}

// Config holds the server configuration that isn't configurable by the user.
type Config struct {
	RunningMode string

	Database          database.Engine
	Provider          auth.Provider
	ModulesResolver   storage.Resolver
	ProvidersResolver storage.Resolver
	Store             session.Store
}

func NewServer(userConfig UserConfig, config Config) (*Server, error) {
	// Set Gin mode based on the configuration
	switch config.RunningMode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "debug":
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

	// Serve static files (frontend) as middleware
	router.Use(static.Serve("/", web.StaticFS()))

	probeGroup := api.NewRouterGroup(router, &api.RouterGroupOptions{
		Prefix: "/check",
	})

	readiness := &atomic.Bool{}

	probeGroup.Register(&controllers.DefaultProbeController{
		Ready: readiness,
	})

	apiV1Group := api.NewRouterGroup(router, &api.RouterGroupOptions{
		Prefix: "/v1",
	})

	jwtManager, err := jwt.New(userConfig.TokenSigningSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %v", err)
	}

	salt, _ := random.String(32)
	exchangeKey, _ := random.String(32)

	// Parse token expiration duration
	tokenExpirationSeconds := services.ParseTokenExpiration(userConfig.AuthTokenExpiration)

	loginService := &services.DefaultLoginService{
		Provider: config.Provider,
		JWT:      jwtManager,

		EncryptSalt:         salt,
		CodeExchangeKey:     exchangeKey,
		TokenExpirationSecs: tokenExpirationSeconds,
	}

	loginController := &controllers.DefaultLoginController{
		Store:        config.Store,
		LoginService: loginService,

		EncryptSalt: salt,
		HostURL:     hostURL,
	}

	apiV1Group.Register(loginController)

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

	enforcer, err := rbac.NewEnforcer(userConfig.RbacPolicyPath, userConfig.RbacDefaultRole)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy enforcer: %v", err)
	}

	authentication := &handlers.Authentication{
		ApiKeyService: apiKeyService,
		JWT:           jwtManager,
		Store:         config.Store,
	}

	authorization := &handlers.Authorization{
		AuthorityService: authorityService,
		Enforcer:         enforcer,
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
		ModuleService:  moduleService,
		Authentication: authentication,
		Authorization:  authorization,
		AnonymousRead:  userConfig.ModulesAnonymousRead,

		HomeDir: userConfig.Home,
	}

	apiV1Group.Register(moduleController)

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
		Authentication:  authentication,
		Authorization:   authorization,
		AnonymousRead:   userConfig.ProvidersAnonymousRead,
	}

	apiV1Group.Register(providerController)

	authorityController := &controllers.DefaultAuthorityController{
		AuthorityService: authorityService,
		ApiKeyService:    apiKeyService,

		Authentication: authentication,
		Authorization:  authorization,
	}

	apiV1Group.Register(authorityController)

	artifactController := &controllers.DefaultArtifactController{
		AuthorityService: authorityService,
		ModuleService:    moduleService,
		ProviderService:  providerService,

		Authentication: authentication,
		Authorization:  authorization,
	}

	apiV1Group.Register(artifactController)

	wellKnownGroup := api.NewRouterGroup(router, &api.RouterGroupOptions{
		Prefix: "/.well-known",
	})

	wellKnownGroup.Register(&controllers.DefaultServiceDiscoveryController{
		AuthorizationEndpoint: apiV1Group.Prefix() + loginController.AuthorizationRoute(),
		TokenEndpoint:         apiV1Group.Prefix() + loginController.TokenRoute(),
		ModuleEndpoint:        apiV1Group.Prefix() + moduleController.TerraformApi(),
		ProviderEndpoint:      apiV1Group.Prefix() + providerController.TerraformApi(),
	})

	internalGroup := api.NewRouterGroup(router, &api.RouterGroupOptions{
		Prefix: "/internal",
	})

	internalGroup.Register(&controllers.DefaultInternalController{
		HostURL:               hostURL.String(),
		CanonicalDomain:       hostURL.Host,
		CustomCompanyName:     userConfig.CustomCompanyName,
		OauthProviders:        []string{userConfig.OauthProvider},
		AuthorizationEndpoint: apiV1Group.Prefix() + loginController.AuthorizationRoute(),
		SessionDetailsRoute:   apiV1Group.Prefix() + loginController.SessionDetailsRoute(),
		ClearSessionRoute:     apiV1Group.Prefix() + loginController.ClearSessionRoute(),
		AuthorizedUsers:       userConfig.AuthorizedUsers,
	})

	return &Server{
		Port:     userConfig.Port,
		CertFile: userConfig.CertFile,
		KeyFile:  userConfig.KeyFile,

		Router: router,

		JWT:      jwtManager,
		Provider: config.Provider,
		Database: config.Database,

		Readiness: readiness,
	}, nil
}

// Start initializes the routes and starts serving.
func (s *Server) Start() error {
	useTLS := s.CertFile != "" && s.KeyFile != ""

	if !useTLS {
		log.Warn().
			Msgf(
				"%s %s",
				"Terralist is running in HTTP mode which is not supported by Terraform.",
				"If you're using a proxy to serve on HTTPS, ignore this warning.",
			)
	}

	// Ensure server gracefully drains connections when stopped
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		// Mark the service as available
		s.Readiness.Store(true)

		log.Info().Msgf("Terralist started, listening on port %v", s.Port)
		var err error

		if useTLS {
			err = s.Router.RunTLS(fmt.Sprintf(":%d", s.Port), s.CertFile, s.KeyFile)
		} else {
			err = s.Router.Run(fmt.Sprintf(":%d", s.Port))
		}

		if err != nil {
			log.Error().AnErr("error", err).Send()
		}
	}()
	<-stop

	log.Warn().Msg("Received interrupt signal, waiting for in-progress operations to complete")
	s.waitForDrain()

	return nil
}

// waitForDrain blocks the process until draining is complete.
func (s *Server) waitForDrain() {
	// Mark the service as unavailable
	s.Readiness.Store(false)

	drainComplete := make(chan bool, 1)

	go func() {
		// TODO: Implement actual draining logic here
		drainComplete <- true
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-drainComplete:
			log.Info().Msg("All in-progress operations completed, shutting down.")
			return
		case <-ticker.C:
			log.Info().Msg("Waiting for in-progress operations to complete...")
		}
	}
}
