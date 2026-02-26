package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobMatchRBACCases(t *testing.T) {
	tests := []struct {
		name     string
		val      string
		pattern  string
		expected bool
	}{
		// Resource matching
		{"resource modules exact", "modules", "modules", true},
		{"resource providers exact", "providers", "providers", true},
		{"resource authorities wildcard", "authorities", "*", true},
		{"resource modules wildcard", "modules", "*", true},
		{"resource not matching", "modules", "providers", false},

		// Action matching
		{"action get exact", "get", "get", true},
		{"action update wildcard", "update", "*", true},
		{"action create no match", "create", "get", false},
		{"action delete wildcard", "delete", "del*", true},

		// Object matching for modules
		{"module slug exact", "myspace/my-module/aws", "myspace/my-module/aws", true},
		{"module slug wildcard namespace", "myspace/my-module/aws", "*/my-module/aws", true},
		{"module slug wildcard module", "myspace/my-module/aws", "myspace/*/aws", true},
		{"module slug wildcard provider", "myspace/my-module/aws", "myspace/my-module/*", true},
		{"module slug full wildcard", "myspace/my-module/aws", "*", true},
		{"module slug no match", "myspace/my-module/aws", "otherspace/*", false},

		// Object matching for providers
		{"provider slug exact", "myspace/aws", "myspace/aws", true},
		{"provider slug wildcard namespace", "myspace/aws", "*/aws", true},
		{"provider slug wildcard provider", "myspace/aws", "myspace/*", true},
		{"provider slug no match", "myspace/aws", "otherspace/*", false},

		// Invalid pattern
		{"invalid pattern", "foo.txt", "[*.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := globMatch(tt.val, tt.pattern)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
