package services

import (
	"strings"
	"terralist/pkg/file"
	"terralist/pkg/vcs"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestBuildProviderCreateDTO(t *testing.T) {
	assetsWithSums := []vcs.ReleaseAsset{
		{Name: "terraform-provider-acme_1.0.0_SHA256SUMS", URL: "https://ex/sums"},
		{Name: "terraform-provider-acme_1.0.0_SHA256SUMS.sig", URL: "https://ex/sig"},
		{Name: "terraform-provider-acme_1.0.0_linux_amd64.zip", URL: "https://ex/linux.zip"},
	}

	t.Run("with SHA256SUMS", func(t *testing.T) {
		mockFetcher := file.NewMockFetcher(t)
		hashLine := strings.Repeat("a", 64) + "  terraform-provider-acme_1.0.0_linux_amd64.zip"
		mockFetcher.
			On("FetchFile", "terraform-provider-acme_1.0.0SHA256SUMS", "https://ex/sums", mock.Anything).
			Return(file.NewInMemoryFile("SHA256SUMS", []byte(hashLine+"\n")), func() {}, nil)

		svc := &DefaultVcsService{
			Fetcher: mockFetcher,
		}
		dto, err := svc.BuildProviderCreateDTO(uuid.New(), "ns", "acme", &vcs.ReleaseEvent{
			SemVer:  "1.0.0",
			Assets:  assetsWithSums,
			Source:  vcs.ReleaseSourceGitHub,
			RepoURL: "https://github.com/acme/acme",
		})
		if err != nil {
			t.Fatal(err)
		}
		want := strings.Repeat("a", 64)
		if len(dto.Platforms) != 1 || dto.Platforms[0].System != "linux" || dto.Platforms[0].Architecture != "amd64" || dto.Platforms[0].ShaSum != want {
			t.Fatalf("%+v", dto.Platforms)
		}
		if dto.ShaSums.URL != "https://ex/sums" || dto.ShaSums.SignatureURL != "https://ex/sig" {
			t.Fatalf("shasums dto: %+v", dto.ShaSums)
		}
	})

	t.Run("without SHA256SUMS", func(t *testing.T) {
		assetsNoSums := []vcs.ReleaseAsset{
			{Name: "terraform-provider-acme_1.0.0_linux_amd64.zip", URL: "https://ex/linux.zip"},
		}
		var provider vcs.Provider
		svc := &DefaultVcsService{
			Provider: provider,
		}
		dto, err := svc.BuildProviderCreateDTO(uuid.New(), "ns", "acme", &vcs.ReleaseEvent{
			SemVer:  "1.0.0",
			Assets:  assetsNoSums,
			Source:  vcs.ReleaseSourceGitHub,
			RepoURL: "https://github.com/acme/acme",
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(dto.Platforms) != 1 || dto.Platforms[0].ShaSum != "" {
			t.Fatalf("%+v", dto.Platforms)
		}
		if dto.ShaSums.URL != "" || dto.ShaSums.SignatureURL != "" {
			t.Fatalf("shasums dto: %+v", dto.ShaSums)
		}
	})
}
