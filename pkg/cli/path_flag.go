package cli

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// PathFlag holds data for the flags with path values.
type PathFlag struct {
	Description  string
	DefaultValue string
	Hidden       bool
	Required     bool

	Value string

	isSet bool
}

func (t *PathFlag) IsHidden() bool {
	return t.Hidden
}

func (t *PathFlag) IsSet() bool {
	return t.isSet
}

func (t *PathFlag) Set(value any) error {
	if value == nil {
		t.Value = t.DefaultValue
		t.isSet = false
	} else {
		v, ok := value.(string)
		if !ok {
			s, ok := value.(string)
			if !ok {
				return fmt.Errorf("unsupported type %T for path flag", value)
			}

			if env, ok := environmentLookup(s); ok {
				s = env
			}

			v = s
		}

		if v == "" {
			if v != t.DefaultValue {
				t.Value = v
			} else {
				t.Value = t.DefaultValue
			}
		} else {
			t.Value = v
			t.isSet = true
		}
	}

	t.Value = sanitizePath(t.Value)

	return nil
}

func (t *PathFlag) Format() string {
	return fmt.Sprintf("%s (default %v)", t.Description, t.DefaultValue)
}

func (t *PathFlag) Validate() error {
	if t.Required && t.isSet {
		return fmt.Errorf("required but not set")
	}

	if !t.Required && t.Value == "" {
		return nil
	}

	if !filepath.IsAbs(t.Value) {
		return fmt.Errorf("not absolute and cannot be resolved to an absolute path")
	}

	return nil
}

func sanitizePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	usr, _ := user.Current()
	dir := usr.HomeDir
	cwd, _ := os.Getwd()

	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(dir, path[2:])
	} else if path == "." {
		path = cwd
	} else if strings.HasPrefix(path, "./") {
		path = filepath.Join(cwd, path[2:])
	}

	return path
}
