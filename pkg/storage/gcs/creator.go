package gcs

import (
	"context"
	"fmt"

	"terralist/pkg/storage"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type Creator struct{}

func (t *Creator) New(config storage.Configurator) (storage.Resolver, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}
	ctx := context.Background()
	var client *gcs.Client
	var err error
	if !cfg.DefaultCredentials {
		client, err = gcs.NewClient(ctx, option.WithCredentialsFile(cfg.ServiceAccountCredFilePath))
		if err != nil {
			return nil, fmt.Errorf("could not login with ServiceAccountCredFilePath %v", err)
		}
	} else {
		client, err = gcs.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not login with default credentials %v", err)
		}
	}

	return &Resolver{
		BucketName:   cfg.BucketName,
		BucketPrefix: cfg.BucketPrefix,
		LinkExpire:   cfg.LinkExpire,

		Client: client,
	}, nil
}
