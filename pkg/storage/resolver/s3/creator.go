package s3

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
	"path"
	"terralist/pkg/storage/resolver"
)

type Creator struct{}

func (t *Creator) New(config resolver.Configurator) (resolver.Resolver, error) {
	cfg, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unsupported configurator")
	}

	cacheDir := path.Join(cfg.HomeDirectory, "s3-cache")
	_ = os.MkdirAll(cacheDir, os.ModePerm)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.BucketRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("could not initiate AWS session: %v", err)
	}

	return &Resolver{
		CacheDir:   cacheDir,
		BucketName: cfg.BucketName,
		LinkExpire: cfg.LinkExpire,

		Session: sess,
	}, nil
}
