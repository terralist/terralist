package services

import (
	"fmt"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/repositories"
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestGetUserDetails(t *testing.T) {
	Convey("Subject: Finding user details for a given API key", t, func() {
		mockAuthorityService := NewMockAuthorityService(t)
		mockApiKeyRepository := repositories.NewMockApiKeyRepository(t)

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

			mockApiKeyRepository.On("Find", apiKey).Return(&authority.ApiKey{AuthorityID: authorityID}, nil)

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
		mockAuthorityService := NewMockAuthorityService(t)
		mockApiKeyRepository := repositories.NewMockApiKeyRepository(t)

		apiKeyService := &DefaultApiKeyService{
			AuthorityService: mockAuthorityService,
			ApiKeyRepository: mockApiKeyRepository,
		}

		Convey("Given a valid authority ID", func() {
			authorityID, _ := uuid.NewRandom()
			apiKeyID, _ := uuid.NewRandom()
			name := "key11"

			Convey("If no expiration is given", func() {
				expireIn := 0

				Convey("When the service is queried", func() {
					mockApiKeyRepository.
						On("Create", &authority.ApiKey{AuthorityID: authorityID, Name: name}).
						Return(&authority.ApiKey{Entity: entity.Entity{ID: apiKeyID}}, nil)
					mockAuthorityService.
						On("GetByID", authorityID).
						Return(&authority.Authority{Entity: entity.Entity{ID: authorityID}, ApiKeys: []authority.ApiKey{}}, nil)

					apiKey, err := apiKeyService.Grant(authorityID, name, expireIn)

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
								On("Create", mock.AnythingOfType("*authority.ApiKey")).
								Return(&authority.ApiKey{Entity: entity.Entity{ID: apiKeyID}}, nil)
							mockAuthorityService.
								On("GetByID", authorityID).
								Return(&authority.Authority{Entity: entity.Entity{ID: authorityID}, ApiKeys: []authority.ApiKey{}}, nil)

							apiKey, err := apiKeyService.Grant(authorityID, "key1", expireIn)

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
		mockAuthorityService := NewMockAuthorityService(t)
		mockApiKeyRepository := repositories.NewMockApiKeyRepository(t)

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
			authorityID, _ := uuid.NewRandom()

			mockApiKeyRepository.On("Find", apiKey).Return(&authority.ApiKey{
				Entity:      entity.Entity{ID: apiKey},
				AuthorityID: authorityID,
			}, nil)
			mockApiKeyRepository.On("Delete", apiKey).Return(nil)
			mockAuthorityService.
				On("GetByID", authorityID).
				Return(&authority.Authority{Entity: entity.Entity{ID: authorityID}, ApiKeys: []authority.ApiKey{}}, nil)

			Convey("When the service is queried", func() {
				err := apiKeyService.Revoke(apiKeyStr)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})
}
