package provider

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/valentindeaconu/terralist/model"
)

// ORM
type Provider struct {
	model.Base
	Name      string
	Namespace string
	Versions  []Version
}

func (Provider) TableName() string {
	return "providers"
}

type Version struct {
	model.Base
	ProviderID uuid.UUID
	Version    string
	Protocols  pq.StringArray `gorm:"type:text[]"`
	Platforms  []Platform
}

func (Version) TableName() string {
	return "provider_versions"
}

type Platform struct {
	model.Base
	VersionID           uuid.UUID
	System              string
	Architecture        string
	FileName            string
	DownloadUrl         string
	ShaSumsUrl          string
	ShaSumsSignatureUrl string
	ShaSum              string
	SigningKeys         []GpgPublicKey
}

func (Platform) TableName() string {
	return "provider_platforms"
}

type GpgPublicKey struct {
	model.Base
	PlatformID     uuid.UUID
	KeyId          string
	AsciiArmor     string
	TrustSignature string
	Source         string
	SourceUrl      string
}

func (GpgPublicKey) TableName() string {
	return "provider_public_keys"
}
