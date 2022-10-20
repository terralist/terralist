package s3

import (
	"fmt"

	"terralist/pkg/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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

	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(cfg.BucketRegion),
			MaxRetries:  aws.Int(1),
			Credentials: creds,
		},
		SharedConfigState: sharedConfig,
	})

	if err != nil {
		return nil, fmt.Errorf("could not initiate AWS session: %v", err)
	}

	return &Resolver{
		BucketName: cfg.BucketName,
		LinkExpire: cfg.LinkExpire,

		Session: sess,
	}, nil
}
