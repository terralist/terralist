package server

import (
	"slices"
	"testing"
)

func TestSplitHostnames(t *testing.T) {
	t.Run("returns the trimmed, lowercased hostnames", func(t *testing.T) {
		hostnames, err := splitHostnames(" registry.terraform.io, Registry.OpenTofu.org ,,")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !slices.Equal(hostnames, []string{"registry.terraform.io", "registry.opentofu.org"}) {
			t.Fatalf("unexpected hostnames %v", hostnames)
		}
	})

	t.Run("rejects an entry that is not a hostname", func(t *testing.T) {
		if _, err := splitHostnames("registry.terraform.io,https://registry.opentofu.org"); err == nil {
			t.Fatal("expected an error for a URL")
		}
	})
}
