package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"

	"terralist/pkg/auth/saml"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRequest_PayloadRoundTrip(t *testing.T) {
	Convey("Subject: Signing and verifying the OAuth state payload", t, func() {
		key := []byte("test-state-key")
		request := Request{
			ClientID:            "terraform-cli",
			CodeChallenge:       "challenge",
			CodeChallengeMethod: "S256",
			RedirectURI:         "http://localhost:10000/login",
			ResponseType:        "code",
			State:               "client-state",
		}

		Convey("Given a payload produced with a key", func() {
			payload, err := request.ToPayload(key)
			So(err, ShouldBeNil)

			Convey("When it is verified with the same key", func() {
				parsed, err := payload.ToRequest(key)

				Convey("Then the original request is returned", func() {
					So(err, ShouldBeNil)
					So(parsed, ShouldResemble, request)
				})
			})

			Convey("When it is verified with a different key", func() {
				_, err := payload.ToRequest([]byte("another-key"))

				Convey("Then it is rejected", func() {
					So(errors.Is(err, ErrInvalidPayload), ShouldBeTrue)
				})
			})

			Convey("When its content is altered", func() {
				raw, err := base64.RawURLEncoding.DecodeString(payload.String())
				So(err, ShouldBeNil)

				var tampered Request
				So(json.Unmarshal(raw[sha256.Size:], &tampered), ShouldBeNil)
				tampered.RedirectURI = "https://evil.example.com/steal"

				data, err := json.Marshal(tampered)
				So(err, ShouldBeNil)

				forged := Payload(base64.RawURLEncoding.EncodeToString(append(raw[:sha256.Size], data...)))
				_, err = forged.ToRequest(key)

				Convey("Then it is rejected", func() {
					So(errors.Is(err, ErrInvalidPayload), ShouldBeTrue)
				})
			})
		})

		Convey("Given a payload with no signature", func() {
			data, err := json.Marshal(request)
			So(err, ShouldBeNil)

			unsigned := Payload(base64.RawURLEncoding.EncodeToString(data))
			_, err = unsigned.ToRequest(key)

			Convey("Then it is rejected", func() {
				So(errors.Is(err, ErrInvalidPayload), ShouldBeTrue)
			})
		})

		Convey("Given a payload shorter than a signature", func() {
			_, err := Payload("QUFBQQ").ToRequest(key)

			Convey("Then it is rejected without panicking", func() {
				So(errors.Is(err, ErrInvalidPayload), ShouldBeTrue)
			})
		})

		Convey("Given an empty payload", func() {
			_, err := Payload("").ToRequest(key)

			Convey("Then it is rejected", func() {
				So(errors.Is(err, ErrInvalidPayload), ShouldBeTrue)
			})
		})

		Convey("Given a payload that is not base64", func() {
			_, err := Payload("not base64!").ToRequest(key)

			Convey("Then it is rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

// The state payload travels through the identity provider and back: the
// GitHub, GitLab and Bitbucket providers (and crewjam/saml for RelayState)
// place it unescaped into a query string, and the SAML flow additionally
// validates it as RelayState. A payload that does not survive this round trip
// fails the callback with HTTP 400.
func TestRequest_PayloadSurvivesIdentityProviderRoundTrip(t *testing.T) {
	Convey("Subject: Transporting the OAuth state payload through an identity provider", t, func() {
		key := []byte("test-state-key")

		// The HMAC makes the encoded output pseudo-random, so varying the
		// client state yields payloads that cover the whole base64 alphabet,
		// including the two characters that need escaping in standard base64.
		var requests []Request
		var payloads []Payload
		for i := 0; i < 64; i++ {
			request := Request{
				ClientID:            "terraform-cli",
				CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
				CodeChallengeMethod: "S256",
				RedirectURI:         "http://localhost:10000/login",
				ResponseType:        "code",
				State:               strings.Repeat("s", i),
			}

			payload, err := request.ToPayload(key)
			So(err, ShouldBeNil)

			requests = append(requests, request)
			payloads = append(payloads, payload)
		}

		Convey("Given payloads that use every character of the encoding alphabet", func() {
			alphabet := map[rune]struct{}{}
			for _, payload := range payloads {
				for _, char := range payload.String() {
					alphabet[char] = struct{}{}
				}
			}
			// >= rather than ==: padded encodings add '=', and the assertions below
			// should report the actual corruption instead of this guard.
			So(len(alphabet), ShouldBeGreaterThanOrEqualTo, 64)

			Convey("When they are placed unescaped into a query string and parsed back", func() {
				Convey("Then they arrive intact and still verify", func() {
					for i, payload := range payloads {
						query, err := url.ParseQuery("state=" + payload.String())
						So(err, ShouldBeNil)
						So(query.Get("state"), ShouldEqual, payload.String())

						parsed, err := Payload(query.Get("state")).ToRequest(key)
						So(err, ShouldBeNil)
						So(parsed, ShouldResemble, requests[i])
					}
				})
			})

			Convey("When they are used as SAML RelayState", func() {
				Convey("Then they pass RelayState validation", func() {
					for _, payload := range payloads {
						So(saml.ValidateRelayState(payload.String()), ShouldBeNil)
					}
				})
			})
		})
	})
}
