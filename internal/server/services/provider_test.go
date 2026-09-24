package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/repositories"
	"terralist/pkg/file"
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
					resp, err := providerService.Get(namespace, name, false)

					Convey("A provider should be returned", func() {
						So(err, ShouldBeNil)
						So(resp, ShouldNotBeNil)
					})
				})
			})

			Convey("If the provider has a version without signature material", func() {
				mockProviderRepository.
					On("Find", namespace, name).
					Return(&provider.Provider{
						Name: name,
						Versions: []provider.Version{
							{Version: "1.0.0", Protocols: "5.0", ShaSumsUrl: "providers/SHA256SUMS", ShaSumsSignatureUrl: "providers/SHA256SUMS.sig"},
							{Version: "1.1.0", Protocols: ""},
						},
					}, nil)

				Convey("When the service is queried", func() {
					resp, err := providerService.Get(namespace, name, false)

					Convey("Only the signed version should be listed", func() {
						So(err, ShouldBeNil)
						So(len(resp.Versions), ShouldEqual, 1)
						So(resp.Versions[0].Version, ShouldEqual, "1.0.0")
					})
				})
			})

			Convey("If the provider does not exist in the database", func() {
				mockProviderRepository.
					On("Find", namespace, name).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					resp, err := providerService.Get(namespace, name, false)

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
					info, err := providerService.GetVersion(namespace, name, version, system, architecture, false)

					Convey("An error should be returned", func() {
						So(info, ShouldBeNil)
						So(err, ShouldNotBeNil)
					})
				})
			})

			Convey("If the version has no signature material", func() {
				mockProviderRepository.
					On("FindVersionPlatform", namespace, name, version, system, architecture).
					Return(&provider.Platform{
						Location: "providers/linux.zip",
						Version:  provider.Version{Version: version},
					}, nil)

				Convey("When the service is queried", func() {
					info, err := providerService.GetVersion(namespace, name, version, system, architecture, false)

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
					On("GetByID", mock.AnythingOfType("uuid.UUID")).
					Return(&authority.Authority{}, nil)

				Convey("If the resolver is not set", func() {
					providerService.Resolver = nil

					location, _ := random.String(16)
					mockProviderPlatform.Location = location

					Convey("When the service is queried", func() {
						info, err := providerService.GetVersion(namespace, name, version, system, architecture, false)

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
							info, err := providerService.GetVersion(namespace, name, version, system, architecture, false)

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
							info, err := providerService.GetVersion(namespace, name, version, system, architecture, false)

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
							info, err := providerService.GetVersion(namespace, name, version, system, architecture, false)

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
							info, err := providerService.GetVersion(namespace, name, version, system, architecture, false)

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

func TestUploadProvider(t *testing.T) {
	Convey("Subject: Upload a provider version", t, func() {
		mockProviderRepository := repositories.NewMockProviderRepository(t)
		mockAuthorityService := NewMockAuthorityService(t)
		mockResolver := storage.NewMockResolver(t)
		mockFetcher := file.NewMockFetcher(t)

		providerService := &DefaultProviderService{
			ProviderRepository: mockProviderRepository,
			AuthorityService:   mockAuthorityService,
			Resolver:           mockResolver,
			Fetcher:            mockFetcher,
		}

		Convey("Given a provider DTO", func() {
			dto := provider.CreateProviderDTO{}

			Convey("If the version is not respecting the semantic format", func() {
				dto.Version = "100%-not-sem-ver-valid"

				Convey("When the service is queried", func() {
					err := providerService.Upload(&dto)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
					})
				})
			})

			Convey("If the version is respecting the semantic format", func() {
				dto.Version = "1.0.0"

				Convey("If the authority does not exist", func() {
					mockAuthorityService.
						On("GetByID", mock.AnythingOfType("uuid.UUID")).
						Return(nil, errors.New(""))

					Convey("When the service is queried", func() {
						err := providerService.Upload(&dto)

						Convey("An error should be returned", func() {
							So(err, ShouldNotBeNil)
						})
					})
				})

				Convey("If the authority exists", func() {
					mockAuthorityService.
						On("GetByID", mock.AnythingOfType("uuid.UUID")).
						Return(&authority.Authority{}, nil)

					Convey("If the provider exists and already has the given version", func() {
						mockProviderRepository.
							On("Find", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
							Return(&provider.Provider{
								Versions: []provider.Version{
									{
										Version: dto.Version,
									},
								},
							}, nil)

						Convey("When the service is queried", func() {
							err := providerService.Upload(&dto)

							Convey("An error should be returned", func() {
								So(err, ShouldNotBeNil)
							})
						})
					})

					similarTestData := []struct {
						Desc string
						Func func()
					}{
						{
							Desc: "If the provider exists and does not have the given version",
							Func: func() {
								mockProviderRepository.
									On("Find", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
									Return(&provider.Provider{}, nil)
							},
						},
						{
							Desc: "If the provider does not exist",
							Func: func() {
								mockProviderRepository.
									On("Find", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
									Return(nil, errors.New(""))
							},
						},
					}

					for _, td := range similarTestData {
						Convey(td.Desc, func() {
							td.Func()

							Convey("If the resolver is not set", func() {
								providerService.Resolver = nil

								mockProviderRepository.
									On("Upsert", mock.AnythingOfType("provider.Provider")).
									Return(&provider.Provider{}, nil)

								Convey("When the service is queried", func() {
									err := providerService.Upload(&dto)

									Convey("No error should be returned", func() {
										So(err, ShouldBeNil)
									})
								})
							})

							Convey("If the resolver is set", func() {
								dto.ShaSums.URL, _ = random.String(16)
								dto.ShaSums.SignatureURL, _ = random.String(16)

								binaryURL, _ := random.String(16)
								dto.Platforms = append(dto.Platforms, provider.CreatePlatformDTO{
									Location: binaryURL,
								})

								Convey("If the provider files cannot be downloaded", func() {
									mockFetcher.
										On(
											"FetchFile",
											mock.AnythingOfType("string"),
											mock.AnythingOfType("string"),
											mock.AnythingOfType("http.Header"),
										).
										Return(nil, nil, errors.New(""))

									Convey("When the service is queried", func() {
										err := providerService.Upload(&dto)

										Convey("An error should be returned", func() {
											So(err, ShouldNotBeNil)
										})
									})
								})

								Convey("If the provider files can be downloaded", func() {
									mockFetcher.
										On(
											"FetchFile",
											mock.AnythingOfType("string"),
											mock.AnythingOfType("string"),
											mock.AnythingOfType("http.Header"),
										).
										Return(file.NewEmptyFile("test.txt"), func() {}, nil)

									mockFetcher.
										On(
											"FetchFileChecksum",
											mock.AnythingOfType("string"),
											mock.AnythingOfType("string"),
											mock.AnythingOfType("string"),
											mock.AnythingOfType("http.Header"),
										).
										Return(file.NewEmptyFile("test-with-checksum.txt"), func() {}, nil)

									Convey("If the resolver cannot store the provider files", func() {
										mockResolver.
											On("Store", mock.AnythingOfType("*storage.StoreInput")).
											Return("", errors.New(""))

										Convey("When the service is queried", func() {
											err := providerService.Upload(&dto)

											Convey("An error should be returned", func() {
												So(err, ShouldNotBeNil)
											})
										})
									})

									Convey("If the resolver successfully stores the provider files", func() {
										mockResolver.
											On("Store", mock.AnythingOfType("*storage.StoreInput")).
											Return("", nil)

										mockProviderRepository.
											On("Upsert", mock.AnythingOfType("provider.Provider")).
											Return(&provider.Provider{}, nil)

										Convey("When the service is queried", func() {
											err := providerService.Upload(&dto)

											Convey("No error should be returned", func() {
												So(err, ShouldBeNil)
											})
										})
									})
								})
							})
						})
					}
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
					On("GetByID", authorityID).
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
					On("GetByID", authorityID).
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
					On("GetByID", authorityID).
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
					On("GetByID", authorityID).
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

func TestListMirrorVersions(t *testing.T) {
	Convey("Subject: List the versions of a provider for the network mirror", t, func() {
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
						Versions: []provider.Version{
							{Version: "1.0.0"},
							{Version: "1.1.0"},
						},
					}, nil)

				Convey("When the service is queried", func() {
					resp, err := providerService.ListMirrorVersions(namespace, name, false)

					Convey("Every version should map to an empty object", func() {
						So(err, ShouldBeNil)
						So(resp.Versions, ShouldResemble, map[string]struct{}{
							"1.0.0": {},
							"1.1.0": {},
						})
					})
				})
			})

			Convey("If the provider does not exist in the database", func() {
				mockProviderRepository.
					On("Find", namespace, name).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					resp, err := providerService.ListMirrorVersions(namespace, name, false)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
						So(resp, ShouldBeNil)
					})
				})
			})
		})
	})
}

func TestListMirrorArchives(t *testing.T) {
	Convey("Subject: List the installation packages of a provider version for the network mirror", t, func() {
		mockProviderRepository := repositories.NewMockProviderRepository(t)
		mockResolver := storage.NewMockResolver(t)

		providerService := &DefaultProviderService{
			ProviderRepository: mockProviderRepository,
			Resolver:           mockResolver,
		}

		Convey("Given a namespace, name and version", func() {
			namespace, _ := random.String(16)
			name, _ := random.String(16)

			Convey("If the provider does not exist in the database", func() {
				mockProviderRepository.
					On("Find", namespace, name).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					resp, err := providerService.ListMirrorArchives(namespace, name, "1.0.0", false)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
						So(resp, ShouldBeNil)
					})
				})
			})

			Convey("If the provider exists in the database", func() {
				mockProviderRepository.
					On("Find", namespace, name).
					Return(&provider.Provider{
						Name: name,
						Versions: []provider.Version{
							{
								Version: "1.0.0",
								Platforms: []provider.Platform{
									{System: "linux", Architecture: "amd64", Location: "providers/linux.zip", ShaSum: "aaaa", H1: "h1:linux"},
									{System: "darwin", Architecture: "arm64", Location: "providers/darwin.zip", ShaSum: "bbbb"},
								},
							},
						},
					}, nil)

				Convey("If the version does not exist", func() {
					Convey("When the service is queried", func() {
						resp, err := providerService.ListMirrorArchives(namespace, name, "9.9.9", false)

						Convey("An error should be returned", func() {
							So(err, ShouldNotBeNil)
							So(resp, ShouldBeNil)
						})
					})
				})

				Convey("If the version exists and the locations can be resolved", func() {
					mockResolver.On("Find", "providers/linux.zip").Return("https://storage/linux.zip", nil)
					mockResolver.On("Find", "providers/darwin.zip").Return("https://storage/darwin.zip", nil)

					Convey("When the service is queried", func() {
						resp, err := providerService.ListMirrorArchives(namespace, name, "1.0.0", false)

						Convey("Every platform should be listed with its resolved url, its h1 hash when known and its zh hash", func() {
							So(err, ShouldBeNil)
							So(resp.Archives, ShouldResemble, map[string]provider.MirrorArchiveDTO{
								"linux_amd64":  {URL: "https://storage/linux.zip", Hashes: []string{"h1:linux", "zh:aaaa"}},
								"darwin_arm64": {URL: "https://storage/darwin.zip", Hashes: []string{"zh:bbbb"}},
							})
						})
					})
				})

				Convey("If a location cannot be resolved", func() {
					mockResolver.On("Find", mock.Anything).Return("", errors.New(""))

					Convey("When the service is queried", func() {
						resp, err := providerService.ListMirrorArchives(namespace, name, "1.0.0", false)

						Convey("An error should be returned", func() {
							So(err, ShouldNotBeNil)
							So(resp, ShouldBeNil)
						})
					})
				})
			})
		})
	})
}

func TestListMirrorArchivesWithoutResolver(t *testing.T) {
	Convey("Subject: List the installation packages of a provider version without a storage resolver", t, func() {
		mockProviderRepository := repositories.NewMockProviderRepository(t)

		providerService := &DefaultProviderService{
			ProviderRepository: mockProviderRepository,
		}

		namespace, _ := random.String(16)
		name, _ := random.String(16)

		mockProviderRepository.
			On("Find", namespace, name).
			Return(&provider.Provider{
				Name: name,
				Versions: []provider.Version{
					{
						Version: "1.0.0",
						Platforms: []provider.Platform{
							{System: "linux", Architecture: "amd64", Location: "https://releases/linux.zip", ShaSum: "aaaa"},
						},
					},
				},
			}, nil)

		Convey("When the service is queried", func() {
			resp, err := providerService.ListMirrorArchives(namespace, name, "1.0.0", false)

			Convey("The stored location should be served as is", func() {
				So(err, ShouldBeNil)
				So(resp.Archives["linux_amd64"].URL, ShouldEqual, "https://releases/linux.zip")
			})
		})
	})
}

func TestListProviderVersions(t *testing.T) {
	Convey("Subject: List every version of a provider", t, func() {
		mockProviderRepository := repositories.NewMockProviderRepository(t)

		providerService := &DefaultProviderService{
			ProviderRepository: mockProviderRepository,
		}

		namespace, _ := random.String(16)
		name, _ := random.String(16)

		Convey("If the provider exists with signed and unsigned versions", func() {
			mockProviderRepository.
				On("Find", namespace, name).
				Return(&provider.Provider{
					Name: name,
					Versions: []provider.Version{
						{Version: "1.0.0", ShaSumsUrl: "providers/SHA256SUMS", ShaSumsSignatureUrl: "providers/SHA256SUMS.sig"},
						{Version: "1.1.0"},
					},
				}, nil)

			Convey("When the service is queried", func() {
				versions, err := providerService.ListVersions(namespace, name)

				Convey("Every version should be listed", func() {
					So(err, ShouldBeNil)
					So(versions, ShouldResemble, []string{"1.0.0", "1.1.0"})
				})
			})
		})

		Convey("If the provider does not exist", func() {
			mockProviderRepository.
				On("Find", namespace, name).
				Return(nil, errors.New(""))

			Convey("When the service is queried", func() {
				versions, err := providerService.ListVersions(namespace, name)

				Convey("An error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(versions, ShouldBeNil)
				})
			})
		})
	})
}

func TestUploadProviderPackages(t *testing.T) {
	Convey("Subject: Upload the packages of a provider version", t, func() {
		mockProviderRepository := repositories.NewMockProviderRepository(t)
		mockAuthorityService := NewMockAuthorityService(t)
		mockResolver := storage.NewMockResolver(t)

		providerService := &DefaultProviderService{
			ProviderRepository: mockProviderRepository,
			AuthorityService:   mockAuthorityService,
			Resolver:           mockResolver,
		}

		authorityID, _ := uuid.NewRandom()
		linux := []byte("linux package")
		darwin := []byte("darwin package")
		linuxSum := sha256Hex(linux)
		darwinSum := sha256Hex(darwin)

		newDTO := func() provider.PackagesUploadDTO {
			return provider.PackagesUploadDTO{
				AuthorityID: authorityID,
				Name:        "null",
				Version:     "3.2.4",
				Metadata: provider.MirrorArchivesDTO{
					Archives: map[string]provider.MirrorArchiveDTO{
						"linux_amd64":  {URL: "terraform-provider-null_3.2.4_linux_amd64.zip", Hashes: []string{"h1:linux"}},
						"darwin_arm64": {URL: "terraform-provider-null_3.2.4_darwin_arm64.zip", Hashes: []string{"h1:darwin"}},
					},
				},
				Archives: []file.File{
					file.NewInMemoryFile("terraform-provider-null_3.2.4_linux_amd64.zip", linux),
					file.NewInMemoryFile("terraform-provider-null_3.2.4_darwin_arm64.zip", darwin),
				},
			}
		}

		expectStore := func() *provider.Provider {
			var saved provider.Provider
			mockAuthorityService.On("GetByID", authorityID).Return(&authority.Authority{Name: "hashicorp"}, nil)
			mockProviderRepository.On("Find", "hashicorp", "null").Return(nil, errors.New("not found"))
			mockResolver.
				On("Store", mock.AnythingOfType("*storage.StoreInput")).
				Return(func(in *storage.StoreInput) (string, error) {
					return in.KeyPrefix + "/" + in.FileName, nil
				})
			mockProviderRepository.
				On("Upsert", mock.AnythingOfType("provider.Provider")).
				Run(func(args mock.Arguments) {
					saved, _ = args.Get(0).(provider.Provider)
				}).
				Return(&provider.Provider{}, nil)

			return &saved
		}

		Convey("Given packages without signature material", func() {
			dto := newDTO()
			saved := expectStore()

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("A mirror-only version should be stored with both hashes per platform", func() {
					So(err, ShouldBeNil)
					So(saved.AuthorityID, ShouldEqual, authorityID)
					So(saved.Name, ShouldEqual, "null")
					So(len(saved.Versions), ShouldEqual, 1)

					v := saved.Versions[0]
					So(v.Version, ShouldEqual, "3.2.4")
					So(v.MirrorOnly(), ShouldBeTrue)
					So(len(v.Platforms), ShouldEqual, 2)

					platforms := map[string]provider.Platform{}
					for _, p := range v.Platforms {
						platforms[p.String()] = p
					}
					So(platforms["linux_amd64"].ShaSum, ShouldEqual, linuxSum)
					So(platforms["linux_amd64"].H1, ShouldEqual, "h1:linux")
					So(platforms["linux_amd64"].Location, ShouldEqual, "providers/hashicorp/null/3.2.4/terraform-provider-null_3.2.4_linux_amd64.zip")
					So(platforms["darwin_arm64"].ShaSum, ShouldEqual, darwinSum)
				})
			})
		})

		Convey("Given packages with a matching SHA256SUMS, its signature and protocols", func() {
			dto := newDTO()
			dto.Protocols = []string{"5.0", "6.0"}
			dto.ShaSums = file.NewInMemoryFile("terraform-provider-null_3.2.4_SHA256SUMS", []byte(
				linuxSum+"  terraform-provider-null_3.2.4_linux_amd64.zip\n"+
					darwinSum+"  terraform-provider-null_3.2.4_darwin_arm64.zip\n",
			))
			dto.ShaSumsSignature = file.NewInMemoryFile("terraform-provider-null_3.2.4_SHA256SUMS.sig", []byte("signature"))
			saved := expectStore()

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("A registry-visible version should be stored", func() {
					So(err, ShouldBeNil)

					v := saved.Versions[0]
					So(v.MirrorOnly(), ShouldBeFalse)
					So(v.Protocols, ShouldEqual, "5.0,6.0")
					So(v.ShaSumsUrl, ShouldEqual, "providers/hashicorp/null/3.2.4/terraform-provider-null_3.2.4_SHA256SUMS")
					So(v.ShaSumsSignatureUrl, ShouldEqual, "providers/hashicorp/null/3.2.4/terraform-provider-null_3.2.4_SHA256SUMS.sig")
				})
			})
		})

		Convey("Given a SHA256SUMS that does not match a package", func() {
			dto := newDTO()
			dto.Protocols = []string{"5.0"}
			dto.ShaSums = file.NewInMemoryFile("SHA256SUMS", []byte(
				linuxSum+"  terraform-provider-null_3.2.4_linux_amd64.zip\n"+
					sha256Hex([]byte("tampered"))+"  terraform-provider-null_3.2.4_darwin_arm64.zip\n",
			))
			dto.ShaSumsSignature = file.NewInMemoryFile("SHA256SUMS.sig", []byte("signature"))
			mockAuthorityService.On("GetByID", authorityID).Return(&authority.Authority{Name: "hashicorp"}, nil)
			mockProviderRepository.On("Find", "hashicorp", "null").Return(nil, errors.New("not found"))

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("The upload should be rejected and nothing stored", func() {
					So(err, ShouldNotBeNil)
					mockResolver.AssertNotCalled(t, "Store", mock.Anything)
				})
			})
		})

		Convey("Given a SHA256SUMS that does not list a package", func() {
			dto := newDTO()
			dto.Protocols = []string{"5.0"}
			dto.ShaSums = file.NewInMemoryFile("SHA256SUMS", []byte(linuxSum+"  terraform-provider-null_3.2.4_linux_amd64.zip\n"))
			dto.ShaSumsSignature = file.NewInMemoryFile("SHA256SUMS.sig", []byte("signature"))
			mockAuthorityService.On("GetByID", authorityID).Return(&authority.Authority{Name: "hashicorp"}, nil)
			mockProviderRepository.On("Find", "hashicorp", "null").Return(nil, errors.New("not found"))

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("The upload should be rejected", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given a SHA256SUMS without its signature", func() {
			dto := newDTO()
			dto.Protocols = []string{"5.0"}
			dto.ShaSums = file.NewInMemoryFile("SHA256SUMS", []byte(linuxSum+"  terraform-provider-null_3.2.4_linux_amd64.zip\n"))

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("The upload should be rejected", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given a SHA256SUMS and signature without protocols", func() {
			dto := newDTO()
			dto.ShaSums = file.NewInMemoryFile("SHA256SUMS", []byte(linuxSum+"  terraform-provider-null_3.2.4_linux_amd64.zip\n"))
			dto.ShaSumsSignature = file.NewInMemoryFile("SHA256SUMS.sig", []byte("signature"))

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("The upload should be rejected", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given a package that is not listed in the metadata", func() {
			dto := newDTO()
			dto.Archives = append(dto.Archives, file.NewInMemoryFile("terraform-provider-null_3.2.4_windows_amd64.zip", []byte("windows")))
			mockAuthorityService.On("GetByID", authorityID).Return(&authority.Authority{Name: "hashicorp"}, nil)
			mockProviderRepository.On("Find", "hashicorp", "null").Return(nil, errors.New("not found"))

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("The upload should be rejected", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given metadata without an h1 hash", func() {
			dto := newDTO()
			dto.Metadata.Archives["linux_amd64"] = provider.MirrorArchiveDTO{URL: "terraform-provider-null_3.2.4_linux_amd64.zip"}
			saved := expectStore()

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("The platform should be stored without an h1 hash", func() {
					So(err, ShouldBeNil)
					for _, p := range saved.Versions[0].Platforms {
						if p.String() == "linux_amd64" {
							So(p.H1, ShouldBeEmpty)
						}
					}
				})
			})
		})

		Convey("Given no packages", func() {
			dto := newDTO()
			dto.Archives = nil

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("The upload should be rejected", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given an invalid version", func() {
			dto := newDTO()
			dto.Version = "latest"

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("The upload should be rejected", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given a version that already exists", func() {
			dto := newDTO()
			mockAuthorityService.On("GetByID", authorityID).Return(&authority.Authority{Name: "hashicorp"}, nil)
			mockProviderRepository.On("Find", "hashicorp", "null").Return(&provider.Provider{
				Name:     "null",
				Versions: []provider.Version{{Version: "3.2.4"}},
			}, nil)

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("The upload should be rejected", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given a provider that already has other versions", func() {
			dto := newDTO()
			var saved provider.Provider
			mockAuthorityService.On("GetByID", authorityID).Return(&authority.Authority{Name: "hashicorp"}, nil)
			mockProviderRepository.On("Find", "hashicorp", "null").Return(&provider.Provider{
				Name:     "null",
				Versions: []provider.Version{{Version: "3.2.3"}},
			}, nil)
			mockResolver.
				On("Store", mock.AnythingOfType("*storage.StoreInput")).
				Return(func(in *storage.StoreInput) (string, error) {
					return in.KeyPrefix + "/" + in.FileName, nil
				})
			mockProviderRepository.
				On("Upsert", mock.AnythingOfType("provider.Provider")).
				Run(func(args mock.Arguments) {
					saved, _ = args.Get(0).(provider.Provider)
				}).
				Return(&provider.Provider{}, nil)

			Convey("When the packages are uploaded", func() {
				err := providerService.UploadPackages(&dto)

				Convey("The version should be appended to the provider", func() {
					So(err, ShouldBeNil)
					So(len(saved.Versions), ShouldEqual, 2)
					So(saved.Versions[1].Version, ShouldEqual, "3.2.4")
				})
			})
		})
	})
}

func TestUploadProviderPackagesWithoutResolver(t *testing.T) {
	Convey("Subject: Upload the packages of a provider version without a storage resolver", t, func() {
		providerService := &DefaultProviderService{
			ProviderRepository: repositories.NewMockProviderRepository(t),
			AuthorityService:   NewMockAuthorityService(t),
		}

		dto := provider.PackagesUploadDTO{
			Name:     "null",
			Version:  "3.2.4",
			Archives: []file.File{file.NewInMemoryFile("terraform-provider-null_3.2.4_linux_amd64.zip", []byte("linux"))},
		}

		Convey("When the packages are uploaded", func() {
			err := providerService.UploadPackages(&dto)

			Convey("The upload should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func sha256Hex(content []byte) string {
	sum := sha256.Sum256(content)

	return hex.EncodeToString(sum[:])
}
