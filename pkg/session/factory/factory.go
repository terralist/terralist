package factory

import (
	"fmt"
	"terralist/pkg/session"
	"terralist/pkg/session/cookie"
	"terralist/pkg/session/database"
)

func NewStore(backend session.Backend, config session.Configurator) (session.Store, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("could not create a new store with invalid configuration: %v", err)
	}

	config.SetDefaults()

	var creator session.Creator

	switch backend {
	case session.COOKIE:
		creator = &cookie.Creator{}
	case session.DATABASE:
		creator = &database.Creator{}
	default:
		return nil, fmt.Errorf("unrecognized backend type")
	}

	return creator.New(config)
}
