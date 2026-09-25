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
	"time"

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
	mux.HandleFunc("/v1/modules/hashicorp/subnets/cidr/versions", func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r)
		_, _ = w.Write(fixture(t, "module_versions.json"))
	})
	mux.HandleFunc("/v1/modules/hashicorp/subnets/cidr/1.0.0/download", func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r)
		w.Header().Set("X-Terraform-Get", "git::https://github.com/hashicorp/terraform-cidr-subnets?ref=52ca061aaea2e8f58c91ac03ca1fae45e44c28bf")
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/v1/modules/hashicorp/subnets/cidr/1.1.0/download", func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r)
		w.Header().Set("X-Terraform-Get", "../../../../../../archives/subnets-1.1.0.tar.gz")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/modules/hashicorp/subnets/cidr/1.2.0/download", func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r)
		w.WriteHeader(http.StatusNoContent)
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
			versions, err := client.ProviderVersions(context.Background(), "hashicorp", "null")

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
			_, _ = client.ProviderVersions(context.Background(), "hashicorp", "null")
			_, _ = client.ProviderVersions(context.Background(), "hashicorp", "null")

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
			_, err := client.ProviderVersions(context.Background(), "hashicorp", "missing")

			Convey("Then it should report the provider as not found", func() {
				So(errors.Is(err, ErrNotFound), ShouldBeTrue)
			})
		})

		Convey("When the registry answers with an invalid document", func() {
			_, err := client.ProviderVersions(context.Background(), "hashicorp", "broken")

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
			_, err := client.ProviderVersions(context.Background(), "hashicorp", "null")

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
				_, err := New(server.URL).ProviderVersions(context.Background(), "hashicorp", "null")

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
				_, err := New(server.URL).ProviderVersions(context.Background(), "hashicorp", "null")

				Convey("Then an error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given a registry without a discovery document", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			Convey("When the versions are requested", func() {
				_, err := New(server.URL).ProviderVersions(context.Background(), "hashicorp", "null")

				Convey("Then an error other than not found should be returned", func() {
					So(err, ShouldNotBeNil)
					So(errors.Is(err, ErrNotFound), ShouldBeFalse)
				})
			})
		})

		Convey("Given a registry announcing services that are not URLs", func() {
			server, _ := newRegistry(t)
			discovery := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"login.v1":{"client":"terraform-cli","grant_types":["authz_code"]},"providers.v1":"` + server.URL + `/v1/providers/"}`))
			}))
			defer discovery.Close()

			Convey("When the versions are requested", func() {
				versions, err := New(discovery.URL).ProviderVersions(context.Background(), "hashicorp", "null")

				Convey("Then the providers service should still be used", func() {
					So(err, ShouldBeNil)
					So(len(versions), ShouldEqual, 2)
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
				versions, err := New(discovery.URL).ProviderVersions(context.Background(), "hashicorp", "null")

				Convey("Then the providers URL should be followed", func() {
					So(err, ShouldBeNil)
					So(len(versions), ShouldEqual, 2)
					So((*requests)[0].URL.Path, ShouldEqual, "/v1/providers/hashicorp/null/versions")
				})
			})
		})
	})
}

func TestConcurrentDiscovery(t *testing.T) {
	Convey("Subject: Service discovery under concurrent requests", t, func() {
		// The discovery document is only served once two requests for it are
		// in flight together, so a client serializing discovery fails.
		arrived := make(chan struct{}, 2)
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/terraform.json", func(w http.ResponseWriter, _ *http.Request) {
			arrived <- struct{}{}
			deadline := time.After(2 * time.Second)
			for len(arrived) < 2 {
				select {
				case <-deadline:
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				case <-time.After(10 * time.Millisecond):
				}
			}
			_, _ = w.Write([]byte(`{"providers.v1":"/v1/providers/"}`))
		})
		mux.HandleFunc("/v1/providers/hashicorp/null/versions", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(fixture(t, "versions.json"))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := New(server.URL)

		Convey("When two requests discover the services at the same time", func() {
			errs := make(chan error, 2)
			for range 2 {
				go func() {
					_, err := client.ProviderVersions(context.Background(), "hashicorp", "null")
					errs <- err
				}()
			}

			Convey("Then neither should wait for the other", func() {
				So(<-errs, ShouldBeNil)
				So(<-errs, ShouldBeNil)
			})
		})
	})
}

func TestOversizedResponse(t *testing.T) {
	Convey("Subject: Reading a response larger than a registry document may be", t, func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/terraform.json", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"providers.v1":"/v1/providers/"}`))
		})
		mux.HandleFunc("/v1/providers/hashicorp/null/versions", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"versions":[` + strings.Repeat(" ", int(maxResponseSize)) + `]}`))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		Convey("When the versions are requested", func() {
			_, err := New(server.URL).ProviderVersions(context.Background(), "hashicorp", "null")

			Convey("Then the response should be refused as too large", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "exceeds")
			})
		})
	})
}

func TestDownload(t *testing.T) {
	Convey("Subject: Fetching the download metadata of an upstream package", t, func() {
		server, _ := newRegistry(t)
		client := New(server.URL)

		Convey("When the metadata of an existing package is requested", func() {
			download, err := client.ProviderDownload(context.Background(), "hashicorp", "null", "3.2.4", "linux", "amd64")

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
			_, err := client.ProviderDownload(context.Background(), "hashicorp", "null", "3.2.4", "plan9", "mips")

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

func TestVerifyArmoredShaSums(t *testing.T) {
	Convey("Subject: Verifying an armored signature of a SHA256SUMS file", t, func() {
		entity, err := openpgp.NewEntity("Signer", "", "signer@example.com", nil)
		So(err, ShouldBeNil)

		var key bytes.Buffer
		encoder, err := armor.Encode(&key, openpgp.PublicKeyType, nil)
		So(err, ShouldBeNil)
		So(entity.Serialize(encoder), ShouldBeNil)
		So(encoder.Close(), ShouldBeNil)

		document := []byte("0000  terraform-provider-null_1.0.0_linux_amd64.zip\n")
		var signature bytes.Buffer
		So(openpgp.ArmoredDetachSign(&signature, entity, bytes.NewReader(document), nil), ShouldBeNil)

		Convey("When the armored signature is checked against the signing key", func() {
			keyID, err := SignatureVerifier{}.VerifyShaSums(document, signature.Bytes(), []GPGPublicKey{{KeyID: entity.PrimaryKey.KeyIdString(), ASCIIArmor: key.String()}})

			Convey("Then it should be accepted", func() {
				So(err, ShouldBeNil)
				So(keyID, ShouldEqual, entity.PrimaryKey.KeyIdString())
			})
		})
	})
}

func TestVerifyShaSums(t *testing.T) {
	Convey("Subject: Verifying the signature of a SHA256SUMS file", t, func() {
		document := fixture(t, "terraform-provider-null_3.2.4_SHA256SUMS")
		signature := fixture(t, "terraform-provider-null_3.2.4_SHA256SUMS.sig")
		// The key HashiCorp advertises for this release expired on 2026-04-19.
		hashicorp := GPGPublicKey{KeyID: "34365D9472D7468F", ASCIIArmor: string(fixture(t, "hashicorp.asc"))}
		other := GPGPublicKey{KeyID: "0000000000000000", ASCIIArmor: otherPublicKey(t)}
		lenient := SignatureVerifier{AcceptExpiredKeys: true}
		strict := SignatureVerifier{}

		Convey("Given a verifier accepting expired keys", func() {
			Convey("When the recorded signature is checked against the HashiCorp key", func() {
				keyID, err := lenient.VerifyShaSums(document, signature, []GPGPublicKey{hashicorp})

				Convey("Then it should verify and report the signing key", func() {
					So(err, ShouldBeNil)
					So(keyID, ShouldEqual, "34365D9472D7468F")
				})
			})

			Convey("When the signing key is not the first one offered", func() {
				keyID, err := lenient.VerifyShaSums(document, signature, []GPGPublicKey{other, hashicorp})

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

				_, err = lenient.VerifyShaSums(document, armored.Bytes(), []GPGPublicKey{hashicorp})

				Convey("Then it should verify", func() {
					So(err, ShouldBeNil)
				})
			})

			Convey("When the document was tampered with", func() {
				tampered := bytes.Replace(document, []byte("9d32ac36"), []byte("00000000"), 1)
				_, err := lenient.VerifyShaSums(tampered, signature, []GPGPublicKey{hashicorp})

				Convey("Then it should be rejected as an invalid signature", func() {
					So(err, ShouldNotBeNil)
					So(errors.Is(err, ErrUnknownIssuer), ShouldBeFalse)
					So(errors.Is(err, ErrKeyExpired), ShouldBeFalse)
					So(err.Error(), ShouldContainSubstring, "invalid signature")
				})
			})

			Convey("When none of the keys issued the signature", func() {
				_, err := lenient.VerifyShaSums(document, signature, []GPGPublicKey{other})

				Convey("Then it should report an unknown issuer", func() {
					So(errors.Is(err, ErrUnknownIssuer), ShouldBeTrue)
				})
			})

			Convey("When no key is offered", func() {
				_, err := lenient.VerifyShaSums(document, signature, nil)

				Convey("Then it should report an unknown issuer", func() {
					So(errors.Is(err, ErrUnknownIssuer), ShouldBeTrue)
				})
			})

			Convey("When a key cannot be decoded", func() {
				_, err := lenient.VerifyShaSums(document, signature, []GPGPublicKey{{KeyID: "bad", ASCIIArmor: "not a key"}})

				Convey("Then an error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})

			Convey("When the signature is garbage", func() {
				_, err := lenient.VerifyShaSums(document, []byte("garbage"), []GPGPublicKey{hashicorp})

				Convey("Then an error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given a verifier rejecting expired keys", func() {
			Convey("When the recorded signature is checked against the expired HashiCorp key", func() {
				_, err := strict.VerifyShaSums(document, signature, []GPGPublicKey{hashicorp})

				Convey("Then it should report the expired key", func() {
					So(errors.Is(err, ErrKeyExpired), ShouldBeTrue)
					So(err.Error(), ShouldContainSubstring, "34365D9472D7468F")
				})
			})

			Convey("When the document was tampered with", func() {
				tampered := bytes.Replace(document, []byte("9d32ac36"), []byte("00000000"), 1)
				_, err := strict.VerifyShaSums(tampered, signature, []GPGPublicKey{hashicorp})

				Convey("Then it should be rejected for the signature, not the expiry", func() {
					So(err, ShouldNotBeNil)
					So(errors.Is(err, ErrKeyExpired), ShouldBeFalse)
				})
			})
		})
	})
}

func TestModuleVersions(t *testing.T) {
	Convey("Subject: Listing the versions of an upstream module", t, func() {
		server, requests := newRegistry(t)
		client := New(server.URL)

		Convey("When the versions of an existing module are requested", func() {
			versions, err := client.ModuleVersions(context.Background(), "hashicorp", "subnets", "cidr")

			Convey("Then the versions should be returned", func() {
				So(err, ShouldBeNil)
				So(versions, ShouldResemble, []ModuleVersion{{Version: "1.0.0"}})
			})

			Convey("Then the modules service should have been discovered and queried", func() {
				So((*requests)[0].URL.Path, ShouldEqual, "/.well-known/terraform.json")
				So((*requests)[1].URL.Path, ShouldEqual, "/v1/modules/hashicorp/subnets/cidr/versions")
			})
		})

		Convey("When the versions of an unknown module are requested", func() {
			_, err := client.ModuleVersions(context.Background(), "hashicorp", "missing", "cidr")

			Convey("Then it should report the module as not found", func() {
				So(errors.Is(err, ErrNotFound), ShouldBeTrue)
			})
		})

		Convey("When providers and modules are both requested", func() {
			_, _ = client.ProviderVersions(context.Background(), "hashicorp", "null")
			_, _ = client.ModuleVersions(context.Background(), "hashicorp", "subnets", "cidr")

			Convey("Then service discovery should still be performed once", func() {
				var discoveries int
				for _, r := range *requests {
					if r.URL.Path == "/.well-known/terraform.json" {
						discoveries++
					}
				}
				So(discoveries, ShouldEqual, 1)
			})
		})
	})
}

func TestModuleLocation(t *testing.T) {
	Convey("Subject: Resolving the source location of an upstream module version", t, func() {
		server, _ := newRegistry(t)
		client := New(server.URL)

		Convey("When the registry answers with a go-getter location", func() {
			location, err := client.ModuleLocation(context.Background(), "hashicorp", "subnets", "cidr", "1.0.0")

			Convey("Then the location should be returned as is", func() {
				So(err, ShouldBeNil)
				So(location, ShouldEqual, "git::https://github.com/hashicorp/terraform-cidr-subnets?ref=52ca061aaea2e8f58c91ac03ca1fae45e44c28bf")
			})
		})

		Convey("When the registry answers with a relative location", func() {
			location, err := client.ModuleLocation(context.Background(), "hashicorp", "subnets", "cidr", "1.1.0")

			Convey("Then the location should be resolved against the download URL", func() {
				So(err, ShouldBeNil)
				So(location, ShouldEqual, server.URL+"/archives/subnets-1.1.0.tar.gz")
			})
		})

		Convey("When the registry answers without a location", func() {
			_, err := client.ModuleLocation(context.Background(), "hashicorp", "subnets", "cidr", "1.2.0")

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(errors.Is(err, ErrNotFound), ShouldBeFalse)
			})
		})

		Convey("When the version does not exist", func() {
			_, err := client.ModuleLocation(context.Background(), "hashicorp", "subnets", "cidr", "9.9.9")

			Convey("Then it should report the version as not found", func() {
				So(errors.Is(err, ErrNotFound), ShouldBeTrue)
			})
		})
	})
}
