package models

type Login struct {
	ClientName            string
	AuthorizationEndpoint string
	TokenEndpoint         string
	Ports                 []int
}

type ServiceDiscovery struct {
	Login            Login
	ModuleEndpoint   string
	ProviderEndpoint string
}
