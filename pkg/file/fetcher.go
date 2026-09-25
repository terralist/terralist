package file

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"

	getter "github.com/hashicorp/go-getter"
)

type Fetcher interface {
	// Fetch resolves a File into a module archive along with a cleanup
	// function. A RemoteFile is downloaded from its URL (through go-getter,
	// with no local-file scheme); any other File is treated as an already
	// uploaded archive and decompressed locally without reaching the network.
	// If the resolved source is a directory it is archived; if it is an
	// archive it is decompressed and then compressed back to zip. The caller
	// must invoke the cleanup function when the file is no longer needed to
	// remove the temporary directory.
	Fetch(name string, f File) (File, func(), error)

	// FetchFile downloads a file from a given URL and returns it
	// along with a cleanup function.
	FetchFile(name string, url string, header http.Header) (File, func(), error)

	// FetchFileChecksum downloads a file from a given URL while
	// checking a given checksum and returns it along with a cleanup function.
	FetchFileChecksum(name string, url string, checksum string, header http.Header) (File, func(), error)

	// FetchDir downloads all files from a given URL and returns
	// them as an archive along with a cleanup function.
	FetchDir(name string, url string, header http.Header) (File, func(), error)

	// FetchDirChecksum downloads all files from a given URL while
	// checking a given checksum and returns them as an archive
	// along with a cleanup function.
	FetchDirChecksum(name string, url string, checksum string, header http.Header) (File, func(), error)

	// CheckUpstreamSource reports whether a module source announced by an
	// upstream registry may be fetched: only over HTTP(S) or git, and, unless
	// private addresses are allowed, from a git host resolving to public
	// addresses only.
	CheckUpstreamSource(src string) error
}

// forcedGetterRegexp finds the getter forced by a go-getter source, as in
// git::https://example.com/repo.
var forcedGetterRegexp = regexp.MustCompile(`^([A-Za-z0-9]+)::(.+)$`)

const (
	file = iota
	dir
	unknown
)

type defaultFetcher struct {
	allowPrivateAddresses bool
}

// NewFetcher creates a Fetcher. Unless allowPrivateAddresses is true, fetches
// over HTTP(S) refuse to connect to private, loopback, link-local and
// unspecified addresses.
func NewFetcher(allowPrivateAddresses bool) Fetcher {
	return &defaultFetcher{
		allowPrivateAddresses: allowPrivateAddresses,
	}
}

func (f *defaultFetcher) Fetch(name string, src File) (File, func(), error) {
	if remote, ok := src.(*RemoteFile); ok {
		return fetch(name, remote.URL(), "", unknown, remote.Header(), f.allowPrivateAddresses)
	}

	return fetchArchive(name, src)
}

func (f *defaultFetcher) FetchFile(name string, url string, header http.Header) (File, func(), error) {
	return fetch(name, url, "", file, header, f.allowPrivateAddresses)
}

func (f *defaultFetcher) FetchFileChecksum(name string, url string, checksum string, header http.Header) (File, func(), error) {
	return fetch(name, url, checksum, file, header, f.allowPrivateAddresses)
}

func (f *defaultFetcher) FetchDir(name string, url string, header http.Header) (File, func(), error) {
	return fetch(name, url, "", dir, header, f.allowPrivateAddresses)
}

func (f *defaultFetcher) FetchDirChecksum(name string, url string, checksum string, header http.Header) (File, func(), error) {
	return fetch(name, url, checksum, dir, header, f.allowPrivateAddresses)
}

// CreateHeader creates an http.Header from a map of key-value strings.
func CreateHeader(headers map[string]string) http.Header {
	if len(headers) < 1 {
		return nil
	}

	header := http.Header{}
	for key, value := range headers {
		header.Add(key, value)
	}

	return header
}

func (f *defaultFetcher) CheckUpstreamSource(src string) error {
	detected, err := getter.Detect(src, "", getter.Detectors)
	if err != nil {
		return fmt.Errorf("invalid source %q: %w", src, err)
	}

	forced := ""
	if ms := forcedGetterRegexp.FindStringSubmatch(detected); ms != nil {
		forced, detected = ms[1], ms[2]
	}

	detected, _ = getter.SourceDirSubdir(detected)

	u, err := url.Parse(detected)
	if err != nil {
		return fmt.Errorf("invalid source %q: %w", src, err)
	}

	switch {
	case forced == "" && (u.Scheme == "http" || u.Scheme == "https"):
		// The HTTP getter refuses private addresses when dialing.
		return nil
	case forced == "git" && (u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ssh"):
		return f.checkHost(u.Hostname())
	default:
		return fmt.Errorf("refusing to fetch %q: only HTTP(S) and git sources are fetched from an upstream", src)
	}
}

// checkHost refuses a host resolving to a private address, unless private
// addresses are allowed.
func (f *defaultFetcher) checkHost(host string) error {
	if f.allowPrivateAddresses {
		return nil
	}

	addresses, err := net.DefaultResolver.LookupNetIP(context.Background(), "ip", host)
	if err != nil {
		return fmt.Errorf("could not resolve %s: %w", host, err)
	}

	for _, address := range addresses {
		if isPrivateAddress(address.Unmap()) {
			return fmt.Errorf("refusing to fetch from non-public address %s of %s", address, host)
		}
	}

	return nil
}
