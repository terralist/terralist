package handlers

import (
	"errors"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/services"
	"terralist/pkg/auth"
	"terralist/pkg/rbac"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCanPerform_AuthorityIsolation(t *testing.T) {
	Convey("Subject: Authority isolation for API key authenticated users", t, func() {
		mockAuthorityService := services.NewMockAuthorityService(t)

		// Create an enforcer with empty policy - users will be denied by default
		enforcer, err := rbac.NewEnforcer("", "readonly")
		So(err, ShouldBeNil)

		authorization := &Authorization{
			Enforcer:         enforcer,
			AuthorityService: mockAuthorityService,
		}

		Convey("Given an API key authenticated user (has AuthorityID)", func() {
			apiKeyUser := auth.User{
				Name:        "api-key-user",
				Email:       "user@example.com",
				Authority:   "my-authority",
				AuthorityID: "some-uuid-value",
			}

			Convey("When accessing modules in their own authority", func() {
				object := "my-authority/my-module/aws"

				Convey("And the authority is public", func() {
					mockAuthorityService.On("GetByName", "my-authority").
						Return(&authority.Authority{Name: "my-authority", Public: true}, nil)

					result := authorization.CanPerform(apiKeyUser, rbac.ResourceModules, rbac.ActionGet, object)

					Convey("Then access should be allowed", func() {
						So(result, ShouldBeTrue)
					})
				})
			})

			Convey("When accessing modules in a different authority", func() {
				object := "other-authority/some-module/aws"

				result := authorization.CanPerform(apiKeyUser, rbac.ResourceModules, rbac.ActionGet, object)

				Convey("Then access should be denied", func() {
					So(result, ShouldBeFalse)
				})
			})

			Convey("When accessing modules in a different authority (case insensitive)", func() {
				object := "MY-AUTHORITY/some-module/aws"

				Convey("And the authority is public", func() {
					mockAuthorityService.On("GetByName", "MY-AUTHORITY").
						Return(&authority.Authority{Name: "MY-AUTHORITY", Public: true}, nil)

					result := authorization.CanPerform(apiKeyUser, rbac.ResourceModules, rbac.ActionGet, object)

					Convey("Then access should be allowed (case insensitive match)", func() {
						So(result, ShouldBeTrue)
					})
				})
			})

			Convey("When accessing providers in their own authority", func() {
				object := "my-authority/my-provider"

				Convey("And the authority is public", func() {
					mockAuthorityService.On("GetByName", "my-authority").
						Return(&authority.Authority{Name: "my-authority", Public: true}, nil)

					result := authorization.CanPerform(apiKeyUser, rbac.ResourceProviders, rbac.ActionGet, object)

					Convey("Then access should be allowed", func() {
						So(result, ShouldBeTrue)
					})
				})
			})

			Convey("When accessing providers in a different authority", func() {
				object := "other-authority/some-provider"

				result := authorization.CanPerform(apiKeyUser, rbac.ResourceProviders, rbac.ActionGet, object)

				Convey("Then access should be denied", func() {
					So(result, ShouldBeFalse)
				})
			})

			Convey("When accessing authorities resource (not modules/providers)", func() {
				object := "other-authority"

				Convey("The isolation check should not apply", func() {
					// For authorities resource, the enforcer handles it
					// With readonly role and get action, this should pass
					result := authorization.CanPerform(apiKeyUser, rbac.ResourceAuthorities, rbac.ActionGet, object)

					Convey("Then access should be allowed by enforcer (readonly role)", func() {
						So(result, ShouldBeTrue)
					})
				})
			})
		})

		Convey("Given a session authenticated user (no AuthorityID)", func() {
			sessionUser := auth.User{
				Name:        "session-user",
				Email:       "session@example.com",
				Authority:   "",
				AuthorityID: "",
			}

			Convey("When accessing modules in any authority", func() {
				object := "any-authority/any-module/aws"

				Convey("And the authority is public", func() {
					mockAuthorityService.On("GetByName", "any-authority").
						Return(&authority.Authority{Name: "any-authority", Public: true}, nil)

					result := authorization.CanPerform(sessionUser, rbac.ResourceModules, rbac.ActionGet, object)

					Convey("Then access should be allowed (no isolation for session users)", func() {
						So(result, ShouldBeTrue)
					})
				})

				Convey("And the authority is private", func() {
					mockAuthorityService.On("GetByName", "any-authority").
						Return(&authority.Authority{Name: "any-authority", Public: false}, nil)

					result := authorization.CanPerform(sessionUser, rbac.ResourceModules, rbac.ActionGet, object)

					Convey("Then access should be allowed by enforcer (readonly role for get)", func() {
						So(result, ShouldBeTrue)
					})
				})
			})
		})

		Convey("Given an anonymous user", func() {
			anonymousUser := auth.User{
				Name: rbac.SubjectAnonymous,
			}

			Convey("When accessing modules in a public authority", func() {
				object := "public-authority/some-module/aws"

				mockAuthorityService.On("GetByName", "public-authority").
					Return(&authority.Authority{Name: "public-authority", Public: true}, nil)

				result := authorization.CanPerform(anonymousUser, rbac.ResourceModules, rbac.ActionGet, object)

				Convey("Then access should be allowed", func() {
					So(result, ShouldBeTrue)
				})
			})

			Convey("When accessing modules in a private authority", func() {
				object := "private-authority/some-module/aws"

				mockAuthorityService.On("GetByName", "private-authority").
					Return(&authority.Authority{Name: "private-authority", Public: false}, nil)

				result := authorization.CanPerform(anonymousUser, rbac.ResourceModules, rbac.ActionGet, object)

				Convey("Then access should be denied (anonymous can't access private)", func() {
					So(result, ShouldBeFalse)
				})
			})
		})
	})
}

func TestCanPerform_EdgeCases(t *testing.T) {
	Convey("Subject: Edge cases for CanPerform authority isolation", t, func() {
		mockAuthorityService := services.NewMockAuthorityService(t)

		enforcer, err := rbac.NewEnforcer("", "readonly")
		So(err, ShouldBeNil)

		authorization := &Authorization{
			Enforcer:         enforcer,
			AuthorityService: mockAuthorityService,
		}

		Convey("Given an API key user", func() {
			apiKeyUser := auth.User{
				Name:        "api-key-user",
				Email:       "user@example.com",
				Authority:   "my-authority",
				AuthorityID: "some-uuid-value",
			}

			Convey("When object is empty", func() {
				object := ""

				// Empty string split produces one empty part, which won't match "my-authority"
				result := authorization.CanPerform(apiKeyUser, rbac.ResourceModules, rbac.ActionGet, object)

				Convey("Then access should be denied (empty namespace doesn't match)", func() {
					So(result, ShouldBeFalse)
				})
			})

			Convey("When object has only namespace (no slash)", func() {
				object := "my-authority"

				mockAuthorityService.On("GetByName", "my-authority").
					Return(&authority.Authority{Name: "my-authority", Public: true}, nil)

				result := authorization.CanPerform(apiKeyUser, rbac.ResourceModules, rbac.ActionGet, object)

				Convey("Then access should be allowed if namespace matches", func() {
					So(result, ShouldBeTrue)
				})
			})

			Convey("When authority service returns an error for public check", func() {
				object := "my-authority/my-module/aws"

				mockAuthorityService.On("GetByName", "my-authority").
					Return(nil, errors.New("database error"))

				result := authorization.CanPerform(apiKeyUser, rbac.ResourceModules, rbac.ActionGet, object)

				Convey("Then it should fall through to enforcer (readonly allows get)", func() {
					So(result, ShouldBeTrue)
				})
			})
		})
	})
}
