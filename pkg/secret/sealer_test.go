package secret

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSealer(t *testing.T) {
	Convey("Subject: Sealing secrets at rest", t, func() {
		sealer := NewSealer("upstream-secret")

		Convey("When a value is sealed and opened with the same secret", func() {
			sealed, err := sealer.Seal("ghp_token")
			So(err, ShouldBeNil)

			opened, err := sealer.Open(sealed)

			Convey("Then the original value should come back", func() {
				So(err, ShouldBeNil)
				So(opened, ShouldEqual, "ghp_token")
			})

			Convey("Then the sealed value should not contain the plaintext", func() {
				So(sealed, ShouldNotContainSubstring, "ghp_token")
			})
		})

		Convey("When the same value is sealed twice", func() {
			first, _ := sealer.Seal("ghp_token")
			second, _ := sealer.Seal("ghp_token")

			Convey("Then the sealed values should differ", func() {
				So(first, ShouldNotEqual, second)
			})
		})

		Convey("When a sealed value is opened with another secret", func() {
			sealed, _ := sealer.Seal("ghp_token")
			_, err := NewSealer("other-secret").Open(sealed)

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When a tampered value is opened", func() {
			sealed, _ := sealer.Seal("ghp_token")
			tampered := sealed[:len(sealed)-2] + "AA"
			_, err := sealer.Open(tampered)

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When garbage is opened", func() {
			_, err := sealer.Open("not base64!")

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
