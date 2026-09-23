package oauth

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrRedirectURINotAllowed = errors.New("redirect uri not allowed")
)

// TerraformPorts is the inclusive range of loopback ports a Terraform client
// may listen on to receive the authorization code, as advertised through the
// service discovery document.
var TerraformPorts = []int{10000, 10010}

var loopbackHosts = []string{"localhost", "127.0.0.1", "::1"}

// ValidateRedirectURI accepts only redirect URIs pointing back to the given
// host or to a loopback address on one of the Terraform ports.
func ValidateRedirectURI(redirectURI string, host string) error {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRedirectURINotAllowed, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: unsupported scheme %q", ErrRedirectURINotAllowed, u.Scheme)
	}

	if u.User != nil {
		return fmt.Errorf("%w: userinfo is not permitted", ErrRedirectURINotAllowed)
	}

	if strings.EqualFold(u.Host, host) {
		return nil
	}

	if isLoopback(u.Hostname()) && isTerraformPort(u.Port()) {
		return nil
	}

	return fmt.Errorf("%w: %q", ErrRedirectURINotAllowed, u.Host)
}

func isLoopback(hostname string) bool {
	for _, h := range loopbackHosts {
		if strings.EqualFold(hostname, h) {
			return true
		}
	}

	return false
}

func isTerraformPort(port string) bool {
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}

	return p >= TerraformPorts[0] && p <= TerraformPorts[1]
}
