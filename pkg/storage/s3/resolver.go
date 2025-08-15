package s3

import (
	"fmt"
	"io"
	"time"

	"terralist/pkg/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// The S3 resolver will download files from the given URL then
// uploads them to an S3 bucket, generating a public download URL.

// Resolver is the concrete implementation of storage.Resolver.
type Resolver struct {
	BucketName   string
	BucketPrefix string
	LinkExpire   int

	ServerSideEncryption string
	DisableACL           bool

	Session *session.Session
}

func (r *Resolver) Store(in *storage.StoreInput) (string, error) {
	key := fmt.Sprintf("%s/%s", in.KeyPrefix, in.FileName)

	serverSideEncryption := aws.String(r.ServerSideEncryption)
	if r.ServerSideEncryption == "none" {
		serverSideEncryption = nil
	}

	// Needed to explicitly rewind the file because it has been entirely consumed before.
	// Otherwise the body will be nil due to the Read method returning nothing.
	if _, err := in.Reader.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("could not upload archive, file can't be rewinded: %w", err)
	}

	putObjectInput := &s3.PutObjectInput{
		Bucket:               aws.String(r.BucketName),
		Key:                  r.withPrefix(key),
		Body:                 in.Reader,
		ContentLength:        aws.Int64(in.Size),
		ContentType:          aws.String(in.ContentType),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: serverSideEncryption,
	}

	// Only set ACL if not disabled (for bucket policy support)
	if !r.DisableACL {
		putObjectInput.ACL = aws.String("private")
	}

	if _, err := s3.New(r.Session).PutObject(putObjectInput); err != nil {
		return "", fmt.Errorf("could not upload archive: %v", err)
	}

	return key, nil
}

func (r *Resolver) Find(key string) (string, error) {
	req, _ := s3.New(r.Session).GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    r.withPrefix(key),
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
		Key:    r.withPrefix(key),
	}); err != nil {
		return fmt.Errorf("could not purge object: %v", err)
	}

	return nil
}

func (r *Resolver) withPrefix(key string) *string {
	return aws.String(fmt.Sprintf("%s%s", r.BucketPrefix, key))
}
