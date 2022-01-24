package server

import (
	"github.com/sirupsen/logrus"
	databaseConfig "github.com/valentindeaconu/terralist/internal/server/database/config"
	oauthConfig "github.com/valentindeaconu/terralist/internal/server/oauth/config"
)

type UserConfig struct {
	DatabaseBackend    string `mapstructure:"database-backend"`
	GitHubClientID     string `mapstructure:"gh-client-id"`
	GitHubClientSecret string `mapstructure:"gh-client-secret"`
	GitHubOrganization string `mapstructure:"gh-organization"`
	LogLevel           string `mapstructure:"log-level"`
	OAuthProvider      string `mapstructure:"oauth-provider"`
	Port               int    `mapstructure:"port"`
	TokenSigningSecret string `mapstructure:"token-signing-secret"`
}

func (u UserConfig) ToLogLevel() logrus.Level {
	switch u.LogLevel {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	}

	return logrus.InfoLevel
}

func (u UserConfig) ToOAuthProviderConfig() oauthConfig.OAuthConfig {
	return oauthConfig.OAuthConfig{
		"GitHubClientID":     u.GitHubClientID,
		"GitHubClientSecret": u.GitHubClientSecret,
		"GitHubOrganization": u.GitHubOrganization,
	}
}

func (u UserConfig) ToDatabaseConfig() databaseConfig.DatabaseConfig {
	return databaseConfig.DatabaseConfig{}
}
