package server

type UserConfig struct {
	LogLevel           string `mapstructure:"log-level"`
	Port               int    `mapstructure:"port"`
	URL                string `mapstructure:"url"`
	TokenSigningSecret string `mapstructure:"token-signing-secret"`
}
