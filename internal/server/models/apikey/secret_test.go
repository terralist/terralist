package apikey

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewSecret(t *testing.T) {
	Convey("Subject: Generating API key secrets", t, func() {
		first, err := NewSecret()
		So(err, ShouldBeNil)

		second, err := NewSecret()
		So(err, ShouldBeNil)

		Convey("Then secrets carry the key prefix", func() {
			So(strings.HasPrefix(first, SecretPrefix), ShouldBeTrue)
		})

		Convey("Then secrets are long enough to hold 32 random bytes", func() {
			So(len(first), ShouldBeGreaterThanOrEqualTo, len(SecretPrefix)+43)
		})

		Convey("Then consecutive secrets differ", func() {
			So(first, ShouldNotEqual, second)
		})
	})
}

func TestHashSecret(t *testing.T) {
	Convey("Subject: Hashing API key secrets", t, func() {
		sum := sha256.Sum256([]byte("tlk_example"))

		Convey("Then the hash is the hex encoded SHA-256 of the secret", func() {
			So(HashSecret("tlk_example"), ShouldEqual, hex.EncodeToString(sum[:]))
		})

		Convey("Then hashing is deterministic", func() {
			So(HashSecret("tlk_example"), ShouldEqual, HashSecret("tlk_example"))
		})
	})
}
