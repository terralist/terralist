package handlers

import (
	"testing"

	"terralist/internal/server/models/provider"
	"terralist/pkg/auth/jwt"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPackageTokens(t *testing.T) {
	Convey("Subject: Capability tokens for package downloads", t, func() {
		tokens, err := NewPackageTokens("signing-secret")
		So(err, ShouldBeNil)

		pkg := provider.Package{Name: "random", Version: "3.6.2", System: "linux", Architecture: "amd64"}

		Convey("When a token is issued and verified for the same package", func() {
			token, err := tokens.Sign("hashicorp", pkg, true)
			So(err, ShouldBeNil)

			fetch, ok := tokens.Verify(token, "hashicorp", pkg)

			Convey("Then it should be accepted with its fetch permission", func() {
				So(ok, ShouldBeTrue)
				So(fetch, ShouldBeTrue)
			})
		})

		Convey("When a read-only token is verified", func() {
			token, _ := tokens.Sign("hashicorp", pkg, false)
			fetch, ok := tokens.Verify(token, "hashicorp", pkg)

			Convey("Then it should be accepted without the fetch permission", func() {
				So(ok, ShouldBeTrue)
				So(fetch, ShouldBeFalse)
			})
		})

		Convey("When a token is verified for another package", func() {
			token, _ := tokens.Sign("hashicorp", pkg, true)
			other := pkg
			other.Architecture = "arm64"
			_, ok := tokens.Verify(token, "hashicorp", other)

			Convey("Then it should be rejected", func() {
				So(ok, ShouldBeFalse)
			})
		})

		Convey("When a token is verified for another authority", func() {
			token, _ := tokens.Sign("hashicorp", pkg, true)
			_, ok := tokens.Verify(token, "other", pkg)

			Convey("Then it should be rejected", func() {
				So(ok, ShouldBeFalse)
			})
		})

		Convey("When a token signed with another secret is verified", func() {
			others, _ := NewPackageTokens("other-secret")
			token, _ := others.Sign("hashicorp", pkg, true)
			_, ok := tokens.Verify(token, "hashicorp", pkg)

			Convey("Then it should be rejected", func() {
				So(ok, ShouldBeFalse)
			})
		})

		Convey("When garbage is verified", func() {
			_, ok := tokens.Verify("garbage", "hashicorp", pkg)

			Convey("Then it should be rejected", func() {
				So(ok, ShouldBeFalse)
			})
		})

		Convey("When a user token signed with the same secret is verified", func() {
			userJWT, _ := jwt.New("signing-secret")
			token, _ := userJWT.Build(map[string]any{"name": "alice"}, 60)
			_, ok := tokens.Verify(token, "hashicorp", pkg)

			Convey("Then it should be rejected", func() {
				So(ok, ShouldBeFalse)
			})
		})
	})
}
