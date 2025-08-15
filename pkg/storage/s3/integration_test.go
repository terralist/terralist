package s3

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestS3DisableACLIntegration(t *testing.T) {
	Convey("Subject: S3 DisableACL configuration integration", t, func() {
		creator := &Creator{}

		Convey("When creating S3 resolver with DisableACL true", func() {
			config := &Config{
				BucketName:      "test-bucket",
				BucketRegion:    "us-east-1",
				LinkExpire:      15,
				DisableACL:      true,
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			}

			err := config.Validate()
			So(err, ShouldBeNil)

			resolver, err := creator.New(config)
			So(err, ShouldBeNil)
			So(resolver, ShouldNotBeNil)

			s3Resolver := resolver.(*Resolver)

			Convey("Should have DisableACL set to true", func() {
				So(s3Resolver.DisableACL, ShouldBeTrue)
			})
		})

		Convey("When creating S3 resolver with DisableACL false (default)", func() {
			config := &Config{
				BucketName:      "test-bucket",
				BucketRegion:    "us-east-1",
				LinkExpire:      15,
				DisableACL:      false,
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			}

			err := config.Validate()
			So(err, ShouldBeNil)

			resolver, err := creator.New(config)
			So(err, ShouldBeNil)
			So(resolver, ShouldNotBeNil)

			s3Resolver := resolver.(*Resolver)

			Convey("Should have DisableACL set to false", func() {
				So(s3Resolver.DisableACL, ShouldBeFalse)
			})
		})
	})
}

func TestS3ConfigValidation(t *testing.T) {
	Convey("Subject: S3 Config validation with DisableACL", t, func() {
		Convey("When config has DisableACL set", func() {
			config := &Config{
				BucketName:      "test-bucket",
				BucketRegion:    "us-east-1", 
				LinkExpire:      15,
				DisableACL:      true,
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			}

			Convey("Should pass validation", func() {
				err := config.Validate()
				So(err, ShouldBeNil)
			})
		})

		Convey("When config has missing required fields", func() {
			config := &Config{
				DisableACL: true,
				LinkExpire: 15,
			}

			Convey("Should fail validation due to missing BucketName", func() {
				err := config.Validate()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "BucketName")
			})
		})
	})
}