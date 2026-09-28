package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/internal/server/repositories"
	"terralist/pkg/storage"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestSubmoduleDocumentationStorage(t *testing.T) {
	Convey("Subject: Submodule documentation kept next to the module archive", t, func() {
		mockModuleRepository := repositories.NewMockModuleRepository(t)
		mockAuthorityService := NewMockAuthorityService(t)
		mockResolver := storage.NewMockResolver(t)

		moduleService := &DefaultModuleService{
			ModuleRepository: mockModuleRepository,
			AuthorityService: mockAuthorityService,
			Resolver:         mockResolver,
		}

		// The archive and its documentation were stored under the names as
		// uploaded, which the request below does not repeat.
		version := module.Version{
			Version:    "1.0.0",
			Location:   "modules/HashiCorp/Dir/Template/1.0.0.zip",
			Submodules: []module.Submodule{{Path: "modules/sub"}},
		}
		docsKey := "modules/HashiCorp/Dir/Template/submodules/1.0.0_modules__sub.md"

		Convey("When the documentation of a submodule is requested in another case", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("# Sub"))
			}))
			defer server.Close()

			mockModuleRepository.On("FindVersion", "hashicorp", "dir", "template", "1.0.0").Return(&version, nil)
			mockResolver.On("Find", docsKey).Return(server.URL, nil)

			doc, err := moduleService.GetSubmoduleDocumentation("hashicorp", "dir", "template", "1.0.0", "modules/sub")

			Convey("Then it should be read from where it was stored", func() {
				So(err, ShouldBeNil)
				So(doc, ShouldEqual, "# Sub")
			})
		})

		Convey("When the module version is deleted", func() {
			authorityID := uuid.New()
			mockAuthorityService.On("GetByID", authorityID).Return(&authority.Authority{Name: "HashiCorp"}, nil)
			mockModuleRepository.On("Find", "HashiCorp", "dir", "template").Return(&module.Module{
				Name:     "Dir",
				Provider: "Template",
				Versions: []module.Version{version, {Version: "1.0.1", Location: "modules/HashiCorp/Dir/Template/1.0.1.zip"}},
			}, nil)
			mockModuleRepository.On("DeleteVersion", mock.AnythingOfType("*module.Version")).Return(nil)
			mockResolver.On("Purge", "modules/HashiCorp/Dir/Template/1.0.0.zip").Return(nil)
			mockResolver.On("Purge", docsKey).Return(nil)

			err := moduleService.DeleteVersion(authorityID, "dir", "template", "1.0.0")

			Convey("Then its submodule documentation should be purged from where it was stored", func() {
				So(err, ShouldBeNil)
				mockResolver.AssertCalled(t, "Purge", docsKey)
			})
		})
	})
}
