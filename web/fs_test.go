package views

import (
	"fmt"
	"html/template"
	"io/fs"
	"testing"

	"github.com/Masterminds/sprig"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCanCompileTemplates(t *testing.T) {
	Convey("Subject: UI templates can compile", t, func() {
		fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, err error) error {
			So(err, ShouldBeNil)

			Convey(fmt.Sprintf("Analyzing the file %s", path), func() {
				_, err := template.New(path).Funcs(sprig.FuncMap()).ParseFS(FS, path)

				Convey("The parser should not return an error", func() {
					So(err, ShouldBeNil)
				})
			})

			return nil
		})
	})
}
