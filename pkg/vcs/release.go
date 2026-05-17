package vcs

import (
	"bufio"
	"strings"
	"terralist/pkg/file"
)

type ReleaseSource string

const (
	ReleaseSourceGitHub ReleaseSource = "github"
)

type ReleaseAsset struct {
	Name string
	URL  string
}

type ReleaseEvent struct {
	Source           ReleaseSource
	Tag              string
	SemVer           string
	Draft            bool
	Prerelease       bool
	ModuleArchiveURL string
	RepoURL          string
	Assets           []ReleaseAsset
}

func ParseSHA256SUMS(data file.File) map[string]string {
	out := make(map[string]string)
	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		hash := parts[0]
		name := parts[len(parts)-1]
		name = strings.TrimPrefix(name, "*")
		if len(hash) == 64 {
			out[name] = hash
		}
	}
	return out
}
