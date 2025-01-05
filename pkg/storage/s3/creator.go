package s3

import (
	"fmt"

	"terralist/pkg/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Creator struct{}

func (t *Creator) New(config storage.Configurator) (storage.Resolver, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	var creds *credentials.Credentials = nil
	var sharedConfig session.SharedConfigState = session.SharedConfigEnable

	if !cfg.DefaultCredentials {
		creds = credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, "")
		sharedConfig = session.SharedConfigDisable
	}

	var endpointResolver endpoints.ResolverFunc = endpoints.DefaultResolver().EndpointFor
	if cfg.Endpoint != "" {
		endpointResolver = endpoints.ResolverFunc(
			func(service, region string, opts ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
				return endpoints.ResolvedEndpoint{
					PartitionID:   "aws",
					URL:           cfg.Endpoint,
					SigningRegion: region,
				}, nil
			},
		)
	}

	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:           aws.String(cfg.BucketRegion),
			MaxRetries:       aws.Int(1),
			Credentials:      creds,
			EndpointResolver: endpointResolver,
			S3ForcePathStyle: aws.Bool(cfg.UsePathStyle),
		},
		SharedConfigState: sharedConfig,
	})

	if err != nil {
		return nil, fmt.Errorf("could not initiate AWS session: %v", err)
	}

	return &Resolver{
		BucketName:   cfg.BucketName,
		BucketPrefix: cfg.BucketPrefix,
		LinkExpire:   cfg.LinkExpire,

		ServerSideEncryption: cfg.ServerSideEncryption,

		Session: sess,
	}, nil
}
