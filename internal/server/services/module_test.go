package services

import (
	"errors"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/internal/server/repositories"
	"terralist/pkg/storage"

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

// TODO: file.Fetch must be mocked
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

		Convey("Given a module DTO and a source download URL", func() {
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
