package factory

import (
	"fmt"

	"terralist/pkg/storage"
	"terralist/pkg/storage/azure"
	"terralist/pkg/storage/gcs"
	"terralist/pkg/storage/local"
	"terralist/pkg/storage/s3"
)

func NewResolver(backend storage.Backend, config storage.Configurator) (storage.Resolver, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("could not create a new resolver with invalid configuration: %v", err)
	}

	config.SetDefaults()

	var creator storage.Creator
	var backendName string

	switch backend {
	case storage.LOCAL:
		creator = &local.Creator{}
		backendName = "local"
	case storage.S3:
		creator = &s3.Creator{}
		backendName = "s3"
	case storage.AZURE:
		creator = &azure.Creator{}
		backendName = "azure"
	case storage.GCS:
		creator = &gcs.Creator{}
		backendName = "gcs"
	default:
		return nil, fmt.Errorf("unrecognized backend type")
	}

	resolver, err := creator.New(config)
	if err != nil {
		return nil, err
	}

	// Wrap resolver with metrics decorator
	return &storage.MetricsResolver{
		Resolver: resolver,
		Backend:  backendName,
	}, nil
}
