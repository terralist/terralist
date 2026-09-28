package module

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestArchiveSubject(t *testing.T) {
	Convey("Subject: Naming a module version for download tokens", t, func() {
		Convey("Then the authority, module and provider names should not depend on case", func() {
			So(ArchiveSubject("HashiCorp", "Dir", "Template", "1.0.0"), ShouldEqual, ArchiveSubject("hashicorp", "dir", "template", "1.0.0"))
		})

		Convey("Then the version should stay exact", func() {
			So(ArchiveSubject("hashicorp", "dir", "template", "1.0.0-RC1"), ShouldNotEqual, ArchiveSubject("hashicorp", "dir", "template", "1.0.0-rc1"))
		})
	})
}
