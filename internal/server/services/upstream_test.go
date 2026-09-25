package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"terralist/internal/server/models/authority"
	"terralist/pkg/cache/memory"
	"terralist/pkg/registry"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

// fakeUpstream is a provider registry serving one provider, hashicorp/null,
// with versions 1.0.0 and 1.1.0 for linux_amd64 and darwin_arm64, signed with
// a key generated for the test.
type fakeUpstream struct {
	server    *httptest.Server
	key       registry.GPGPublicKey
	shaSums   map[string][]byte // version → SHA256SUMS document
	signature map[string][]byte // version → detached signature
	packages  map[string][]byte // file name → archive bytes
	requests  atomic.Int32
	failing   atomic.Bool
	tokens    []string
}

func newFakeUpstream(t *testing.T) *fakeUpstream {
	t.Helper()

	entity, err := openpgp.NewEntity("Upstream", "", "upstream@example.com", nil)
	if err != nil {
		t.Fatalf("could not generate key: %v", err)
	}

	var armored bytes.Buffer
	encoder, _ := armor.Encode(&armored, openpgp.PublicKeyType, nil)
	_ = entity.Serialize(encoder)
	_ = encoder.Close()

	u := &fakeUpstream{
		key:       registry.GPGPublicKey{KeyID: entity.PrimaryKey.KeyIdString(), ASCIIArmor: armored.String()},
		shaSums:   map[string][]byte{},
		signature: map[string][]byte{},
		packages:  map[string][]byte{},
	}

	for _, version := range []string{"1.0.0", "1.1.0"} {
		var sums strings.Builder
		for _, platform := range []string{"linux_amd64", "darwin_arm64"} {
			name := fmt.Sprintf("terraform-provider-null_%s_%s.zip", version, platform)
			content := []byte("package " + name)
			u.packages[name] = content
			digest := sha256.Sum256(content)
			fmt.Fprintf(&sums, "%s  %s\n", hex.EncodeToString(digest[:]), name)
		}

		u.shaSums[version] = []byte(sums.String())

		var sig bytes.Buffer
		if err := openpgp.DetachSign(&sig, entity, strings.NewReader(sums.String()), nil); err != nil {
			t.Fatalf("could not sign: %v", err)
		}
		u.signature[version] = sig.Bytes()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/terraform.json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"providers.v1":"/v1/providers/","modules.v1":"/v1/modules/"}`))
	})
	mux.HandleFunc("/v1/modules/hashicorp/dir/template/versions", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"modules":[{"source":"hashicorp/dir/template","versions":[{"version":"1.0.1"},{"version":"1.0.2"}]}]}`))
	})
	mux.HandleFunc("/v1/modules/hashicorp/dir/template/1.0.2/download", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Terraform-Get", "git::https://github.com/hashicorp/terraform-template-dir?ref=v1.0.2")
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/v1/providers/hashicorp/null/versions", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"versions": []map[string]any{
				{"version": "1.0.0", "protocols": []string{"5.0"}, "platforms": []map[string]string{{"os": "linux", "arch": "amd64"}, {"os": "darwin", "arch": "arm64"}}},
				{"version": "1.1.0", "protocols": []string{"5.0", "6.0"}, "platforms": []map[string]string{{"os": "linux", "arch": "amd64"}, {"os": "darwin", "arch": "arm64"}}},
			},
		})
	})
	mux.HandleFunc("/v1/providers/hashicorp/null/", func(w http.ResponseWriter, r *http.Request) {
		// /v1/providers/hashicorp/null/<version>/download/<os>/<arch>
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/providers/hashicorp/null/"), "/")
		if len(parts) != 4 || parts[1] != "download" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		version, os, arch := parts[0], parts[2], parts[3]
		name := fmt.Sprintf("terraform-provider-null_%s_%s_%s.zip", version, os, arch)
		content, ok := u.packages[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		digest := sha256.Sum256(content)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"protocols":             []string{"5.0"},
			"os":                    os,
			"arch":                  arch,
			"filename":              name,
			"download_url":          u.server.URL + "/files/" + name,
			"shasums_url":           u.server.URL + "/files/SHA256SUMS-" + version,
			"shasums_signature_url": u.server.URL + "/files/SHA256SUMS-" + version + ".sig",
			"shasum":                hex.EncodeToString(digest[:]),
			"signing_keys":          map[string]any{"gpg_public_keys": []registry.GPGPublicKey{u.key}},
		})
	})
	mux.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/files/")
		switch {
		case strings.HasSuffix(name, ".sig"):
			_, _ = w.Write(u.signature[strings.TrimSuffix(strings.TrimPrefix(name, "SHA256SUMS-"), ".sig")])
		case strings.HasPrefix(name, "SHA256SUMS-"):
			_, _ = w.Write(u.shaSums[strings.TrimPrefix(name, "SHA256SUMS-")])
		default:
			_, _ = w.Write(u.packages[name])
		}
	})

	u.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u.requests.Add(1)
		if token := r.Header.Get("Authorization"); token != "" {
			u.tokens = append(u.tokens, token)
		}
		if u.failing.Load() {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(u.server.Close)

	return u
}

func (u *fakeUpstream) authority() *authority.Authority {
	hostname := "registry.terraform.io"
	namespace := "hashicorp"
	url := u.server.URL

	return &authority.Authority{
		Name:                  "hashicorp",
		UpstreamHostname:      &hostname,
		UpstreamNamespace:     &namespace,
		UpstreamURL:           &url,
		UpstreamEnabled:       true,
		UpstreamDefaultPolicy: authority.PolicyAllow,
	}
}

func newUpstreamService(t *testing.T, now *time.Time) *DefaultUpstreamService {
	t.Helper()

	c, err := (&memory.Creator{}).New(&memory.Config{SweepInterval: time.Hour})
	if err != nil {
		t.Fatalf("could not create cache: %v", err)
	}

	return &DefaultUpstreamService{
		Cache:      c,
		TTL:        5 * time.Minute,
		Retention:  24 * time.Hour,
		HTTPClient: http.DefaultClient,
		Verifier:   registry.SignatureVerifier{AcceptExpiredKeys: true},
		now:        func() time.Time { return *now },
	}
}

func TestUpstreamProviderVersions(t *testing.T) {
	Convey("Subject: Listing the versions an authority may serve from its upstream", t, func() {
		upstream := newFakeUpstream(t)
		now := time.Now()
		service := newUpstreamService(t, &now)
		a := upstream.authority()
		a.ID = uuid.New()

		Convey("When the versions are requested", func() {
			versions, err := service.ProviderVersions(a, "null")

			Convey("Then every upstream version should be returned with its platforms", func() {
				So(err, ShouldBeNil)
				So(len(versions), ShouldEqual, 2)
				So(versions[1].Version, ShouldEqual, "1.1.0")
				So(versions[1].Protocols, ShouldResemble, []string{"5.0", "6.0"})
				So(versions[1].Platforms, ShouldContain, registry.Platform{OS: "darwin", Arch: "arm64"})
			})
		})

		Convey("When a deny rule matches a version", func() {
			a.Rules = []authority.Rule{{Kind: authority.RuleKindProvider, Name: "null", Version: "1.1.*", Effect: authority.EffectDeny}}
			versions, err := service.ProviderVersions(a, "null")

			Convey("Then that version should be left out", func() {
				So(err, ShouldBeNil)
				So(len(versions), ShouldEqual, 1)
				So(versions[0].Version, ShouldEqual, "1.0.0")
			})
		})

		Convey("When the upstream is disabled", func() {
			a.UpstreamEnabled = false
			versions, err := service.ProviderVersions(a, "null")

			Convey("Then nothing should be returned and the upstream not called", func() {
				So(err, ShouldBeNil)
				So(versions, ShouldBeEmpty)
				So(upstream.requests.Load(), ShouldEqual, 0)
			})
		})

		Convey("When the versions are requested twice within the TTL", func() {
			_, _ = service.ProviderVersions(a, "null")
			requests := upstream.requests.Load()
			_, err := service.ProviderVersions(a, "null")

			Convey("Then the second answer should come from the cache", func() {
				So(err, ShouldBeNil)
				So(upstream.requests.Load(), ShouldEqual, requests)
			})
		})

		Convey("When the TTL passed and the upstream fails", func() {
			_, _ = service.ProviderVersions(a, "null")
			now = now.Add(10 * time.Minute)
			upstream.failing.Store(true)
			versions, err := service.ProviderVersions(a, "null")

			Convey("Then the stale answer should be served", func() {
				So(err, ShouldBeNil)
				So(len(versions), ShouldEqual, 2)
			})
		})

		Convey("When the TTL passed and the upstream answers", func() {
			_, _ = service.ProviderVersions(a, "null")
			now = now.Add(10 * time.Minute)
			requests := upstream.requests.Load()
			_, err := service.ProviderVersions(a, "null")

			Convey("Then the upstream should be queried again", func() {
				So(err, ShouldBeNil)
				So(upstream.requests.Load(), ShouldBeGreaterThan, requests)
			})
		})

		Convey("When the upstream fails and nothing is cached", func() {
			upstream.failing.Store(true)
			_, err := service.ProviderVersions(a, "null")

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When the provider is unknown upstream", func() {
			versions, err := service.ProviderVersions(a, "missing")

			Convey("Then no versions should be returned without an error", func() {
				So(err, ShouldBeNil)
				So(versions, ShouldBeEmpty)
			})
		})
	})
}

func TestUpstreamProviderVersion(t *testing.T) {
	Convey("Subject: Reading the verified metadata of an upstream version", t, func() {
		upstream := newFakeUpstream(t)
		now := time.Now()
		service := newUpstreamService(t, &now)
		a := upstream.authority()
		a.ID = uuid.New()

		Convey("When the metadata of an allowed version is requested", func() {
			metadata, err := service.ProviderVersion(a, "null", "1.0.0")

			Convey("Then the digests, documents and keys should be returned", func() {
				So(err, ShouldBeNil)
				So(metadata.Protocols, ShouldResemble, []string{"5.0"})
				So(len(metadata.ShaSums), ShouldEqual, 2)
				So(metadata.ShaSums["terraform-provider-null_1.0.0_linux_amd64.zip"], ShouldNotBeEmpty)
				So(metadata.ShaSumsDocument, ShouldResemble, upstream.shaSums["1.0.0"])
				So(metadata.Signature, ShouldResemble, upstream.signature["1.0.0"])
				So(metadata.SigningKeys[0].KeyID, ShouldEqual, upstream.key.KeyID)
			})
		})

		Convey("When the metadata is requested twice", func() {
			_, _ = service.ProviderVersion(a, "null", "1.0.0")
			requests := upstream.requests.Load()
			_, err := service.ProviderVersion(a, "null", "1.0.0")

			Convey("Then the second answer should come from the cache", func() {
				So(err, ShouldBeNil)
				So(upstream.requests.Load(), ShouldEqual, requests)
			})
		})

		Convey("When a deny rule matches the version", func() {
			a.Rules = []authority.Rule{{Kind: authority.RuleKindProvider, Name: "*", Version: "1.0.0", Effect: authority.EffectDeny}}
			_, err := service.ProviderVersion(a, "null", "1.0.0")

			Convey("Then it should be denied", func() {
				So(errors.Is(err, ErrUpstreamDenied), ShouldBeTrue)
			})
		})

		Convey("When the SHA256SUMS signature does not verify", func() {
			upstream.signature["1.0.0"] = upstream.signature["1.1.0"]
			_, err := service.ProviderVersion(a, "null", "1.0.0")

			Convey("Then an error should be returned and nothing cached", func() {
				So(err, ShouldNotBeNil)
				upstream.failing.Store(true)
				_, err = service.ProviderVersion(a, "null", "1.0.0")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When the version is unknown upstream", func() {
			_, err := service.ProviderVersion(a, "null", "9.9.9")

			Convey("Then it should be reported as not found", func() {
				So(errors.Is(err, registry.ErrNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestUpstreamProviderPackage(t *testing.T) {
	Convey("Subject: Locating an upstream package", t, func() {
		upstream := newFakeUpstream(t)
		now := time.Now()
		service := newUpstreamService(t, &now)
		a := upstream.authority()
		a.ID = uuid.New()

		Convey("When an allowed package is located", func() {
			pkg, err := service.ProviderPackage(a, "null", "1.0.0", "linux", "amd64")

			Convey("Then its download URL and verified digest should be returned", func() {
				So(err, ShouldBeNil)
				So(pkg.URL, ShouldEqual, upstream.server.URL+"/files/terraform-provider-null_1.0.0_linux_amd64.zip")
				So(pkg.FileName, ShouldEqual, "terraform-provider-null_1.0.0_linux_amd64.zip")
				digest := sha256.Sum256(upstream.packages[pkg.FileName])
				So(pkg.ShaSum, ShouldEqual, hex.EncodeToString(digest[:]))
			})
		})

		Convey("When the advertised digest does not match the SHA256SUMS file", func() {
			upstream.packages["terraform-provider-null_1.0.0_linux_amd64.zip"] = []byte("tampered")
			_, err := service.ProviderPackage(a, "null", "1.0.0", "linux", "amd64")

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When the platform is unknown upstream", func() {
			_, err := service.ProviderPackage(a, "null", "1.0.0", "plan9", "mips")

			Convey("Then it should be reported as not found", func() {
				So(errors.Is(err, registry.ErrNotFound), ShouldBeTrue)
			})
		})

		Convey("When a deny rule matches the version", func() {
			a.Rules = []authority.Rule{{Kind: authority.RuleKindProvider, Name: "null", Version: "*", Effect: authority.EffectDeny}}
			_, err := service.ProviderPackage(a, "null", "1.0.0", "linux", "amd64")

			Convey("Then it should be denied", func() {
				So(errors.Is(err, ErrUpstreamDenied), ShouldBeTrue)
			})
		})
	})
}

func TestUpstreamToken(t *testing.T) {
	Convey("Subject: Authenticating against a private upstream", t, func() {
		upstream := newFakeUpstream(t)
		now := time.Now()
		service := newUpstreamService(t, &now)
		a := upstream.authority()
		a.ID = uuid.New()

		Convey("Given a sealed token and the matching sealer", func() {
			sealer := newTestSealer()
			sealed, _ := sealer.Seal("ghp_upstream")
			a.UpstreamToken = &sealed
			service.Sealer = sealer

			_, err := service.ProviderVersions(a, "null")

			Convey("Then every upstream request should carry the token", func() {
				So(err, ShouldBeNil)
				So(upstream.tokens, ShouldNotBeEmpty)
				for _, token := range upstream.tokens {
					So(token, ShouldEqual, "Bearer ghp_upstream")
				}
			})
		})

		Convey("Given a sealed token without a sealer", func() {
			sealed := "c2VhbGVk"
			a.UpstreamToken = &sealed

			_, err := service.ProviderVersions(a, "null")

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestUpstreamModuleVersions(t *testing.T) {
	Convey("Subject: Listing the module versions an authority may serve from its upstream", t, func() {
		upstream := newFakeUpstream(t)
		now := time.Now()
		service := newUpstreamService(t, &now)
		a := upstream.authority()
		a.ID = uuid.New()

		Convey("When the versions are requested", func() {
			versions, err := service.ModuleVersions(a, "dir", "template")

			Convey("Then every upstream version should be returned", func() {
				So(err, ShouldBeNil)
				So(versions, ShouldResemble, []string{"1.0.1", "1.0.2"})
			})
		})

		Convey("When a module rule denies a version", func() {
			a.Rules = []authority.Rule{{Kind: authority.RuleKindModule, Name: "dir/template", Version: "1.0.1", Effect: authority.EffectDeny}}
			versions, err := service.ModuleVersions(a, "dir", "template")

			Convey("Then that version should be left out", func() {
				So(err, ShouldBeNil)
				So(versions, ShouldResemble, []string{"1.0.2"})
			})
		})

		Convey("When a provider rule denies everything", func() {
			a.Rules = []authority.Rule{{Kind: authority.RuleKindProvider, Name: "*", Version: "*", Effect: authority.EffectDeny}}
			versions, err := service.ModuleVersions(a, "dir", "template")

			Convey("Then modules are not affected", func() {
				So(err, ShouldBeNil)
				So(len(versions), ShouldEqual, 2)
			})
		})

		Convey("When the versions are requested twice", func() {
			_, _ = service.ModuleVersions(a, "dir", "template")
			requests := upstream.requests.Load()
			_, err := service.ModuleVersions(a, "dir", "template")

			Convey("Then the second answer should come from the cache", func() {
				So(err, ShouldBeNil)
				So(upstream.requests.Load(), ShouldEqual, requests)
			})
		})

		Convey("When the module is unknown upstream", func() {
			versions, err := service.ModuleVersions(a, "missing", "template")

			Convey("Then no versions should be returned without an error", func() {
				So(err, ShouldBeNil)
				So(versions, ShouldBeEmpty)
			})
		})

		Convey("When the upstream is disabled", func() {
			a.UpstreamEnabled = false
			versions, err := service.ModuleVersions(a, "dir", "template")

			Convey("Then nothing should be returned", func() {
				So(err, ShouldBeNil)
				So(versions, ShouldBeEmpty)
			})
		})
	})
}

func TestUpstreamModuleLocation(t *testing.T) {
	Convey("Subject: Locating an upstream module version", t, func() {
		upstream := newFakeUpstream(t)
		now := time.Now()
		service := newUpstreamService(t, &now)
		a := upstream.authority()
		a.ID = uuid.New()

		Convey("When an allowed version is located", func() {
			location, err := service.ModuleLocation(a, "dir", "template", "1.0.2")

			Convey("Then its go-getter source should be returned", func() {
				So(err, ShouldBeNil)
				So(location, ShouldEqual, "git::https://github.com/hashicorp/terraform-template-dir?ref=v1.0.2")
			})
		})

		Convey("When a rule denies the version", func() {
			a.Rules = []authority.Rule{{Kind: authority.RuleKindModule, Name: "dir/*", Version: "*", Effect: authority.EffectDeny}}
			_, err := service.ModuleLocation(a, "dir", "template", "1.0.2")

			Convey("Then it should be denied", func() {
				So(errors.Is(err, ErrUpstreamDenied), ShouldBeTrue)
			})
		})

		Convey("When the version is unknown upstream", func() {
			_, err := service.ModuleLocation(a, "dir", "template", "9.9.9")

			Convey("Then it should be reported as not found", func() {
				So(errors.Is(err, registry.ErrNotFound), ShouldBeTrue)
			})
		})
	})
}
