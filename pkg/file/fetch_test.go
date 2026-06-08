package file

import (
	"archive/zip"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFetchArchive_LoadsLocalArchive(t *testing.T) {
	// Build a zip archive on disk, mimicking an uploaded module file.
	tempDir, err := os.MkdirTemp("", "test-fetch-archive-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "module.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Failed to create zip: %v", err)
	}
	zw := zip.NewWriter(zf)
	w, err := zw.Create("main.tf")
	if err != nil {
		t.Fatalf("Failed to add zip entry: %v", err)
	}
	if _, err := w.Write([]byte("# module content")); err != nil {
		t.Fatalf("Failed to write zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}
	zf.Close()

	uploaded, err := LoadFromDisk("module.zip", zipPath)
	if err != nil {
		t.Fatalf("Failed to load uploaded archive: %v", err)
	}

	// Loading the archive must not require any URL scheme.
	result, cleanup, err := fetchArchive("module", uploaded)
	if err != nil {
		t.Fatalf("fetchArchive failed: %v", err)
	}
	defer cleanup()
	defer result.Close()

	archiveFile, ok := result.(*ArchiveFile)
	if !ok {
		t.Fatalf("Expected *ArchiveFile, got %T", result)
	}

	if _, found := archiveFile.FS().files["main.tf"]; !found {
		t.Errorf("Expected archive FS to contain main.tf, got files: %v", archiveFile.FS().files)
	}
}

func TestFetch_RejectsFileScheme(t *testing.T) {
	// Create a local file to read
	tempFile, err := os.CreateTemp("", "test-fetch-file-scheme-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString("sensitive content"); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tempFile.Close()

	// Fetching through the file:// scheme must not be allowed
	_, cleanup, err := fetch("test.txt", "file://"+tempFile.Name(), "", file, nil, false)
	if cleanup != nil {
		cleanup()
	}

	if err == nil {
		t.Fatal("Expected fetch to reject the file:// scheme, but it succeeded")
	}

	if !errors.Is(err, ErrDownloadFailure) {
		t.Errorf("Expected ErrDownloadFailure, got: %v", err)
	}
}

func TestFetch_BlocksPrivateAddressByDefault(t *testing.T) {
	// httptest binds to a loopback address, which the guard must reject
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("internal content"))
	}))
	defer server.Close()

	_, cleanup, err := fetch("test.txt", server.URL, "", file, nil, false)
	if cleanup != nil {
		cleanup()
	}

	if err == nil {
		t.Fatal("Expected fetch to reject the private address, but it succeeded")
	}

	if !errors.Is(err, ErrDownloadFailure) {
		t.Errorf("Expected ErrDownloadFailure, got: %v", err)
	}
}

func TestFetch_AllowsPrivateAddressWhenEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("internal content"))
	}))
	defer server.Close()

	result, cleanup, err := fetch("test.txt", server.URL, "", file, nil, true)
	if err != nil {
		t.Fatalf("fetch failed with private addresses allowed: %v", err)
	}
	defer cleanup()
	defer result.Close()
}

func TestFetch_CleanupRemovesTempDir(t *testing.T) {
	// Serve a file over HTTP to fetch
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test content"))
	}))
	defer server.Close()

	// Fetch the file
	result, cleanup, err := fetch("test.txt", server.URL, "", file, nil, true)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	defer result.Close()

	// The cleanup function should be non-nil
	if cleanup == nil {
		t.Fatal("Expected non-nil cleanup function")
	}

	// Find the temp dir by checking that tl-fetch dirs exist
	matches, _ := filepath.Glob(filepath.Join(os.TempDir(), "tl-fetch*"))
	if len(matches) == 0 {
		t.Fatal("Expected at least one tl-fetch temp dir to exist before cleanup")
	}

	// Run cleanup
	cleanup()

	// Verify temp dirs created by this test are gone
	// (we can't be 100% sure which one is ours, but the count should decrease)
	matchesAfter, _ := filepath.Glob(filepath.Join(os.TempDir(), "tl-fetch*"))
	if len(matchesAfter) >= len(matches) {
		t.Error("Expected cleanup to remove the temp dir")
	}
}

func TestArchiveDir_StripRootFolder(t *testing.T) {
	// Create a temporary directory structure that mimics a GitHub archive
	tempDir, err := os.MkdirTemp("", "test-archive-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a root folder (like terraform-aws-eks-21.15.1)
	rootFolder := "terraform-aws-eks-21.15.1"
	rootPath := filepath.Join(tempDir, rootFolder)

	// Create directory structure
	modulesPath := filepath.Join(rootPath, "modules", "karpenter")
	if err := os.MkdirAll(modulesPath, 0755); err != nil {
		t.Fatalf("Failed to create modules dir: %v", err)
	}

	// Create a main.tf file in the root
	rootMainTf := filepath.Join(rootPath, "main.tf")
	if err := os.WriteFile(rootMainTf, []byte("# Root module"), 0644); err != nil {
		t.Fatalf("Failed to create root main.tf: %v", err)
	}

	// Create a main.tf file in the submodule
	submoduleMainTf := filepath.Join(modulesPath, "main.tf")
	if err := os.WriteFile(submoduleMainTf, []byte("# Karpenter submodule"), 0644); err != nil {
		t.Fatalf("Failed to create submodule main.tf: %v", err)
	}

	// Archive the directory
	archive, err := archiveDir("test.zip", tempDir)
	if err != nil {
		t.Fatalf("archiveDir failed: %v", err)
	}
	defer archive.Close()

	// Verify it's an ArchiveFile
	archiveFile, ok := archive.(*ArchiveFile)
	if !ok {
		t.Fatal("Expected ArchiveFile type")
	}

	// Get the filesystem from the archive
	fs := archiveFile.FS()

	// Check if files have the root folder stripped
	foundRootMain := false
	foundSubmoduleMain := false
	hasRootFolder := false

	for name := range fs.files {
		t.Logf("Archive contains: %s", name)

		// Check if any file still has the root folder prefix
		if strings.HasPrefix(name, rootFolder+"/") {
			hasRootFolder = true
		}

		// Check for expected files without root folder
		if name == "main.tf" {
			foundRootMain = true
		}
		if name == "modules/karpenter/main.tf" {
			foundSubmoduleMain = true
		}
	}

	if hasRootFolder {
		t.Errorf("Archive still contains root folder prefix %s/", rootFolder)
	}

	if !foundRootMain {
		t.Error("Archive missing main.tf at root")
	}

	if !foundSubmoduleMain {
		t.Error("Archive missing modules/karpenter/main.tf")
	}
}

func TestArchiveDir_NoStripWhenMultipleRoots(t *testing.T) {
	// Create a temporary directory with multiple root-level items
	tempDir, err := os.MkdirTemp("", "test-archive-multi-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create files at different root levels
	file1 := filepath.Join(tempDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	dir1 := filepath.Join(tempDir, "dir1")
	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}

	file2 := filepath.Join(dir1, "file2.txt")
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Archive the directory
	archive, err := archiveDir("test.zip", tempDir)
	if err != nil {
		t.Fatalf("archiveDir failed: %v", err)
	}
	defer archive.Close()

	// Verify it's an ArchiveFile
	archiveFile, ok := archive.(*ArchiveFile)
	if !ok {
		t.Fatal("Expected ArchiveFile type")
	}

	// Get the filesystem from the archive
	fs := archiveFile.FS()

	// Check that files are NOT stripped (because there are multiple roots)
	foundFile1 := false
	foundFile2InDir := false

	for name := range fs.files {
		t.Logf("Archive contains: %s", name)

		if name == "file1.txt" {
			foundFile1 = true
		}
		if name == "dir1/file2.txt" {
			foundFile2InDir = true
		}
	}

	if !foundFile1 {
		t.Error("Archive should contain file1.txt at root level")
	}

	if !foundFile2InDir {
		t.Error("Archive should contain dir1/file2.txt")
	}
}
