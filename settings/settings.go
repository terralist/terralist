package settings

import "github.com/valentindeaconu/terralist/model"

var ServiceDiscovery model.ServiceDiscovery = model.ServiceDiscovery{
	LoginEndpoint:    "/v1/login",
	ModuleEndpoint:   "/v1/modules",
	ProviderEndpoint: "/v1/providers",
}
 