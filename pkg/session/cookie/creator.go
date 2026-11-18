package cookie

import (
	"fmt"
	"net/http"

	"terralist/pkg/session"

	"github.com/gorilla/sessions"
)

type Creator struct{}

func (c *Creator) New(config session.Configurator) (session.Store, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	store := sessions.NewCookieStore([]byte(cfg.Secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	return &Store{
		name: cfg.Name,

		store: store,
	}, nil
}
