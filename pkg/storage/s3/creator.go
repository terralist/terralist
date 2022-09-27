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

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.BucketRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
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
