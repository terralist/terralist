package services

import (
	"errors"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/internal/server/repositories"
	"terralist/pkg/database/entity"
	"terralist/pkg/file"
	"terralist/pkg/storage"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

type moduleFixture struct {
	repo      *repositories.MockModuleRepository
	authority *MockAuthorityService
	upstream  *MockUpstreamService
	resolver  *storage.MockResolver
	fetcher   *file.MockFetcher
	service   *DefaultModuleService
	auth      *authority.Authority
}

func newModuleFixture(t *testing.T) *moduleFixture {
	t.Helper()

	f := &moduleFixture{
		repo:      repositories.NewMockModuleRepository(t),
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
	}
	f.authority.On("GetByName", "hashicorp").Return(f.auth, nil).Maybe()
	f.authority.On("GetByID", f.auth.ID).Return(f.auth, nil).Maybe()

	f.service = &DefaultModuleService{
		ModuleRepository: f.repo,
		AuthorityService: f.authority,
		Resolver:         f.resolver,
		Fetcher:          f.fetcher,
		Upstream:         f.upstream,
		ArchiveBaseURL:   "https://terralist.example.com/v1/modules",
	}

	return f
}

func (f *moduleFixture) localModule() *module.Module {
	return &module.Module{
		AuthorityID: f.auth.ID,
		Name:        "dir",
		Provider:    "template",
		Versions:    []module.Version{{Version: "1.0.0", Location: "modules/hashicorp/dir/template/1.0.0.zip"}},
	}
}

func TestGetModuleWithUpstream(t *testing.T) {
	Convey("Subject: Listing module versions merged with the upstream", t, func() {
		f := newModuleFixture(t)

		Convey("Given local and upstream versions and a caller who may fetch", func() {
			f.repo.On("Find", "hashicorp", "dir", "template").Return(f.localModule(), nil)
			f.upstream.On("ModuleVersions", f.auth, "dir", "template").Return([]string{"1.0.0", "1.0.2"}, nil)

			dto, err := f.service.Get("hashicorp", "dir", "template", true)

			Convey("Then the local version wins and the upstream ones fill the gaps", func() {
				So(err, ShouldBeNil)
				So(dto.Modules[0].Versions, ShouldResemble, []module.VersionListDTO{{Version: "1.0.0"}, {Version: "1.0.2"}})
			})
		})

		Convey("Given a caller who may not fetch", func() {
			f.repo.On("Find", "hashicorp", "dir", "template").Return(f.localModule(), nil)

			dto, err := f.service.Get("hashicorp", "dir", "template", false)

			Convey("Then only the local versions are listed", func() {
				So(err, ShouldBeNil)
				So(len(dto.Modules[0].Versions), ShouldEqual, 1)
				f.upstream.AssertNotCalled(t, "ModuleVersions", mock.Anything, mock.Anything, mock.Anything)
			})
		})

		Convey("Given a module that exists upstream only", func() {
			f.repo.On("Find", "hashicorp", "dir", "template").Return(nil, errors.New("not found"))
			f.upstream.On("ModuleVersions", f.auth, "dir", "template").Return([]string{"1.0.2"}, nil)

			dto, err := f.service.Get("hashicorp", "dir", "template", true)

			Convey("Then the upstream versions are listed", func() {
				So(err, ShouldBeNil)
				So(dto.Modules[0].Versions, ShouldResemble, []module.VersionListDTO{{Version: "1.0.2"}})
			})
		})

		Convey("Given a module that exists nowhere", func() {
			f.repo.On("Find", "hashicorp", "dir", "template").Return(nil, errors.New("not found"))
			f.upstream.On("ModuleVersions", f.auth, "dir", "template").Return(nil, nil)

			_, err := f.service.Get("hashicorp", "dir", "template", true)

			Convey("Then it is not found", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given the upstream fails and the module exists locally", func() {
			f.repo.On("Find", "hashicorp", "dir", "template").Return(f.localModule(), nil)
			f.upstream.On("ModuleVersions", f.auth, "dir", "template").Return(nil, errors.New("upstream down"))

			dto, err := f.service.Get("hashicorp", "dir", "template", true)

			Convey("Then the local versions are still served", func() {
				So(err, ShouldBeNil)
				So(dto.Modules[0].Versions, ShouldResemble, []module.VersionListDTO{{Version: "1.0.0"}})
			})
		})

		Convey("Given the upstream fails and the module is not held locally", func() {
			f.repo.On("Find", "hashicorp", "dir", "template").Return(nil, repositories.ErrNotFound)
			f.upstream.On("ModuleVersions", f.auth, "dir", "template").Return(nil, errors.New("upstream down"))

			_, err := f.service.Get("hashicorp", "dir", "template", true)

			Convey("Then the upstream is reported unavailable", func() {
				So(errors.Is(err, ErrUpstreamUnavailable), ShouldBeTrue)
			})
		})
	})
}

func TestGetModuleVersionURLFromUpstream(t *testing.T) {
	Convey("Subject: The download location of a module version not stored yet", t, func() {
		f := newModuleFixture(t)
		f.repo.On("FindVersionLocation", "hashicorp", "dir", "template", "1.0.2").Return(nil, repositories.ErrNotFound)

		Convey("Given a caller who may fetch and an upstream offering the version", func() {
			f.upstream.On("ModuleVersions", f.auth, "dir", "template").Return([]string{"1.0.2"}, nil)

			location, err := f.service.GetVersionURL("hashicorp", "dir", "template", "1.0.2", true)

			Convey("Then the location points at the archive route that fetches on demand", func() {
				So(err, ShouldBeNil)
				So(*location, ShouldEqual, "https://terralist.example.com/v1/modules/hashicorp/dir/template/1.0.2/archive")
			})
		})

		Convey("Given an upstream that does not offer the version", func() {
			f.upstream.On("ModuleVersions", f.auth, "dir", "template").Return([]string{"1.0.1"}, nil)

			_, err := f.service.GetVersionURL("hashicorp", "dir", "template", "1.0.2", true)

			Convey("Then it is not found", func() {
				So(errors.Is(err, repositories.ErrNotFound), ShouldBeTrue)
			})
		})

		Convey("Given the upstream fails", func() {
			f.upstream.On("ModuleVersions", f.auth, "dir", "template").Return(nil, errors.New("upstream down"))

			_, err := f.service.GetVersionURL("hashicorp", "dir", "template", "1.0.2", true)

			Convey("Then the upstream is reported unavailable", func() {
				So(errors.Is(err, ErrUpstreamUnavailable), ShouldBeTrue)
			})
		})

		Convey("Given a caller who may not fetch", func() {
			_, err := f.service.GetVersionURL("hashicorp", "dir", "template", "1.0.2", false)

			Convey("Then it is not found", func() {
				So(errors.Is(err, repositories.ErrNotFound), ShouldBeTrue)
			})
		})
	})
}

func TestDownloadModule(t *testing.T) {
	Convey("Subject: Resolving a module archive for download, fetching it from the upstream on a miss", t, func() {
		f := newModuleFixture(t)

		Convey("Given the version is stored", func() {
			location := "modules/hashicorp/dir/template/1.0.0.zip"
			f.repo.On("FindVersionLocation", "hashicorp", "dir", "template", "1.0.0").Return(&location, nil)
			f.resolver.On("Find", location).Return("https://storage/1.0.0.zip", nil)

			url, err := f.service.Download("hashicorp", "dir", "template", "1.0.0", false)

			Convey("Then its storage URL is returned", func() {
				So(err, ShouldBeNil)
				So(url, ShouldEqual, "https://storage/1.0.0.zip")
			})
		})

		Convey("Given the version is not stored", func() {
			f.repo.On("FindVersionLocation", "hashicorp", "dir", "template", "1.0.2").Return(nil, repositories.ErrNotFound).Once()

			Convey("When the caller may not fetch", func() {
				_, err := f.service.Download("hashicorp", "dir", "template", "1.0.2", false)

				Convey("Then fetching is refused", func() {
					So(errors.Is(err, ErrFetchRequiresCreate), ShouldBeTrue)
				})
			})

			Convey("When the authority has no enabled upstream and the caller may not fetch", func() {
				f.auth.UpstreamEnabled = false
				_, err := f.service.Download("hashicorp", "dir", "template", "1.0.2", false)

				Convey("Then the version is not found", func() {
					So(errors.Is(err, repositories.ErrNotFound), ShouldBeTrue)
				})
			})

			Convey("When the caller may fetch", func() {
				f.upstream.On("ModuleLocation", f.auth, "dir", "template", "1.0.2").Return("git::https://github.com/hashicorp/terraform-template-dir?ref=v1.0.2", nil)
				// Checked once more under the single flight, in case another
				// request stored the version meanwhile.
				f.repo.On("FindVersionLocation", "hashicorp", "dir", "template", "1.0.2").Return(nil, repositories.ErrNotFound).Once()
				f.repo.On("Find", "hashicorp", "dir", "template").Return(f.localModule(), nil)
				f.fetcher.
					On("Fetch", "1.0.2", mock.MatchedBy(func(src file.File) bool {
						remote, ok := src.(*file.RemoteFile)
						return ok && remote.URL() == "git::https://github.com/hashicorp/terraform-template-dir?ref=v1.0.2"
					})).
					Return(file.NewInMemoryFile("1.0.2.zip", []byte("archive")), func() {}, nil)
				f.resolver.
					On("Store", mock.AnythingOfType("*storage.StoreInput")).
					Return(func(in *storage.StoreInput) (string, error) { return in.KeyPrefix + "/" + in.FileName, nil })
				var saved module.Module
				f.repo.
					On("Upsert", mock.AnythingOfType("module.Module")).
					Run(func(args mock.Arguments) { saved, _ = args.Get(0).(module.Module) }).
					Return(&module.Module{}, nil)
				stored := "modules/hashicorp/dir/template/1.0.2.zip"
				f.repo.On("FindVersionLocation", "hashicorp", "dir", "template", "1.0.2").Return(&stored, nil).Once()
				f.resolver.On("Find", stored).Return("https://storage/1.0.2.zip", nil)

				url, err := f.service.Download("hashicorp", "dir", "template", "1.0.2", true)

				Convey("Then the module is fetched from its upstream source, stored as an upstream version and served from storage", func() {
					So(err, ShouldBeNil)
					So(url, ShouldEqual, "https://storage/1.0.2.zip")
					So(len(saved.Versions), ShouldEqual, 2)
					So(saved.Versions[1].Version, ShouldEqual, "1.0.2")
					So(saved.Versions[1].Origin, ShouldEqual, module.OriginUpstream)
				})
			})

			Convey("When another request stores the version at the same time", func() {
				f.upstream.On("ModuleLocation", f.auth, "dir", "template", "1.0.2").Return("git::https://github.com/hashicorp/terraform-template-dir?ref=v1.0.2", nil)
				stored := false
				concurrent := "modules/hashicorp/dir/template/concurrent.zip"
				f.repo.On("FindVersionLocation", "hashicorp", "dir", "template", "1.0.2").Return(func(string, string, string, string) (*string, error) {
					if stored {
						return &concurrent, nil
					}

					return nil, repositories.ErrNotFound
				})
				f.repo.On("Find", "hashicorp", "dir", "template").Return(f.localModule(), nil)
				f.fetcher.On("Fetch", "1.0.2", mock.Anything).Return(file.NewInMemoryFile("1.0.2.zip", []byte("archive")), func() {}, nil)
				f.resolver.
					On("Store", mock.AnythingOfType("*storage.StoreInput")).
					Return(func(in *storage.StoreInput) (string, error) { return in.KeyPrefix + "/" + in.FileName, nil })
				f.repo.
					On("Upsert", mock.AnythingOfType("module.Module")).
					Run(func(mock.Arguments) { stored = true }).
					Return(nil, repositories.ErrAlreadyExists)
				f.resolver.On("Find", concurrent).Return("https://storage/concurrent.zip", nil)

				url, err := f.service.Download("hashicorp", "dir", "template", "1.0.2", true)

				Convey("Then the version stored by the other request is served", func() {
					So(err, ShouldBeNil)
					So(url, ShouldEqual, "https://storage/concurrent.zip")
				})
			})

			Convey("When another request committed the version right before the upload", func() {
				f.upstream.On("ModuleLocation", f.auth, "dir", "template", "1.0.2").Return("git::https://github.com/hashicorp/terraform-template-dir?ref=v1.0.2", nil)
				concurrent := "modules/hashicorp/dir/template/concurrent.zip"
				f.repo.On("FindVersionLocation", "hashicorp", "dir", "template", "1.0.2").Return(nil, repositories.ErrNotFound).Once()
				f.repo.On("FindVersionLocation", "hashicorp", "dir", "template", "1.0.2").Return(&concurrent, nil)
				withVersion := f.localModule()
				withVersion.Versions = append(withVersion.Versions, module.Version{Version: "1.0.2", Location: concurrent})
				f.repo.On("Find", "hashicorp", "dir", "template").Return(withVersion, nil)
				f.resolver.On("Find", concurrent).Return("https://storage/concurrent.zip", nil)

				url, err := f.service.Download("hashicorp", "dir", "template", "1.0.2", true)

				Convey("Then the version stored by the other request is served", func() {
					So(err, ShouldBeNil)
					So(url, ShouldEqual, "https://storage/concurrent.zip")
					f.fetcher.AssertNotCalled(t, "Fetch", mock.Anything, mock.Anything)
				})
			})

			Convey("When the upstream source may not be fetched", func() {
				f.upstream.On("ModuleLocation", f.auth, "dir", "template", "1.0.2").Return("s3::https://s3.amazonaws.com/bucket/dir.zip", nil)
				f.repo.On("FindVersionLocation", "hashicorp", "dir", "template", "1.0.2").Return(nil, repositories.ErrNotFound).Once()
				f.repo.On("Find", "hashicorp", "dir", "template").Return(f.localModule(), nil)
				f.fetcher.On("Fetch", "1.0.2", mock.Anything).Return(nil, nil, errors.New("refusing to fetch"))

				_, err := f.service.Download("hashicorp", "dir", "template", "1.0.2", true)

				Convey("Then the refusal is reported and nothing stored", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "refusing to fetch")
					f.resolver.AssertNotCalled(t, "Store", mock.Anything)
				})
			})

			Convey("When the upstream denies the version", func() {
				f.upstream.On("ModuleLocation", f.auth, "dir", "template", "1.0.2").Return("", ErrUpstreamDenied)

				_, err := f.service.Download("hashicorp", "dir", "template", "1.0.2", true)

				Convey("Then the denial is reported and nothing stored", func() {
					So(errors.Is(err, ErrUpstreamDenied), ShouldBeTrue)
					f.resolver.AssertNotCalled(t, "Store", mock.Anything)
				})
			})

			Convey("When the authority has no enabled upstream", func() {
				f.auth.UpstreamEnabled = false

				_, err := f.service.Download("hashicorp", "dir", "template", "1.0.2", true)

				Convey("Then the version is not found", func() {
					So(errors.Is(err, repositories.ErrNotFound), ShouldBeTrue)
				})
			})
		})
	})
}
