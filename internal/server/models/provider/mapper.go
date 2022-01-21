package provider

import "fmt"

func (m *Provider) ToVersionListProvider() VersionListProviderDTO {
	out := VersionListProviderDTO{}

	for _, version := range m.Versions {
		v := VersionListVersionDTO{
			Version:   version.Version,
			Protocols: version.Protocols,
		}

		for _, platform := range version.Platforms {
			p := VersionListPlatformDTO{
				System:       platform.System,
				Architecture: platform.Architecture,
			}

			v.Platforms = append(v.Platforms, p)
		}

		out.Versions = append(out.Versions, v)
	}

	return out
}

func (m *Version) ToDownloadProvider(system string, architecture string) (DownloadProviderDTO, error) {
	out := DownloadProviderDTO{
		System:       system,
		Architecture: architecture,
	}

	for _, platform := range m.Platforms {
		if platform.System == system && platform.Architecture == architecture {
			out.FileName = platform.FileName
			out.DownloadUrl = platform.DownloadUrl
			out.ShaSumsUrl = platform.ShaSumsUrl
			out.ShaSumsSignatureUrl = platform.ShaSumsSignatureUrl
			out.ShaSum = platform.ShaSum
			out.Protocols = m.Protocols

			for _, publicKey := range platform.SigningKeys {
				pk := GpgPublicKeyDTO{
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

func FromCreateDTO(dto CreateProviderDTO) Provider {
	out := Provider{
		Name:      dto.Name,
		Namespace: dto.Namespace,
		Versions: []Version{
			{
				Version:   dto.Version,
				Protocols: dto.Protocols,
			},
		},
	}

	for _, platform := range dto.Platforms {
		p := Platform{
			System:              platform.System,
			Architecture:        platform.Architecture,
			FileName:            platform.FileName,
			DownloadUrl:         platform.DownloadUrl,
			ShaSumsUrl:          platform.ShaSumsUrl,
			ShaSumsSignatureUrl: platform.ShaSumsSignatureUrl,
			ShaSum:              platform.ShaSum,
		}

		for _, signingKey := range platform.SigningKeys.GpgPublicKeys {
			pk := GpgPublicKey{
				KeyId:          signingKey.KeyId,
				AsciiArmor:     signingKey.AsciiArmor,
				TrustSignature: signingKey.TrustSignature,
				Source:         signingKey.Source,
				SourceUrl:      signingKey.SourceUrl,
			}

			p.SigningKeys = append(p.SigningKeys, pk)
		}

		out.Versions[0].Platforms = append(out.Versions[0].Platforms, p)
	}

	return out
}
