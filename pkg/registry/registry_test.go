package registry

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	. "github.com/smartystreets/goconvey/convey"
)

// otherPublicKey returns the armored public key of a freshly generated
// identity that never signed anything.
func otherPublicKey(t *testing.T) string {
	t.Helper()

	entity, err := openpgp.NewEntity("Someone Else", "", "else@example.com", nil)
	if err != nil {
		t.Fatalf("could not generate key: %v", err)
	}

	var armored bytes.Buffer
	encoder, err := armor.Encode(&armored, openpgp.PublicKeyType, nil)
	if err != nil {
		t.Fatalf("could not armor key: %v", err)
	}
	if err := entity.Serialize(encoder); err != nil {
		t.Fatalf("could not serialize key: %v", err)
	}
	if err := encoder.Close(); err != nil {
		t.Fatalf("could not close armor: %v", err)
	}

	return armored.String()
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("could not read fixture %s: %v", name, err)
	}

	return content
}

// newRegistry starts a server replaying the recorded registry.terraform.io
// payloads and records the requests it receives.
func newRegistry(t *testing.T) (*httptest.Server, *[]*http.Request) {
	t.Helper()

	var requests []*http.Request

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/terraform.json", func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r)
		_, _ = w.Write(fixture(t, "discovery.json"))
	})
	mux.HandleFunc("/v1/providers/hashicorp/null/versions", func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r)
		_, _ = w.Write(fixture(t, "versions.json"))
	})
	mux.HandleFunc("/v1/providers/hashicorp/null/3.2.4/download/linux/amd64", func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r)
		_, _ = w.Write(fixture(t, "download.json"))
	})
	mux.HandleFunc("/v1/providers/hashicorp/broken/versions", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r)
		w.WriteHeader(http.StatusNotFound)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server, &requests
}

func TestVersions(t *testing.T) {
	Convey("Subject: Listing the versions of an upstream provider", t, func() {
		server, requests := newRegistry(t)
		client := New(server.URL)

		Convey("When the versions of an existing provider are requested", func() {
			versions, err := client.Versions(context.Background(), "hashicorp", "null")

			Convey("Then the versions should be returned with their protocols and platforms", func() {
				So(err, ShouldBeNil)
				So(len(versions), ShouldEqual, 2)
				So(versions[0].Version, ShouldEqual, "3.2.4")
				So(versions[0].Protocols, ShouldResemble, []string{"5.0"})
				So(len(versions[0].Platforms), ShouldEqual, 11)
				So(versions[0].Platforms, ShouldContain, Platform{OS: "linux", Arch: "amd64"})
			})

			Convey("Then service discovery should have been performed first", func() {
				So((*requests)[0].URL.Path, ShouldEqual, "/.well-known/terraform.json")
				So((*requests)[1].URL.Path, ShouldEqual, "/v1/providers/hashicorp/null/versions")
			})
		})

		Convey("When the versions are requested twice", func() {
			_, _ = client.Versions(context.Background(), "hashicorp", "null")
			_, _ = client.Versions(context.Background(), "hashicorp", "null")

			Convey("Then service discovery should be performed once", func() {
				var discoveries int
				for _, r := range *requests {
					if r.URL.Path == "/.well-known/terraform.json" {
						discoveries++
					}
				}
				So(discoveries, ShouldEqual, 1)
			})
		})

		Convey("When the versions of an unknown provider are requested", func() {
			_, err := client.Versions(context.Background(), "hashicorp", "missing")

			Convey("Then it should report the provider as not found", func() {
				So(errors.Is(err, ErrNotFound), ShouldBeTrue)
			})
		})

		Convey("When the registry answers with an invalid document", func() {
			_, err := client.Versions(context.Background(), "hashicorp", "broken")

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(errors.Is(err, ErrNotFound), ShouldBeFalse)
			})
		})
	})
}

func TestVersionsWithToken(t *testing.T) {
	Convey("Subject: Authenticating against an upstream registry", t, func() {
		server, requests := newRegistry(t)
		client := New(server.URL, WithToken("secret"))

		Convey("When the versions are requested", func() {
			_, err := client.Versions(context.Background(), "hashicorp", "null")

			Convey("Then every request should carry the bearer token", func() {
				So(err, ShouldBeNil)
				for _, r := range *requests {
					So(r.Header.Get("Authorization"), ShouldEqual, "Bearer secret")
				}
			})
		})
	})
}

func TestDiscoveryFailures(t *testing.T) {
	Convey("Subject: Service discovery failures", t, func() {
		Convey("Given a registry without a providers service", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"modules.v1":"/v1/modules/"}`))
			}))
			defer server.Close()

			Convey("When the versions are requested", func() {
				_, err := New(server.URL).Versions(context.Background(), "hashicorp", "null")

				Convey("Then an error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given a registry whose discovery document is unreachable", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			Convey("When the versions are requested", func() {
				_, err := New(server.URL).Versions(context.Background(), "hashicorp", "null")

				Convey("Then an error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given a registry announcing an absolute providers URL", func() {
			server, requests := newRegistry(t)
			discovery := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"providers.v1":"` + server.URL + `/v1/providers/"}`))
			}))
			defer discovery.Close()

			Convey("When the versions are requested", func() {
				versions, err := New(discovery.URL).Versions(context.Background(), "hashicorp", "null")

				Convey("Then the providers URL should be followed", func() {
					So(err, ShouldBeNil)
					So(len(versions), ShouldEqual, 2)
					So((*requests)[0].URL.Path, ShouldEqual, "/v1/providers/hashicorp/null/versions")
				})
			})
		})
	})
}

func TestDownload(t *testing.T) {
	Convey("Subject: Fetching the download metadata of an upstream package", t, func() {
		server, _ := newRegistry(t)
		client := New(server.URL)

		Convey("When the metadata of an existing package is requested", func() {
			download, err := client.Download(context.Background(), "hashicorp", "null", "3.2.4", "linux", "amd64")

			Convey("Then the metadata should be returned", func() {
				So(err, ShouldBeNil)
				So(download.OS, ShouldEqual, "linux")
				So(download.Arch, ShouldEqual, "amd64")
				So(download.Filename, ShouldEqual, "terraform-provider-null_3.2.4_linux_amd64.zip")
				So(download.ShaSum, ShouldEqual, "9d32ac3619cfc93eb3c4f423492a8e0f79db05fec58e449dee9b2d5873d5f69f")
				So(download.DownloadURL, ShouldStartWith, "https://")
				So(download.ShaSumsURL, ShouldEndWith, "terraform-provider-null_3.2.4_SHA256SUMS")
				So(download.ShaSumsSignatureURL, ShouldContainSubstring, "terraform-provider-null_3.2.4_SHA256SUMS.")
				So(download.ShaSumsSignatureURL, ShouldEndWith, ".sig")
				So(download.Protocols, ShouldResemble, []string{"5.0"})
				So(len(download.SigningKeys.GPGPublicKeys), ShouldEqual, 1)
				So(download.SigningKeys.GPGPublicKeys[0].KeyID, ShouldEqual, "34365D9472D7468F")
				So(download.SigningKeys.GPGPublicKeys[0].ASCIIArmor, ShouldStartWith, "-----BEGIN PGP PUBLIC KEY BLOCK-----")
			})
		})

		Convey("When the metadata of an unknown package is requested", func() {
			_, err := client.Download(context.Background(), "hashicorp", "null", "3.2.4", "plan9", "mips")

			Convey("Then it should report the package as not found", func() {
				So(errors.Is(err, ErrNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestParseShaSums(t *testing.T) {
	Convey("Subject: Parsing a SHA256SUMS file", t, func() {
		Convey("When the recorded SHA256SUMS file is parsed", func() {
			sums, err := ParseShaSums(bytes.NewReader(fixture(t, "terraform-provider-null_3.2.4_SHA256SUMS")))

			Convey("Then every package should map to its digest", func() {
				So(err, ShouldBeNil)
				So(len(sums), ShouldEqual, 12)
				So(sums["terraform-provider-null_3.2.4_linux_amd64.zip"], ShouldEqual, "9d32ac3619cfc93eb3c4f423492a8e0f79db05fec58e449dee9b2d5873d5f69f")
			})
		})

		Convey("When a line has a binary marker before the file name", func() {
			sums, err := ParseShaSums(strings.NewReader("abcd *terraform-provider-null_3.2.4_linux_amd64.zip\n"))

			Convey("Then the marker should be stripped", func() {
				So(err, ShouldBeNil)
				So(sums, ShouldResemble, map[string]string{"terraform-provider-null_3.2.4_linux_amd64.zip": "abcd"})
			})
		})

		Convey("When a line is malformed", func() {
			_, err := ParseShaSums(strings.NewReader("this is not a sums line\n"))

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestVerifyShaSums(t *testing.T) {
	Convey("Subject: Verifying the signature of a SHA256SUMS file", t, func() {
		document := fixture(t, "terraform-provider-null_3.2.4_SHA256SUMS")
		signature := fixture(t, "terraform-provider-null_3.2.4_SHA256SUMS.sig")
		hashicorp := GPGPublicKey{KeyID: "34365D9472D7468F", ASCIIArmor: string(fixture(t, "hashicorp.asc"))}
		other := GPGPublicKey{KeyID: "0000000000000000", ASCIIArmor: otherPublicKey(t)}

		Convey("When the recorded signature is checked against the HashiCorp key", func() {
			keyID, err := VerifyShaSums(document, signature, []GPGPublicKey{hashicorp})

			Convey("Then it should verify and report the signing key", func() {
				So(err, ShouldBeNil)
				So(keyID, ShouldEqual, "34365D9472D7468F")
			})
		})

		Convey("When the signing key is not the first one offered", func() {
			keyID, err := VerifyShaSums(document, signature, []GPGPublicKey{other, hashicorp})

			Convey("Then it should still verify", func() {
				So(err, ShouldBeNil)
				So(keyID, ShouldEqual, "34365D9472D7468F")
			})
		})

		Convey("When the signature is armored", func() {
			var armored bytes.Buffer
			encoder, err := armor.Encode(&armored, "PGP SIGNATURE", nil)
			So(err, ShouldBeNil)
			_, err = encoder.Write(signature)
			So(err, ShouldBeNil)
			So(encoder.Close(), ShouldBeNil)

			_, err = VerifyShaSums(document, armored.Bytes(), []GPGPublicKey{hashicorp})

			Convey("Then it should verify", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When the document was tampered with", func() {
			tampered := bytes.Replace(document, []byte("9d32ac36"), []byte("00000000"), 1)
			_, err := VerifyShaSums(tampered, signature, []GPGPublicKey{hashicorp})

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
				So(errors.Is(err, ErrUnknownIssuer), ShouldBeFalse)
			})
		})

		Convey("When none of the keys issued the signature", func() {
			_, err := VerifyShaSums(document, signature, []GPGPublicKey{other})

			Convey("Then it should report an unknown issuer", func() {
				So(errors.Is(err, ErrUnknownIssuer), ShouldBeTrue)
			})
		})

		Convey("When no key is offered", func() {
			_, err := VerifyShaSums(document, signature, nil)

			Convey("Then it should report an unknown issuer", func() {
				So(errors.Is(err, ErrUnknownIssuer), ShouldBeTrue)
			})
		})

		Convey("When a key cannot be decoded", func() {
			_, err := VerifyShaSums(document, signature, []GPGPublicKey{{KeyID: "bad", ASCIIArmor: "not a key"}})

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When the signature is garbage", func() {
			_, err := VerifyShaSums(document, []byte("garbage"), []GPGPublicKey{hashicorp})

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
