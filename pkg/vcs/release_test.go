package vcs

import (
	"terralist/pkg/file"
	"testing"
)

func TestRepoURLsMatch(t *testing.T) {
	if !RepoURLsMatch("https://github.com/A/B", "git@github.com:a/b.git") {
		t.Fatal("expected match")
	}
	if RepoURLsMatch("https://github.com/a/b", "https://github.com/a/c") {
		t.Fatal("expected mismatch")
	}
}

func TestParseSHA256SUMS(t *testing.T) {
	hash := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	if len(hash) != 64 {
		t.Fatal("test hash length")
	}
	data := []byte(hash + "  *terraform-provider-x_1.0.0_linux_amd64.zip\n")
	m := ParseSHA256SUMS(file.NewInMemoryFile("SHA256SUMS", data))
	if m["terraform-provider-x_1.0.0_linux_amd64.zip"] == "" {
		t.Fatalf("%v", m)
	}
}
