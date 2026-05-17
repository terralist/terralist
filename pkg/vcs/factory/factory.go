package factory

import (
	"fmt"

	"terralist/pkg/vcs"
	"terralist/pkg/vcs/github"

	"github.com/rs/zerolog/log"
)

func NewProvider(backend vcs.Backend, config vcs.Configurator) (vcs.Provider, error) {
	backendName := backendLabel(backend)

	if err := config.Validate(); err != nil {
		log.Error().
			Err(err).
			Str("vcs_backend", backendName).
			Msg("vcs provider config validation failed")
		return nil, fmt.Errorf("could not create a new resolver with invalid configuration: %v", err)
	}

	config.SetDefaults()

	var creator vcs.Creator

	switch backend {
	case vcs.GITHUB:
		creator = &github.Creator{}
	default:
		return nil, fmt.Errorf("unrecognized backend type: %d", backend)
	}

	provider, err := creator.New(config)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func backendLabel(backend vcs.Backend) string {
	switch backend {
	case vcs.GITHUB:
		return "github"
	default:
		return fmt.Sprintf("unknown_%d", backend)
	}
}
