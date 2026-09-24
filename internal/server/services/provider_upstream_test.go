package services

import (
	"errors"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/repositories"
	"terralist/pkg/database/entity"
	"terralist/pkg/file"
	"terralist/pkg/registry"
	"terralist/pkg/storage"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

type pullThroughFixture struct {
	repo      *repositories.MockProviderRepository
	authority *MockAuthorityService
	upstream  *MockUpstreamService
	resolver  *storage.MockResolver
	fetcher   *file.MockFetcher
	service   *DefaultProviderService
	auth      *authority.Authority
}

func newPullThroughFixture(t *testing.T) *pullThroughFixture {
	t.Helper()

	f := &pullThroughFixture{
		repo:      repositories.NewMockProviderRepository(t),
		authority: NewMockAuthorityService(t),
		upstream:  NewMockUpstreamService(t),
		resolver:  storage.NewMockResolver(t),
		fetcher:   file.NewMockFetcher(t),
	}

	hostname := "registry.terraform.io"
	f.auth = &authority.Authority{
		Entity:           entity.Entity{ID: uuid.New()},
		Name:             "hashicorp",
		UpstreamHostname: &hostname,
		UpstreamEnabled:  true,
		Keys:             []authority.Key{{KeyId: "LOCALKEY", AsciiArmor: "local-armor"}},
	}
	f.authority.On("GetByName", "hashicorp").Return(f.auth, nil).Maybe()
	f.authority.On("GetByID", f.auth.ID).Return(f.auth, nil).Maybe()

	f.service = &DefaultProviderService{
		ProviderRepository: f.repo,
		AuthorityService:   f.authority,
		Resolver:           f.resolver,
		Fetcher:            f.fetcher,
		Upstream:           f.upstream,
		MirrorBaseURL:      "https://terralist.example.com/providers/terralist.example.com",
	}

	return f
}

func (f *pullThroughFixture) localProvider() *provider.Provider {
	return &provider.Provider{
		AuthorityID: f.auth.ID,
		Name:        "null",
		Versions: []provider.Version{
			{
				Version:             "3.2.4",
				Protocols:           "5.0",
				ShaSumsUrl:          "providers/hashicorp/null/3.2.4/SHA256SUMS",
				ShaSumsSignatureUrl: "providers/hashicorp/null/3.2.4/SHA256SUMS.sig",
				Platforms: []provider.Platform{
					{System: "linux", Architecture: "amd64", Location: "providers/hashicorp/null/3.2.4/linux.zip", ShaSum: "aaaa"},
				},
			},
		},
	}
}

var upstreamVersions = []UpstreamVersion{
	{Version: "3.2.4", Protocols: []string{"5.0"}, Platforms: []registry.Platform{{OS: "linux", Arch: "amd64"}, {OS: "darwin", Arch: "arm64"}}},
	{Version: "3.2.5", Protocols: []string{"5.0", "6.0"}, Platforms: []registry.Platform{{OS: "linux", Arch: "amd64"}}},
}

var upstreamMetadata = &UpstreamVersionMetadata{
	Protocols: []string{"5.0"},
	ShaSums: map[string]string{
		"terraform-provider-null_3.2.4_linux_amd64.zip":  "aaaa",
		"terraform-provider-null_3.2.4_darwin_arm64.zip": "bbbb",
	},
	ShaSumsDocument: []byte("sums"),
	Signature:       []byte("sig"),
	SigningKeys:     []registry.GPGPublicKey{{KeyID: "UPSTREAMKEY", ASCIIArmor: "upstream-armor"}},
}

func TestGetProviderWithUpstream(t *testing.T) {
	Convey("Subject: Listing registry versions merged with the upstream", t, func() {
		f := newPullThroughFixture(t)

		Convey("Given local and upstream versions", func() {
			f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil)
			f.upstream.On("ProviderVersions", f.auth, "null").Return(upstreamVersions, nil)

			Convey("When the caller may fetch", func() {
				dto, err := f.service.Get("hashicorp", "null", true)

				Convey("Then the local version wins and the upstream ones fill the gaps", func() {
					So(err, ShouldBeNil)
					So(len(dto.Versions), ShouldEqual, 2)
					So(dto.Versions[0].Version, ShouldEqual, "3.2.4")
					So(len(dto.Versions[0].Platforms), ShouldEqual, 1)
					So(dto.Versions[1].Version, ShouldEqual, "3.2.5")
					So(dto.Versions[1].Protocols, ShouldResemble, []string{"5.0", "6.0"})
					So(dto.Versions[1].Platforms, ShouldResemble, []provider.VersionListPlatformDTO{{System: "linux", Architecture: "amd64"}})
				})
			})
		})

		Convey("Given a mirror-only local version that also exists upstream", func() {
			local := f.localProvider()
			local.Versions = append(local.Versions, provider.Version{Version: "3.2.5"})
			f.repo.On("Find", "hashicorp", "null").Return(local, nil)
			f.upstream.On("ProviderVersions", f.auth, "null").Return(upstreamVersions, nil)

			dto, err := f.service.Get("hashicorp", "null", true)

			Convey("Then the uploaded version wins and stays hidden from the registry", func() {
				So(err, ShouldBeNil)
				So(len(dto.Versions), ShouldEqual, 1)
				So(dto.Versions[0].Version, ShouldEqual, "3.2.4")
			})
		})

		Convey("Given the caller may not fetch", func() {
			f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil)

			dto, err := f.service.Get("hashicorp", "null", false)

			Convey("Then only the local versions are listed and the upstream is not consulted", func() {
				So(err, ShouldBeNil)
				So(len(dto.Versions), ShouldEqual, 1)
				f.upstream.AssertNotCalled(t, "ProviderVersions", mock.Anything, mock.Anything)
			})
		})

		Convey("Given a provider that exists upstream only", func() {
			f.repo.On("Find", "hashicorp", "random").Return(nil, errors.New("not found"))
			f.upstream.On("ProviderVersions", f.auth, "random").Return(upstreamVersions, nil)

			dto, err := f.service.Get("hashicorp", "random", true)

			Convey("Then the upstream versions are listed", func() {
				So(err, ShouldBeNil)
				So(len(dto.Versions), ShouldEqual, 2)
			})
		})

		Convey("Given a provider that exists nowhere", func() {
			f.repo.On("Find", "hashicorp", "missing").Return(nil, errors.New("not found"))
			f.upstream.On("ProviderVersions", f.auth, "missing").Return(nil, nil)

			_, err := f.service.Get("hashicorp", "missing", true)

			Convey("Then it is not found", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given the upstream fails and the provider exists locally", func() {
			f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil)
			f.upstream.On("ProviderVersions", f.auth, "null").Return(nil, errors.New("upstream down"))

			dto, err := f.service.Get("hashicorp", "null", true)

			Convey("Then the local versions are still served", func() {
				So(err, ShouldBeNil)
				So(len(dto.Versions), ShouldEqual, 1)
			})
		})
	})
}

func TestUpstreamRequiresResolver(t *testing.T) {
	Convey("Subject: Pulling through without a storage backend", t, func() {
		f := newPullThroughFixture(t)
		f.service.Resolver = nil
		f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil)

		dto, err := f.service.Get("hashicorp", "null", true)

		Convey("Then only the local versions are served and the upstream is not consulted", func() {
			So(err, ShouldBeNil)
			So(len(dto.Versions), ShouldEqual, 1)
			f.upstream.AssertNotCalled(t, "ProviderVersions", mock.Anything, mock.Anything)
		})
	})
}

func TestListMirrorVersionsWithUpstream(t *testing.T) {
	Convey("Subject: Listing mirror versions merged with the upstream", t, func() {
		f := newPullThroughFixture(t)
		f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil)
		f.upstream.On("ProviderVersions", f.auth, "null").Return(upstreamVersions, nil)

		dto, err := f.service.ListMirrorVersions("hashicorp", "null", true)

		Convey("Then the union of versions is listed", func() {
			So(err, ShouldBeNil)
			So(dto.Versions, ShouldResemble, map[string]struct{}{"3.2.4": {}, "3.2.5": {}})
		})
	})
}

func TestListMirrorArchivesWithUpstream(t *testing.T) {
	Convey("Subject: Listing mirror packages merged with the upstream", t, func() {
		f := newPullThroughFixture(t)

		Convey("Given a version with one stored and one upstream platform", func() {
			f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil)
			f.resolver.On("Find", "providers/hashicorp/null/3.2.4/linux.zip").Return("https://storage/linux.zip", nil)
			f.upstream.On("ProviderVersions", f.auth, "null").Return(upstreamVersions, nil)
			f.upstream.On("ProviderVersion", f.auth, "null", "3.2.4").Return(upstreamMetadata, nil)

			dto, err := f.service.ListMirrorArchives("hashicorp", "null", "3.2.4", true)

			Convey("Then the stored platform keeps its storage URL and the upstream one points back at the mirror", func() {
				So(err, ShouldBeNil)
				So(dto.Archives["linux_amd64"], ShouldResemble, provider.MirrorArchiveDTO{URL: "https://storage/linux.zip", Hashes: []string{"zh:aaaa"}})
				So(dto.Archives["darwin_arm64"], ShouldResemble, provider.MirrorArchiveDTO{URL: "terraform-provider-null_3.2.4_darwin_arm64.zip", Hashes: []string{"zh:bbbb"}})
			})
		})

		Convey("Given a version that exists upstream only", func() {
			f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil)
			f.upstream.On("ProviderVersions", f.auth, "null").Return(upstreamVersions, nil)
			f.upstream.On("ProviderVersion", f.auth, "null", "3.2.5").Return(&UpstreamVersionMetadata{
				ShaSums: map[string]string{"terraform-provider-null_3.2.5_linux_amd64.zip": "cccc"},
			}, nil)

			dto, err := f.service.ListMirrorArchives("hashicorp", "null", "3.2.5", true)

			Convey("Then every upstream platform points back at the mirror", func() {
				So(err, ShouldBeNil)
				So(dto.Archives, ShouldResemble, map[string]provider.MirrorArchiveDTO{
					"linux_amd64": {URL: "terraform-provider-null_3.2.5_linux_amd64.zip", Hashes: []string{"zh:cccc"}},
				})
			})
		})

		Convey("Given the caller may not fetch", func() {
			f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil)
			f.resolver.On("Find", "providers/hashicorp/null/3.2.4/linux.zip").Return("https://storage/linux.zip", nil)

			dto, err := f.service.ListMirrorArchives("hashicorp", "null", "3.2.4", false)

			Convey("Then only the stored platforms are listed", func() {
				So(err, ShouldBeNil)
				So(len(dto.Archives), ShouldEqual, 1)
			})
		})

		Convey("Given the upstream denies the version", func() {
			f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil)
			f.upstream.On("ProviderVersions", f.auth, "null").Return(nil, nil)

			_, err := f.service.ListMirrorArchives("hashicorp", "null", "3.2.5", true)

			Convey("Then the version is not found", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestGetVersionFromUpstream(t *testing.T) {
	Convey("Subject: Registry download metadata for a platform not stored yet", t, func() {
		f := newPullThroughFixture(t)
		f.repo.On("FindVersionPlatform", "hashicorp", "null", "3.2.4", "darwin", "arm64").Return(nil, repositories.ErrNotFound).Maybe()

		Convey("Given the caller may fetch and the upstream knows the version", func() {
			local := f.localProvider()
			local.Versions[0].Origin = provider.OriginUpstream
			f.repo.On("Find", "hashicorp", "null").Return(local, nil)
			f.upstream.On("ProviderVersion", f.auth, "null", "3.2.4").Return(upstreamMetadata, nil)
			f.resolver.On("Find", "providers/hashicorp/null/3.2.4/SHA256SUMS").Return("https://storage/SHA256SUMS", nil)
			f.resolver.On("Find", "providers/hashicorp/null/3.2.4/SHA256SUMS.sig").Return("https://storage/SHA256SUMS.sig", nil)

			dto, err := f.service.GetVersion("hashicorp", "null", "3.2.4", "darwin", "arm64", true)

			Convey("Then the metadata points the download at the mirror and carries the digest", func() {
				So(err, ShouldBeNil)
				So(dto.DownloadUrl, ShouldEqual, "https://terralist.example.com/providers/terralist.example.com/hashicorp/null/terraform-provider-null_3.2.4_darwin_arm64.zip")
				So(dto.ShaSum, ShouldEqual, "bbbb")
				So(dto.FileName, ShouldEqual, "terraform-provider-null_3.2.4_darwin_arm64.zip")
				So(dto.ShaSumsUrl, ShouldEqual, "https://storage/SHA256SUMS")
				So(dto.ShaSumsSignatureUrl, ShouldEqual, "https://storage/SHA256SUMS.sig")
				So(dto.Protocols, ShouldResemble, []string{"5.0"})
			})
		})

		Convey("Given a version that exists upstream only", func() {
			f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil)
			metadata := &UpstreamVersionMetadata{
				Protocols:       []string{"6.0"},
				ShaSums:         map[string]string{"terraform-provider-null_3.2.5_darwin_arm64.zip": "dddd"},
				ShaSumsDocument: []byte("sums"),
				Signature:       []byte("sig"),
				SigningKeys:     upstreamMetadata.SigningKeys,
			}
			f.repo.On("FindVersionPlatform", "hashicorp", "null", "3.2.5", "darwin", "arm64").Return(nil, repositories.ErrNotFound)
			f.upstream.On("ProviderVersion", f.auth, "null", "3.2.5").Return(metadata, nil)
			f.resolver.
				On("Store", mock.AnythingOfType("*storage.StoreInput")).
				Return(func(in *storage.StoreInput) (string, error) { return in.KeyPrefix + "/" + in.FileName, nil })
			var saved provider.Provider
			f.repo.
				On("Upsert", mock.AnythingOfType("provider.Provider")).
				Run(func(args mock.Arguments) { saved, _ = args.Get(0).(provider.Provider) }).
				Return(&provider.Provider{}, nil)
			f.resolver.On("Find", mock.Anything).Return("https://storage/resolved", nil)

			dto, err := f.service.GetVersion("hashicorp", "null", "3.2.5", "darwin", "arm64", true)

			Convey("Then the version is created from the upstream with its signature material and keys", func() {
				So(err, ShouldBeNil)
				So(len(saved.Versions), ShouldEqual, 2)
				v := saved.Versions[1]
				So(v.Version, ShouldEqual, "3.2.5")
				So(v.Origin, ShouldEqual, provider.OriginUpstream)
				So(v.Protocols, ShouldEqual, "6.0")
				So(v.ShaSumsUrl, ShouldEqual, "providers/hashicorp/null/3.2.5/terraform-provider-null_3.2.5_SHA256SUMS")
				So(v.ShaSumsSignatureUrl, ShouldEqual, "providers/hashicorp/null/3.2.5/terraform-provider-null_3.2.5_SHA256SUMS.sig")
				So(v.SigningKeys, ShouldContainSubstring, "UPSTREAMKEY")
				So(len(v.Platforms), ShouldEqual, 0)
				So(dto.SigningKeys.Keys[0].KeyId, ShouldEqual, "UPSTREAMKEY")
			})
		})

		Convey("Given the version was uploaded by an operator", func() {
			f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil)

			_, err := f.service.GetVersion("hashicorp", "null", "3.2.4", "darwin", "arm64", true)

			Convey("Then the missing platform is not completed from the upstream", func() {
				So(errors.Is(err, repositories.ErrNotFound), ShouldBeTrue)
				f.upstream.AssertNotCalled(t, "ProviderVersion", mock.Anything, mock.Anything, mock.Anything)
			})
		})

		Convey("Given the caller may not fetch", func() {
			_, err := f.service.GetVersion("hashicorp", "null", "3.2.4", "darwin", "arm64", false)

			Convey("Then the platform is not found", func() {
				So(errors.Is(err, repositories.ErrNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestDownloadProviderPackage(t *testing.T) {
	Convey("Subject: Resolving a package for download, fetching it from the upstream on a miss", t, func() {
		f := newPullThroughFixture(t)

		Convey("Given the platform is stored", func() {
			f.repo.On("FindVersionPlatform", "hashicorp", "null", "3.2.4", "linux", "amd64").Return(&provider.Platform{
				Location: "providers/hashicorp/null/3.2.4/linux.zip",
				Version:  provider.Version{Version: "3.2.4"},
			}, nil)
			f.resolver.On("Find", "providers/hashicorp/null/3.2.4/linux.zip").Return("https://storage/linux.zip", nil)

			url, err := f.service.Download("hashicorp", "null", "3.2.4", "linux", "amd64", false)

			Convey("Then its storage URL is returned without touching the upstream", func() {
				So(err, ShouldBeNil)
				So(url, ShouldEqual, "https://storage/linux.zip")
			})
		})

		Convey("Given the platform is not stored", func() {
			f.repo.On("FindVersionPlatform", "hashicorp", "null", "3.2.4", "darwin", "arm64").Return(nil, repositories.ErrNotFound)

			Convey("When the caller may not fetch", func() {
				_, err := f.service.Download("hashicorp", "null", "3.2.4", "darwin", "arm64", false)

				Convey("Then fetching is refused", func() {
					So(errors.Is(err, ErrFetchRequiresCreate), ShouldBeTrue)
				})
			})

			Convey("When the authority has no enabled upstream", func() {
				f.auth.UpstreamEnabled = false
				_, err := f.service.Download("hashicorp", "null", "3.2.4", "darwin", "arm64", true)

				Convey("Then the package is not found", func() {
					So(errors.Is(err, repositories.ErrNotFound), ShouldBeTrue)
				})
			})

			Convey("When the caller may fetch", func() {
				f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil).Maybe()
				f.upstream.On("ProviderPackage", f.auth, "null", "3.2.4", "darwin", "arm64").Return(&UpstreamPackage{
					FileName: "terraform-provider-null_3.2.4_darwin_arm64.zip",
					URL:      "https://releases.example.com/darwin.zip",
					ShaSum:   "bbbb",
				}, nil)
				f.upstream.On("ProviderVersion", f.auth, "null", "3.2.4").Return(upstreamMetadata, nil).Maybe()
				f.fetcher.
					On("FetchFileChecksum", "terraform-provider-null_3.2.4_darwin_arm64.zip", "https://releases.example.com/darwin.zip", "bbbb", mock.Anything).
					Return(file.NewInMemoryFile("terraform-provider-null_3.2.4_darwin_arm64.zip", []byte("zip")), func() {}, nil)
				f.resolver.
					On("Store", mock.AnythingOfType("*storage.StoreInput")).
					Return(func(in *storage.StoreInput) (string, error) { return in.KeyPrefix + "/" + in.FileName, nil })
				var saved provider.Provider
				f.repo.
					On("Upsert", mock.AnythingOfType("provider.Provider")).
					Run(func(args mock.Arguments) { saved, _ = args.Get(0).(provider.Provider) }).
					Return(&provider.Provider{}, nil)
				f.resolver.On("Find", "providers/hashicorp/null/3.2.4/terraform-provider-null_3.2.4_darwin_arm64.zip").Return("https://storage/darwin.zip", nil)

				url, err := f.service.Download("hashicorp", "null", "3.2.4", "darwin", "arm64", true)

				Convey("Then the package is fetched with its digest enforced, stored and recorded as an upstream platform", func() {
					So(err, ShouldBeNil)
					So(url, ShouldEqual, "https://storage/darwin.zip")

					v := saved.Versions[0]
					So(len(v.Platforms), ShouldEqual, 2)
					p := v.Platforms[1]
					So(p.System, ShouldEqual, "darwin")
					So(p.Architecture, ShouldEqual, "arm64")
					So(p.ShaSum, ShouldEqual, "bbbb")
					So(p.Origin, ShouldEqual, provider.OriginUpstream)
					So(p.Location, ShouldEqual, "providers/hashicorp/null/3.2.4/terraform-provider-null_3.2.4_darwin_arm64.zip")
				})
			})

			Convey("When the upstream fetch fails", func() {
				f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil).Maybe()
				f.upstream.On("ProviderPackage", f.auth, "null", "3.2.4", "darwin", "arm64").Return(nil, errors.New("upstream down"))

				_, err := f.service.Download("hashicorp", "null", "3.2.4", "darwin", "arm64", true)

				Convey("Then nothing is stored and the failure is reported", func() {
					So(err, ShouldNotBeNil)
					f.resolver.AssertNotCalled(t, "Store", mock.Anything)
					f.repo.AssertNotCalled(t, "Upsert", mock.Anything)
				})
			})
		})
	})
}

func TestFetchProviderPackages(t *testing.T) {
	Convey("Subject: Fetching packages from the upstream on demand", t, func() {
		f := newPullThroughFixture(t)
		f.repo.On("FindVersionPlatform", "hashicorp", "null", "3.2.4", "linux", "amd64").Return(&provider.Platform{
			Location: "providers/hashicorp/null/3.2.4/linux.zip",
		}, nil)
		f.resolver.On("Find", "providers/hashicorp/null/3.2.4/linux.zip").Return("https://storage/linux.zip", nil)
		f.repo.On("FindVersionPlatform", "hashicorp", "null", "3.2.4", "darwin", "arm64").Return(nil, repositories.ErrNotFound)
		f.repo.On("Find", "hashicorp", "null").Return(f.localProvider(), nil).Maybe()
		f.upstream.On("ProviderPackage", f.auth, "null", "3.2.4", "darwin", "arm64").Return(nil, errors.New("upstream down"))

		results := f.service.Fetch("hashicorp", "null", "3.2.4", []string{"linux_amd64", "darwin_arm64", "bogus"})

		Convey("Then every platform reports its own outcome", func() {
			So(len(results), ShouldEqual, 3)
			So(results[0], ShouldResemble, provider.FetchResultDTO{Platform: "linux_amd64"})
			So(results[1].Platform, ShouldEqual, "darwin_arm64")
			So(results[1].Error, ShouldContainSubstring, "upstream down")
			So(results[2].Platform, ShouldEqual, "bogus")
			So(results[2].Error, ShouldNotBeEmpty)
		})
	})
}
