package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

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
				raw, err := base64.StdEncoding.DecodeString(payload.String())
				So(err, ShouldBeNil)

				var tampered Request
				So(json.Unmarshal(raw[sha256.Size:], &tampered), ShouldBeNil)
				tampered.RedirectURI = "https://evil.example.com/steal"

				data, err := json.Marshal(tampered)
				So(err, ShouldBeNil)

				forged := Payload(base64.StdEncoding.EncodeToString(append(raw[:sha256.Size], data...)))
				_, err = forged.ToRequest(key)

				Convey("Then it is rejected", func() {
					So(errors.Is(err, ErrInvalidPayload), ShouldBeTrue)
				})
			})
		})

		Convey("Given a payload with no signature", func() {
			data, err := json.Marshal(request)
			So(err, ShouldBeNil)

			unsigned := Payload(base64.StdEncoding.EncodeToString(data))
			_, err = unsigned.ToRequest(key)

			Convey("Then it is rejected", func() {
				So(errors.Is(err, ErrInvalidPayload), ShouldBeTrue)
			})
		})

		Convey("Given a payload shorter than a signature", func() {
			_, err := Payload("QUFBQQ==").ToRequest(key)

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
