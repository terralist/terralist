package oauth

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateRedirectURI(t *testing.T) {
	host := "registry.example.com"

	testCases := []struct {
		title       string
		redirectURI string
		allowed     bool
	}{
		{
			title:       "own host over https",
			redirectURI: "https://registry.example.com",
			allowed:     true,
		},
		{
			title:       "own host with a path",
			redirectURI: "https://registry.example.com/dashboard",
			allowed:     true,
		},
		{
			title:       "own host in a different case",
			redirectURI: "https://Registry.Example.com/",
			allowed:     true,
		},
		{
			title:       "localhost on the lowest terraform port",
			redirectURI: "http://localhost:10000/login",
			allowed:     true,
		},
		{
			title:       "ipv4 loopback on the highest terraform port",
			redirectURI: "http://127.0.0.1:10010/login",
			allowed:     true,
		},
		{
			title:       "ipv6 loopback on a terraform port",
			redirectURI: "http://[::1]:10005/login",
			allowed:     true,
		},
		{
			title:       "localhost below the terraform port range",
			redirectURI: "http://localhost:9999/login",
			allowed:     false,
		},
		{
			title:       "localhost above the terraform port range",
			redirectURI: "http://localhost:10011/login",
			allowed:     false,
		},
		{
			title:       "localhost without a port",
			redirectURI: "http://localhost/login",
			allowed:     false,
		},
		{
			title:       "foreign host",
			redirectURI: "https://evil.example.com/steal",
			allowed:     false,
		},
		{
			title:       "own host as a subdomain of a foreign host",
			redirectURI: "https://registry.example.com.evil.example.com/",
			allowed:     false,
		},
		{
			title:       "own host as a prefix of a foreign host",
			redirectURI: "https://registry.example.community/",
			allowed:     false,
		},
		{
			title:       "loopback address in the userinfo of a foreign host",
			redirectURI: "http://localhost:10000@evil.example.com/",
			allowed:     false,
		},
		{
			title:       "userinfo on an allowed host",
			redirectURI: "http://user@localhost:10000/login",
			allowed:     false,
		},
		{
			title:       "javascript scheme",
			redirectURI: "javascript:alert(1)",
			allowed:     false,
		},
		{
			title:       "non http scheme on own host",
			redirectURI: "ftp://registry.example.com/",
			allowed:     false,
		},
		{
			title:       "scheme relative url",
			redirectURI: "//evil.example.com/",
			allowed:     false,
		},
		{
			title:       "relative path",
			redirectURI: "/dashboard",
			allowed:     false,
		},
		{
			title:       "empty",
			redirectURI: "",
			allowed:     false,
		},
		{
			title:       "unparsable",
			redirectURI: "http://[::1:10000/",
			allowed:     false,
		},
	}

	Convey("Subject: Validating the OAuth redirect URI", t, func() {
		for _, test := range testCases {
			Convey("Given "+test.title, func() {
				err := ValidateRedirectURI(test.redirectURI, host)

				if test.allowed {
					Convey("Then it is allowed", func() {
						So(err, ShouldBeNil)
					})
				} else {
					Convey("Then it is rejected", func() {
						So(errors.Is(err, ErrRedirectURINotAllowed), ShouldBeTrue)
					})
				}
			})
		}
	})
}
