package s3

import (
	"context"
	"fmt"

	"terralist/pkg/storage"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Creator struct{}

func (t *Creator) New(configurator storage.Configurator) (storage.Resolver, error) {
	options, ok := configurator.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	awsConfigResolvers := []func(*config.LoadOptions) error{
		config.WithRegion(options.BucketRegion),
		config.WithRetryMaxAttempts(1),
	}

	if !options.DefaultCredentials {
		credsProvider := credentials.NewStaticCredentialsProvider(options.AccessKeyID, options.SecretAccessKey, "")
		awsConfigResolvers = append(awsConfigResolvers, config.WithCredentialsProvider(credsProvider))
	}

	var endpointOptions = func(o *s3.Options) {}
	if options.Endpoint != "" {
		endpointOptions = func(o *s3.Options) {
			o.BaseEndpoint = &options.Endpoint
			o.Region = options.BucketRegion
			o.UsePathStyle = true
		}
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), awsConfigResolvers...)
	if err != nil {
		return nil, fmt.Errorf("could not load AWS config: %v", err)
	}

	client := s3.NewFromConfig(cfg, endpointOptions)

	return &Resolver{
		BucketName:           options.BucketName,
		BucketPrefix:         options.BucketPrefix,
		LinkExpire:           options.LinkExpire,
		ServerSideEncryption: options.ServerSideEncryption,
		DisableACL:           options.DisableACL,

		Client:    client,
		Presigner: s3.NewPresignClient(client),
	}, nil
}
