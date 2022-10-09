package services

import (
	"errors"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/repositories"
	"terralist/pkg/storage"

	"github.com/google/uuid"
	"github.com/mazen160/go-random"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestGetProvider(t *testing.T) {
	Convey("Subject: Find a provider", t, func() {
		mockProviderRepository := repositories.NewMockProviderRepository(t)

		providerService := &DefaultProviderService{
			ProviderRepository: mockProviderRepository,
		}

		Convey("Given a namespace and name", func() {
			namespace, _ := random.String(16)
			name, _ := random.String(16)

			Convey("If the provider exists in the database", func() {
				mockProviderRepository.
					On("Find", namespace, name).
					Return(&provider.Provider{
						Name: name,
					}, nil)

				Convey("When the service is queried", func() {
					resp, err := providerService.Get(namespace, name)

					Convey("A provider should be returned", func() {
						So(err, ShouldBeNil)
						So(resp, ShouldNotBeNil)
					})
				})
			})

			Convey("If the provider does not exist in the database", func() {
				mockProviderRepository.
					On("Find", namespace, name).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					resp, err := providerService.Get(namespace, name)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
						So(resp, ShouldBeNil)
					})
				})
			})
		})
	})
}

func TestGetProviderVersionDownloadInfo(t *testing.T) {
	Convey("Subject: Find a provider version download info", t, func() {
		mockProviderRepository := repositories.NewMockProviderRepository(t)
		mockAuthorityService := NewMockAuthorityService(t)
		mockResolver := storage.NewMockResolver(t)

		providerService := &DefaultProviderService{
			ProviderRepository: mockProviderRepository,
			AuthorityService:   mockAuthorityService,
			Resolver:           mockResolver,
		}

		Convey("Given a namespace, name, version, system and architecture", func() {
			namespace, _ := random.String(16)
			name, _ := random.String(16)
			version, _ := random.String(16)
			system, _ := random.String(16)
			architecture, _ := random.String(16)

			Convey("If the resource does not exist in the database", func() {
				mockProviderRepository.
					On("FindVersionPlatform", namespace, name, version, system, architecture).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					info, err := providerService.GetVersion(namespace, name, version, system, architecture)

					Convey("An error should be returned", func() {
						So(info, ShouldBeNil)
						So(err, ShouldNotBeNil)
					})
				})
			})

			Convey("If the resource exists in the database", func() {
				shaSumsKey, _ := random.String(16)
				shaSumsSigKey, _ := random.String(16)
				binaryKey, _ := random.String(16)

				mockProviderPlatform := provider.Platform{
					Location: binaryKey,
					Version: provider.Version{
						ShaSumsUrl:          shaSumsKey,
						ShaSumsSignatureUrl: shaSumsSigKey,
					},
				}

				mockProviderRepository.
					On("FindVersionPlatform", namespace, name, version, system, architecture).
					Return(&mockProviderPlatform, nil)

				mockAuthorityService.
					On("Get", mock.AnythingOfType("uuid.UUID")).
					Return(&authority.Authority{}, nil)

				Convey("If the resolver is not set", func() {
					providerService.Resolver = nil

					location, _ := random.String(16)
					mockProviderPlatform.Location = location

					Convey("When the service is queried", func() {
						info, err := providerService.GetVersion(namespace, name, version, system, architecture)

						Convey("A response with download info should be returned", func() {
							So(info, ShouldNotBeNil)
							So(err, ShouldBeNil)
							So(info.DownloadUrl, ShouldEqual, location)
						})
					})
				})

				Convey("If the resolver is set", func() {
					// Set by default

					Convey("If the shasum location cannot be resolved", func() {
						mockResolver.
							On("Find", shaSumsKey).
							Return("", errors.New(""))

						Convey("When the service is queried", func() {
							info, err := providerService.GetVersion(namespace, name, version, system, architecture)

							Convey("An error should be returned", func() {
								So(info, ShouldBeNil)
								So(err, ShouldNotBeNil)
							})
						})
					})

					Convey("If the shasum signature location cannot be resolved", func() {
						shaSumsLocation, _ := random.String(16)

						mockResolver.
							On("Find", shaSumsKey).
							Return(shaSumsLocation, nil)

						mockResolver.
							On("Find", shaSumsSigKey).
							Return("", errors.New(""))

						Convey("When the service is queried", func() {
							info, err := providerService.GetVersion(namespace, name, version, system, architecture)

							Convey("An error should be returned", func() {
								So(info, ShouldBeNil)
								So(err, ShouldNotBeNil)
							})
						})
					})

					Convey("If the binary location cannot be resolved", func() {
						shaSumsLocation, _ := random.String(16)
						shaSumsSigLocation, _ := random.String(16)

						mockResolver.
							On("Find", shaSumsKey).
							Return(shaSumsLocation, nil)

						mockResolver.
							On("Find", shaSumsSigKey).
							Return(shaSumsSigLocation, nil)

						mockResolver.
							On("Find", binaryKey).
							Return("", errors.New(""))

						Convey("When the service is queried", func() {
							info, err := providerService.GetVersion(namespace, name, version, system, architecture)

							Convey("An error should be returned", func() {
								So(info, ShouldBeNil)
								So(err, ShouldNotBeNil)
							})
						})
					})

					Convey("If shasums, shasums signature and the binary location can be resolved", func() {
						shaSumsLocation, _ := random.String(16)
						shaSumsSigLocation, _ := random.String(16)
						binaryLocation, _ := random.String(16)

						mockResolver.
							On("Find", shaSumsKey).
							Return(shaSumsLocation, nil)

						mockResolver.
							On("Find", shaSumsSigKey).
							Return(shaSumsSigLocation, nil)

						mockResolver.
							On("Find", binaryKey).
							Return(binaryLocation, nil)

						Convey("When the service is queried", func() {
							info, err := providerService.GetVersion(namespace, name, version, system, architecture)

							Convey("A response with download info should be returned", func() {
								So(info, ShouldNotBeNil)
								So(err, ShouldBeNil)
								So(info.ShaSumsUrl, ShouldEqual, shaSumsLocation)
								So(info.ShaSumsSignatureUrl, ShouldEqual, shaSumsSigLocation)
								So(info.DownloadUrl, ShouldEqual, binaryLocation)
							})
						})
					})
				})
			})
		})
	})
}

func TestDeleteProvider(t *testing.T) {
	Convey("Subject: Delete a provider", t, func() {
		mockProviderRepository := repositories.NewMockProviderRepository(t)
		mockAuthorityService := NewMockAuthorityService(t)
		mockResolver := storage.NewMockResolver(t)

		providerService := &DefaultProviderService{
			ProviderRepository: mockProviderRepository,
			AuthorityService:   mockAuthorityService,
			Resolver:           mockResolver,
		}

		Convey("Given an authority ID and a provider name", func() {
			authorityID, _ := uuid.NewRandom()
			name, _ := random.String(16)

			Convey("If the authority does not exist", func() {
				mockAuthorityService.
					On("Get", authorityID).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					err := providerService.Delete(authorityID, name)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
					})
				})
			})

			Convey("If the authority exists", func() {
				mockAuthorityService.
					On("Get", authorityID).
					Return(&authority.Authority{}, nil)

				Convey("If the provider does not exist", func() {
					mockProviderRepository.
						On("Find", mock.AnythingOfType("string"), name).
						Return(nil, errors.New(""))

					Convey("When the service is queried", func() {
						err := providerService.Delete(authorityID, name)

						Convey("An error should be returned", func() {
							So(err, ShouldNotBeNil)
						})
					})
				})

				Convey("If the provider exists", func() {
					mockProvider := provider.Provider{
						AuthorityID: authorityID,
						Name:        name,
						Versions: []provider.Version{
							{
								Platforms: []provider.Platform{
									{}, // Add one platform so we can mock the resolver purge call
								},
							},
						},
					}

					mockProviderRepository.
						On("Find", mock.AnythingOfType("string"), name).
						Return(&mockProvider, nil)

					mockProviderRepository.
						On("Delete", &mockProvider).
						Return(nil)

					Convey("If the resolver is not set", func() {
						providerService.Resolver = nil

						Convey("When the service is queried", func() {
							err := providerService.Delete(authorityID, name)

							Convey("No error should be returned", func() {
								So(err, ShouldBeNil)
							})
						})
					})

					Convey("If the resolver is set", func() {
						mockResolver.
							On("Purge", mock.AnythingOfType("string")).
							Return(nil)

						Convey("When the service is queried", func() {
							err := providerService.Delete(authorityID, name)

							Convey("No error should be returned", func() {
								So(err, ShouldBeNil)
							})
						})
					})
				})
			})
		})
	})
}

func TestDeleteProviderVersion(t *testing.T) {
	Convey("Subject: Delete a provider version", t, func() {
		mockProviderRepository := repositories.NewMockProviderRepository(t)
		mockAuthorityService := NewMockAuthorityService(t)
		mockResolver := storage.NewMockResolver(t)

		providerService := &DefaultProviderService{
			ProviderRepository: mockProviderRepository,
			AuthorityService:   mockAuthorityService,
			Resolver:           mockResolver,
		}

		Convey("Given an authority ID a provider name and version", func() {
			authorityID, _ := uuid.NewRandom()
			name, _ := random.String(16)
			version := "1.0.0"

			Convey("If the authority does not exist", func() {
				mockAuthorityService.
					On("Get", authorityID).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					err := providerService.DeleteVersion(authorityID, name, version)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
					})
				})
			})

			Convey("If the authority exists", func() {
				mockAuthorityService.
					On("Get", authorityID).
					Return(&authority.Authority{}, nil)

				Convey("If the provider does not exist", func() {
					mockProviderRepository.
						On("Find", mock.AnythingOfType("string"), name).
						Return(nil, errors.New(""))

					Convey("When the service is queried", func() {
						err := providerService.DeleteVersion(authorityID, name, version)

						Convey("An error should be returned", func() {
							So(err, ShouldNotBeNil)
						})
					})
				})

				Convey("If the provider exists", func() {
					mockProvider := provider.Provider{
						AuthorityID: authorityID,
						Name:        name,
						Versions: []provider.Version{
							{
								Platforms: []provider.Platform{
									{},
								},
							},
						},
					}

					mockProviderRepository.
						On("Find", mock.AnythingOfType("string"), name).
						Return(&mockProvider, nil)

					Convey("If the provider does not have the given version", func() {
						mockProvider.Versions[0].Version = "1.0.1" // Not the same with version

						Convey("When the service is queried", func() {
							err := providerService.DeleteVersion(authorityID, name, version)

							Convey("An error should be returned", func() {
								So(err, ShouldNotBeNil)
							})
						})
					})

					Convey("If the provider has the given version", func() {
						mockProvider.Versions[0].Version = version

						mockProviderRepository.
							On("DeleteVersion", &mockProvider, version).
							Return(nil)

						Convey("If the resolver is not set", func() {
							providerService.Resolver = nil

							Convey("When the service is queried", func() {
								err := providerService.DeleteVersion(authorityID, name, version)

								Convey("No error should be returned while trying to delete the provider version", func() {
									So(err, ShouldBeNil)
								})
							})
						})

						Convey("If the resolver is set", func() {
							mockResolver.
								On("Purge", mock.AnythingOfType("string")).
								Return(nil)

							Convey("When the service is queried", func() {
								err := providerService.DeleteVersion(authorityID, name, version)

								Convey("No error should be returned while trying to delete the provider version", func() {
									So(err, ShouldBeNil)
								})
							})
						})
					})
				})
			})
		})
	})
}
