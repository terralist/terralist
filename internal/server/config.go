package server

type UserConfig struct {
	LogLevel               string `mapstructure:"log-level"`
	Port                   int    `mapstructure:"port"`
	URL                    string `mapstructure:"url"`
	CertFile               string `mapstructure:"cert-file"`
	KeyFile                string `mapstructure:"key-file"`
	TokenSigningSecret     string `mapstructure:"token-signing-secret"`
	OauthProvider          string `mapstructure:"oauth-provider"`
	CustomCompanyName      string `mapstructure:"custom-company-name"`
	ModulesAnonymousRead   bool   `mapstructure:"modules-anonymous-read"`
	ProvidersAnonymousRead bool   `mapstructure:"providers-anonymous-read"`
}
