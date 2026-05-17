package services

import (
	"fmt"
	"strings"
	"terralist/internal/server/models/provider"
	"terralist/pkg/file"
	"terralist/pkg/vcs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	DefaultProviderProtocols = "5.0"
)

type VcsService interface {
	GetHeaders() map[string]string
	ParseModuleReleaseWebhook(ctx *gin.Context, vcsName string, namespace string, name string, provider string) (*vcs.ReleaseEvent, error)
	ParseProviderReleaseWebhook(ctx *gin.Context, vcsName string, namespace string, name string) (*vcs.ReleaseEvent, error)

	BuildProviderCreateDTO(authorityID uuid.UUID, namespace string, name string, ev *vcs.ReleaseEvent) (*provider.CreateProviderDTO, error)
}

type DefaultVcsService struct {
	Provider vcs.Provider
	Fetcher  file.Fetcher
}

func (s *DefaultVcsService) GetHeaders() map[string]string {
	return s.Provider.GetHeaders()
}

func (s *DefaultVcsService) ParseModuleReleaseWebhook(ctx *gin.Context, vcsName string, namespace string, name string, provider string) (*vcs.ReleaseEvent, error) {
	body, err := ctx.GetRawData()
	if err != nil {
		return nil, err
	}

	if err := s.Provider.Authenticate(ctx, body); err != nil {
		return nil, err
	}

	ev, err := s.Provider.BuildReleaseEventFromWebhook(body)
	if err != nil {
		return nil, err
	}
	if ev == nil {
		return nil, nil
	}

	return ev, nil
}

func (s *DefaultVcsService) ParseProviderReleaseWebhook(ctx *gin.Context, vcsName string, namespace string, name string) (*vcs.ReleaseEvent, error) {
	body, err := ctx.GetRawData()
	if err != nil {
		return nil, err
	}

	if err := s.Provider.Authenticate(ctx, body); err != nil {
		return nil, err
	}

	ev, err := s.Provider.BuildReleaseEventFromWebhook(body)
	if err != nil {
		return nil, err
	}
	if ev == nil {
		return nil, nil
	}

	return ev, nil
}

func (s *DefaultVcsService) BuildProviderCreateDTO(authorityID uuid.UUID, namespace string, name string, ev *vcs.ReleaseEvent) (*provider.CreateProviderDTO, error) {
	if ev == nil {
		return nil, fmt.Errorf("no release event to build provider from")
	}

	prefix := fmt.Sprintf("terraform-provider-%s_%s", name, ev.SemVer)

	var shasumsURL, shasumsSigURL string
	zips := make(map[string]vcs.ReleaseAsset)

	for _, a := range ev.Assets {
		n := a.Name
		switch {
		case strings.HasPrefix(n, prefix) && strings.HasSuffix(n, ".zip"):
			base := strings.TrimSuffix(n, ".zip")
			rest := strings.TrimPrefix(base, prefix)
			if rest == "" {
				continue
			}
			zips[rest] = a
		case n == prefix+"SHA256SUMS":
			shasumsURL = a.URL
		case n == prefix+"SHA256SUMS.sig":
			shasumsSigURL = a.URL
		}
	}

	if len(zips) == 0 {
		return nil, fmt.Errorf("no terraform provider zip assets found for %s", prefix)
	}

	var sums map[string]string
	if shasumsURL != "" {
		sumsBody, cleanup, err := s.Fetcher.FetchFile(fmt.Sprintf("%sSHA256SUMS", prefix), shasumsURL, nil)
		if err != nil {
			return nil, fmt.Errorf("fetch SHA256SUMS: %v", err)
		}
		defer cleanup()
		sums = vcs.ParseSHA256SUMS(sumsBody)
	}

	var platforms []provider.CreatePlatformDTO
	for key, asset := range zips {
		osName, arch, ok := splitOSArchFromZipBase(key)
		if !ok {
			continue
		}
		var sum string
		if sums != nil {
			var found bool
			sum, found = sums[asset.Name]
			if !found {
				return nil, fmt.Errorf("no sha256 entry for %s", asset.Name)
			}
		}
		platforms = append(platforms, provider.CreatePlatformDTO{
			System:       osName,
			Architecture: arch,
			Location:     asset.URL,
			ShaSum:       sum,
		})
	}

	if len(platforms) == 0 {
		return nil, fmt.Errorf("no recognized provider platforms in release assets")
	}

	var shaDTO provider.CreateProviderShaSumsDTO
	if shasumsURL != "" {
		shaDTO.URL = shasumsURL
		shaDTO.SignatureURL = shasumsSigURL
	}

	dto := &provider.CreateProviderDTO{
		AuthorityID: authorityID,
		Name:        name,
		Version:     ev.SemVer,
		Protocols:   []string{DefaultProviderProtocols},
		ShaSums:     shaDTO,
		Platforms:   platforms,
		Headers:     s.Provider.GetHeaders(),
	}
	return dto, nil
}

func splitOSArchFromZipBase(zipBase string) (osName, arch string, ok bool) {
	knownArch := []string{"amd64", "arm64", "386", "arm"}
	for _, a := range knownArch {
		suffix := "_" + a
		if strings.HasSuffix(zipBase, suffix) {
			osPart := strings.TrimSuffix(zipBase, suffix)
			if osPart == "" {
				return "", "", false
			}
			return osPart, a, true
		}
	}
	return "", "", false
}
