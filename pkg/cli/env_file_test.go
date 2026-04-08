package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadEnvFilesSetsValueFromFile(t *testing.T) {
	t.Setenv("TERRALIST_OI_CLIENT_ID", "")
	t.Setenv("TERRALIST_OI_CLIENT_ID_FILE", "")

	secretPath := filepath.Join(t.TempDir(), "client-id")
	if err := os.WriteFile(secretPath, []byte("client-id-from-file"), 0o600); err != nil {
		t.Fatalf("write secret file: %v", err)
	}

	if err := os.Setenv("TERRALIST_OI_CLIENT_ID_FILE", secretPath); err != nil {
		t.Fatalf("set _FILE env: %v", err)
	}

	if err := os.Unsetenv("TERRALIST_OI_CLIENT_ID"); err != nil {
		t.Fatalf("unset target env: %v", err)
	}

	if err := LoadEnvFiles("TERRALIST_OI_CLIENT_ID"); err != nil {
		t.Fatalf("LoadEnvFiles() error = %v", err)
	}

	if got := os.Getenv("TERRALIST_OI_CLIENT_ID"); got != "client-id-from-file" {
		t.Fatalf("TERRALIST_OI_CLIENT_ID = %q, want %q", got, "client-id-from-file")
	}
}

func TestLoadEnvFilesErrorsWhenBothEnvAndFileAreSet(t *testing.T) {
	t.Setenv("TERRALIST_OI_CLIENT_SECRET", "inline-secret")

	secretPath := filepath.Join(t.TempDir(), "client-secret")
	if err := os.WriteFile(secretPath, []byte("secret-from-file"), 0o600); err != nil {
		t.Fatalf("write secret file: %v", err)
	}

	t.Setenv("TERRALIST_OI_CLIENT_SECRET_FILE", secretPath)

	err := LoadEnvFiles("TERRALIST_OI_CLIENT_SECRET")
	if err == nil {
		t.Fatal("LoadEnvFiles() error = nil, want conflict error")
	}

	if !strings.Contains(err.Error(), "cannot both be set") {
		t.Fatalf("LoadEnvFiles() error = %q, want conflict message", err)
	}
}
