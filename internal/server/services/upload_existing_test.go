package services

import (
	"errors"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/repositories"
	"terralist/pkg/file"
	"terralist/pkg/storage"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUploadOverExistingVersion(t *testing.T) {
	Convey("Subject: Uploading a version Terralist already holds", t, func() {
		authorityID := uuid.New()
		mockAuthorityService := NewMockAuthorityService(t)
		mockAuthorityService.On("GetByID", authorityID).Return(&authority.Authority{Name: "hashicorp"}, nil)

		for _, tc := range []struct {
			origin string
			want   string
		}{
			{provider.OriginManual, "version 1.0.0 already exists"},
			{provider.OriginUpstream, "version 1.0.0 already exists: it was pulled from the upstream, delete it before uploading your own"},
		} {
			Convey("Given a provider version of origin "+tc.origin, func() {
				mockProviderRepository := repositories.NewMockProviderRepository(t)
				mockProviderRepository.On("Find", "hashicorp", "null").Return(&provider.Provider{
					Name:     "null",
					Versions: []provider.Version{{Version: "1.0.0", Origin: tc.origin}},
				}, nil)
				service := &DefaultProviderService{
					ProviderRepository: mockProviderRepository,
					AuthorityService:   mockAuthorityService,
					Resolver:           storage.NewMockResolver(t),
				}

				err := service.Upload(&provider.CreateProviderDTO{AuthorityID: authorityID, Name: "null", Version: "1.0.0"})

				Convey("Then the upload should be rejected with the reason", func() {
					So(errors.Is(err, repositories.ErrAlreadyExists), ShouldBeTrue)
					So(err.Error(), ShouldEqual, tc.want)
				})
			})

			Convey("Given a module version of origin "+tc.origin, func() {
				mockModuleRepository := repositories.NewMockModuleRepository(t)
				mockModuleRepository.On("Find", "hashicorp", "vpc", "aws").Return(&module.Module{
					Name:     "vpc",
					Provider: "aws",
					Versions: []module.Version{{Version: "1.0.0", Origin: tc.origin}},
				}, nil)
				service := &DefaultModuleService{
					ModuleRepository: mockModuleRepository,
					AuthorityService: mockAuthorityService,
					Resolver:         storage.NewMockResolver(t),
					Fetcher:          file.NewMockFetcher(t),
				}

				err := service.Upload(&module.CreateDTO{
					AuthorityID:      authorityID,
					Name:             "vpc",
					Provider:         "aws",
					VersionCreateDTO: module.VersionCreateDTO{Version: "1.0.0"},
				}, file.NewRemoteFile("https://example.com/vpc.zip", nil))

				Convey("Then the upload should be rejected with the reason", func() {
					So(errors.Is(err, repositories.ErrAlreadyExists), ShouldBeTrue)
					So(err.Error(), ShouldEqual, tc.want)
				})
			})
		}
	})
}
