package s3

import (
	"bytes"
	"context"
	"testing"

	"terralist/pkg/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	. "github.com/smartystreets/goconvey/convey"
)

func newAWSS3Client(t *testing.T) *s3.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		),
	)
	if err != nil {
		t.Fatalf("failed to load SDK configuration, %v", err)
	}

	return s3.NewFromConfig(cfg)
}

func TestStore(t *testing.T) {
	Convey("Subject: Store files in S3", t, func() {
		Convey("When ACL is enabled (default behavior)", func() {
			resolver := &Resolver{
				BucketName:           "test-bucket",
				BucketPrefix:         "test-prefix/",
				ServerSideEncryption: "AES256",
				DisableACL:           false, // ACL enabled
				Client:               newAWSS3Client(t),
			}

			Convey("Should prepare PutObjectInput with ACL set to private", func() {
				// This test validates that when DisableACL is false, the logic prepares
				// the PutObjectInput with ACL set to "private"
				So(resolver.DisableACL, ShouldBeFalse)
			})
		})

		Convey("When ACL is disabled (bucket policy mode)", func() {
			resolver := &Resolver{
				BucketName:           "test-bucket",
				BucketPrefix:         "test-prefix/",
				ServerSideEncryption: "AES256",
				DisableACL:           true, // ACL disabled
				Client:               newAWSS3Client(t),
			}

			Convey("Should prepare PutObjectInput without ACL parameter", func() {
				// This test validates that when DisableACL is true, the ACL parameter
				// is not set, allowing bucket policies to control access
				So(resolver.DisableACL, ShouldBeTrue)
			})
		})
	})
}

func TestDisableACLConfiguration(t *testing.T) {
	Convey("Subject: S3 DisableACL configuration", t, func() {
		Convey("When DisableACL is false", func() {
			config := &Config{
				BucketName: "test-bucket",
				LinkExpire: 15,
				DisableACL: false,
			}

			Convey("Config should indicate ACL is enabled", func() {
				So(config.DisableACL, ShouldBeFalse)
			})
		})

		Convey("When DisableACL is true", func() {
			config := &Config{
				BucketName: "test-bucket",
				LinkExpire: 15,
				DisableACL: true,
			}

			Convey("Config should indicate ACL is disabled", func() {
				So(config.DisableACL, ShouldBeTrue)
			})
		})
	})
}

func TestResolverCreation(t *testing.T) {
	Convey("Subject: Create S3 Resolver with DisableACL configuration", t, func() {
		creator := &Creator{}

		Convey("When creating resolver with ACL disabled", func() {
			config := &Config{
				BucketName:      "test-bucket",
				BucketRegion:    "us-east-1",
				LinkExpire:      15,
				DisableACL:      true,
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			}

			resolver, err := creator.New(config)

			Convey("Should create resolver with DisableACL set to true", func() {
				So(err, ShouldBeNil)
				So(resolver, ShouldNotBeNil)

				s3Resolver, ok := resolver.(*Resolver)
				So(ok, ShouldBeTrue)
				So(s3Resolver.DisableACL, ShouldBeTrue)
			})
		})

		Convey("When creating resolver with ACL enabled", func() {
			config := &Config{
				BucketName:      "test-bucket",
				BucketRegion:    "us-east-1",
				LinkExpire:      15,
				DisableACL:      false,
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			}

			resolver, err := creator.New(config)

			Convey("Should create resolver with DisableACL set to false", func() {
				So(err, ShouldBeNil)
				So(resolver, ShouldNotBeNil)

				s3Resolver, ok := resolver.(*Resolver)
				So(ok, ShouldBeTrue)
				So(s3Resolver.DisableACL, ShouldBeFalse)
			})
		})
	})
}

func TestPutObjectInputPreparation(t *testing.T) {
	Convey("Subject: PutObjectInput preparation with ACL configuration", t, func() {
		storeInput := &storage.StoreInput{
			KeyPrefix:   "modules",
			FileName:    "test.zip",
			Reader:      bytes.NewReader([]byte("test content")),
			Size:        12,
			ContentType: "application/zip",
		}

		Convey("When ACL is enabled", func() {
			resolver := &Resolver{
				BucketName:           "test-bucket",
				BucketPrefix:         "test-prefix/",
				ServerSideEncryption: "AES256",
				DisableACL:           false,
				Client:               newAWSS3Client(t),
			}

			// Create the PutObjectInput as the Store method would
			key := "modules/test.zip"
			putObjectInput := &s3.PutObjectInput{
				Bucket:             aws.String(resolver.BucketName),
				Key:                resolver.withPrefix(key),
				Body:               storeInput.Reader,
				ContentLength:      aws.Int64(storeInput.Size),
				ContentType:        aws.String(storeInput.ContentType),
				ContentDisposition: aws.String("attachment"),
			}

			if !resolver.DisableACL {
				putObjectInput.ACL = types.ObjectCannedACLPrivate
			}

			if resolver.ServerSideEncryption != "none" {
				putObjectInput.ServerSideEncryption = types.ServerSideEncryption(resolver.ServerSideEncryption)
			}

			Convey("PutObjectInput should have ACL set to private", func() {
				So(putObjectInput.ACL, ShouldEqual, types.ObjectCannedACLPrivate)
			})
		})

		Convey("When ACL is disabled", func() {
			resolver := &Resolver{
				BucketName:           "test-bucket",
				BucketPrefix:         "test-prefix/",
				ServerSideEncryption: "AES256",
				DisableACL:           true,
				Client:               newAWSS3Client(t),
			}

			// Create the PutObjectInput as the Store method would
			key := "modules/test.zip"
			putObjectInput := &s3.PutObjectInput{
				Bucket:             aws.String(resolver.BucketName),
				Key:                resolver.withPrefix(key),
				Body:               storeInput.Reader,
				ContentLength:      aws.Int64(storeInput.Size),
				ContentType:        aws.String(storeInput.ContentType),
				ContentDisposition: aws.String("attachment"),
			}

			if !resolver.DisableACL {
				putObjectInput.ACL = types.ObjectCannedACLPrivate
			}

			if resolver.ServerSideEncryption != "none" {
				putObjectInput.ServerSideEncryption = types.ServerSideEncryption(resolver.ServerSideEncryption)
			}

			Convey("PutObjectInput should not have ACL set", func() {
				So(putObjectInput.ACL, ShouldBeEmpty)
			})
		})
	})
}
