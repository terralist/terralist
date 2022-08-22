package s3

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"terralist/pkg/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// The S3 resolver will download files from the given URL then
// uploads them to an S3 bucket, generating a public download URL

// Resolver is the concrete implementation of storage.Resolver
type Resolver struct {
	CacheDir   string
	BucketName string
	LinkExpire int

	Session *session.Session
}

func (r *Resolver) Store(in *storage.StoreInput) (string, error) {
	key := fmt.Sprintf("%s/%s", in.KeyPrefix, in.FileName)
	if _, err := s3.New(r.Session).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(r.BucketName),
		Key:                  aws.String(key),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(in.Content),
		ContentLength:        aws.Int64(int64(len(in.Content))),
		ContentType:          aws.String(http.DetectContentType(in.Content)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	}); err != nil {
		return "", fmt.Errorf("could not upload archive: %v", err)
	}

	return key, nil
}

func (r *Resolver) Find(key string) (string, error) {
	req, _ := s3.New(r.Session).GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(key),
	})

	url, err := req.Presign(time.Duration(r.LinkExpire) * time.Minute)
	if err != nil {
		return "", fmt.Errorf("could not generate URL for %v: %v", key, err)
	}

	return url, nil
}

func (r *Resolver) Purge(key string) error {
	if _, err := s3.New(r.Session).DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("could not purge object: %v", err)
	}

	return nil
}
