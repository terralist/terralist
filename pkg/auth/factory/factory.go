package factory

import (
	"fmt"

	"terralist/pkg/auth"
	"terralist/pkg/auth/bitbucket"
	"terralist/pkg/auth/github"
	"terralist/pkg/auth/gitlab"
)

func NewProvider(backend auth.Backend, config auth.Configurator) (auth.Provider, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("could not create a new provider with invalid configuration: %v", err)
	}

	config.SetDefaults()

	var creator auth.Creator

	switch backend {
	case auth.GITHUB:
		creator = &github.Creator{}
	case auth.BITBUCKET:
		creator = &bitbucket.Creator{}
	case auth.GITLAB:
		creator = &gitlab.Creator{}
	default:
		return nil, fmt.Errorf("unrecognized backend type")
	}

	return creator.New(config)
}
