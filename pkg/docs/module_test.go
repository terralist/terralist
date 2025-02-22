package docs

import (
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
	}

	for i, test := range testData {
		result, err := GetModuleDocumentation(test.fs, test.relativePath)

		if err != nil && !test.shouldError {
			t.Fatalf("#%d (%v): expected result, but got error: %v", i, test.title, err)
		}

		if err == nil && test.shouldError {
			t.Fatalf("#%d (%v): expected error, but got result: %v", i, test.title, result)
		}

		if test.expectedFn == nil && result != test.expected {
			t.Fatalf("#%d (%v): expected `%v`, but got `%v`", i, test.title, test.expected, result)
		}

		if test.expectedFn != nil && test.expected == "" && !test.expectedFn(result) {
			t.Fatalf("#%d (%v): result `%v` could not pass the expected func", i, test.title, result)
		}
	}
}
