package services

import (
	"errors"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/internal/server/repositories"
	"terralist/pkg/storage"

	"github.com/google/uuid"
	"github.com/mazen160/go-random"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestGetModule(t *testing.T) {
	Convey("Subject: Find a module", t, func() {
		mockModuleRepository := repositories.NewMockModuleRepository(t)

		moduleService := &DefaultModuleService{
			ModuleRepository: mockModuleRepository,
		}

		Convey("Given a namespace, name, and provider", func() {
			namespace, _ := random.String(16)
			name, _ := random.String(16)
			provider, _ := random.String(16)

			Convey("If the module exists in the database", func() {
				mockModuleRepository.
					On("Find", namespace, name, provider).
					Return(&module.Module{
						Name:     name,
						Provider: provider,
					}, nil)

				Convey("When the service is queried", func() {
					resp, err := moduleService.Get(namespace, name, provider)

					Convey("A module should be returned", func() {
						So(err, ShouldBeNil)
						So(resp, ShouldNotBeNil)
						So(len(resp.Modules), ShouldEqual, 1)
					})
				})
			})

			Convey("If the module does not exist in the database", func() {
				mockModuleRepository.
					On("Find", namespace, name, provider).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					resp, err := moduleService.Get(namespace, name, provider)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
						So(resp, ShouldBeNil)
					})
				})
			})
		})
	})
}

func TestGetModuleDownloadLocation(t *testing.T) {
	Convey("Subject: Find a module", t, func() {
		mockModuleRepository := repositories.NewMockModuleRepository(t)
		mockResolver := storage.NewMockResolver(t)

		moduleService := &DefaultModuleService{
			ModuleRepository: mockModuleRepository,
			Resolver:         mockResolver,
		}

		Convey("Given a namespace, name, provider, and version", func() {
			namespace, _ := random.String(16)
			name, _ := random.String(16)
			provider, _ := random.String(16)
			version, _ := random.String(16)

			Convey("If the module exists in the database", func() {
				locationKey, _ := random.String(16)

				mockModuleRepository.
					On("FindVersionLocation", namespace, name, provider, version).
					Return(&locationKey, nil)

				Convey("If the resolver is not set", func() {
					moduleService.Resolver = nil

					Convey("When the service is queried", func() {
						url, err := moduleService.GetVersion(namespace, name, provider, version)

						Convey("A download URL should be returned", func() {
							So(url, ShouldNotBeNil)
							So(err, ShouldBeNil)
							So(*url, ShouldEqual, locationKey)
						})
					})
				})

				Convey("If the resolver can resolve the location path", func() {
					location, _ := random.String(16)

					mockResolver.
						On("Find", locationKey).
						Return(location, nil)

					Convey("When the service is queried", func() {
						url, err := moduleService.GetVersion(namespace, name, provider, version)

						Convey("A download URL should be returned", func() {
							So(url, ShouldNotBeNil)
							So(err, ShouldBeNil)
							So(*url, ShouldEqual, location)
						})
					})
				})

				Convey("If the resolver cannot resolve the location path", func() {
					mockResolver.
						On("Find", locationKey).
						Return("", errors.New(""))

					Convey("When the service is queried", func() {
						url, err := moduleService.GetVersion(namespace, name, provider, version)

						Convey("An error should be returned", func() {
							So(url, ShouldBeNil)
							So(err, ShouldNotBeNil)
						})
					})
				})
			})

			Convey("If the module does not exist in the database", func() {
				mockModuleRepository.
					On("FindVersionLocation", namespace, name, provider, version).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					url, err := moduleService.GetVersion(namespace, name, provider, version)

					Convey("An error should be returned", func() {
						So(url, ShouldBeNil)
						So(err, ShouldNotBeNil)
					})
				})
			})
		})
	})
}

func TestUploadModule(t *testing.T) {
	Convey("Subject: Upload a new module version", t, func() {
		mockModuleRepository := repositories.NewMockModuleRepository(t)
		mockAuthorityService := NewMockAuthorityService(t)
		mockResolver := storage.NewMockResolver(t)

		moduleService := &DefaultModuleService{
			ModuleRepository: mockModuleRepository,
			AuthorityService: mockAuthorityService,
			Resolver:         mockResolver,
		}

		// TODO: file.Fetch must be mocked
		SkipConvey("Given a module DTO and a source download URL", func() {
			dto := module.CreateDTO{}
			url, _ := random.String(16)

			Convey("If the version is not respecting the semantic format", func() {
				dto.Version = "100%-not-sem-ver-valid"

				Convey("When the service is queried", func() {
					err := moduleService.Upload(&dto, url)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
					})
				})
			})

			Convey("If the version is respecting the semantic format", func() {
				dto.Version = "1.0.0"

				Convey("If the authority does not exist", func() {
					mockAuthorityService.
						On("Get", mock.AnythingOfType("uuid.UUID")).
						Return(nil, errors.New(""))

					Convey("When the service is queried", func() {
						err := moduleService.Upload(&dto, url)

						Convey("An error should be returned", func() {
							So(err, ShouldNotBeNil)
						})
					})
				})

				Convey("If the authority exists", func() {
					mockAuthorityService.
						On("Get", mock.AnythingOfType("uuid.UUID")).
						Return(&authority.Authority{}, nil)

					Convey("If the module exists and already have the given version", func() {
						mockModuleRepository.
							On("Find", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
							Return(&module.Module{
								Versions: []module.Version{
									{
										Version: dto.Version,
									},
								},
							}, nil)

						Convey("When the service is queried", func() {
							err := moduleService.Upload(&dto, url)

							Convey("An error should be returned", func() {
								So(err, ShouldNotBeNil)
							})
						})
					})

					Convey("If the module exists and does not have the given version", func() {
						mockModuleRepository.
							On("Find", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
							Return(&module.Module{}, nil)

						Convey("If the resolver is not set", func() {
							moduleService.Resolver = nil

							mockModuleRepository.
								On("Upsert", mock.AnythingOfType("module.Module")).
								Return(&module.Module{}, nil)

							Convey("When the service is queried", func() {
								err := moduleService.Upload(&dto, url)

								Convey("No error should be returned", func() {
									So(err, ShouldBeNil)
								})
							})
						})

						Convey("If the resolver fails to store the module files", func() {
							mockResolver.
								On("Store", mock.AnythingOfType("*storage.StoreInput")).
								Return("", errors.New(""))

							Convey("When the service is queried", func() {
								err := moduleService.Upload(&dto, url)

								Convey("An error should be returned", func() {
									So(err, ShouldNotBeNil)
								})
							})
						})

						Convey("If the resolver is successfully stores the module files", func() {
							location, _ := random.String(16)

							mockResolver.
								On("Store", mock.AnythingOfType("*storage.StoreInput")).
								Return(location, nil)

							mockModuleRepository.
								On("Upsert", mock.AnythingOfType("module.Module")).
								Return(&module.Module{}, nil)

							Convey("When the service is queried", func() {
								err := moduleService.Upload(&dto, url)

								Convey("No error should be returned", func() {
									So(err, ShouldBeNil)
								})
							})
						})
					})

					Convey("If the module does not exist", func() {
						mockModuleRepository.
							On("Find", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
							Return(nil, errors.New(""))

						Convey("If the resolver is not set", func() {
							moduleService.Resolver = nil

							mockModuleRepository.
								On("Upsert", mock.AnythingOfType("module.Module")).
								Return(&module.Module{}, nil)

							Convey("When the service is queried", func() {
								err := moduleService.Upload(&dto, url)

								Convey("No error should be returned", func() {
									So(err, ShouldBeNil)
								})
							})
						})

						Convey("If the resolver fails to store the module files", func() {
							mockResolver.
								On("Store", mock.AnythingOfType("*storage.StoreInput")).
								Return("", errors.New(""))

							Convey("When the service is queried", func() {
								err := moduleService.Upload(&dto, url)

								Convey("An error should be returned", func() {
									So(err, ShouldNotBeNil)
								})
							})
						})

						Convey("If the resolver is successfully stores the module files", func() {
							location, _ := random.String(16)

							mockResolver.
								On("Store", mock.AnythingOfType("*storage.StoreInput")).
								Return(location, nil)

							mockModuleRepository.
								On("Upsert", mock.AnythingOfType("module.Module")).
								Return(&module.Module{}, nil)

							Convey("When the service is queried", func() {
								err := moduleService.Upload(&dto, url)

								Convey("No error should be returned", func() {
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

func TestDeleteModule(t *testing.T) {
	Convey("Subject: Delete a module", t, func() {
		mockModuleRepository := repositories.NewMockModuleRepository(t)
		mockAuthorityService := NewMockAuthorityService(t)
		mockResolver := storage.NewMockResolver(t)

		moduleService := &DefaultModuleService{
			ModuleRepository: mockModuleRepository,
			AuthorityService: mockAuthorityService,
			Resolver:         mockResolver,
		}

		Convey("Given an authority ID a module name and provider", func() {
			authorityID, _ := uuid.NewRandom()
			name, _ := random.String(16)
			provider, _ := random.String(16)

			Convey("If the authority does not exist", func() {
				mockAuthorityService.
					On("Get", authorityID).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					err := moduleService.Delete(authorityID, name, provider)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
					})
				})
			})

			Convey("If the authority exists", func() {
				mockAuthorityService.
					On("Get", authorityID).
					Return(&authority.Authority{}, nil)

				Convey("If the module does not exist", func() {
					mockModuleRepository.
						On("Find", mock.AnythingOfType("string"), name, provider).
						Return(nil, errors.New(""))

					Convey("When the service is queried", func() {
						err := moduleService.Delete(authorityID, name, provider)

						Convey("An error should be returned", func() {
							So(err, ShouldNotBeNil)
						})
					})
				})

				Convey("If the module exists", func() {
					mockModule := module.Module{
						AuthorityID: authorityID,
						Name:        name,
						Provider:    provider,
						Versions: []module.Version{
							{}, // Add one version so we can mock the resolver purge call
						},
					}

					mockModuleRepository.
						On("Find", mock.AnythingOfType("string"), name, provider).
						Return(&mockModule, nil)

					mockModuleRepository.
						On("Delete", &mockModule).
						Return(nil)

					Convey("If the resolver is not set", func() {
						moduleService.Resolver = nil

						Convey("When the service is queried", func() {
							err := moduleService.Delete(authorityID, name, provider)

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
							err := moduleService.Delete(authorityID, name, provider)

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

func TestDeleteModuleVersion(t *testing.T) {
	Convey("Subject: Delete a module version", t, func() {
		mockModuleRepository := repositories.NewMockModuleRepository(t)
		mockAuthorityService := NewMockAuthorityService(t)
		mockResolver := storage.NewMockResolver(t)

		moduleService := &DefaultModuleService{
			ModuleRepository: mockModuleRepository,
			AuthorityService: mockAuthorityService,
			Resolver:         mockResolver,
		}

		Convey("Given an authority ID a module name, provider and version", func() {
			authorityID, _ := uuid.NewRandom()
			name, _ := random.String(16)
			provider, _ := random.String(16)
			version := "1.0.0"

			Convey("If the authority does not exist", func() {
				mockAuthorityService.
					On("Get", authorityID).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					err := moduleService.DeleteVersion(authorityID, name, provider, version)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
					})
				})
			})

			Convey("If the authority exists", func() {
				mockAuthorityService.
					On("Get", authorityID).
					Return(&authority.Authority{}, nil)

				Convey("If the module does not exist", func() {
					mockModuleRepository.
						On("Find", mock.AnythingOfType("string"), name, provider).
						Return(nil, errors.New(""))

					Convey("When the service is queried", func() {
						err := moduleService.DeleteVersion(authorityID, name, provider, version)

						Convey("An error should be returned", func() {
							So(err, ShouldNotBeNil)
						})
					})
				})

				Convey("If the module exists", func() {
					mockModule := module.Module{
						AuthorityID: authorityID,
						Name:        name,
						Provider:    provider,
						Versions: []module.Version{
							{},
						},
					}

					mockModuleRepository.
						On("Find", mock.AnythingOfType("string"), name, provider).
						Return(&mockModule, nil)

					Convey("If the module does not have the given version", func() {
						mockModule.Versions[0].Version = "1.0.1" // Not the same with version

						Convey("When the service is queried", func() {
							err := moduleService.DeleteVersion(authorityID, name, provider, version)

							Convey("An error should be returned", func() {
								So(err, ShouldNotBeNil)
							})
						})
					})

					Convey("If the module has the given version", func() {
						mockModule.Versions[0].Version = version

						Convey("If the resolver is not set", func() {
							moduleService.Resolver = nil

							Convey("If the module has only one version", func() {
								// Only one is set by default

								mockModuleRepository.
									On("Delete", &mockModule).
									Return(nil)

								Convey("When the service is queried", func() {
									err := moduleService.DeleteVersion(authorityID, name, provider, version)

									Convey("No error should be returned, while trying to delete the module", func() {
										So(err, ShouldBeNil)
									})
								})
							})

							Convey("If the module has more than one version", func() {
								mockModule.Versions = append(mockModule.Versions, module.Version{})

								mockModuleRepository.
									On("DeleteVersion", mock.AnythingOfType("*module.Version")).
									Return(nil)

								Convey("When the service is queried", func() {
									err := moduleService.DeleteVersion(authorityID, name, provider, version)

									Convey("No error should be returned while trying to delete the module version", func() {
										So(err, ShouldBeNil)
									})
								})
							})
						})

						Convey("If the resolver is set", func() {
							// Set by default

							Convey("If the module has only one version", func() {
								// Only one is set by default

								mockResolver.
									On("Purge", mock.AnythingOfType("string")).
									Return(nil)

								mockModuleRepository.
									On("Delete", &mockModule).
									Return(nil)

								Convey("When the service is queried", func() {
									err := moduleService.DeleteVersion(authorityID, name, provider, version)

									Convey("No error should be returned, while trying to delete the module", func() {
										So(err, ShouldBeNil)
									})
								})
							})

							Convey("If the module has more than one version", func() {
								mockModule.Versions = append(mockModule.Versions, module.Version{})

								mockResolver.
									On("Purge", mock.AnythingOfType("string")).
									Return(nil)

								mockModuleRepository.
									On("DeleteVersion", mock.AnythingOfType("*module.Version")).
									Return(nil)

								Convey("When the service is queried", func() {
									err := moduleService.DeleteVersion(authorityID, name, provider, version)

									Convey("No error should be returned while trying to delete the module version", func() {
										So(err, ShouldBeNil)
									})
								})
							})
						})
					})
				})
			})
		})
	})
}
