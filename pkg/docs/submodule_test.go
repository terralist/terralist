package docs

import (
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
