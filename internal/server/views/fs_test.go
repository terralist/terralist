package views

import (
	"html/template"
	"io/fs"
	"strings"
	"testing"

	"github.com/Masterminds/sprig"
)

func Test_CanCompileTemplates(t *testing.T) {
	err := fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".tpl") {
			_, err := template.New(path).Funcs(sprig.FuncMap()).ParseFS(FS, path)
			if err != nil {
				t.Errorf("cannot parse template %s: %v", path, err)
			}
		}

		return nil
	})

	if err != nil {
		t.Errorf("%v", err)
	}
}
