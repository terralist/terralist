package file

import (
	"net"
	"net/http"
	"time"
)

// NewHTTPClient returns an HTTP client with the given overall timeout. Unless
// allowPrivateAddresses is true, it refuses to connect to private, loopback,
// link-local and unspecified addresses, the same way the fetcher does.
func NewHTTPClient(allowPrivateAddresses bool, timeout time.Duration) *http.Client {
	client := &http.Client{Timeout: timeout}

	if !allowPrivateAddresses {
		defaultTransport, _ := http.DefaultTransport.(*http.Transport)
		transport := defaultTransport.Clone()
		transport.DialContext = (&net.Dialer{Control: privateAddressGuard}).DialContext
		client.Transport = transport
	}

	return client
}
