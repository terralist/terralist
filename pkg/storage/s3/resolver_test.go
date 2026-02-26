package s3

import (
	"bytes"
	"fmt"
	"testing"

	"terralist/pkg/storage"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	. "github.com/smartystreets/goconvey/convey"
	mock "github.com/stretchr/testify/mock"
)

func TestStore(t *testing.T) {
	Convey("Subject: Store files in S3", t, func() {
		client := NewMockS3Client(t)
		presigner := NewMockPresignClient(t)

		storeInput := &storage.StoreInput{
			KeyPrefix:   "test",
			FileName:    "test.txt",
			Reader:      bytes.NewReader([]byte("test content")),
			Size:        int64(12),
			ContentType: "text/plain",
		}

		expectedKey := "test/test.txt"
		expectedPrefixedKey := "test-prefix/" + expectedKey

		Convey("When ACL is enabled", func() {
			resolver := &Resolver{
				BucketName:           "test-bucket",
				BucketPrefix:         "test-prefix/",
				ServerSideEncryption: "AES256",
				UseACLs:              true,
				Client:               client,
				Presigner:            presigner,
			}

			Convey("Should succeed and set ACL and SSE", func() {
				client.
					On("PutObject", mock.Anything, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
						So(*input.Bucket, ShouldEqual, "test-bucket")
						So(*input.Key, ShouldEqual, expectedPrefixedKey)
						So(input.ACL, ShouldEqual, types.ObjectCannedACLPrivate)
						So(input.ServerSideEncryption, ShouldEqual, types.ServerSideEncryptionAes256)
						return true
					})).
					Return(&s3.PutObjectOutput{}, nil).
					Once()

				key, err := resolver.Store(storeInput)
				So(err, ShouldBeNil)
				So(key, ShouldEqual, expectedKey)
				client.AssertExpectations(t)
			})

			Convey("Should fail and return error from S3", func() {
				client.
					On("PutObject", mock.Anything, mock.AnythingOfType("*s3.PutObjectInput")).
					Return(nil, fmt.Errorf("s3 error")).
					Once()

				key, err := resolver.Store(storeInput)
				So(err, ShouldNotBeNil)
				So(key, ShouldBeEmpty)
			})
		})

		Convey("When ACL is disabled (default behavior - bucket policy mode)", func() {
			resolver := &Resolver{
				BucketName:           "test-bucket",
				BucketPrefix:         "test-prefix/",
				ServerSideEncryption: "AES256",
				UseACLs:              false,
				Client:               client,
				Presigner:            presigner,
			}

			Convey("Should succeed and not set ACL, but set SSE", func() {
				client.
					On("PutObject", mock.Anything, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
						So(*input.Bucket, ShouldEqual, "test-bucket")
						So(*input.Key, ShouldEqual, expectedPrefixedKey)
						So(input.ACL, ShouldBeEmpty)
						So(input.ServerSideEncryption, ShouldEqual, types.ServerSideEncryptionAes256)
						return true
					})).
					Return(&s3.PutObjectOutput{}, nil).
					Once()

				key, err := resolver.Store(storeInput)
				So(err, ShouldBeNil)
				So(key, ShouldEqual, expectedKey)
				client.AssertExpectations(t)
			})

			Convey("Should fail and return error from S3", func() {
				client.
					On("PutObject", mock.Anything, mock.AnythingOfType("*s3.PutObjectInput")).
					Return(nil, fmt.Errorf("s3 error")).
					Once()

				key, err := resolver.Store(storeInput)
				So(err, ShouldNotBeNil)
				So(key, ShouldBeEmpty)
				client.AssertExpectations(t)
			})
		})

		Convey("When ServerSideEncryption is 'none'", func() {
			resolver := &Resolver{
				BucketName:           "test-bucket",
				BucketPrefix:         "test-prefix/",
				ServerSideEncryption: "none",
				UseACLs:              false,
				Client:               client,
				Presigner:            presigner,
			}

			Convey("Should not set ServerSideEncryption", func() {
				client.
					On("PutObject", mock.Anything, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
						So(input.ServerSideEncryption, ShouldBeEmpty)
						return true
					})).
					Return(&s3.PutObjectOutput{}, nil).
					Once()

				_, err := resolver.Store(storeInput)
				So(err, ShouldBeNil)

				client.AssertExpectations(t)
			})
		})
	})
}

func TestFind(t *testing.T) {
	Convey("Subject: Find files in S3 (presigned URL)", t, func() {
		client := NewMockS3Client(t)
		presigner := NewMockPresignClient(t)

		expectedKey := "test/test.txt"
		expectedPrefixedKey := "test-prefix/" + expectedKey
		expectedURL := "https://example.com/s3/test-prefix/test/test.txt"

		resolver := &Resolver{
			BucketName:   "test-bucket",
			BucketPrefix: "test-prefix/",
			LinkExpire:   15,
			Client:       client,
			Presigner:    presigner,
		}

		Convey("Should succeed and return presigned URL", func() {
			presigner.
				On("PresignGetObject", mock.Anything, mock.MatchedBy(func(input *s3.GetObjectInput) bool {
					So(*input.Bucket, ShouldEqual, "test-bucket")
					So(*input.Key, ShouldEqual, expectedPrefixedKey)
					return true
				}), mock.Anything).
				Return(&v4.PresignedHTTPRequest{URL: expectedURL}, nil).
				Once()

			url, err := resolver.Find(expectedKey)
			So(err, ShouldBeNil)
			So(url, ShouldEqual, expectedURL)
			presigner.AssertExpectations(t)
		})

		Convey("Should fail and return error from presigner", func() {
			presigner.
				On("PresignGetObject", mock.Anything, mock.AnythingOfType("*s3.GetObjectInput"), mock.Anything).
				Return(nil, fmt.Errorf("presign error")).
				Once()

			url, err := resolver.Find(expectedKey)
			So(err, ShouldNotBeNil)
			So(url, ShouldBeEmpty)
			presigner.AssertExpectations(t)
		})
	})
}

func TestPurge(t *testing.T) {
	Convey("Subject: Purge files from S3", t, func() {
		client := NewMockS3Client(t)
		presigner := NewMockPresignClient(t)

		expectedKey := "test/test.txt"
		expectedPrefixedKey := "test-prefix/" + expectedKey

		resolver := &Resolver{
			BucketName:   "test-bucket",
			BucketPrefix: "test-prefix/",
			Client:       client,
			Presigner:    presigner,
		}

		Convey("Should succeed and delete the object", func() {
			client.
				On("DeleteObject", mock.Anything, mock.MatchedBy(func(input *s3.DeleteObjectInput) bool {
					So(*input.Bucket, ShouldEqual, "test-bucket")
					So(*input.Key, ShouldEqual, expectedPrefixedKey)
					return true
				})).
				Return(&s3.DeleteObjectOutput{}, nil).
				Once()

			err := resolver.Purge(expectedKey)
			So(err, ShouldBeNil)
			client.AssertExpectations(t)
		})

		Convey("Should fail and return error from S3", func() {
			client.
				On("DeleteObject", mock.Anything, mock.AnythingOfType("*s3.DeleteObjectInput")).
				Return(nil, fmt.Errorf("delete error")).
				Once()

			err := resolver.Purge(expectedKey)
			So(err, ShouldNotBeNil)
			client.AssertExpectations(t)
		})
	})
}
