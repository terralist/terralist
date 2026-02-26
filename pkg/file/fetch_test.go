package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
