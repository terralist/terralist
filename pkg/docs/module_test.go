package docs

import (
	"errors"
	"strings"
	"testing"

	"terralist/pkg/file"
)

func TestGetModuleDocumentation(t *testing.T) {
	testData := []struct {
		title        string
		fs           *file.FS
		relativePath string
		expected     string
		expectedFn   func(string) bool
		shouldError  bool
	}{
		{
			title: "README.md UTF-8 with emoji",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("README.md", []byte("# hello 😀\n")),
			}),
			expected: "# hello 😀\n",
		},
		{
			title: "README.md with UTF-8 BOM is normalized",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("README.md", append([]byte{0xEF, 0xBB, 0xBF}, []byte("# bom utf8\n")...)),
			}),
			expected: "# bom utf8\n",
		},
		{
			title: "README.md with UTF-16LE BOM is decoded (with emoji)",
			fs: func() *file.FS {
				// Encode a string as UTF-16LE with BOM
				s := "# utf16le 😀\n"
				// Build bytes: BOM 0xFF,0xFE then little-endian 16-bit code units
				// Use Go's rune encoding into UTF-16LE manually to avoid pulling encoder in tests
				// Simple encoder for test purposes
				var data []byte
				data = append(data, 0xFF, 0xFE)
				for _, r := range s {
					if r < 0x10000 {
						data = append(data, byte(r), byte(r>>8))
					} else {
						// encode surrogate pair
						rPrime := r - 0x10000
						hi := 0xD800 + ((rPrime >> 10) & 0x3FF)
						lo := 0xDC00 + (rPrime & 0x3FF)
						data = append(data, byte(hi), byte(hi>>8))
						data = append(data, byte(lo), byte(lo>>8))
					}
				}
				return file.MustNewFS([]file.File{file.NewInMemoryFile("README.md", data)})
			}(),
			expected: "# utf16le 😀\n",
		},
		{
			title: "README.md with UTF-16BE BOM is decoded",
			fs: func() *file.FS {
				s := "# utf16be\n"
				var data []byte
				data = append(data, 0xFE, 0xFF)
				for _, r := range s {
					if r < 0x10000 {
						data = append(data, byte(r>>8), byte(r))
					} else {
						rPrime := r - 0x10000
						hi := 0xD800 + ((rPrime >> 10) & 0x3FF)
						lo := 0xDC00 + (rPrime & 0x3FF)
						data = append(data, byte(hi>>8), byte(hi))
						data = append(data, byte(lo>>8), byte(lo))
					}
				}
				return file.MustNewFS([]file.File{file.NewInMemoryFile("README.md", data)})
			}(),
			expected: "# utf16be\n",
		},
		{
			title: "README.md with malformed UTF-16LE BOM errors",
			fs: file.MustNewFS([]file.File{
				// BOM then odd trailing byte to force decode error
				file.NewInMemoryFile("README.md", []byte{0xFF, 0xFE, 0x61}),
			}),
			shouldError: true,
		},
		{
			title: "Module with only main.tf",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`
					variable "test" {}
					output "test" {}
					`)),
			}),
			expected: "\n# Module `.`\n\n## Input Variables\n* `test` (required)\n\n## Output Values\n* `test`\n\n",
		},
		{
			title: "Module with both main.tf and README.md",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("main.tf", []byte(`
					variable "test" {}
					output "test" {}
					`)),
				file.NewInMemoryFile("README.md", []byte(`# my module\n`)),
			}),
			expected: `# my module\n`,
		},
		{
			title: "Module with subdirectory with only main.tf",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("subdir/main.tf", []byte(`
					variable "test" {}
					output "test" {}
					`)),
			}),
			expected: "\n# Module `subdir`\n\n## Input Variables\n* `test` (required)\n\n## Output Values\n* `test`\n\n",
		},
		{
			title: "Module with subdirectory with main.tf and README.md in root",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("subdir/main.tf", []byte(`
					variable "test" {}
					output "test" {}
					`)),
				file.NewInMemoryFile("README.md", []byte(`# my module\n`)),
			}),
			expected: `# my module\n`,
		},
		{
			title: "Module with two subdirectories with both main.tf and README.md",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("subdir1/main.tf", []byte(`
				variable "test1" {}
				output "test1" {}
				`)),
				file.NewInMemoryFile("subdir1/README.md", []byte(`# my module1\n`)),
				file.NewInMemoryFile("subdir2/main.tf", []byte(`
				variable "test2" {}
				output "test2" {}
				`)),
				file.NewInMemoryFile("subdir2/README.md", []byte(`# my module2\n`)),
			}),
			// This behavior is undefined, either of the modules can be found first
			// depending on how the OS parses the directory. We need to validate both.
			expectedFn: func(s string) bool {
				return s == `# my module1\n` || s == `# my module2\n`
			},
		},
		{
			title: "Module with two subdirectories with both main.tf and README.md when root dir is known",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("subdir1/main.tf", []byte(`
				variable "test1" {}
				output "test1" {}
				`)),
				file.NewInMemoryFile("subdir1/README.md", []byte(`# my module1\n`)),
				file.NewInMemoryFile("subdir2/main.tf", []byte(`
				variable "test2" {}
				output "test2" {}
				`)),
				file.NewInMemoryFile("subdir2/README.md", []byte(`# my module2\n`)),
			}),
			relativePath: "subdir2",
			expected:     `# my module2\n`,
		},
		{
			title: "Module with multiple .tf files and a README.md but no main.tf because it wasn't authored that way",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("rds.tf", []byte(`
				resource "postgresql_role" "service_role" {}
				`)),
				file.NewInMemoryFile("variables.tf", []byte(`variable "test" {}`)),
				file.NewInMemoryFile("outputs.tf", []byte(`output "test" {}`)),
				file.NewInMemoryFile("README.md", []byte(`# My Custom Module Readme`)),
			}),
			expected: `# My Custom Module Readme`,
		},
		{
			title: "Module with no README.md and no main.tf",
			fs: file.MustNewFS([]file.File{
				file.NewInMemoryFile("other.tf", []byte(`resource "null_resource" "foo" {}`)),
			}),
			shouldError: true,
		},
	}

	for i, test := range testData {
		result, err := GetModuleDocumentation(test.fs, test.relativePath)

		if err != nil && !test.shouldError {
			t.Fatalf("#%d (%v): expected result, but got error: %v", i, test.title, err)
		}

		if err == nil && test.shouldError {
			t.Fatalf("#%d (%v): expected error, but got result: %v", i, test.title, result)
		}

		// Verify that the error for missing entrypoint provides helpful information
		if test.shouldError && err != nil {
			if errors.Is(err, ErrNoEntrypointFound) {
				if !strings.Contains(err.Error(), "main.tf") || !strings.Contains(err.Error(), "README.md") {
					t.Errorf("#%d (%v): error should mention both main.tf and README.md, got: %v", i, test.title, err)
				}
			}
		}

		if test.expectedFn == nil && result != test.expected {
			t.Fatalf("#%d (%v): expected `%v`, but got `%v`", i, test.title, test.expected, result)
		}

		if test.expectedFn != nil && test.expected == "" && !test.expectedFn(result) {
			t.Fatalf("#%d (%v): result `%v` could not pass the expected func", i, test.title, result)
		}
	}
}
