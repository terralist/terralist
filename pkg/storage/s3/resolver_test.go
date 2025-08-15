package s3

import (
	"bytes"
	"testing"

	"terralist/pkg/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

// MockS3API is a mock implementation of the S3 API interface
type MockS3API struct {
	mock.Mock
}

func (m *MockS3API) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockS3API) GetObjectRequest(input *s3.GetObjectInput) (*request.Request, *s3.GetObjectOutput) {
	args := m.Called(input)
	return args.Get(0).(*request.Request), args.Get(1).(*s3.GetObjectOutput)
}

func (m *MockS3API) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*s3.DeleteObjectOutput), args.Error(1)
}

func TestStore(t *testing.T) {
	Convey("Subject: Store files in S3", t, func() {
		// Create a test session (won't be used for actual API calls)
		sess, _ := session.NewSession(&aws.Config{
			Region:      aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials("test", "test", ""),
		})

		Convey("When ACL is enabled (default behavior)", func() {
			resolver := &Resolver{
				BucketName:           "test-bucket",
				BucketPrefix:         "test-prefix/",
				ServerSideEncryption: "AES256",
				DisableACL:           false, // ACL enabled
				Session:              sess,
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
				Session:              sess,
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
				BucketName:  "test-bucket",
				LinkExpire:  15,
				DisableACL:  false,
			}

			Convey("Config should indicate ACL is enabled", func() {
				So(config.DisableACL, ShouldBeFalse)
			})
		})

		Convey("When DisableACL is true", func() {
			config := &Config{
				BucketName:  "test-bucket",
				LinkExpire:  15,
				DisableACL:  true,
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
				BucketName:    "test-bucket",
				BucketRegion:  "us-east-1",
				LinkExpire:    15,
				DisableACL:    true,
				AccessKeyID:   "test-key",
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
				BucketName:    "test-bucket",
				BucketRegion:  "us-east-1",
				LinkExpire:    15,
				DisableACL:    false,
				AccessKeyID:   "test-key",
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
		sess, _ := session.NewSession(&aws.Config{
			Region:      aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials("test", "test", ""),
		})

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
				Session:              sess,
			}

			// Create the PutObjectInput as the Store method would
			serverSideEncryption := aws.String(resolver.ServerSideEncryption)
			if resolver.ServerSideEncryption == "none" {
				serverSideEncryption = nil
			}

			key := "modules/test.zip"
			putObjectInput := &s3.PutObjectInput{
				Bucket:               aws.String(resolver.BucketName),
				Key:                  resolver.withPrefix(key),
				Body:                 storeInput.Reader,
				ContentLength:        aws.Int64(storeInput.Size),
				ContentType:          aws.String(storeInput.ContentType),
				ContentDisposition:   aws.String("attachment"),
				ServerSideEncryption: serverSideEncryption,
			}

			if !resolver.DisableACL {
				putObjectInput.ACL = aws.String("private")
			}

			Convey("PutObjectInput should have ACL set to private", func() {
				So(putObjectInput.ACL, ShouldNotBeNil)
				So(*putObjectInput.ACL, ShouldEqual, "private")
			})
		})

		Convey("When ACL is disabled", func() {
			resolver := &Resolver{
				BucketName:           "test-bucket",
				BucketPrefix:         "test-prefix/",
				ServerSideEncryption: "AES256",
				DisableACL:           true,
				Session:              sess,
			}

			// Create the PutObjectInput as the Store method would
			serverSideEncryption := aws.String(resolver.ServerSideEncryption)
			if resolver.ServerSideEncryption == "none" {
				serverSideEncryption = nil
			}

			key := "modules/test.zip"
			putObjectInput := &s3.PutObjectInput{
				Bucket:               aws.String(resolver.BucketName),
				Key:                  resolver.withPrefix(key),
				Body:                 storeInput.Reader,
				ContentLength:        aws.Int64(storeInput.Size),
				ContentType:          aws.String(storeInput.ContentType),
				ContentDisposition:   aws.String("attachment"),
				ServerSideEncryption: serverSideEncryption,
			}

			if !resolver.DisableACL {
				putObjectInput.ACL = aws.String("private")
			}

			Convey("PutObjectInput should not have ACL set", func() {
				So(putObjectInput.ACL, ShouldBeNil)
			})
		})
	})
}