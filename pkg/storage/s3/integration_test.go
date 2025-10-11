package s3

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestS3UseACLsIntegration(t *testing.T) {
	Convey("Subject: S3 UseACLs configuration integration", t, func() {
		creator := &Creator{}

		Convey("When creating S3 resolver with UseACLs true", func() {
			config := &Config{
				BucketName:      "test-bucket",
				BucketRegion:    "us-east-1",
				LinkExpire:      15,
				UseACLs:         true,
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			}

			err := config.Validate()
			So(err, ShouldBeNil)

			resolver, err := creator.New(config)
			So(err, ShouldBeNil)
			So(resolver, ShouldNotBeNil)

			s3Resolver, ok := resolver.(*Resolver)
			So(ok, ShouldBeTrue)

			Convey("Should have UseACLs set to true", func() {
				So(s3Resolver.UseACLs, ShouldBeTrue)
			})
		})

		Convey("When creating S3 resolver with UseACLs false (default)", func() {
			config := &Config{
				BucketName:      "test-bucket",
				BucketRegion:    "us-east-1",
				LinkExpire:      15,
				UseACLs:         false,
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			}

			err := config.Validate()
			So(err, ShouldBeNil)

			resolver, err := creator.New(config)
			So(err, ShouldBeNil)
			So(resolver, ShouldNotBeNil)

			s3Resolver, ok := resolver.(*Resolver)
			So(ok, ShouldBeTrue)

			Convey("Should have UseACLs set to false", func() {
				So(s3Resolver.UseACLs, ShouldBeFalse)
			})
		})
	})
}

func TestS3ConfigValidation(t *testing.T) {
	Convey("Subject: S3 Config validation with UseACLs", t, func() {
		Convey("When config has UseACLs set", func() {
			config := &Config{
				BucketName:      "test-bucket",
				BucketRegion:    "us-east-1",
				LinkExpire:      15,
				UseACLs:         true,
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
				UseACLs:    true,
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
