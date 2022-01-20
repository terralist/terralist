package settings

import "github.com/valentindeaconu/terralist/models"

var ServiceDiscovery models.ServiceDiscovery = models.ServiceDiscovery{
	Login: models.Login{
		ClientName:            "terraform-cli",
		AuthorizationEndpoint: "/oauth/authorization",
		TokenEndpoint:         "/oauth/token",
		Ports:                 []int{10000, 10010},
	},
	ModuleEndpoint:   "/v1/modules",
	ProviderEndpoint: "/v1/providers",
}

// sha256.Sum256([]byte("abcd"))
var EncryptSalt string = "88d4266fd4e6338d13b845fcf289579d209c897823b9217da3e161936f031589"

// sha256.Sum256([]byte("abcd"))
var CodeExchangeKey string = "88d4266fd4e6338d13b845fcf289579d209c897823b9217da3e161936f031589"

// sha256.Sum256([]byte("abcd"))
var TokenSigningSecret []byte = []byte("88d4266fd4e6338d13b845fcf289579d209c897823b9217da3e161936f031589")
