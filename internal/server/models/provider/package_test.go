package provider

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParsePackageFileName(t *testing.T) {
	Convey("Subject: Parsing a provider package file name", t, func() {
		Convey("A registry package name yields its coordinates", func() {
			pkg, ok := ParsePackageFileName("terraform-provider-null_3.2.4_linux_amd64.zip")
			So(ok, ShouldBeTrue)
			So(pkg, ShouldResemble, Package{Name: "null", Version: "3.2.4", System: "linux", Architecture: "amd64"})
		})

		Convey("A prerelease version is kept whole", func() {
			pkg, ok := ParsePackageFileName("terraform-provider-aws_5.0.0-beta1_darwin_arm64.zip")
			So(ok, ShouldBeTrue)
			So(pkg.Version, ShouldEqual, "5.0.0-beta1")
		})

		Convey("A name with dashes is kept whole", func() {
			pkg, ok := ParsePackageFileName("terraform-provider-my-thing_1.0.0_windows_386.zip")
			So(ok, ShouldBeTrue)
			So(pkg.Name, ShouldEqual, "my-thing")
		})

		Convey("Other names are rejected", func() {
			for _, name := range []string{"3.2.4.json", "terraform-provider-null_3.2.4.zip", "null_3.2.4_linux_amd64.zip", "terraform-provider-null_3.2.4_linux_amd64.tar"} {
				_, ok := ParsePackageFileName(name)
				So(ok, ShouldBeFalse)
			}
		})

		Convey("The file name round trips", func() {
			So(PackageFileName("null", "3.2.4", "linux", "amd64"), ShouldEqual, "terraform-provider-null_3.2.4_linux_amd64.zip")
		})
	})
}

func TestPackageSubject(t *testing.T) {
	Convey("Subject: Naming a package for download tokens", t, func() {
		pkg := Package{Name: "AWS", Version: "5.0.0-RC1", System: "linux", Architecture: "amd64"}

		Convey("Then the authority and provider names should not depend on case", func() {
			So(pkg.Subject("HashiCorp"), ShouldEqual, Package{Name: "aws", Version: "5.0.0-RC1", System: "linux", Architecture: "amd64"}.Subject("hashicorp"))
		})

		Convey("Then the version should stay exact", func() {
			So(pkg.Subject("hashicorp"), ShouldNotEqual, Package{Name: "aws", Version: "5.0.0-rc1", System: "linux", Architecture: "amd64"}.Subject("hashicorp"))
		})
	})
}
