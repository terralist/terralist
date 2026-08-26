package oauth

import "time"

// Code is a short-lived, single-use authorization code together with the
// identity claims it authorizes. It is persisted so that any server replica can
// resolve a code issued by another replica during the token exchange.
type Code struct {
	Code       string    `gorm:"primaryKey"`
	Components string    `gorm:"type:text;not null"`
	ExpiresAt  time.Time `gorm:"not null"`
}

func (Code) TableName() string {
	return "oauth_codes"
}
