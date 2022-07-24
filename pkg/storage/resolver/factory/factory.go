package factory

import (
	"fmt"

	"terralist/pkg/storage/resolver"
	"terralist/pkg/storage/resolver/local"
	"terralist/pkg/storage/resolver/proxy"
	"terralist/pkg/storage/resolver/s3"
)

func NewResolver(backend resolver.Backend, config resolver.Configurator) (resolver.Resolver, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("could not create a new resolver with invalid configuration: %v", err)
	}

	config.SetDefaults()

	var creator resolver.Creator

	switch backend {
	case resolver.PROXY:
		creator = &proxy.Creator{}
	case resolver.LOCAL:
		creator = &local.Creator{}
	case resolver.S3:
		creator = &s3.Creator{}
	default:
		return nil, fmt.Errorf("unrecognized backend type")
	}

	return creator.New(config)
}
