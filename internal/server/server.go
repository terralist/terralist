package server

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"

	"terralist/internal/server/controllers"
	"terralist/internal/server/handlers"
	"terralist/internal/server/models/oauth"
	"terralist/internal/server/repositories"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/auth/saml"
	"terralist/pkg/database"
	"terralist/pkg/file"
	"terralist/pkg/metrics"
	"terralist/pkg/rbac"
	"terralist/pkg/session"
	"terralist/pkg/storage"
	"terralist/pkg/storage/local"
	"terralist/web"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	random "github.com/mazen160/go-random"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

// Server represents the Terralist server.
type Server struct {
	Port        int
	MetricsPort int
	CertFile    string
	KeyFile     string

	Router        *gin.Engine
	MetricsRouter *http.ServeMux

	JWT      jwt.JWT
	Provider auth.Provider
	Database database.Engine
	Resolver storage.Resolver

	Readiness *atomic.Bool
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

	// Get SQL DB for metrics
	var sqlDB *sql.DB
	if config.Database != nil {
		sqlDB, _ = config.Database.Handler().DB()
	}

	// Create Prometheus metrics registry
	metricsRegistry := metrics.NewRegistry(&metrics.RegistryConfig{
		SqlDB: sqlDB,
	})

	router := gin.New()
	router.Use(handlers.PrometheusMetrics(metricsRegistry))
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

	// Prometheus metrics endpoint
	var metricsRouter *http.ServeMux
	if userConfig.MetricsPort > 0 {
		metricsRouter = http.NewServeMux()
		metricsRouter.Handle("/metrics", promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{}))
	} else {
		router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})))
	}

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
		Provider:  config.Provider,
		JWT:       jwtManager,
		CodeStore: services.NewInMemoryOAuthCodeStore(2 * time.Minute),

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

	// Register SAML endpoints only when SAML is configured as the auth provider.
	if samlProvider, ok := config.Provider.(interface {
		GetSPMetadata() ([]byte, error)
	}); ok {
		acsRateLimiter := handlers.NewRateLimiter(10, 1*time.Minute)
		samlEndpoints := apiV1Group.RouterGroup().Group("/api/auth/saml")

		// Enforce HTTPS transport for SAML endpoints.
		samlEndpoints.Use(func(ctx *gin.Context) {
			scheme := ctx.GetHeader("X-Forwarded-Proto")
			if scheme == "" {
				if ctx.Request.TLS != nil {
					scheme = "https"
				} else {
					scheme = "http"
				}
			}

			if strings.ToLower(scheme) != "https" {
				log.Error().
					Str("scheme", scheme).
					Str("path", ctx.Request.URL.Path).
					Str("remote_addr", ctx.ClientIP()).
					Msg("SAML endpoint accessed over non-HTTPS connection - rejecting request")
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":             "forbidden",
					"error_description": "SAML endpoints require HTTPS transport for security",
				})
				return
			}

			ctx.Next()
		})

		samlEndpoints.GET("/metadata", func(ctx *gin.Context) {
			metadata, err := samlProvider.GetSPMetadata()
			if err != nil {
				log.Error().
					AnErr("Error", err).
					Msg("Failed to generate SAML metadata")
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			ctx.Data(http.StatusOK, "application/xml; charset=utf-8", metadata)
		})

		samlEndpoints.POST("/acs", handlers.RateLimitMiddleware(acsRateLimiter), func(ctx *gin.Context) {
			samlResponse := ctx.PostForm("SAMLResponse")
			relayState := ctx.PostForm("RelayState")

			if samlResponse == "" || relayState == "" {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}

			if err := saml.ValidateRelayState(relayState); err != nil {
				log.Warn().
					AnErr("error", err).
					Str("source", "relaystate_validation").
					Int("relaystate_size", len(relayState)).
					Str("client_ip", ctx.ClientIP()).
					Msg("SAML ACS endpoint: invalid RelayState detected (potential CSRF attack)")
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}

			r, err := oauth.Payload(relayState).ToRequest(salt)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}

			codeComponents, erro := loginService.UnpackCode(samlResponse, &r)
			if erro != nil {
				ctx.Redirect(http.StatusFound, redirectWithError(r.RedirectURI, r.State, erro))
				return
			}

			var userDetails auth.User
			if err := loginService.Provider.GetUserDetails(samlResponse, &userDetails); err != nil {
				userDetails = auth.User{
					Name:        codeComponents.UserName,
					Email:       codeComponents.UserEmail,
					Groups:      codeComponents.UserGroups,
					Authority:   codeComponents.UserAuthority,
					AuthorityID: codeComponents.UserAuthorityID,
				}
			}

			uri, err := url.Parse(r.RedirectURI)
			if err != nil {
				log.Warn().
					AnErr("Error", err).
					Str("RedirectURI", r.RedirectURI).
					Msg("An invalid redirect URI was detected during the SAML callback.")
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}

			if uri.Host == hostURL.Host {
				sess, err := config.Store.Get(ctx.Request)
				if err != nil {
					log.Error().
						Err(err).
						Str("user", userDetails.Email).
						Msg("SAML ACS: failed to fetch session")

					ctx.Redirect(http.StatusFound, redirectWithError(
						uri.String(),
						"",
						oauth.WrapError(fmt.Errorf("could not fetch the session"), oauth.ServerError),
					))
					return
				}

				sess.Set("user", &auth.User{
					Name:        userDetails.Name,
					Email:       userDetails.Email,
					Groups:      userDetails.Groups,
					Authority:   userDetails.Authority,
					AuthorityID: userDetails.AuthorityID,
				})

				if err := config.Store.Save(ctx.Request, ctx.Writer, sess); err != nil {
					log.Error().
						Err(err).
						Str("user", userDetails.Email).
						Int("groups_count", len(userDetails.Groups)).
						Msg("SAML ACS: failed to save session")

					ctx.Redirect(http.StatusFound, redirectWithError(
						uri.String(),
						"",
						oauth.WrapError(fmt.Errorf("could not save session"), oauth.ServerError),
					))
					return
				}

				ctx.Redirect(http.StatusFound, uri.String())
				return
			}

			redirectURL, erro := loginService.Redirect(codeComponents, &r)
			if erro != nil {
				ctx.Redirect(http.StatusFound, redirectWithError(r.RedirectURI, r.State, erro))
				return
			}

			ctx.Redirect(http.StatusFound, redirectURL)
		})
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

	enforcer, err := rbac.NewEnforcer(userConfig.RbacPolicyPath, userConfig.RbacDefaultRole)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy enforcer: %v", err)
	}

	standaloneApiKeyRepository := &repositories.DefaultStandaloneApiKeyRepository{
		Database: config.Database,
	}

	standaloneApiKeyService := &services.DefaultStandaloneApiKeyService{
		Repository: standaloneApiKeyRepository,
	}

	authentication := &handlers.Authentication{
		ApiKeyService:           apiKeyService,
		StandaloneApiKeyService: standaloneApiKeyService,
		MasterApiKey:            userConfig.MasterApiKey,
		JWT:                     jwtManager,
		Store:                   config.Store,
	}

	authorization := &handlers.Authorization{
		AuthorityService: authorityService,
		Enforcer:         enforcer,
	}

	settingsCapabilityController := &controllers.DefaultSettingsCapabilityController{
		Authentication: authentication,
		Authorization:  authorization,
	}

	apiV1Group.Register(settingsCapabilityController)

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
		ModuleService:    moduleService,
		AuthorityService: authorityService,
		Authentication:   authentication,
		Authorization:    authorization,
		AnonymousRead:    userConfig.ModulesAnonymousRead,
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
		ProviderService:  providerService,
		AuthorityService: authorityService,
		Authentication:   authentication,
		Authorization:    authorization,
		AnonymousRead:    userConfig.ProvidersAnonymousRead,
	}

	apiV1Group.Register(providerController)

	authorityController := &controllers.DefaultAuthorityController{
		AuthorityService: authorityService,
		ApiKeyService:    apiKeyService,

		Authentication: authentication,
		Authorization:  authorization,
	}

	apiV1Group.Register(authorityController)

	apiKeyController := &controllers.DefaultApiKeyController{
		Service:        standaloneApiKeyService,
		Authentication: authentication,
		Authorization:  authorization,
	}

	apiV1Group.Register(apiKeyController)

	artifactController := &controllers.DefaultArtifactController{
		AuthorityService: authorityService,
		ModuleService:    moduleService,
		ProviderService:  providerService,

		Authentication: authentication,
		Authorization:  authorization,
	}

	apiV1Group.Register(artifactController)

	modulesLocal := local.UnwrapResolver(config.ModulesResolver)
	providersLocal := local.UnwrapResolver(config.ProvidersResolver)
	if modulesLocal != nil || providersLocal != nil {
		localJWTManager, err := jwt.New(userConfig.LocalTokenSigningSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to create local JWT manager: %v", err)
		}

		filesController := &controllers.DefaultFileServer{
			ModulesResolver:   config.ModulesResolver,
			ProvidersResolver: config.ProvidersResolver,
			JWT:               localJWTManager,
		}

		apiV1Group.Register(filesController)
	}

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
		SamlDisplayName:       userConfig.SamlDisplayName,
	})

	return &Server{
		Port:        userConfig.Port,
		MetricsPort: userConfig.MetricsPort,
		CertFile:    userConfig.CertFile,
		KeyFile:     userConfig.KeyFile,

		Router:        router,
		MetricsRouter: metricsRouter,

		JWT:      jwtManager,
		Provider: config.Provider,
		Database: config.Database,

		Readiness: readiness,
	}, nil
}

// redirectWithError creates a redirect URL with OAuth error parameters.
// Error messages are sanitized to prevent information leakage.
func redirectWithError(uri string, state string, err oauth.Error) string {
	stateQuery := ""
	if state != "" {
		stateQuery = fmt.Sprintf("&state=%s", state)
	}

	errorMsg := err.Error()
	if sanitized := saml.SanitizeErrorForURL(errors.New(errorMsg)); sanitized != errorMsg {
		errorMsg = sanitized
	}

	return fmt.Sprintf(
		"%s?error=%s&error_description=%s%s",
		uri,
		err.Kind(),
		url.QueryEscape(errorMsg),
		stateQuery,
	)
}

// Start initializes the routes and starts serving.
func (s *Server) Start() error {
	useTLS := s.CertFile != "" && s.KeyFile != ""

	if _, ok := s.Provider.(interface {
		GetSPMetadata() ([]byte, error)
	}); ok && !useTLS {
		log.Warn().
			Msg("SAML authentication requires TLS/HTTPS transport for security. Please ensure your reverse proxy terminates TLS, or configure cert-file and key-file for direct TLS support.")
	}

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

	// Start metrics server on a separate port if configured.
	if s.MetricsRouter != nil && s.MetricsPort > 0 {
		go func() {
			addr := fmt.Sprintf(":%d", s.MetricsPort)
			log.Info().Msgf("Metrics server listening on port %v", s.MetricsPort)
			if err := http.ListenAndServe(addr, s.MetricsRouter); err != nil {
				log.Error().Err(err).Msg("Metrics server failed")
			}
		}()
	}

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
