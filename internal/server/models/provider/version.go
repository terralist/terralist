package provider

import (
	"encoding/json"
	"strings"

	"terralist/internal/server/models/artifact"
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

const (
	// OriginManual marks artifacts uploaded by an operator, OriginUpstream
	// artifacts fetched from the upstream registry of the authority.
	OriginManual   = "manual"
	OriginUpstream = "upstream"
)

type Version struct {
	entity.Entity
	ProviderID          uuid.UUID
	Provider            Provider
	Version             string     `gorm:"not null"`
	Protocols           string     `gorm:"not null"`
	ShaSumsUrl          string     `gorm:"shasums_url"`
	ShaSumsSignatureUrl string     `gorm:"shasums_signature_url"`
	Platforms           []Platform `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	// SigningKeys is the JSON signing_keys block advertised by the upstream
	// registry for a version fetched from it; empty for uploaded versions,
	// which are signed with the authority keys.
	SigningKeys string
	Origin      string `gorm:"not null;default:manual"`
}

// GetPlatform returns the platform built for the given system and
// architecture, or nil.
func (v Version) GetPlatform(system, architecture string) *Platform {
	for i := range v.Platforms {
		if v.Platforms[i].System == system && v.Platforms[i].Architecture == architecture {
			return &v.Platforms[i]
		}
	}

	return nil
}

// SigningKeysDTO returns the signing keys stored with the version, if any.
func (v Version) SigningKeysDTO() (SigningKeysDTO, bool) {
	if v.SigningKeys == "" {
		return SigningKeysDTO{}, false
	}

	var keys SigningKeysDTO
	if err := json.Unmarshal([]byte(v.SigningKeys), &keys); err != nil {
		return SigningKeysDTO{}, false
	}

	return keys, true
}

func (Version) TableName() string {
	return "provider_versions"
}

// MirrorOnly reports whether the version lacks the SHA256SUMS file and its
// signature. Terraform refuses unsigned packages from a provider registry, so
// such a version is served through the network mirror protocol only.
func (v Version) MirrorOnly() bool {
	return v.ShaSumsUrl == "" || v.ShaSumsSignatureUrl == ""
}

func (v Version) ToVersionListVersionDTO() VersionListVersionDTO {
	var platforms []VersionListPlatformDTO
	for _, p := range v.Platforms {
		platforms = append(platforms, p.ToVersionListPlatformDTO())
	}

	return VersionListVersionDTO{
		Version:   v.Version,
		Protocols: strings.Split(v.Protocols, ","),
		Platforms: platforms,
	}
}

func (v Version) ToArtifactVersion() artifact.Version {
	return artifact.Version{
		Tag: v.Version,
	}
}

type VersionListVersionDTO struct {
	Version   string                   `json:"version"`
	Protocols []string                 `json:"protocols"`
	Platforms []VersionListPlatformDTO `json:"platforms"`
}
