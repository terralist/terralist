package server

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
