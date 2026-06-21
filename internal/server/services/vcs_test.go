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

	vcsHeaders := map[string]string{"Accept": "application/json"}

	mockFetcher := file.NewMockFetcher(t)
	mockProvider := vcs.NewMockProvider(t)
	mockProvider.On("GetHeaders").Return(vcsHeaders)

	hashLine := strings.Repeat("a", 64) + "  terraform-provider-acme_1.0.0_linux_amd64.zip"
	mockFetcher.
		On("FetchFile", "terraform-provider-acme_1.0.0_SHA256SUMS", "https://ex/sums", mock.Anything).
		Return(file.NewInMemoryFile("SHA256SUMS", []byte(hashLine+"\n")), func() {}, nil)

	svc := &DefaultVcsService{
		Provider: mockProvider,
		Fetcher:  mockFetcher,
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
	if dto.Headers["Accept"] != vcsHeaders["Accept"] {
		t.Fatalf("headers dto: %+v", dto.Headers)
	}

}
