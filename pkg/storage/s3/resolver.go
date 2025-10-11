package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"terralist/pkg/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Client is a wrapper interface around the S3 client methods used by the Resolver.
type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// PresignClient is a wrapper interface around the S3 Presign client methods used by the Resolver.
type PresignClient interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// Resolver is the concrete implementation of storage.Resolver.
// The S3 resolver will download files from the given URL then
// uploads them to an S3 bucket, generating a public download URL.
type Resolver struct {
	BucketName   string
	BucketPrefix string
	LinkExpire   int

	ServerSideEncryption string
	DisableACL           bool

	Client    S3Client
	Presigner PresignClient
}

func (r *Resolver) Store(in *storage.StoreInput) (string, error) {
	key := fmt.Sprintf("%s/%s", in.KeyPrefix, in.FileName)

	// Preemptively rewind the file to make sure all content is available.
	if _, err := in.Reader.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("could not upload archive, file can't be rewinded: %w", err)
	}

	putObjectInput := &s3.PutObjectInput{
		Bucket:             aws.String(r.BucketName),
		Key:                r.withPrefix(key),
		Body:               in.Reader,
		ContentLength:      aws.Int64(in.Size),
		ContentType:        aws.String(in.ContentType),
		ContentDisposition: aws.String("attachment"),
	}

	// Only set ACL if not disabled (for bucket policy support)
	if !r.DisableACL {
		putObjectInput.ACL = types.ObjectCannedACLPrivate
	}

	if r.ServerSideEncryption != "none" {
		putObjectInput.ServerSideEncryption = types.ServerSideEncryption(r.ServerSideEncryption)
	}

	if _, err := r.Client.PutObject(context.TODO(), putObjectInput); err != nil {
		return "", fmt.Errorf("could not upload archive: %v", err)
	}

	return key, nil
}

func (r *Resolver) Find(key string) (string, error) {
	req, err := r.Presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(r.BucketName),
		Key:    r.withPrefix(key),
	}, func(po *s3.PresignOptions) {
		po.Expires = time.Duration(r.LinkExpire) * time.Minute
	})
	if err != nil {
		return "", fmt.Errorf("could not generate URL for %v: %v", key, err)
	}

	return req.URL, nil
}

func (r *Resolver) Purge(key string) error {
	if _, err := r.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
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
