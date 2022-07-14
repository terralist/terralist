package version

import (
	"testing"
)

type versionTestData struct {
	version string
	major   string
	minor   string
	patch   string
	pre     string
	meta    string
}

var (
	validSemVer = []versionTestData{
		{"0.0.4", "0", "0", "4", "", ""},
		{"1.2.3", "1", "2", "3", "", ""},
		{"10.20.30", "10", "20", "30", "", ""},
		{"1.1.2-prerelease+meta", "1", "1", "2", "prerelease", "meta"},
		{"1.1.2+meta", "1", "1", "2", "", "meta"},
		{"1.1.2+meta-valid", "1", "1", "2", "", "meta-valid"},
		{"1.0.0-alpha", "1", "0", "0", "alpha", ""},
		{"1.0.0-beta", "1", "0", "0", "beta", ""},
		{"1.0.0-alpha.beta", "1", "0", "0", "alpha.beta", ""},
		{"1.0.0-alpha.beta.1", "1", "0", "0", "alpha.beta.1", ""},
		{"1.0.0-alpha.1", "1", "0", "0", "alpha.1", ""},
		{"1.0.0-alpha0.valid", "1", "0", "0", "alpha0.valid", ""},
		{"1.0.0-alpha.0valid", "1", "0", "0", "alpha.0valid", ""},
		{"1.0.0-alpha-a.b-c-somethinglong+build.1-aef.1-its-okay", "1", "0", "0", "alpha-a.b-c-somethinglong", "build.1-aef.1-its-okay"},
		{"1.0.0-rc.1+build.1", "1", "0", "0", "rc.1", "build.1"},
		{"2.0.0-rc.1+build.123", "2", "0", "0", "rc.1", "build.123"},
		{"1.2.3-beta", "1", "2", "3", "beta", ""},
		{"10.2.3-DEV-SNAPSHOT", "10", "2", "3", "DEV-SNAPSHOT", ""},
		{"1.2.3-SNAPSHOT-123", "1", "2", "3", "SNAPSHOT-123", ""},
		{"1.0.0", "1", "0", "0", "", ""},
		{"2.0.0", "2", "0", "0", "", ""},
		{"1.1.7", "1", "1", "7", "", ""},
		{"2.0.0+build.1848", "2", "0", "0", "", "build.1848"},
		{"2.0.1-alpha.1227", "2", "0", "1", "alpha.1227", ""},
		{"1.0.0-alpha+beta", "1", "0", "0", "alpha", "beta"},
		{"1.2.3----RC-SNAPSHOT.12.9.1--.12+788", "1", "2", "3", "---RC-SNAPSHOT.12.9.1--.12", "788"},
		{"1.2.3----R-S.12.9.1--.12+meta", "1", "2", "3", "---R-S.12.9.1--.12", "meta"},
		{"1.2.3----RC-SNAPSHOT.12.9.1--.12", "1", "2", "3", "---RC-SNAPSHOT.12.9.1--.12", ""},
		{"1.0.0+0.build.1-rc.10000aaa-kk-0.1", "1", "0", "0", "", "0.build.1-rc.10000aaa-kk-0.1"},
		{"99999999999999999999999.999999999999999999.99999999999999999", "99999999999999999999999", "999999999999999999", "99999999999999999", "", ""},
		{"1.0.0-0A.is.legal", "1", "0", "0", "0A.is.legal", ""},
	}

	invalidSemVer = []versionTestData{
		{"1", "", "", "", "", ""},
		{"1.2", "", "", "", "", ""},
		{"1.2.3-0123", "", "", "", "", ""},
		{"1.2.3-0123.0123", "", "", "", "", ""},
		{"1.1.2+.123", "", "", "", "", ""},
		{"+invalid", "", "", "", "", ""},
		{"-invalid", "", "", "", "", ""},
		{"-invalid+invalid", "", "", "", "", ""},
		{"-invalid.01", "", "", "", "", ""},
		{"alpha", "", "", "", "", ""},
		{"alpha.beta", "", "", "", "", ""},
		{"alpha.beta.1", "", "", "", "", ""},
		{"alpha.1", "", "", "", "", ""},
		{"alpha+beta", "", "", "", "", ""},
		{"alpha_beta", "", "", "", "", ""},
		{"alpha.", "", "", "", "", ""},
		{"alpha..", "", "", "", "", ""},
		{"beta", "", "", "", "", ""},
		{"1.0.0-alpha_beta", "", "", "", "", ""},
		{"-alpha.", "", "", "", "", ""},
		{"1.0.0-alpha..", "", "", "", "", ""},
		{"1.0.0-alpha..1", "", "", "", "", ""},
		{"1.0.0-alpha...1", "", "", "", "", ""},
		{"1.0.0-alpha....1", "", "", "", "", ""},
		{"1.0.0-alpha.....1", "", "", "", "", ""},
		{"1.0.0-alpha......1", "", "", "", "", ""},
		{"1.0.0-alpha.......1", "", "", "", "", ""},
		{"01.1.1", "", "", "", "", ""},
		{"1.01.1", "", "", "", "", ""},
		{"1.1.01", "", "", "", "", ""},
		{"1.2", "", "", "", "", ""},
		{"1.2.3.DEV", "", "", "", "", ""},
		{"1.2-SNAPSHOT", "", "", "", "", ""},
		{"1.2.31.2.3----RC-SNAPSHOT.12.09.1--..12+788", "", "", "", "", ""},
		{"1.2-RC-SNAPSHOT", "", "", "", "", ""},
		{"-1.0.3-gamma+b7718", "", "", "", "", ""},
		{"+justmeta", "", "", "", "", ""},
		{"9.8.7+meta+meta", "", "", "", "", ""},
		{"9.8.7-whatever+meta+meta", "", "", "", "", ""},
		{"99999999999999999999999.999999999999999999.99999999999999999----RC-SNAPSHOT.12.09.1--------------------------------..12", "", "", "", "", ""},
	}
)

func assertPanic(f func(), g func()) {
	defer func() {
		if r := recover(); r == nil {
			g()
		}
	}()

	f()
}

func TestVersion_Valid(t *testing.T) {
	for _, data := range validSemVer {
		ver := Version(data.version)

		if !ver.Valid() {
			t.Errorf("Version %v should be valid.", ver)
		}
	}

	for _, data := range invalidSemVer {
		ver := Version(data.version)

		if ver.Valid() {
			t.Errorf("Version %v should not be valid", ver)
		}
	}
}

func TestVersion_Major(t *testing.T) {
	for _, data := range validSemVer {
		ver := Version(data.version)

		got := ver.Major()
		want := data.major

		if got != want {
			t.Errorf("Version %v major is not correctly returned: got %v, expected %v", ver, got, want)
		}
	}

	for _, data := range invalidSemVer {
		ver := Version(data.version)

		assertPanic(func() {
			ver.Major()
		}, func() {
			t.Errorf("Version %v Major() should panic, but returned successfully.", ver)
		})
	}
}

func TestVersion_Minor(t *testing.T) {
	for _, data := range validSemVer {
		ver := Version(data.version)

		got := ver.Minor()
		want := data.minor

		if got != want {
			t.Errorf("Version %v minor is not correctly returned: got %v, expected %v", ver, got, want)
		}
	}

	for _, data := range invalidSemVer {
		ver := Version(data.version)

		assertPanic(func() {
			ver.Minor()
		}, func() {
			t.Errorf("Version %v Minor() should panic, but returned successfully.", ver)
		})
	}
}

func TestVersion_Patch(t *testing.T) {
	for _, data := range validSemVer {
		ver := Version(data.version)

		got := ver.Patch()
		want := data.patch

		if got != want {
			t.Errorf("Version %v patch is not correctly returned: got %v, expected %v", ver, got, want)
		}
	}

	for _, data := range invalidSemVer {
		ver := Version(data.version)

		assertPanic(func() {
			ver.Patch()
		}, func() {
			t.Errorf("Version %v Patch() should panic, but returned successfully.", ver)
		})
	}
}

func TestVersion_PreRelease(t *testing.T) {
	for _, data := range validSemVer {
		ver := Version(data.version)

		got := ver.PreRelease()
		want := data.pre

		if got != want {
			t.Errorf("Version %v pre-release is not correctly returned: got %v, expected %v", ver, got, want)
		}
	}

	for _, data := range invalidSemVer {
		ver := Version(data.version)

		got := ver.PreRelease()
		want := data.pre

		if got != want {
			t.Errorf("Version %v pre-release is not correctly returned: got %v, expected %v", ver, got, want)
		}
	}
}

func TestVersion_BuildMetadata(t *testing.T) {
	for _, data := range validSemVer {
		ver := Version(data.version)

		got := ver.BuildMetadata()
		want := data.meta

		if got != want {
			t.Errorf("Version %v build metadata is not correctly returned: got %v, expected %v", ver, got, want)
		}
	}

	for _, data := range invalidSemVer {
		ver := Version(data.version)

		got := ver.BuildMetadata()
		want := data.meta

		if got != want {
			t.Errorf("Version %v build metadata is not correctly returned: got %v, expected %v", ver, got, want)
		}
	}
}
