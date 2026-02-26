package sqlite

import (
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite/lib"
)

func TestConfigSetDefaults_UsesHome(t *testing.T) {
	home := t.TempDir()
	cfg := &Config{Home: home}
	cfg.SetDefaults()

	want := filepath.Join(home, "data", "storage.db")
	if cfg.Path != want {
		t.Fatalf("Config.SetDefaults: got %q, want %q", cfg.Path, want)
	}
}

func TestConfigSetDefaults_PreservesExplicitPath(t *testing.T) {
	home := t.TempDir()
	explicit := filepath.Join(home, "custom.db")
	cfg := &Config{Path: explicit, Home: home}
	cfg.SetDefaults()

	if cfg.Path != explicit {
		t.Fatalf("Config.SetDefaults cleared explicit path: got %q, want %q", cfg.Path, explicit)
	}
}

func TestConfigValidate_CreatesDirectory(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nested", "sub", "storage.db")
	cfg := &Config{Path: path}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	dir := filepath.Dir(path)
	st, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected directory %q to exist: %v", dir, err)
	}
	if !st.IsDir() {
		t.Fatalf("expected %q to be a directory", dir)
	}
}

func TestConfigValidate_FailsWhenParentIsFile(t *testing.T) {
	tmp := t.TempDir()
	// Create a file where a directory is expected.
	blocker := filepath.Join(tmp, "blocked")
	if err := os.WriteFile(blocker, []byte("I block the path"), 0644); err != nil {
		t.Fatalf("failed to create blocker file: %v", err)
	}

	// Use the file path as the parent directory for the DB path.
	dbPath := filepath.Join(blocker, "storage.db")
	cfg := &Config{Path: dbPath}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("Validate should have failed when parent path is a file")
	}
}

func TestConfigDSN_IncludesTimeFormatAndPath(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "storage.db")

	cfg := &Config{Path: dbPath}
	dsn := cfg.DSN()

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("failed to parse DSN %q: %v", dsn, err)
	}

	if u.Path != cfg.Path {
		t.Fatalf("DSN path = %q, want %q", u.Path, cfg.Path)
	}

	q := u.Query()
	if q.Get("_time_format") != "sqlite" {
		t.Fatalf("DSN query _time_format = %q, want %q", q.Get("_time_format"), "sqlite")
	}
}

// Integration test: ensure the sqlite driver can open the DSN and perform basic CRUD.
func TestSQLiteDriver_ConnectAndWrite(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "integration.db")

	cfg := &Config{Path: dbPath}
	// Ensure parent dir exists
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	dsn := cfg.DSN()

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("failed to open sqlite db with DSN %q: %v", dsn, err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS kv (k TEXT PRIMARY KEY, v TEXT);`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Insert a row
	_, err = db.Exec(`INSERT INTO kv(k, v) VALUES(?, ?)`, "foo", "bar")
	if err != nil {
		t.Fatalf("failed to insert row: %v", err)
	}

	// Read it back
	var v string
	err = db.QueryRow(`SELECT v FROM kv WHERE k = ?`, "foo").Scan(&v)
	if err != nil {
		t.Fatalf("failed to query row: %v", err)
	}
	if v != "bar" {
		t.Fatalf("unexpected value from db: got %q, want %q", v, "bar")
	}
}
