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
