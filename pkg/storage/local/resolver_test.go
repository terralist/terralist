package local

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"terralist/pkg/storage"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStore(t *testing.T) {
	Convey("Subject: Store files on the disk", t, func() {
		root := t.TempDir()
		resolver := &Resolver{RegistryDir: filepath.Join(root, "registry")}

		Convey("When a file is stored", func() {
			key, err := resolver.Store(&storage.StoreInput{
				KeyPrefix: "providers/hashicorp/null/1.0.0",
				FileName:  "SHA256SUMS",
				Reader:    bytes.NewReader([]byte("content")),
			})

			Convey("Then it should be written under the registry directory", func() {
				So(err, ShouldBeNil)
				So(key, ShouldEqual, "providers/hashicorp/null/1.0.0/SHA256SUMS")

				content, err := os.ReadFile(filepath.Join(root, "registry", key))
				So(err, ShouldBeNil)
				So(string(content), ShouldEqual, "content")
			})
		})

		Convey("When the file name escapes the registry directory", func() {
			_, err := resolver.Store(&storage.StoreInput{
				KeyPrefix: "providers/hashicorp/null/1.0.0",
				FileName:  "../../../../../escaped",
				Reader:    bytes.NewReader([]byte("content")),
			})

			Convey("Then it should be rejected and nothing written", func() {
				So(err, ShouldNotBeNil)

				_, statErr := os.Stat(filepath.Join(root, "escaped"))
				So(os.IsNotExist(statErr), ShouldBeTrue)
			})
		})
	})
}

func TestGetObject(t *testing.T) {
	Convey("Subject: Read files from the disk", t, func() {
		root := t.TempDir()
		resolver := &Resolver{RegistryDir: filepath.Join(root, "registry")}
		So(os.WriteFile(filepath.Join(root, "outside"), []byte("secret"), 0600), ShouldBeNil)

		Convey("When the key escapes the registry directory", func() {
			_, err := resolver.GetObject("../outside")

			Convey("Then it should be rejected", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
