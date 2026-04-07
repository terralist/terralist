package database

import (
	"fmt"
	"time"

	"terralist/pkg/session"

	"github.com/rs/zerolog/log"
)

type Creator struct{}

func (c *Creator) New(config session.Configurator) (session.Store, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	// Auto-migrate the sessions table.
	if err := cfg.Database.Handler().AutoMigrate(&SessionRecord{}); err != nil {
		return nil, fmt.Errorf("failed to migrate sessions table: %w", err)
	}

	store := &Store{
		cookieName: cfg.CookieName,
		secret:     cfg.Secret,
		maxAge:     cfg.MaxAge,
		database:   cfg.Database,
	}

	// Start a background goroutine to clean up expired sessions.
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			store.Cleanup()
		}
	}()

	log.Info().Msg("Database session store initialized")

	return store, nil
}
