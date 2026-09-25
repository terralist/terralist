package secret

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const testSecret = "upstream-secret-of-at-least-32-chars"

func TestNewSealer(t *testing.T) {
	Convey("Subject: Deriving the sealing key from the server secret", t, func() {
		Convey("When the secret is shorter than 32 characters", func() {
			_, err := NewSealer(strings.Repeat("a", 31))

			Convey("Then it should be refused", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When the secret is 32 characters long", func() {
			sealer, err := NewSealer(strings.Repeat("a", 32))

			Convey("Then a sealer should be returned", func() {
				So(err, ShouldBeNil)
				So(sealer, ShouldNotBeNil)
			})
		})
	})
}

func TestSealer(t *testing.T) {
	Convey("Subject: Sealing secrets at rest", t, func() {
		sealer, err := NewSealer(testSecret)
		So(err, ShouldBeNil)

		Convey("When a value is sealed and opened with the same secret and scope", func() {
			sealed, err := sealer.Seal("ghp_token", "authority-1")
			So(err, ShouldBeNil)

			opened, err := sealer.Open(sealed, "authority-1")

			Convey("Then the original value should come back", func() {
				So(err, ShouldBeNil)
				So(opened, ShouldEqual, "ghp_token")
			})

			Convey("Then the sealed value should not contain the plaintext", func() {
				So(sealed, ShouldNotContainSubstring, "ghp_token")
			})
		})

		Convey("When a sealed value is opened under another scope", func() {
			sealed, _ := sealer.Seal("ghp_token", "authority-1")
			_, err := sealer.Open(sealed, "authority-2")

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When the same value is sealed twice", func() {
			first, _ := sealer.Seal("ghp_token", "authority-1")
			second, _ := sealer.Seal("ghp_token", "authority-1")

			Convey("Then the sealed values should differ", func() {
				So(first, ShouldNotEqual, second)
			})
		})

		Convey("When a sealed value is opened with another secret", func() {
			sealed, _ := sealer.Seal("ghp_token", "authority-1")
			other, err := NewSealer("another-upstream-secret-of-32-chars")
			So(err, ShouldBeNil)
			_, err = other.Open(sealed, "authority-1")

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When a tampered value is opened", func() {
			sealed, _ := sealer.Seal("ghp_token", "authority-1")
			tampered := sealed[:len(sealed)-2] + "AA"
			_, err := sealer.Open(tampered, "authority-1")

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When garbage is opened", func() {
			_, err := sealer.Open("not base64!", "authority-1")

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
