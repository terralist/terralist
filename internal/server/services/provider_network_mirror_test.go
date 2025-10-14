package services

import (
	"errors"
	"testing"

	"terralist/internal/server/models/provider"
	"terralist/internal/server/repositories"
	"terralist/pkg/storage"

	"github.com/mazen160/go-random"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetVersionAllPlatforms(t *testing.T) {
	Convey("Subject: Get all platforms for a provider version", t, func() {
		mockProviderRepository := repositories.NewMockProviderRepository(t)
		mockResolver := storage.NewMockResolver(t)

		providerService := &DefaultProviderService{
			ProviderRepository: mockProviderRepository,
			Resolver:           mockResolver,
		}

		Convey("Given a namespace, name, and version", func() {
			namespace, _ := random.String(16)
			name, _ := random.String(16)
			version := "1.0.0"

			Convey("If the provider does not exist", func() {
				mockProviderRepository.
					On("Find", namespace, name).
					Return(nil, errors.New("provider not found"))

				Convey("When the service is queried", func() {
					result, err := providerService.GetVersionAllPlatforms(namespace, name, version)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
						So(result, ShouldBeNil)
					})
				})
			})

			Convey("If the provider exists", func() {
				Convey("If the version does not exist", func() {
					mockProviderRepository.
						On("Find", namespace, name).
						Return(&provider.Provider{
							Name: name,
							Versions: []provider.Version{
								{Version: "2.0.0"},
							},
						}, nil)

					Convey("When the service is queried", func() {
						result, err := providerService.GetVersionAllPlatforms(namespace, name, version)

						Convey("An error should be returned", func() {
							So(err, ShouldNotBeNil)
							So(result, ShouldBeNil)
							So(err.Error(), ShouldContainSubstring, "not found")
						})
					})
				})

				Convey("If the version exists", func() {
					Convey("With empty protocols", func() {
						mockProviderRepository.
							On("Find", namespace, name).
							Return(&provider.Provider{
								Name: name,
								Versions: []provider.Version{
									{
										Version:   version,
										Protocols: "",
										Platforms: []provider.Platform{
											{
												System:       "linux",
												Architecture: "amd64",
												Location:     "key1",
												ShaSum:       "abc123",
											},
										},
									},
								},
							}, nil)

						Convey("If the resolver is not set", func() {
							providerService.Resolver = nil

							Convey("When the service is queried", func() {
								result, err := providerService.GetVersionAllPlatforms(namespace, name, version)

								Convey("A valid response should be returned with raw locations", func() {
									So(err, ShouldBeNil)
									So(result, ShouldNotBeNil)
									So(result.Version, ShouldEqual, version)
									So(result.Protocols, ShouldResemble, []string{})
									So(len(result.Platforms), ShouldEqual, 1)
									So(result.Platforms[0].OS, ShouldEqual, "linux")
									So(result.Platforms[0].Arch, ShouldEqual, "amd64")
									So(result.Platforms[0].DownloadURL, ShouldEqual, "key1")
									So(result.Platforms[0].Shasum, ShouldEqual, "abc123")
								})
							})
						})

						Convey("If the resolver is set", func() {
							Convey("If the resolver fails to resolve a platform location", func() {
								mockResolver.
									On("Find", "key1").
									Return("", errors.New("resolution failed"))

								Convey("When the service is queried", func() {
									result, err := providerService.GetVersionAllPlatforms(namespace, name, version)

									Convey("An error should be returned", func() {
										So(err, ShouldNotBeNil)
										So(result, ShouldBeNil)
										So(err.Error(), ShouldContainSubstring, "could not resolve")
									})
								})
							})

							Convey("If the resolver successfully resolves platform locations", func() {
								resolvedURL, _ := random.String(16)
								mockResolver.
									On("Find", "key1").
									Return(resolvedURL, nil)

								Convey("When the service is queried", func() {
									result, err := providerService.GetVersionAllPlatforms(namespace, name, version)

									Convey("A valid response with resolved URLs should be returned", func() {
										So(err, ShouldBeNil)
										So(result, ShouldNotBeNil)
										So(result.Platforms[0].DownloadURL, ShouldEqual, resolvedURL)
									})
								})
							})
						})
					})

					Convey("With multiple protocols", func() {
						mockProviderRepository.
							On("Find", namespace, name).
							Return(&provider.Provider{
								Name: name,
								Versions: []provider.Version{
									{
										Version:   version,
										Protocols: "5.0,6.0",
										Platforms: []provider.Platform{
											{
												System:       "darwin",
												Architecture: "arm64",
												Location:     "key2",
												ShaSum:       "def456",
											},
											{
												System:       "linux",
												Architecture: "amd64",
												Location:     "key3",
												ShaSum:       "ghi789",
											},
										},
									},
								},
							}, nil)

						providerService.Resolver = nil

						Convey("When the service is queried", func() {
							result, err := providerService.GetVersionAllPlatforms(namespace, name, version)

							Convey("Protocols should be correctly split", func() {
								So(err, ShouldBeNil)
								So(result, ShouldNotBeNil)
								So(result.Protocols, ShouldResemble, []string{"5.0", "6.0"})
							})

							Convey("All platforms should be included", func() {
								So(len(result.Platforms), ShouldEqual, 2)
								So(result.Platforms[0].OS, ShouldEqual, "darwin")
								So(result.Platforms[0].Arch, ShouldEqual, "arm64")
								So(result.Platforms[1].OS, ShouldEqual, "linux")
								So(result.Platforms[1].Arch, ShouldEqual, "amd64")
							})
						})
					})

					Convey("With ShaSumsUrl", func() {
						shaSumsLocation := "providers/test/provider/1.0.0/SHA256SUMS"
						resolvedShaSumsURL := "https://s3.example.com/providers/test/provider/1.0.0/SHA256SUMS?signed"

						mockProviderRepository.
							On("Find", namespace, name).
							Return(&provider.Provider{
								Name: name,
								Versions: []provider.Version{
									{
										Version:    version,
										Protocols:  "5.0",
										ShaSumsUrl: shaSumsLocation,
										Platforms: []provider.Platform{
											{
												System:       "linux",
												Architecture: "amd64",
												Location:     "key1",
												ShaSum:       "abc123",
											},
										},
									},
								},
							}, nil)

						Convey("If resolver is set", func() {
							mockResolver.
								On("Find", shaSumsLocation).
								Return(resolvedShaSumsURL, nil)

							mockResolver.
								On("Find", "key1").
								Return("https://s3.example.com/key1", nil)

							Convey("When the service is queried", func() {
								result, err := providerService.GetVersionAllPlatforms(namespace, name, version)

								Convey("ShaSumsUrl should be resolved", func() {
									So(err, ShouldBeNil)
									So(result, ShouldNotBeNil)
									So(result.ShaSumsUrl, ShouldEqual, resolvedShaSumsURL)
								})
							})
						})

						Convey("If resolver fails to resolve ShaSumsUrl", func() {
							mockResolver.
								On("Find", shaSumsLocation).
								Return("", errors.New("shasums resolution failed"))

							Convey("When the service is queried", func() {
								result, err := providerService.GetVersionAllPlatforms(namespace, name, version)

								Convey("An error should be returned", func() {
									So(err, ShouldNotBeNil)
									So(result, ShouldBeNil)
									So(err.Error(), ShouldContainSubstring, "could not resolve shasums")
								})
							})
						})
					})
				})
			})
		})
	})
}
