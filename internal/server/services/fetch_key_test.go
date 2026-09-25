package services

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFetchKeys(t *testing.T) {
	Convey("Subject: Coalescing concurrent fetches of one artifact", t, func() {
		Convey("Then a provider package should be fetched once whatever the case of its names", func() {
			So(packageFetchKey("HashiCorp", "Null", "3.2.4", "linux", "amd64"), ShouldEqual, packageFetchKey("hashicorp", "null", "3.2.4", "linux", "amd64"))
			So(packageFetchKey("hashicorp", "null", "3.2.4-RC1", "linux", "amd64"), ShouldNotEqual, packageFetchKey("hashicorp", "null", "3.2.4-rc1", "linux", "amd64"))
		})

		Convey("Then a module version should be fetched once whatever the case of its names", func() {
			So(moduleFetchKey("HashiCorp", "Dir", "Template", "1.0.2"), ShouldEqual, moduleFetchKey("hashicorp", "dir", "template", "1.0.2"))
			So(moduleFetchKey("hashicorp", "dir", "template", "1.0.2-RC1"), ShouldNotEqual, moduleFetchKey("hashicorp", "dir", "template", "1.0.2-rc1"))
		})
	})
}
