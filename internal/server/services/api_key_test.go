package services

import (
	"fmt"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/repositories"
	"terralist/pkg/auth"
	"terralist/pkg/database/entity"

	mockRepositories "terralist/mocks/server/repositories"
	mockServices "terralist/mocks/server/services"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestGetUserDetails(t *testing.T) {
	Convey("Subject: Finding user details for a given API key", t, func() {
		mockAuthorityService := mockServices.NewAuthorityService(t)
		mockApiKeyRepository := mockRepositories.NewApiKeyRepository(t)

		apiKeyService := &DefaultApiKeyService{
			AuthorityService: mockAuthorityService,
			ApiKeyRepository: mockApiKeyRepository,
		}

		Convey("Given an invalid API key", func() {
			apiKey := "100%-not-valid-uuid"

			Convey("When the service is queried", func() {
				user, err := apiKeyService.GetUserDetails(apiKey)

				Convey("Then an invalid key error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(user, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "invalid UUID")
				})
			})
		})

		Convey("Given a valid but expired API key", func() {
			apiKey := uuid.Must(uuid.NewRandom())
			apiKeyStr := apiKey.String()

			mockApiKeyRepository.On("Find", apiKey).Return(nil, repositories.ErrApiKeyExpired)

			Convey("When the service is queried", func() {
				user, err := apiKeyService.GetUserDetails(apiKeyStr)

				Convey("Then a expire error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(user, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, repositories.ErrApiKeyExpired.Error())
				})
			})
		})

		Convey("Given a valid API key", func() {
			apiKey, _ := uuid.NewRandom()
			apiKeyStr := apiKey.String()
			authorityID, _ := uuid.NewRandom()
			userEmail := "test@example.com"

			mockApiKeyRepository.On("Find", apiKey).Return(&auth.ApiKey{AuthorityID: authorityID}, nil)

			Convey("If the API key is associated to an invalid authority", func() {
				mockAuthorityService.On("GetByID", authorityID).Return(nil, repositories.ErrNotFound)

				Convey("When the service is queried", func() {
					user, err := apiKeyService.GetUserDetails(apiKeyStr)

					Convey("Then a not found error should be returned", func() {
						So(err, ShouldNotBeNil)
						So(user, ShouldBeNil)
						So(err.Error(), ShouldContainSubstring, repositories.ErrNotFound.Error())
					})
				})
			})

			Convey("If the API key is associated to a valid authority", func() {
				mockAuthorityService.On("GetByID", authorityID).Return(&authority.Authority{Owner: userEmail}, nil)

				Convey("When the service is queried", func() {
					user, err := apiKeyService.GetUserDetails(apiKeyStr)

					Convey("Then the user e-mail and the authority ID should be returned successfully", func() {
						So(err, ShouldBeNil)
						So(user.Email, ShouldEqual, userEmail)
						So(user.AuthorityID, ShouldEqual, authorityID.String())
					})
				})
			})
		})
	})
}

func TestGrant(t *testing.T) {
	Convey("Subject: Generating a new API key", t, func() {
		mockAuthorityService := mockServices.NewAuthorityService(t)
		mockApiKeyRepository := mockRepositories.NewApiKeyRepository(t)

		apiKeyService := &DefaultApiKeyService{
			AuthorityService: mockAuthorityService,
			ApiKeyRepository: mockApiKeyRepository,
		}

		Convey("Given a valid authority ID", func() {
			authorityID, _ := uuid.NewRandom()
			apiKeyID, _ := uuid.NewRandom()

			Convey("If no expiration is given", func() {
				expireIn := 0

				Convey("When the service is queried", func() {
					mockApiKeyRepository.
						On("Create", &auth.ApiKey{AuthorityID: authorityID}).
						Return(&auth.ApiKey{Entity: entity.Entity{ID: apiKeyID}}, nil)

					apiKey, err := apiKeyService.Grant(authorityID, expireIn)

					Convey("Then a valid API key should be returned", func() {
						So(err, ShouldBeNil)
						So(apiKey, ShouldNotBeNil)
						So(apiKey, ShouldEqual, apiKeyID.String())
						So(func() { uuid.MustParse(apiKey) }, ShouldNotPanic)
					})
				})
			})

			testData := []int{1, 2, 4, 8, 12, 18, 24, 48, 72}
			multipliers := []int{1, -1}

			for _, expireIn := range testData {
				for _, multiplier := range multipliers {
					expireIn *= multiplier

					Convey(fmt.Sprintf("If expiration is set to %d hours", expireIn), func() {

						Convey("When the service is queried", func() {
							mockApiKeyRepository.
								On("Create", mock.AnythingOfType("*auth.ApiKey")).
								Return(&auth.ApiKey{Entity: entity.Entity{ID: apiKeyID}}, nil)

							apiKey, err := apiKeyService.Grant(authorityID, expireIn)

							Convey("Then a valid API key should be returned", func() {
								So(err, ShouldBeNil)
								So(apiKey, ShouldNotBeNil)
								So(apiKey, ShouldEqual, apiKeyID.String())
								So(func() { uuid.MustParse(apiKey) }, ShouldNotPanic)
							})
						})
					})
				}
			}
		})
	})
}

func TestRevoke(t *testing.T) {
	Convey("Subject: Revoking access from an existing API key", t, func() {
		mockAuthorityService := mockServices.NewAuthorityService(t)
		mockApiKeyRepository := mockRepositories.NewApiKeyRepository(t)

		apiKeyService := &DefaultApiKeyService{
			AuthorityService: mockAuthorityService,
			ApiKeyRepository: mockApiKeyRepository,
		}

		Convey("Given an invalid API key", func() {
			apiKey := "100%-not-valid-uuid"

			Convey("When the service is queried", func() {
				err := apiKeyService.Revoke(apiKey)

				Convey("Then an invalid key error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "invalid UUID")
				})
			})

		})

		Convey("Given a valid API key", func() {
			apiKey, _ := uuid.NewRandom()
			apiKeyStr := apiKey.String()

			mockApiKeyRepository.On("Delete", apiKey).Return(nil)

			Convey("When the service is queried", func() {
				err := apiKeyService.Revoke(apiKeyStr)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})
}
