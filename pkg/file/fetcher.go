package file

import "net/http"

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
}

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
