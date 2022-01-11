package settings

import "github.com/valentindeaconu/terralist/models"

var ServiceDiscovery models.ServiceDiscovery = models.ServiceDiscovery{
	LoginEndpoint:    "/v1/login",
	ModuleEndpoint:   "/v1/modules",
	ProviderEndpoint: "/v1/providers",
}
