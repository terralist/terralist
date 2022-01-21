package provider

type VersionListProviderDTO struct {
	Versions []VersionListVersionDTO `json:"versions"`
}

type CreateProviderDTO struct {
	Name      string              `json:"name"`
	Namespace string              `json:"namespace"`
	Version   string              `json:"version"`
	Protocols []string            `json:"protocols"`
	Platforms []CreatePlatformDTO `json:"platforms"`
}

type VersionListVersionDTO struct {
	Version   string                   `json:"version"`
	Protocols []string                 `json:"protocols"`
	Platforms []VersionListPlatformDTO `json:"platforms"`
}

type VersionListPlatformDTO struct {
	System       string `json:"os"`
	Architecture string `json:"arch"`
}

type CreatePlatformDTO struct {
	System              string         `json:"os"`
	Architecture        string         `json:"arch"`
	FileName            string         `json:"filename"`
	DownloadUrl         string         `json:"download_url"`
	ShaSumsUrl          string         `json:"shasums_url"`
	ShaSumsSignatureUrl string         `json:"shasums_signature_url"`
	ShaSum              string         `json:"shasum"`
	SigningKeys         SigningKeysDTO `json:"signing_keys"`
}

type DownloadProviderDTO struct {
	Protocols           []string       `json:"protocols"`
	System              string         `json:"os"`
	Architecture        string         `json:"arch"`
	FileName            string         `json:"filename"`
	DownloadUrl         string         `json:"download_url"`
	ShaSumsUrl          string         `json:"shasums_url"`
	ShaSumsSignatureUrl string         `json:"shasums_signature_url"`
	ShaSum              string         `json:"shasum"`
	SigningKeys         SigningKeysDTO `json:"signing_keys"`
}

type SigningKeysDTO struct {
	GpgPublicKeys []GpgPublicKeyDTO `json:"gpg_public_keys"`
}

type GpgPublicKeyDTO struct {
	KeyId          string `json:"key_id"`
	AsciiArmor     string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
	Source         string `json:"source"`
	SourceUrl      string `json:"source_url"`
}
