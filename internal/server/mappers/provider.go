package mappers

import (
	"fmt"

	models "github.com/valentindeaconu/terralist/internal/server/models/provider"
)

type ProviderMapper struct{}

func (p *ProviderMapper) ProviderToVersionListProviderDTO(provider models.Provider) models.VersionListProviderDTO {
	out := models.VersionListProviderDTO{}

	for _, version := range provider.Versions {
		v := models.VersionListVersionDTO{
			Version:   version.Version,
			Protocols: version.Protocols,
		}

		for _, platform := range version.Platforms {
			plat := models.VersionListPlatformDTO{
				System:       platform.System,
				Architecture: platform.Architecture,
			}

			v.Platforms = append(v.Platforms, plat)
		}

		out.Versions = append(out.Versions, v)
	}

	return out
}

func (p *ProviderMapper) VersionToDownloadProviderDTO(version models.Version, system string, architecture string) (models.DownloadProviderDTO, error) {
	out := models.DownloadProviderDTO{
		System:       system,
		Architecture: architecture,
	}

	for _, platform := range version.Platforms {
		if platform.System == system && platform.Architecture == architecture {
			out.FileName = platform.FileName
			out.DownloadUrl = platform.DownloadUrl
			out.ShaSumsUrl = platform.ShaSumsUrl
			out.ShaSumsSignatureUrl = platform.ShaSumsSignatureUrl
			out.ShaSum = platform.ShaSum
			out.Protocols = version.Protocols

			for _, publicKey := range platform.SigningKeys {
				pk := models.GpgPublicKeyDTO{
					KeyId:          publicKey.KeyId,
					AsciiArmor:     publicKey.AsciiArmor,
					TrustSignature: publicKey.TrustSignature,
					Source:         publicKey.Source,
					SourceUrl:      publicKey.SourceUrl,
				}

				out.SigningKeys.GpgPublicKeys = append(out.SigningKeys.GpgPublicKeys, pk)
			}

			return out, nil
		}
	}

	return out, fmt.Errorf("no platform found for %s_%s machine", system, architecture)
}

func (p *ProviderMapper) CreateProviderDTOToProvider(dto models.CreateProviderDTO) models.Provider {
	out := models.Provider{
		Name:      dto.Name,
		Namespace: dto.Namespace,
		Versions: []models.Version{
			{
				Version:   dto.Version,
				Protocols: dto.Protocols,
			},
		},
	}

	for _, platform := range dto.Platforms {
		plat := models.Platform{
			System:              platform.System,
			Architecture:        platform.Architecture,
			FileName:            platform.FileName,
			DownloadUrl:         platform.DownloadUrl,
			ShaSumsUrl:          platform.ShaSumsUrl,
			ShaSumsSignatureUrl: platform.ShaSumsSignatureUrl,
			ShaSum:              platform.ShaSum,
		}

		for _, signingKey := range platform.SigningKeys.GpgPublicKeys {
			pk := models.GpgPublicKey{
				KeyId:          signingKey.KeyId,
				AsciiArmor:     signingKey.AsciiArmor,
				TrustSignature: signingKey.TrustSignature,
				Source:         signingKey.Source,
				SourceUrl:      signingKey.SourceUrl,
			}

			plat.SigningKeys = append(plat.SigningKeys, pk)
		}

		out.Versions[0].Platforms = append(out.Versions[0].Platforms, plat)
	}

	return out
}
