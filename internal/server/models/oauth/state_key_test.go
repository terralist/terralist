package oauth

import (
	"crypto/sha256"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeriveStateKey(t *testing.T) {
	Convey("Subject: Deriving the state signing key from the token signing secret", t, func() {
		Convey("Given the same secret", func() {
			first := DeriveStateKey("token-signing-secret")
			second := DeriveStateKey("token-signing-secret")

			Convey("Then the derived key is stable and never equals the secret itself", func() {
				So(first, ShouldResemble, second)
				So(len(first), ShouldEqual, sha256.Size)
				So(string(first), ShouldNotEqual, "token-signing-secret")
			})
		})

		Convey("Given different secrets", func() {
			first := DeriveStateKey("first-secret")
			second := DeriveStateKey("second-secret")

			Convey("Then the derived keys differ", func() {
				So(first, ShouldNotResemble, second)
			})
		})
	})
}
