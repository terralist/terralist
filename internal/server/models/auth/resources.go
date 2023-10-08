package auth

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidResources = errors.New("cannot decompose resources")
)

type ResourceType string

const (
	ResourceModule    = "module"
	ResourceProvider  = "provider"
	ResourceAuthority = "authority"
	ResourceApiKey    = "apiKey"
)

func AnyOfResources(t ResourceType) Resource {
	return Resource(fmt.Sprintf("%s:*", string(t)))
}

type Resource = string

func ComposeResource(resource ResourceType, identifier string) Resource {
	return Resource(fmt.Sprintf("%s:%s", resource, identifier))
}

func DecomposeResource(resource Resource) (ResourceType, string) {
	tokens := strings.SplitN(resource, ":", 2)
	return ResourceType(tokens[0]), tokens[1]
}
