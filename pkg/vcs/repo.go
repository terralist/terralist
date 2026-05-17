package vcs

import (
	"fmt"
	"net/url"
	"strings"
)

type normalizedRepo struct {
	host string
	path string
}

func normalizeRepoString(s string) (normalizedRepo, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return normalizedRepo{}, fmt.Errorf("empty repository URL")
	}
	if strings.HasPrefix(s, "git@") {
		rest := strings.TrimPrefix(s, "git@")
		i := strings.Index(rest, ":")
		if i < 0 {
			return normalizedRepo{}, fmt.Errorf("invalid git ssh URL")
		}
		host := strings.ToLower(strings.TrimSpace(rest[:i]))
		path := strings.TrimSpace(rest[i+1:])
		path = strings.TrimSuffix(path, ".git")
		path = strings.Trim(path, "/")
		return normalizedRepo{host: host, path: strings.ToLower(path)}, nil
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return normalizedRepo{}, fmt.Errorf("invalid repository URL")
	}
	path := strings.TrimSuffix(strings.Trim(u.Path, "/"), ".git")
	return normalizedRepo{
		host: strings.ToLower(u.Host),
		path: strings.ToLower(path),
	}, nil
}

// CanonicalVCSRepoURL returns a stable lowercase https-style identity string for logging and storage.
func CanonicalVCSRepoURL(s string) (string, error) {
	n, err := normalizeRepoString(s)
	if err != nil {
		return "", err
	}
	return "https://" + n.host + "/" + n.path, nil
}

// RepoURLsMatch reports whether two repository references denote the same project.
func RepoURLsMatch(a, b string) bool {
	if strings.TrimSpace(a) == "" || strings.TrimSpace(b) == "" {
		return false
	}
	na, err := normalizeRepoString(a)
	if err != nil {
		return false
	}
	nb, err := normalizeRepoString(b)
	if err != nil {
		return false
	}
	return na.host == nb.host && na.path == nb.path
}
