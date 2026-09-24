package handlers

import (
	"testing"

	"terralist/pkg/auth/jwt"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDownloadTokens(t *testing.T) {
	Convey("Subject: Capability tokens for artifact download links", t, func() {
		tokens, err := NewDownloadTokens("signing-secret")
		So(err, ShouldBeNil)

		subject := "providers/hashicorp/random/3.6.2/linux_amd64"

		Convey("When a token is issued and verified for the same subject", func() {
			token, err := tokens.Sign(subject, true)
			So(err, ShouldBeNil)

			fetch, ok := tokens.Verify(token, subject)

			Convey("Then it should be accepted with its fetch permission", func() {
				So(ok, ShouldBeTrue)
				So(fetch, ShouldBeTrue)
			})
		})

		Convey("When a read-only token is verified", func() {
			token, _ := tokens.Sign(subject, false)
			fetch, ok := tokens.Verify(token, subject)

			Convey("Then it should be accepted without the fetch permission", func() {
				So(ok, ShouldBeTrue)
				So(fetch, ShouldBeFalse)
			})
		})

		Convey("When a token is verified for another subject", func() {
			token, _ := tokens.Sign(subject, true)
			_, ok := tokens.Verify(token, "providers/hashicorp/random/3.6.2/darwin_arm64")

			Convey("Then it should be rejected", func() {
				So(ok, ShouldBeFalse)
			})
		})

		Convey("When a token signed with another secret is verified", func() {
			others, _ := NewDownloadTokens("other-secret")
			token, _ := others.Sign(subject, true)
			_, ok := tokens.Verify(token, subject)

			Convey("Then it should be rejected", func() {
				So(ok, ShouldBeFalse)
			})
		})

		Convey("When garbage is verified", func() {
			_, ok := tokens.Verify("garbage", subject)

			Convey("Then it should be rejected", func() {
				So(ok, ShouldBeFalse)
			})
		})

		Convey("When a user token signed with the same secret is verified", func() {
			userJWT, _ := jwt.New("signing-secret")
			token, _ := userJWT.Build(map[string]any{"name": "alice"}, 60)
			_, ok := tokens.Verify(token, subject)

			Convey("Then it should be rejected", func() {
				So(ok, ShouldBeFalse)
			})
		})
	})
}
