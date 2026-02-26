package docs

import (
	"errors"
	"strings"
	"testing"

	"terralist/pkg/file"
)

func TestFindSubmodules(t *testing.T) {
	tests := []struct {
		name          string
		fs            *file.FS
		expectedCount int
		expectedPaths []string
	}{
		{
			name: "No submodules directory",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`
					variable "test" {}
					output "test" {}
				`)),
				file.NewInMemoryFile("README.md", []byte(`# my module\n`)),
			}),
			expectedCount: 0,
			expectedPaths: []string{},
		},
		{
			name: "Empty modules directory",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`
					variable "test" {}
				`)),
				file.NewInMemoryFile("modules/.gitkeep", []byte("")),
			}),
			expectedCount: 0,
			expectedPaths: []string{},
		},
		{
			name: "Single submodule in modules directory",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`
					variable "test" {}
				`)),
				file.NewInMemoryFile("modules/vpc/main.tf", []byte(`
					variable "cidr" {}
					output "vpc_id" {}
				`)),
				file.NewInMemoryFile("modules/vpc/README.md", []byte(`# VPC Module\n`)),
			}),
			expectedCount: 1,
			expectedPaths: []string{"modules/vpc"},
		},
		{
			name: "Multiple submodules in modules directory",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`
					variable "test" {}
				`)),
				file.NewInMemoryFile("modules/vpc/main.tf", []byte(`
					variable "cidr" {}
				`)),
				file.NewInMemoryFile("modules/vpc/README.md", []byte(`# VPC Module\n`)),
				file.NewInMemoryFile("modules/subnet/main.tf", []byte(`
					variable "vpc_id" {}
				`)),
				file.NewInMemoryFile("modules/subnet/README.md", []byte(`# Subnet Module\n`)),
			}),
			expectedCount: 2,
			expectedPaths: []string{"modules/vpc", "modules/subnet"},
		},
		{
			name: "Submodules in submodules directory",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`
					variable "test" {}
				`)),
				file.NewInMemoryFile("submodules/storage/main.tf", []byte(`
					variable "bucket_name" {}
				`)),
				file.NewInMemoryFile("submodules/storage/README.md", []byte(`# Storage Module\n`)),
			}),
			expectedCount: 1,
			expectedPaths: []string{"submodules/storage"},
		},
		{
			name: "Nested submodules",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`
					variable "test" {}
				`)),
				file.NewInMemoryFile("modules/networking/vpc/main.tf", []byte(`
					variable "cidr" {}
				`)),
				file.NewInMemoryFile("modules/networking/subnet/main.tf", []byte(`
					variable "vpc_id" {}
				`)),
			}),
			expectedCount: 2,
			expectedPaths: []string{"modules/networking/vpc", "modules/networking/subnet"},
		},
		{
			name: "Submodule without main.tf but with other tf files",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`
					variable "test" {}
				`)),
				file.NewInMemoryFile("modules/config/variables.tf", []byte(`
					variable "setting" {}
				`)),
				file.NewInMemoryFile("modules/config/outputs.tf", []byte(`
					output "value" {}
				`)),
				file.NewInMemoryFile("modules/config/README.md", []byte(`# Config Module\n`)),
			}),
			expectedCount: 1,
			expectedPaths: []string{"modules/config"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submodules, err := FindSubmodules(tt.fs)
			if err != nil {
				t.Fatalf("FindSubmodules() returned unexpected error: %v", err)
			}

			if len(submodules) != tt.expectedCount {
				t.Errorf("FindSubmodules() returned %d submodules, expected %d", len(submodules), tt.expectedCount)
			}

			// Check if all expected paths are present
			foundPaths := make(map[string]bool)
			for _, sm := range submodules {
				foundPaths[sm.Path] = true
			}

			for _, expectedPath := range tt.expectedPaths {
				if !foundPaths[expectedPath] {
					t.Errorf("Expected submodule path %s not found", expectedPath)
				}
			}
		})
	}
}

func TestFindSubmodules_WithDocumentation(t *testing.T) {
	fs := file.MustNewFS([]file.File{
		file.NewInMemoryFile("main.tf", []byte(`
			variable "root_var" {}
		`)),
		file.NewInMemoryFile("modules/vpc/main.tf", []byte(`
			variable "cidr" {
				description = "VPC CIDR block"
			}
			output "vpc_id" {
				description = "The VPC ID"
			}
		`)),
		file.NewInMemoryFile("modules/vpc/README.md", []byte(`# VPC Module

This is a VPC submodule.
`)),
	})

	submodules, err := FindSubmodules(fs)
	if err != nil {
		t.Fatalf("FindSubmodules() returned unexpected error: %v", err)
	}

	if len(submodules) != 1 {
		t.Fatalf("Expected 1 submodule, got %d", len(submodules))
	}

	sm := submodules[0]
	if sm.Path != "modules/vpc" {
		t.Errorf("Expected path 'modules/vpc', got '%s'", sm.Path)
	}

	// Should use the README.md content
	expectedDoc := "# VPC Module\n\nThis is a VPC submodule.\n"
	if sm.Documentation != expectedDoc {
		t.Errorf("Expected documentation to match README.md content.\nExpected: %q\nGot: %q", expectedDoc, sm.Documentation)
	}
}

func TestFindSubmodules_NoCollisions(t *testing.T) {
	// Test case to ensure no filename collisions can occur
	// when converting submodule paths to filenames
	fs := file.MustNewFS([]file.File{
		file.NewInMemoryFile("main.tf", []byte(`
			variable "root_var" {}
		`)),
		// These would collide if using single underscore: modules_net_vpc
		file.NewInMemoryFile("modules/net/vpc/main.tf", []byte(`
			variable "cidr" {
				description = "VPC in net/vpc directory"
			}
		`)),
		file.NewInMemoryFile("modules/net_vpc/main.tf", []byte(`
			variable "vpc_cidr" {
				description = "VPC in net_vpc directory"
			}
		`)),
		// These would also collide: modules_networking_subnet
		file.NewInMemoryFile("modules/networking/subnet/main.tf", []byte(`
			variable "subnet_id" {}
		`)),
		file.NewInMemoryFile("modules/networking_subnet/main.tf", []byte(`
			variable "other_id" {}
		`)),
	})

	submodules, err := FindSubmodules(fs)
	if err != nil {
		t.Fatalf("FindSubmodules() returned unexpected error: %v", err)
	}

	if len(submodules) != 4 {
		t.Fatalf("Expected 4 submodules, got %d", len(submodules))
	}

	// Verify all expected submodules are present
	expectedPaths := map[string]bool{
		"modules/net/vpc":           false,
		"modules/net_vpc":           false,
		"modules/networking/subnet": false,
		"modules/networking_subnet": false,
	}

	for _, sm := range submodules {
		if _, exists := expectedPaths[sm.Path]; exists {
			expectedPaths[sm.Path] = true
		} else {
			t.Errorf("Unexpected submodule path: %s", sm.Path)
		}
	}

	for path, found := range expectedPaths {
		if !found {
			t.Errorf("Expected submodule path %s not found", path)
		}
	}

	// Test that filename generation doesn't create collisions
	// Using double underscore: "/" -> "__"
	fileNames := make(map[string][]string)
	version := "1.0.0"

	for _, sm := range submodules {
		fileName := version + "_" + strings.ReplaceAll(sm.Path, "/", "__") + ".md"
		fileNames[fileName] = append(fileNames[fileName], sm.Path)
	}

	// Check for collisions
	for fileName, paths := range fileNames {
		if len(paths) > 1 {
			t.Errorf("Filename collision detected for '%s': %v", fileName, paths)
		}
	}

	// Verify expected unique filenames
	expectedFileNames := map[string]bool{
		"1.0.0_modules__net__vpc.md":           true,
		"1.0.0_modules__net_vpc.md":            true,
		"1.0.0_modules__networking__subnet.md": true,
		"1.0.0_modules__networking_subnet.md":  true,
	}

	for fileName := range fileNames {
		if !expectedFileNames[fileName] {
			t.Errorf("Unexpected filename: %s", fileName)
		}
	}
}

func TestFindSubmodules_ErrorHandling(t *testing.T) {
	tests := []struct {
		name                string
		fs                  *file.FS
		expectedSubmodules  int
		expectedDocContains string
		shouldHaveMessage   bool
	}{
		{
			name: "Submodule with main.tf generates documentation",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`variable "root" {}`)),
				file.NewInMemoryFile("modules/working/main.tf", []byte(`
					variable "test" {
						description = "Test variable"
					}
				`)),
			}),
			expectedSubmodules:  1,
			expectedDocContains: "Input Variables",
			shouldHaveMessage:   false,
		},
		{
			name: "Submodule with README.md uses it",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`variable "root" {}`)),
				file.NewInMemoryFile("modules/documented/README.md", []byte(`# Custom Docs`)),
			}),
			expectedSubmodules:  1,
			expectedDocContains: "# Custom Docs",
			shouldHaveMessage:   false,
		},
		{
			name: "Empty submodule directory generates default documentation",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`variable "root" {}`)),
				file.NewInMemoryFile("modules/empty/.gitkeep", []byte(``)),
			}),
			expectedSubmodules: 1,
			// Even empty directories get some auto-generated documentation
			expectedDocContains: "Module",
			shouldHaveMessage:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submodules, err := FindSubmodules(tt.fs)
			if err != nil {
				t.Fatalf("FindSubmodules() returned unexpected error: %v", err)
			}

			if len(submodules) != tt.expectedSubmodules {
				t.Errorf("Expected %d submodules, got %d", tt.expectedSubmodules, len(submodules))
				return
			}

			if len(submodules) > 0 {
				doc := submodules[0].Documentation

				if tt.shouldHaveMessage {
					if !strings.Contains(doc, tt.expectedDocContains) {
						t.Errorf("Expected documentation to contain %q, got: %q", tt.expectedDocContains, doc)
					}
				} else {
					if !strings.Contains(doc, tt.expectedDocContains) {
						t.Errorf("Expected documentation to contain %q, got: %q", tt.expectedDocContains, doc)
					}
				}
			}
		})
	}
}

// TestGetModuleDocumentation_ErrorMessages tests that error messages are informative.
func TestGetModuleDocumentation_ErrorMessages(t *testing.T) {
	// Test that missing entrypoint error includes file names
	fs := file.MustNewFS([]file.File{
		file.NewInMemoryFile("other.tf", []byte(`resource "null_resource" "foo" {}`)),
	})

	_, err := GetModuleDocumentation(fs, "")
	if err == nil {
		t.Fatal("Expected error for missing entrypoint, got nil")
	}

	if !errors.Is(err, ErrNoEntrypointFound) {
		t.Errorf("Expected ErrNoEntrypointFound, got: %v", err)
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "main.tf") {
		t.Errorf("Error message should mention main.tf, got: %v", errMsg)
	}
	if !strings.Contains(errMsg, "README.md") {
		t.Errorf("Error message should mention README.md, got: %v", errMsg)
	}
}
