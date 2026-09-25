package services

import (
	"errors"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/repositories"
	"terralist/pkg/database/entity"
	"terralist/pkg/secret"

	"github.com/google/uuid"
	"github.com/mazen160/go-random"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestFindAuthorities(t *testing.T) {
	Convey("Subject: Finding authorities", t, func() {
		mockAuthorityRepository := repositories.NewMockAuthorityRepository(t)

		authorityService := &DefaultAuthorityService{
			AuthorityRepository: mockAuthorityRepository,
		}

		Convey("Given a valid authority ID", func() {
			authorityID, _ := uuid.NewRandom()

			Convey("If the authority exists in the database", func() {
				mockAuthorityRepository.
					On("FindByID", authorityID).
					Return(&authority.Authority{Entity: entity.Entity{ID: authorityID}}, nil)

				Convey("When the service is queried", func() {
					authority, err := authorityService.GetByID(authorityID)

					Convey("The authority should be returned", func() {
						So(authority, ShouldNotBeNil)
						So(err, ShouldBeNil)
					})
				})
			})

			Convey("If the authority does not exist in the database", func() {
				mockAuthorityRepository.
					On("FindByID", authorityID).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					authority, err := authorityService.GetByID(authorityID)

					Convey("An error should be returned", func() {
						So(authority, ShouldBeNil)
						So(err, ShouldNotBeNil)
					})
				})
			})
		})

		Convey("Given a valid authority name", func() {
			name, _ := random.String(16)

			Convey("If the authority exists in the database", func() {
				mockAuthorityRepository.
					On("FindByName", name).
					Return(&authority.Authority{
						Entity: entity.Entity{ID: uuid.Must(uuid.NewRandom())},
						Name:   name,
					}, nil)

				Convey("When the service is queried", func() {
					authority, err := authorityService.GetByName(name)

					Convey("The authority should be returned", func() {
						So(authority, ShouldNotBeNil)
						So(err, ShouldBeNil)
						So(authority.ID, ShouldNotBeNil)
						So(authority.Name, ShouldEqual, name)
					})
				})
			})

			Convey("If the authority does not exist in the database", func() {
				mockAuthorityRepository.
					On("FindByName", name).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					authority, err := authorityService.GetByName(name)

					Convey("An error should be returned", func() {
						So(authority, ShouldBeNil)
						So(err, ShouldNotBeNil)
					})
				})
			})
		})

		Convey("Given an owner's e-mail address", func() {
			owner := "test@example.com"

			Convey("If the owner is associated with one or more authorities", func() {
				mockAuthorityRepository.
					On("FindAllByOwner", owner).
					Return([]*authority.Authority{{Owner: owner}}, nil)

				Convey("When the service is queried", func() {
					authorities, err := authorityService.GetAllByOwner(owner)

					Convey("A list with a single authority should be returned", func() {
						So(authorities, ShouldNotBeNil)
						So(len(authorities), ShouldBeGreaterThan, 0)
						So(err, ShouldBeNil)
					})
				})
			})

			Convey("If the owner is not associated with any authority", func() {
				mockAuthorityRepository.
					On("FindAllByOwner", owner).
					Return([]*authority.Authority{}, nil)

				Convey("When the service is queried", func() {
					authorities, err := authorityService.GetAllByOwner(owner)

					Convey("An empty list should be returned", func() {
						So(authorities, ShouldNotBeNil)
						So(len(authorities), ShouldEqual, 0)
						So(err, ShouldBeNil)
					})
				})
			})
		})
	})
}

func TestCreateAuthority(t *testing.T) {
	Convey("Subject: Creating authorities", t, func() {
		mockAuthorityRepository := repositories.NewMockAuthorityRepository(t)

		authorityService := &DefaultAuthorityService{
			AuthorityRepository: mockAuthorityRepository,
		}

		Convey("Given a valid authority create DTO", func() {
			dto := authority.AuthorityCreateDTO{
				Name:      "Test",
				PolicyURL: "https://example.com/test.html",
				Owner:     "test@example.com",
			}

			mockAuthorityRepository.
				On("Upsert", mock.AnythingOfType("authority.Authority")).
				Return(&authority.Authority{
					Name:      dto.Name,
					PolicyURL: dto.PolicyURL,
				}, nil)

			Convey("When the service is queried", func() {
				created, err := authorityService.Create(dto)

				Convey("No error should be returned", func() {
					So(err, ShouldBeNil)
				})

				Convey("Created authority should have and ID", func() {
					So(created.ID, ShouldNotBeEmpty)
				})

				Convey("Created authority should have the same attributes", func() {
					So(created.Name, ShouldEqual, dto.Name)
					So(created.PolicyURL, ShouldEqual, created.PolicyURL)
				})
			})
		})
	})
}

func TestAddKey(t *testing.T) {
	Convey("Subject: Create authority keys", t, func() {
		mockAuthorityRepository := repositories.NewMockAuthorityRepository(t)

		authorityService := &DefaultAuthorityService{
			AuthorityRepository: mockAuthorityRepository,
		}

		Convey("Given an authority ID and a key", func() {
			authorityID, _ := uuid.NewRandom()
			dto := authority.KeyDTO{}

			Convey("If the authority exists", func() {
				mockAuthorityRepository.
					On("FindByID", authorityID).
					Return(&authority.Authority{}, nil)

				mockAuthorityRepository.
					On("Upsert", mock.AnythingOfType("authority.Authority")).
					Return(&authority.Authority{}, nil)

				Convey("When the service is queried", func() {
					result, err := authorityService.AddKey(authorityID, dto)

					Convey("No error should be returned", func() {
						So(err, ShouldBeNil)
					})

					Convey("The returned key should have the same attributes", func() {
						So(result.KeyId, ShouldEqual, dto.KeyId)
						So(result.AsciiArmor, ShouldEqual, dto.AsciiArmor)
						So(result.TrustSignature, ShouldEqual, dto.TrustSignature)
					})
				})
			})

			Convey("If the authority does not exist", func() {
				mockAuthorityRepository.
					On("FindByID", authorityID).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					result, err := authorityService.AddKey(authorityID, dto)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
					})

					Convey("The returned object should be nil", func() {
						So(result, ShouldBeNil)
					})
				})
			})
		})
	})
}

func TestRemoveKey(t *testing.T) {
	Convey("Subject: Delete authority keys", t, func() {
		mockAuthorityRepository := repositories.NewMockAuthorityRepository(t)

		authorityService := &DefaultAuthorityService{
			AuthorityRepository: mockAuthorityRepository,
		}

		Convey("Given an authority ID and a key ID", func() {
			authorityID, _ := uuid.NewRandom()
			keyID, _ := uuid.NewRandom()

			Convey("If the authority exists and the key is associated with the authority", func() {
				mockAuthorityRepository.
					On("FindByID", authorityID).
					Return(&authority.Authority{
						Keys: []authority.Key{},
					}, nil)

				Convey("When the service is queried", func() {
					err := authorityService.RemoveKey(authorityID, keyID)

					Convey("A key not found error should be returned", func() {
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldContainSubstring, ErrKeyNotFound.Error())
					})
				})
			})

			Convey("If the authority exists and the key is the only key associated with the authority", func() {
				mockAuthorityRepository.
					On("FindByID", authorityID).
					Return(&authority.Authority{
						Keys: []authority.Key{
							{
								Entity: entity.Entity{
									ID: keyID,
								},
							},
						},
					}, nil)

				mockAuthorityRepository.
					On("Delete", authorityID).
					Return(nil)

				Convey("When the service is queried", func() {
					err := authorityService.RemoveKey(authorityID, keyID)

					Convey("No error should be returned", func() {
						So(err, ShouldBeNil)
					})
				})
			})

			Convey("If the authority exists and the key is not the only key associated with the authority", func() {
				otherKeyID, _ := uuid.NewRandom()

				mockAuthorityRepository.
					On("FindByID", authorityID).
					Return(&authority.Authority{
						Keys: []authority.Key{
							{
								Entity: entity.Entity{
									ID: keyID,
								},
							},
							{
								Entity: entity.Entity{
									ID: otherKeyID,
								},
							},
						},
					}, nil)

				mockAuthorityRepository.
					On("Upsert", mock.AnythingOfType("authority.Authority")).
					Return(&authority.Authority{}, nil)

				Convey("When the service is queried", func() {
					err := authorityService.RemoveKey(authorityID, keyID)

					Convey("No error should be returned", func() {
						So(err, ShouldBeNil)
					})

				})

			})

			Convey("If the authority does not exists", func() {
				mockAuthorityRepository.
					On("FindByID", authorityID).
					Return(nil, errors.New(""))

				Convey("When the service is queried", func() {
					err := authorityService.RemoveKey(authorityID, keyID)

					Convey("An error should be returned", func() {
						So(err, ShouldNotBeNil)
					})
				})
			})
		})
	})
}

func TestDeleteAuthority(t *testing.T) {
	Convey("Subject: Creating authorities", t, func() {
		mockAuthorityRepository := repositories.NewMockAuthorityRepository(t)

		authorityService := &DefaultAuthorityService{
			AuthorityRepository: mockAuthorityRepository,
		}

		Convey("Given an authority ID", func() {
			authorityID, _ := uuid.NewRandom()

			mockAuthorityRepository.
				On("Delete", authorityID).
				Return(nil)

			Convey("When the service is queried", func() {
				err := authorityService.Delete(authorityID)

				Convey("No error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})
}

func TestCreateAuthorityWithUpstream(t *testing.T) {
	Convey("Subject: Creating authorities with an upstream identity", t, func() {
		mockAuthorityRepository := repositories.NewMockAuthorityRepository(t)

		authorityService := &DefaultAuthorityService{
			AuthorityRepository: mockAuthorityRepository,
		}

		Convey("Given a hostname and no namespace", func() {
			dto := authority.AuthorityCreateDTO{
				Name:             "hashicorp",
				PolicyURL:        "https://example.com/hashicorp",
				Owner:            "test@example.com",
				UpstreamHostname: "Registry.Terraform.io",
			}

			var saved authority.Authority
			mockAuthorityRepository.
				On("Upsert", mock.AnythingOfType("authority.Authority")).
				Run(func(args mock.Arguments) {
					saved, _ = args.Get(0).(authority.Authority)
				}).
				Return(&authority.Authority{Name: dto.Name}, nil)

			Convey("When the authority is created", func() {
				_, err := authorityService.Create(dto)

				Convey("The hostname should be lowercased and the namespace default to the name", func() {
					So(err, ShouldBeNil)
					So(*saved.UpstreamHostname, ShouldEqual, "registry.terraform.io")
					So(*saved.UpstreamNamespace, ShouldEqual, "hashicorp")
				})
			})
		})

		Convey("Given no hostname", func() {
			dto := authority.AuthorityCreateDTO{
				Name:      "hashicorp",
				PolicyURL: "https://example.com/hashicorp",
				Owner:     "test@example.com",
			}

			var saved authority.Authority
			mockAuthorityRepository.
				On("Upsert", mock.AnythingOfType("authority.Authority")).
				Run(func(args mock.Arguments) {
					saved, _ = args.Get(0).(authority.Authority)
				}).
				Return(&authority.Authority{Name: dto.Name}, nil)

			Convey("When the authority is created", func() {
				_, err := authorityService.Create(dto)

				Convey("No upstream identity should be stored", func() {
					So(err, ShouldBeNil)
					So(saved.UpstreamHostname, ShouldBeNil)
					So(saved.UpstreamNamespace, ShouldBeNil)
				})
			})
		})

		Convey("Given a namespace without a hostname", func() {
			dto := authority.AuthorityCreateDTO{
				Name:              "hashicorp",
				PolicyURL:         "https://example.com/hashicorp",
				Owner:             "test@example.com",
				UpstreamNamespace: "hashicorp",
			}

			Convey("When the authority is created", func() {
				_, err := authorityService.Create(dto)

				Convey("An error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given an invalid hostname", func() {
			dto := authority.AuthorityCreateDTO{
				Name:             "hashicorp",
				PolicyURL:        "https://example.com/hashicorp",
				Owner:            "test@example.com",
				UpstreamHostname: "https://registry.terraform.io",
			}

			Convey("When the authority is created", func() {
				_, err := authorityService.Create(dto)

				Convey("An error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given an invalid namespace", func() {
			dto := authority.AuthorityCreateDTO{
				Name:              "hashicorp",
				PolicyURL:         "https://example.com/hashicorp",
				Owner:             "test@example.com",
				UpstreamHostname:  "registry.terraform.io",
				UpstreamNamespace: "hashi corp",
			}

			Convey("When the authority is created", func() {
				_, err := authorityService.Create(dto)

				Convey("An error should be returned", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}

func TestUpdateAuthorityUpstream(t *testing.T) {
	Convey("Subject: Updating the upstream identity of an authority", t, func() {
		mockAuthorityRepository := repositories.NewMockAuthorityRepository(t)

		authorityService := &DefaultAuthorityService{
			AuthorityRepository: mockAuthorityRepository,
		}

		id, _ := uuid.NewRandom()
		mockAuthorityRepository.On("FindByID", id).Return(nil, errors.New("")).Maybe()

		Convey("Given a hostname with a port and an explicit namespace", func() {
			dto := authority.AuthorityDTO{
				Name:              "mirror",
				PolicyURL:         "https://example.com/mirror",
				UpstreamHostname:  "Registry.Example.com:8443",
				UpstreamNamespace: "hashicorp",
			}

			var saved authority.Authority
			mockAuthorityRepository.
				On("Upsert", mock.AnythingOfType("authority.Authority")).
				Run(func(args mock.Arguments) {
					saved, _ = args.Get(0).(authority.Authority)
				}).
				Return(&authority.Authority{Name: dto.Name}, nil)

			Convey("When the authority is updated", func() {
				_, err := authorityService.Update(id, dto)

				Convey("The identity should be stored normalized", func() {
					So(err, ShouldBeNil)
					So(*saved.UpstreamHostname, ShouldEqual, "registry.example.com:8443")
					So(*saved.UpstreamNamespace, ShouldEqual, "hashicorp")
				})
			})
		})

		Convey("Given an empty hostname and namespace", func() {
			dto := authority.AuthorityDTO{
				Name:      "mirror",
				PolicyURL: "https://example.com/mirror",
			}

			var saved authority.Authority
			mockAuthorityRepository.
				On("Upsert", mock.AnythingOfType("authority.Authority")).
				Run(func(args mock.Arguments) {
					saved, _ = args.Get(0).(authority.Authority)
				}).
				Return(&authority.Authority{Name: dto.Name}, nil)

			Convey("When the authority is updated", func() {
				_, err := authorityService.Update(id, dto)

				Convey("The identity should be cleared", func() {
					So(err, ShouldBeNil)
					So(saved.UpstreamHostname, ShouldBeNil)
					So(saved.UpstreamNamespace, ShouldBeNil)
				})
			})
		})
	})
}

func TestGetAuthorityByUpstream(t *testing.T) {
	Convey("Subject: Finding an authority by its upstream identity", t, func() {
		mockAuthorityRepository := repositories.NewMockAuthorityRepository(t)

		authorityService := &DefaultAuthorityService{
			AuthorityRepository: mockAuthorityRepository,
		}

		Convey("If the repository finds the authority", func() {
			mockAuthorityRepository.
				On("FindByUpstream", "registry.terraform.io", "hashicorp").
				Return(&authority.Authority{Name: "hashicorp"}, nil)

			Convey("When the service is queried", func() {
				a, err := authorityService.GetByUpstream("registry.terraform.io", "hashicorp")

				Convey("The authority should be returned", func() {
					So(err, ShouldBeNil)
					So(a.Name, ShouldEqual, "hashicorp")
				})
			})
		})

		Convey("If the repository does not find the authority", func() {
			mockAuthorityRepository.
				On("FindByUpstream", "registry.terraform.io", "hashicorp").
				Return(nil, errors.New(""))

			Convey("When the service is queried", func() {
				a, err := authorityService.GetByUpstream("registry.terraform.io", "hashicorp")

				Convey("An error should be returned", func() {
					So(err, ShouldNotBeNil)
					So(a, ShouldBeNil)
				})
			})
		})
	})
}

func TestAuthorityUpstreamSettings(t *testing.T) {
	Convey("Subject: Validating the upstream connection settings of an authority", t, func() {
		mockAuthorityRepository := repositories.NewMockAuthorityRepository(t)
		sealer := newTestSealer()

		authorityService := &DefaultAuthorityService{
			AuthorityRepository: mockAuthorityRepository,
			Sealer:              sealer,
		}

		var saved authority.Authority
		expectUpsert := func() {
			mockAuthorityRepository.
				On("Upsert", mock.AnythingOfType("authority.Authority")).
				Run(func(args mock.Arguments) {
					saved, _ = args.Get(0).(authority.Authority)
				}).
				Return(&authority.Authority{Name: "hashicorp"}, nil)
		}

		base := func() authority.AuthorityCreateDTO {
			return authority.AuthorityCreateDTO{
				Name:             "hashicorp",
				PolicyURL:        "https://example.com/hashicorp",
				Owner:            "test@example.com",
				UpstreamHostname: "registry.terraform.io",
			}
		}

		Convey("Given no default policy", func() {
			expectUpsert()
			_, err := authorityService.Create(base())

			Convey("Then the policy should default to allow", func() {
				So(err, ShouldBeNil)
				So(saved.UpstreamDefaultPolicy, ShouldEqual, authority.PolicyAllow)
			})
		})

		Convey("Given an unknown default policy", func() {
			dto := base()
			dto.UpstreamPolicy = "sometimes"
			_, err := authorityService.Create(dto)

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given an upstream URL", func() {
			Convey("When it is a valid https URL", func() {
				dto := base()
				dto.UpstreamURL = "https://mirror.example.com/registry/"
				expectUpsert()
				_, err := authorityService.Create(dto)

				Convey("Then it should be stored without the trailing slash", func() {
					So(err, ShouldBeNil)
					So(*saved.UpstreamURL, ShouldEqual, "https://mirror.example.com/registry")
				})
			})

			Convey("When it is not an http URL", func() {
				dto := base()
				dto.UpstreamURL = "ftp://mirror.example.com"
				_, err := authorityService.Create(dto)

				Convey("Then it should be rejected", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("Given the upstream is enabled without a hostname", func() {
			dto := base()
			dto.UpstreamHostname = ""
			dto.UpstreamEnabled = true
			_, err := authorityService.Create(dto)

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Given a token", func() {
			dto := base()
			dto.UpstreamToken = "ghp_secret"
			expectUpsert()
			_, err := authorityService.Create(dto)

			Convey("Then it should be sealed for the authority before it is stored", func() {
				So(err, ShouldBeNil)
				So(saved.Empty(), ShouldBeFalse)
				So(*saved.UpstreamToken, ShouldNotContainSubstring, "ghp_secret")
				opened, err := sealer.Open(*saved.UpstreamToken, saved.ID.String())
				So(err, ShouldBeNil)
				So(opened, ShouldEqual, "ghp_secret")
			})
		})

		Convey("Given a token without a sealer configured", func() {
			authorityService.Sealer = nil
			dto := base()
			dto.UpstreamToken = "ghp_secret"
			_, err := authorityService.Create(dto)

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "upstream-secret")
			})
		})
	})
}

func TestUpdateAuthorityKeepsToken(t *testing.T) {
	Convey("Subject: Updating an authority that has an upstream token", t, func() {
		mockAuthorityRepository := repositories.NewMockAuthorityRepository(t)
		sealer := newTestSealer()

		authorityService := &DefaultAuthorityService{
			AuthorityRepository: mockAuthorityRepository,
			Sealer:              sealer,
		}

		id, _ := uuid.NewRandom()
		sealed, _ := sealer.Seal("ghp_old", id.String())
		hostname := "registry.terraform.io"

		mockAuthorityRepository.On("FindByID", id).Return(&authority.Authority{
			Entity:           entity.Entity{ID: id},
			Name:             "hashicorp",
			UpstreamHostname: &hostname,
			UpstreamToken:    &sealed,
		}, nil).Maybe()

		var saved authority.Authority
		mockAuthorityRepository.
			On("Upsert", mock.AnythingOfType("authority.Authority")).
			Run(func(args mock.Arguments) {
				saved, _ = args.Get(0).(authority.Authority)
			}).
			Return(&authority.Authority{Name: "hashicorp"}, nil)

		Convey("When the update carries no token", func() {
			_, err := authorityService.Update(id, authority.AuthorityDTO{
				Name:             "hashicorp",
				PolicyURL:        "https://example.com/hashicorp",
				UpstreamHostname: hostname,
				UpstreamEnabled:  true,
			})

			Convey("Then the stored token should be kept", func() {
				So(err, ShouldBeNil)
				So(*saved.UpstreamToken, ShouldEqual, sealed)
				So(saved.UpstreamEnabled, ShouldBeTrue)
			})
		})

		Convey("When the update carries a new token", func() {
			_, err := authorityService.Update(id, authority.AuthorityDTO{
				Name:             "hashicorp",
				PolicyURL:        "https://example.com/hashicorp",
				UpstreamHostname: hostname,
				UpstreamToken:    "ghp_new",
			})

			Convey("Then the new token should be sealed and stored", func() {
				So(err, ShouldBeNil)
				opened, _ := sealer.Open(*saved.UpstreamToken, id.String())
				So(opened, ShouldEqual, "ghp_new")
			})
		})

		Convey("When the update drops the upstream hostname", func() {
			_, err := authorityService.Update(id, authority.AuthorityDTO{
				Name:      "hashicorp",
				PolicyURL: "https://example.com/hashicorp",
			})

			Convey("Then the token should be dropped with it", func() {
				So(err, ShouldBeNil)
				So(saved.UpstreamToken, ShouldBeNil)
			})
		})
	})
}

func TestAuthorityRules(t *testing.T) {
	Convey("Subject: Managing the upstream rules of an authority", t, func() {
		mockAuthorityRepository := repositories.NewMockAuthorityRepository(t)

		authorityService := &DefaultAuthorityService{
			AuthorityRepository: mockAuthorityRepository,
		}

		id, _ := uuid.NewRandom()
		ruleID, _ := uuid.NewRandom()
		existing := authority.Rule{
			Entity:  entity.Entity{ID: ruleID},
			Kind:    authority.RuleKindProvider,
			Name:    "aws",
			Version: "*",
			Effect:  authority.EffectDeny,
		}

		Convey("When a valid rule is added", func() {
			mockAuthorityRepository.On("FindByID", id).Return(&authority.Authority{Entity: entity.Entity{ID: id}}, nil)

			var saved authority.Authority
			mockAuthorityRepository.
				On("Upsert", mock.AnythingOfType("authority.Authority")).
				Run(func(args mock.Arguments) {
					saved, _ = args.Get(0).(authority.Authority)
				}).
				Return(&authority.Authority{Rules: []authority.Rule{existing}}, nil)

			dto, err := authorityService.AddRule(id, authority.RuleDTO{Kind: "provider", Name: "aws", Version: "*", Effect: "deny"})

			Convey("Then it should be stored and returned", func() {
				So(err, ShouldBeNil)
				So(len(saved.Rules), ShouldEqual, 1)
				So(saved.Rules[0].Name, ShouldEqual, "aws")
				So(dto.ID, ShouldEqual, ruleID.String())
			})
		})

		Convey("When an invalid rule is added", func() {
			mockAuthorityRepository.On("FindByID", id).Return(&authority.Authority{Entity: entity.Entity{ID: id}}, nil).Maybe()

			_, err := authorityService.AddRule(id, authority.RuleDTO{Kind: "bucket", Name: "aws", Version: "*", Effect: "deny"})

			Convey("Then it should be rejected without storing", func() {
				So(err, ShouldNotBeNil)
				mockAuthorityRepository.AssertNotCalled(t, "Upsert", mock.Anything)
			})
		})

		Convey("When an existing rule is removed", func() {
			mockAuthorityRepository.On("FindByID", id).Return(&authority.Authority{Entity: entity.Entity{ID: id}, Rules: []authority.Rule{existing}}, nil)
			mockAuthorityRepository.On("DeleteRule", ruleID).Return(nil)

			err := authorityService.RemoveRule(id, ruleID)

			Convey("Then it should be deleted", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When an unknown rule is removed", func() {
			mockAuthorityRepository.On("FindByID", id).Return(&authority.Authority{Entity: entity.Entity{ID: id}}, nil)

			err := authorityService.RemoveRule(id, ruleID)

			Convey("Then it should be reported as not found", func() {
				So(errors.Is(err, ErrRuleNotFound), ShouldBeTrue)
			})
		})
	})
}

func newTestSealer() *secret.Sealer {
	sealer, err := secret.NewSealer("test-secret-of-at-least-32-characters")
	if err != nil {
		panic(err)
	}

	return sealer
}
