package factory

import (
	"fmt"

	"terralist/pkg/storage"
	"terralist/pkg/storage/local"
	"terralist/pkg/storage/proxy"
	"terralist/pkg/storage/s3"
)

func NewResolver(backend storage.Backend, config storage.Configurator) (storage.Resolver, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("could not create a new resolver with invalid configuration: %v", err)
	}

	config.SetDefaults()

	var creator storage.Creator

	switch backend {
	case storage.PROXY:
		creator = &proxy.Creator{}
	case storage.LOCAL:
		creator = &local.Creator{}
	case storage.S3:
		creator = &s3.Creator{}
	default:
		return nil, fmt.Errorf("unrecognized backend type")
	}

	return creator.New(config)
}
