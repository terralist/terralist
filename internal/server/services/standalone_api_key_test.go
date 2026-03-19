package services

import (
	"testing"

	"terralist/internal/server/models/apikey"
	"terralist/internal/server/repositories"
	"terralist/pkg/auth"
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestAuthenticate(t *testing.T) {
	Convey("Subject: Authenticating with a standalone API key", t, func() {
		mockRepo := repositories.NewMockStandaloneApiKeyRepository(t)

		service := &DefaultStandaloneApiKeyService{
			Repository: mockRepo,
		}

		Convey("Given an invalid API key", func() {
			Convey("When the service is queried", func() {
				user, err := service.Authenticate("not-a-uuid")

				Convey("Then a parse error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(user, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "cannot parse")
				})
			})
		})

		Convey("Given an expired API key", func() {
			keyID := uuid.Must(uuid.NewRandom())

			mockRepo.On("FindWithPolicies", keyID).Return(nil, repositories.ErrApiKeyExpired)

			Convey("When the service is queried", func() {
				user, err := service.Authenticate(keyID.String())

				Convey("Then an invalid key error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(user, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "invalid key")
				})
			})
		})

		Convey("Given a valid API key with policies", func() {
			keyID := uuid.Must(uuid.NewRandom())
			createdBy := "test@example.com"

			expectedPolicies := []apikey.Policy{
				{Resource: "modules", Action: "get", Object: "my-authority/*", Effect: "allow"},
			}

			mockRepo.On("FindWithPolicies", keyID).Return(&apikey.ApiKey{
				Entity:    entity.Entity{ID: keyID},
				Name:      "test-key",
				CreatedBy: createdBy,
				Policies:  expectedPolicies,
			}, nil)

			Convey("When the service is queried", func() {
				user, err := service.Authenticate(keyID.String())

				Convey("Then the user with inline policies should be returned", func() {
					So(err, ShouldBeNil)
					So(user, ShouldNotBeNil)
					So(user.Name, ShouldEqual, "apikey:"+keyID.String())
					So(user.Email, ShouldEqual, createdBy)
					So(user.Authority, ShouldBeEmpty)
					So(user.AuthorityID, ShouldBeEmpty)
					So(user.InlinePolicies, ShouldResemble, []auth.Policy{
						{Resource: "modules", Action: "get", Object: "my-authority/*", Effect: "allow"},
					})
				})
			})
		})
	})
}

func TestStandaloneCreate(t *testing.T) {
	Convey("Subject: Creating a standalone API key", t, func() {
		mockRepo := repositories.NewMockStandaloneApiKeyRepository(t)

		service := &DefaultStandaloneApiKeyService{
			Repository: mockRepo,
		}

		Convey("Given valid policies", func() {
			keyID := uuid.Must(uuid.NewRandom())
			policies := []apikey.Policy{
				{Resource: "modules", Action: "get", Object: "my-authority/*", Effect: "allow"},
				{Resource: "providers", Action: "*", Object: "*", Effect: "allow"},
			}

			mockRepo.
				On("Create", mock.AnythingOfType("*apikey.ApiKey")).
				Return(&apikey.ApiKey{Entity: entity.Entity{ID: keyID}}, nil)

			Convey("When the service is queried with no expiration", func() {
				key, err := service.Create("ci-key", "test@example.com", 0, policies)

				Convey("Then a valid API key should be returned", func() {
					So(err, ShouldBeNil)
					So(key, ShouldEqual, keyID.String())
				})
			})

			Convey("When the service is queried with expiration", func() {
				key, err := service.Create("ci-key", "test@example.com", 24, policies)

				Convey("Then a valid API key should be returned", func() {
					So(err, ShouldBeNil)
					So(key, ShouldEqual, keyID.String())
				})
			})
		})

		Convey("Given an invalid resource in policy", func() {
			policies := []apikey.Policy{
				{Resource: "invalid-resource", Action: "get", Object: "*", Effect: "allow"},
			}

			Convey("When the service is queried", func() {
				key, err := service.Create("ci-key", "test@example.com", 0, policies)

				Convey("Then a validation error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(key, ShouldBeEmpty)
					So(err.Error(), ShouldContainSubstring, "invalid resource")
				})
			})
		})

		Convey("Given an invalid action in policy", func() {
			policies := []apikey.Policy{
				{Resource: "modules", Action: "invalid-action", Object: "*", Effect: "allow"},
			}

			Convey("When the service is queried", func() {
				key, err := service.Create("ci-key", "test@example.com", 0, policies)

				Convey("Then a validation error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(key, ShouldBeEmpty)
					So(err.Error(), ShouldContainSubstring, "invalid action")
				})
			})
		})

		Convey("Given an invalid effect in policy", func() {
			policies := []apikey.Policy{
				{Resource: "modules", Action: "get", Object: "*", Effect: "maybe"},
			}

			Convey("When the service is queried", func() {
				key, err := service.Create("ci-key", "test@example.com", 0, policies)

				Convey("Then a validation error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(key, ShouldBeEmpty)
					So(err.Error(), ShouldContainSubstring, "invalid effect")
				})
			})
		})

		Convey("Given an empty object in policy", func() {
			policies := []apikey.Policy{
				{Resource: "modules", Action: "get", Object: "", Effect: "allow"},
			}

			Convey("When the service is queried", func() {
				key, err := service.Create("ci-key", "test@example.com", 0, policies)

				Convey("Then a validation error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(key, ShouldBeEmpty)
					So(err.Error(), ShouldContainSubstring, "empty object")
				})
			})
		})

		Convey("Given wildcard resource and action", func() {
			keyID := uuid.Must(uuid.NewRandom())
			policies := []apikey.Policy{
				{Resource: "*", Action: "*", Object: "*", Effect: "allow"},
			}

			mockRepo.
				On("Create", mock.AnythingOfType("*apikey.ApiKey")).
				Return(&apikey.ApiKey{Entity: entity.Entity{ID: keyID}}, nil)

			Convey("When the service is queried", func() {
				key, err := service.Create("admin-key", "admin@example.com", 0, policies)

				Convey("Then a valid API key should be returned", func() {
					So(err, ShouldBeNil)
					So(key, ShouldEqual, keyID.String())
				})
			})
		})
	})
}

func TestStandaloneDelete(t *testing.T) {
	Convey("Subject: Deleting a standalone API key", t, func() {
		mockRepo := repositories.NewMockStandaloneApiKeyRepository(t)

		service := &DefaultStandaloneApiKeyService{
			Repository: mockRepo,
		}

		Convey("Given an invalid API key", func() {
			Convey("When the service is queried", func() {
				err := service.Delete("not-a-uuid")

				Convey("Then a parse error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "cannot parse")
				})
			})
		})

		Convey("Given a valid API key", func() {
			keyID := uuid.Must(uuid.NewRandom())

			mockRepo.On("Delete", keyID).Return(nil)

			Convey("When the service is queried", func() {
				err := service.Delete(keyID.String())

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})
}
