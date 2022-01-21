package server

import "github.com/sirupsen/logrus"

type UserConfig struct {
	GitHubClientID     string `mapstructure:"gh-client-id"`
	GitHubClientSecret string `mapstructure:"gh-client-secret"`
	GitHubOrganization string `mapstructure:"gh-organization"`
	LogLevel           string `mapstructure:"log-level"`
	OAuthProvider      string `mapstructure:"oauth-provider"`
	Port               int    `mapstructure:"port"`
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
