package server

type UserConfig struct {
	LogLevel                string `mapstructure:"log-level"`
	Port                    int    `mapstructure:"port"`
	MetricsPort             int    `mapstructure:"metrics-port"`
	URL                     string `mapstructure:"url"`
	Home                    string `mapstructure:"home"`
	CertFile                string `mapstructure:"cert-file"`
	KeyFile                 string `mapstructure:"key-file"`
	TokenSigningSecret      string `mapstructure:"token-signing-secret"`
	OauthProvider           string `mapstructure:"oauth-provider"`
	CustomCompanyName       string `mapstructure:"custom-company-name"`
	ModulesAnonymousRead    bool   `mapstructure:"modules-anonymous-read"`
	ProvidersAnonymousRead  bool   `mapstructure:"providers-anonymous-read"`
	LocalTokenSigningSecret string `mapstructure:"local-token-signing-secret"`
	SamlDisplayName         string `mapstructure:"saml-display-name"`
	RbacPolicyPath          string `mapstructure:"rbac-policy-path"`
	RbacDefaultRole         string `mapstructure:"rbac-default-role"`
	MasterApiKey            string `mapstructure:"master-api-key"`
	AuthTokenExpiration     string `mapstructure:"auth-token-expiration"`
}
