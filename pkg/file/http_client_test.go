package file

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPClientGuardsPrivateAddresses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if _, err := NewHTTPClient(false, time.Second).Get(server.URL); err == nil {
		t.Fatalf("expected the loopback address to be refused")
	}

	resp, err := NewHTTPClient(true, time.Second).Get(server.URL)
	if err != nil {
		t.Fatalf("expected the loopback address to be allowed, got: %v", err)
	}
	resp.Body.Close()
}
