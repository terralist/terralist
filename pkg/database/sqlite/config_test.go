package sqlite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPathUsesProvidedHome(t *testing.T) {
	home := "/custom/home"
	want := filepath.Join(home, "data", "storage.db")

	if got := DefaultPath(home); got != want {
		t.Fatalf("DefaultPath(%q) = %q, want %q", home, got, want)
	}
}

func TestDefaultPathUsesEnvWhenHomeEmpty(t *testing.T) {
	const envHome = "/env/terralist"

	t.Setenv("TERRALIST_HOME", envHome)

	got := DefaultPath("")
	want := filepath.Join(envHome, "data", "storage.db")
	if got != want {
		t.Fatalf("DefaultPath(empty) = %q, want %q", got, want)
	}
}

func TestDefaultPathFallsBackToUserHome(t *testing.T) {
	t.Setenv("TERRALIST_HOME", "")

	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not determine user home dir: %v", err)
	}

	got := DefaultPath("")
	want := filepath.Join(userHome, "data", "storage.db")
	if got != want {
		t.Fatalf("DefaultPath(empty) = %q, want %q", got, want)
	}
}

func TestConfigSetDefaults(t *testing.T) {
	const home = "/example/home"

	cfg := &Config{Home: home}
	cfg.SetDefaults()

	want := filepath.Join(home, "data", "storage.db")
	if cfg.Path != want {
		t.Fatalf("Config.SetDefaults: got %q, want %q", cfg.Path, want)
	}

	preexisting := &Config{
		Path: "/data/db.sqlite",
		Home: home,
	}
	preexisting.SetDefaults()
	if preexisting.Path != "/data/db.sqlite" {
		t.Fatalf("Config.SetDefaults overwrote path: got %q, want %q", preexisting.Path, "/data/db.sqlite")
	}
}

func TestConfigSetDefaultsHonorsExplicitPath(t *testing.T) {
	const home = "/custom/home"
	const explicit = "/tmp/terralist-data/storage.db"

	cfg := &Config{
		Path: explicit,
		Home: home,
	}

	cfg.SetDefaults()
	if cfg.Path != explicit {
		t.Fatalf("Config.SetDefaults cleared explicit path: got %q, want %q", cfg.Path, explicit)
	}
}
