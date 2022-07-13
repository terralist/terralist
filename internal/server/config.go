package server

type UserConfig struct {
	LogLevel           string `mapstructure:"log-level"`
	Port               int    `mapstructure:"port"`
	TokenSigningSecret string `mapstructure:"token-signing-secret"`
}
