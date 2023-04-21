package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var FS embed.FS

var (
	distFS, _ = fs.Sub(FS, "dist")
)

type fileSystem struct {
	http.FileSystem
}

func StaticFS() *fileSystem {
	return &fileSystem{
		FileSystem: http.FS(distFS),
	}
}

func (*fileSystem) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		if p == "" {
			// On nothing, serve index.html
			return true
		}

		if _, err := fs.Stat(distFS, p); err != nil {
			return false
		}

		return true
	}

	return false
}
