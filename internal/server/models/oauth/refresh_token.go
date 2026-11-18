package oauth

import "time"

// RefreshToken is a refresh session issued to an OAuth client. Only hashes of
// the Terralist refresh token are stored; the upstream grant is sealed with the
// refresh secret, scoped to the session ID.
type RefreshToken struct {
	ID string `gorm:"primaryKey;size:36"`

	// TokenHash is the SHA-256 hash of the refresh token the client holds.
	TokenHash string `gorm:"size:64;uniqueIndex;not null"`

	// PreviousTokenHash is the hash of the token rotated out by the last
	// refresh. Presenting it again means the token leaked.
	PreviousTokenHash string `gorm:"size:64;index"`

	ClientID string `gorm:"not null"`
	Subject  string `gorm:"not null"`

	UpstreamGrant string `gorm:"type:text;not null"`

	ExpiresAt        time.Time `gorm:"not null"`
	SessionExpiresAt *time.Time
	CreatedAt        time.Time
}

func (RefreshToken) TableName() string {
	return "oauth_refresh_tokens"
}
