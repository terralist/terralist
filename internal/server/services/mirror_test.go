package services

import (
	"archive/zip"
	"bytes"
	"errors"
	"testing"

	"terralist/internal/server/models/mirror"
	"terralist/internal/server/repositories"
	"terralist/pkg/file"
	"terralist/pkg/storage"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

// fixtureH1 is the h1 hash of the archive produced by fixtureArchive.
const fixtureH1 = "h1:BE037OLpevcFng+h3LIvMEHxpIBGEv+aM2XiVAN8/A4="

// fixtureArchive builds a deterministic provider package archive.
func fixtureArchive(name string) file.File {
	var buf bytes.Buffer

	w := zip.NewWriter(&buf)
	for _, entry := range [][2]string{
		{"terraform-provider-null_v3.2.4_x5", "binary-content"},
		{"LICENSE", "license-content"},
	} {
		f, _ := w.Create(entry[0])
		_, _ = f.Write([]byte(entry[1]))
	}
	_ = w.Close()

	return file.NewInMemoryFile(name, buf.Bytes())
}

func fixtureMetadata(archives map[string]string) mirror.ArchivesDTO {
	dto := mirror.ArchivesDTO{Archives: map[string]mirror.ArchiveDTO{}}
	for key, url := range archives {
		dto.Archives[key] = mirror.ArchiveDTO{
			URL:    url,
			Hashes: []string{fixtureH1},
		}
	}

	return dto
}

func fixtureProvider() *mirror.Provider {
	p := &mirror.Provider{
		Hostname:  "registry.terraform.io",
		Namespace: "hashicorp",
		Name:      "null",
		Versions: []mirror.Version{
			{
				Version: "3.2.4",
				Platforms: []mirror.Platform{
					{
						System:       "linux",
						Architecture: "amd64",
						Location:     "mirror/registry.terraform.io/hashicorp/null/3.2.4/terraform-provider-null_3.2.4_linux_amd64.zip",
						Hashes:       fixtureH1 + ",zh:abc",
					},
				},
			},
		},
	}
	p.ID = uuid.New()
	p.Versions[0].ID = uuid.New()
	p.Versions[0].Platforms[0].ID = uuid.New()

	return p
}

func TestMirrorListVersions(t *testing.T) {
	Convey("Subject: List the versions of a mirrored provider", t, func() {
		mockRepository := repositories.NewMockMirrorRepository(t)

		service := &DefaultMirrorService{
			MirrorRepository: mockRepository,
		}

		Convey("If the provider exists", func() {
			mockRepository.
				On("Find", "registry.terraform.io", "hashicorp", "null").
				Return(fixtureProvider(), nil)

			Convey("When the service is queried", func() {
				resp, err := service.ListVersions("registry.terraform.io", "hashicorp", "null")

				Convey("The versions should be returned", func() {
					So(err, ShouldBeNil)
					So(resp.Versions, ShouldContainKey, "3.2.4")
					So(len(resp.Versions), ShouldEqual, 1)
				})
			})
		})

		Convey("If the provider does not exist", func() {
			mockRepository.
				On("Find", "registry.terraform.io", "hashicorp", "null").
				Return(nil, errors.New(""))

			Convey("When the service is queried", func() {
				resp, err := service.ListVersions("registry.terraform.io", "hashicorp", "null")

				Convey("An error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(resp, ShouldBeNil)
				})
			})
		})
	})
}

func TestMirrorGetVersion(t *testing.T) {
	Convey("Subject: Get the installation packages of a mirrored provider version", t, func() {
		mockRepository := repositories.NewMockMirrorRepository(t)
		mockResolver := storage.NewMockResolver(t)

		service := &DefaultMirrorService{
			MirrorRepository: mockRepository,
			Resolver:         mockResolver,
		}

		Convey("If the provider does not exist", func() {
			mockRepository.
				On("Find", "registry.terraform.io", "hashicorp", "null").
				Return(nil, errors.New(""))

			Convey("When the service is queried", func() {
				resp, err := service.GetVersion("registry.terraform.io", "hashicorp", "null", "3.2.4")

				Convey("An error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(resp, ShouldBeNil)
				})
			})
		})

		Convey("If the provider exists", func() {
			mockRepository.
				On("Find", "registry.terraform.io", "hashicorp", "null").
				Return(fixtureProvider(), nil)

			Convey("If the version does not exist", func() {
				Convey("When the service is queried", func() {
					resp, err := service.GetVersion("registry.terraform.io", "hashicorp", "null", "9.9.9")

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
						So(resp, ShouldBeNil)
					})
				})
			})

			Convey("If the version exists", func() {
				Convey("If the package location cannot be resolved", func() {
					mockResolver.
						On("Find", mock.Anything).
						Return("", errors.New(""))

					Convey("When the service is queried", func() {
						resp, err := service.GetVersion("registry.terraform.io", "hashicorp", "null", "3.2.4")

						Convey("An error should be returned", func() {
							So(err, ShouldNotBeNil)
							So(resp, ShouldBeNil)
						})
					})
				})

				Convey("If the package location can be resolved", func() {
					mockResolver.
						On("Find", "mirror/registry.terraform.io/hashicorp/null/3.2.4/terraform-provider-null_3.2.4_linux_amd64.zip").
						Return("https://storage.example.com/signed-url", nil)

					Convey("When the service is queried", func() {
						resp, err := service.GetVersion("registry.terraform.io", "hashicorp", "null", "3.2.4")

						Convey("The archives should be returned with resolved URLs and hashes", func() {
							So(err, ShouldBeNil)
							So(resp.Archives, ShouldContainKey, "linux_amd64")
							So(resp.Archives["linux_amd64"].URL, ShouldEqual, "https://storage.example.com/signed-url")
							So(resp.Archives["linux_amd64"].Hashes, ShouldResemble, []string{fixtureH1, "zh:abc"})
						})
					})
				})
			})
		})
	})
}

func TestMirrorUpload(t *testing.T) {
	Convey("Subject: Upload mirrored provider packages", t, func() {
		mockRepository := repositories.NewMockMirrorRepository(t)
		mockResolver := storage.NewMockResolver(t)

		service := &DefaultMirrorService{
			MirrorRepository: mockRepository,
			Resolver:         mockResolver,
		}

		archiveName := "terraform-provider-null_3.2.4_linux_amd64.zip"
		metadata := fixtureMetadata(map[string]string{
			"linux_amd64":  archiveName,
			"darwin_arm64": "terraform-provider-null_3.2.4_darwin_arm64.zip",
		})

		Convey("If the version is not a valid semantic version", func() {
			err := service.Upload("registry.terraform.io", "hashicorp", "null", "latest", metadata, []file.File{fixtureArchive(archiveName)})

			Convey("An error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("If the hostname is not valid", func() {
			err := service.Upload("not a hostname/../", "hashicorp", "null", "3.2.4", metadata, []file.File{fixtureArchive(archiveName)})

			Convey("An error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("If the namespace or name are not valid", func() {
			err := service.Upload("registry.terraform.io", "hashi/corp", "null", "3.2.4", metadata, []file.File{fixtureArchive(archiveName)})

			Convey("An error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("If no archive is uploaded", func() {
			err := service.Upload("registry.terraform.io", "hashicorp", "null", "3.2.4", metadata, nil)

			Convey("An error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("If an uploaded archive is not listed in the metadata", func() {
			err := service.Upload("registry.terraform.io", "hashicorp", "null", "3.2.4", metadata, []file.File{fixtureArchive("unknown.zip")})

			Convey("An error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("If the metadata lists no h1 hash for the archive", func() {
			noH1 := fixtureMetadata(map[string]string{"linux_amd64": archiveName})
			noH1.Archives["linux_amd64"] = mirror.ArchiveDTO{URL: archiveName, Hashes: []string{"zh:abc"}}

			err := service.Upload("registry.terraform.io", "hashicorp", "null", "3.2.4", noH1, []file.File{fixtureArchive(archiveName)})

			Convey("An error should be returned", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("If the archive content does not match the h1 hash", func() {
			tampered := fixtureMetadata(map[string]string{"linux_amd64": archiveName})
			tampered.Archives["linux_amd64"] = mirror.ArchiveDTO{URL: archiveName, Hashes: []string{"h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}

			err := service.Upload("registry.terraform.io", "hashicorp", "null", "3.2.4", tampered, []file.File{fixtureArchive(archiveName)})

			Convey("An error should be returned and nothing stored", func() {
				So(err, ShouldNotBeNil)
				mockResolver.AssertNotCalled(t, "Store", mock.Anything)
			})
		})

		Convey("If the archives are valid", func() {
			Convey("If the provider does not exist yet", func() {
				mockRepository.
					On("Find", "registry.terraform.io", "hashicorp", "null").
					Return(nil, errors.New(""))

				mockResolver.
					On("Store", mock.MatchedBy(func(in *storage.StoreInput) bool {
						return in.KeyPrefix == "mirror/registry.terraform.io/hashicorp/null/3.2.4" &&
							in.FileName == archiveName
					})).
					Return("mirror/registry.terraform.io/hashicorp/null/3.2.4/"+archiveName, nil)

				var saved mirror.Provider
				mockRepository.
					On("Upsert", mock.AnythingOfType("mirror.Provider")).
					Run(func(args mock.Arguments) {
						saved, _ = args.Get(0).(mirror.Provider)
					}).
					Return(&mirror.Provider{}, nil)

				Convey("When the upload is requested", func() {
					err := service.Upload("Registry.Terraform.IO", "hashicorp", "null", "3.2.4", metadata, []file.File{fixtureArchive(archiveName)})

					Convey("A provider with the uploaded platform only should be created", func() {
						So(err, ShouldBeNil)
						So(saved.Hostname, ShouldEqual, "registry.terraform.io")
						So(saved.Namespace, ShouldEqual, "hashicorp")
						So(saved.Name, ShouldEqual, "null")
						So(len(saved.Versions), ShouldEqual, 1)
						So(saved.Versions[0].Version, ShouldEqual, "3.2.4")
						So(len(saved.Versions[0].Platforms), ShouldEqual, 1)
						So(saved.Versions[0].Platforms[0].System, ShouldEqual, "linux")
						So(saved.Versions[0].Platforms[0].Architecture, ShouldEqual, "amd64")
						So(saved.Versions[0].Platforms[0].Location, ShouldEqual, "mirror/registry.terraform.io/hashicorp/null/3.2.4/"+archiveName)
						So(saved.Versions[0].Platforms[0].Hashes, ShouldEqual, fixtureH1)
					})
				})
			})

			Convey("If the provider and version exist", func() {
				mockRepository.
					On("Find", "registry.terraform.io", "hashicorp", "null").
					Return(fixtureProvider(), nil)

				Convey("If the platform already exists", func() {
					Convey("When the upload is requested", func() {
						err := service.Upload("registry.terraform.io", "hashicorp", "null", "3.2.4", metadata, []file.File{fixtureArchive(archiveName)})

						Convey("A conflict error should be returned and nothing stored", func() {
							So(err, ShouldNotBeNil)
							mockResolver.AssertNotCalled(t, "Store", mock.Anything)
						})
					})
				})

				Convey("If the platform is new", func() {
					darwinName := "terraform-provider-null_3.2.4_darwin_arm64.zip"

					mockResolver.
						On("Store", mock.MatchedBy(func(in *storage.StoreInput) bool {
							return in.FileName == darwinName
						})).
						Return("mirror/registry.terraform.io/hashicorp/null/3.2.4/"+darwinName, nil)

					var saved mirror.Provider
					mockRepository.
						On("Upsert", mock.AnythingOfType("mirror.Provider")).
						Run(func(args mock.Arguments) {
							saved, _ = args.Get(0).(mirror.Provider)
						}).
						Return(&mirror.Provider{}, nil)

					Convey("When the upload is requested", func() {
						err := service.Upload("registry.terraform.io", "hashicorp", "null", "3.2.4", metadata, []file.File{fixtureArchive(darwinName)})

						Convey("The platform should be added to the existing version", func() {
							So(err, ShouldBeNil)
							So(len(saved.Versions), ShouldEqual, 1)
							So(len(saved.Versions[0].Platforms), ShouldEqual, 2)
							So(saved.Versions[0].Platforms[1].System, ShouldEqual, "darwin")
							So(saved.Versions[0].Platforms[1].Architecture, ShouldEqual, "arm64")
						})
					})
				})
			})

			Convey("If the provider exists but the version is new", func() {
				mockRepository.
					On("Find", "registry.terraform.io", "hashicorp", "null").
					Return(fixtureProvider(), nil)

				mockResolver.
					On("Store", mock.MatchedBy(func(in *storage.StoreInput) bool {
						return in.KeyPrefix == "mirror/registry.terraform.io/hashicorp/null/3.2.5"
					})).
					Return("mirror/registry.terraform.io/hashicorp/null/3.2.5/"+archiveName, nil)

				var saved mirror.Provider
				mockRepository.
					On("Upsert", mock.AnythingOfType("mirror.Provider")).
					Run(func(args mock.Arguments) {
						saved, _ = args.Get(0).(mirror.Provider)
					}).
					Return(&mirror.Provider{}, nil)

				Convey("When the upload is requested", func() {
					err := service.Upload("registry.terraform.io", "hashicorp", "null", "3.2.5", metadata, []file.File{fixtureArchive(archiveName)})

					Convey("The version should be added to the existing provider", func() {
						So(err, ShouldBeNil)
						So(len(saved.Versions), ShouldEqual, 2)
						So(saved.Versions[1].Version, ShouldEqual, "3.2.5")
					})
				})
			})

			Convey("If storing an archive fails", func() {
				mockRepository.
					On("Find", "registry.terraform.io", "hashicorp", "null").
					Return(nil, errors.New(""))

				mockResolver.
					On("Store", mock.Anything).
					Return("", errors.New(""))

				Convey("When the upload is requested", func() {
					err := service.Upload("registry.terraform.io", "hashicorp", "null", "3.2.4", metadata, []file.File{fixtureArchive(archiveName)})

					Convey("An error should be returned and nothing saved", func() {
						So(err, ShouldNotBeNil)
						mockRepository.AssertNotCalled(t, "Upsert", mock.Anything)
					})
				})
			})
		})
	})
}

func TestMirrorDeleteVersion(t *testing.T) {
	Convey("Subject: Delete a mirrored provider version", t, func() {
		mockRepository := repositories.NewMockMirrorRepository(t)
		mockResolver := storage.NewMockResolver(t)

		service := &DefaultMirrorService{
			MirrorRepository: mockRepository,
			Resolver:         mockResolver,
		}

		Convey("If the provider does not exist", func() {
			mockRepository.
				On("Find", "registry.terraform.io", "hashicorp", "null").
				Return(nil, errors.New(""))

			Convey("When the deletion is requested", func() {
				err := service.DeleteVersion("registry.terraform.io", "hashicorp", "null", "3.2.4")

				Convey("An error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("If the provider exists", func() {
			p := fixtureProvider()
			mockRepository.
				On("Find", "registry.terraform.io", "hashicorp", "null").
				Return(p, nil)

			Convey("If the version does not exist", func() {
				Convey("When the deletion is requested", func() {
					err := service.DeleteVersion("registry.terraform.io", "hashicorp", "null", "9.9.9")

					Convey("An error should be returned and nothing purged", func() {
						So(err, ShouldNotBeNil)
						mockResolver.AssertNotCalled(t, "Purge", mock.Anything)
					})
				})
			})

			Convey("If the version exists", func() {
				mockResolver.
					On("Purge", p.Versions[0].Platforms[0].Location).
					Return(nil)

				mockRepository.
					On("DeleteVersion", p, "3.2.4").
					Return(nil)

				Convey("When the deletion is requested", func() {
					err := service.DeleteVersion("registry.terraform.io", "hashicorp", "null", "3.2.4")

					Convey("The packages should be purged and the version removed", func() {
						So(err, ShouldBeNil)
					})
				})
			})
		})
	})
}

func TestMirrorDelete(t *testing.T) {
	Convey("Subject: Delete a mirrored provider", t, func() {
		mockRepository := repositories.NewMockMirrorRepository(t)
		mockResolver := storage.NewMockResolver(t)

		service := &DefaultMirrorService{
			MirrorRepository: mockRepository,
			Resolver:         mockResolver,
		}

		Convey("If the provider does not exist", func() {
			mockRepository.
				On("Find", "registry.terraform.io", "hashicorp", "null").
				Return(nil, errors.New(""))

			Convey("When the deletion is requested", func() {
				err := service.Delete("registry.terraform.io", "hashicorp", "null")

				Convey("An error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("If the provider exists", func() {
			p := fixtureProvider()
			mockRepository.
				On("Find", "registry.terraform.io", "hashicorp", "null").
				Return(p, nil)

			mockResolver.
				On("Purge", p.Versions[0].Platforms[0].Location).
				Return(nil)

			mockRepository.
				On("Delete", p).
				Return(nil)

			Convey("When the deletion is requested", func() {
				err := service.Delete("registry.terraform.io", "hashicorp", "null")

				Convey("The packages should be purged and the provider removed", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})
}

func TestMirrorDeleteNamespace(t *testing.T) {
	Convey("Subject: Delete all mirrored providers under a namespace", t, func() {
		mockRepository := repositories.NewMockMirrorRepository(t)
		mockResolver := storage.NewMockResolver(t)

		service := &DefaultMirrorService{
			MirrorRepository: mockRepository,
			Resolver:         mockResolver,
		}

		Convey("If no provider exists under the namespace", func() {
			mockRepository.
				On("FindByNamespace", "registry.terraform.io", "hashicorp").
				Return([]mirror.Provider{}, nil)

			Convey("When the deletion is requested", func() {
				err := service.DeleteNamespace("registry.terraform.io", "hashicorp")

				Convey("An error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("If providers exist under the namespace", func() {
			first := fixtureProvider()
			second := fixtureProvider()
			second.Name = "random"
			second.Versions[0].Platforms[0].Location = "mirror/registry.terraform.io/hashicorp/random/3.2.4/random.zip"

			mockRepository.
				On("FindByNamespace", "registry.terraform.io", "hashicorp").
				Return([]mirror.Provider{*first, *second}, nil)

			mockResolver.On("Purge", first.Versions[0].Platforms[0].Location).Return(nil)
			mockResolver.On("Purge", second.Versions[0].Platforms[0].Location).Return(nil)

			mockRepository.On("Delete", mock.MatchedBy(func(p *mirror.Provider) bool { return p.Name == "null" })).Return(nil)
			mockRepository.On("Delete", mock.MatchedBy(func(p *mirror.Provider) bool { return p.Name == "random" })).Return(nil)

			Convey("When the deletion is requested", func() {
				err := service.DeleteNamespace("registry.terraform.io", "hashicorp")

				Convey("All providers should be purged and removed", func() {
					So(err, ShouldBeNil)
					mockRepository.AssertNumberOfCalls(t, "Delete", 2)
				})
			})
		})
	})
}

func TestMirrorDeleteHostname(t *testing.T) {
	Convey("Subject: Delete all mirrored providers under a hostname", t, func() {
		mockRepository := repositories.NewMockMirrorRepository(t)
		mockResolver := storage.NewMockResolver(t)

		service := &DefaultMirrorService{
			MirrorRepository: mockRepository,
			Resolver:         mockResolver,
		}

		Convey("If no provider exists under the hostname", func() {
			mockRepository.
				On("FindByHostname", "registry.terraform.io").
				Return([]mirror.Provider{}, nil)

			Convey("When the deletion is requested", func() {
				err := service.DeleteHostname("registry.terraform.io")

				Convey("An error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("If providers exist under the hostname", func() {
			p := fixtureProvider()

			mockRepository.
				On("FindByHostname", "registry.terraform.io").
				Return([]mirror.Provider{*p}, nil)

			mockResolver.On("Purge", p.Versions[0].Platforms[0].Location).Return(nil)
			mockRepository.On("Delete", mock.AnythingOfType("*mirror.Provider")).Return(nil)

			Convey("When the deletion is requested", func() {
				err := service.DeleteHostname("registry.terraform.io")

				Convey("All providers should be purged and removed", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})
}
