package file

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"time"
)

// RemoteFile is a File that references content located at a URL. It is a marker
// that a Fetcher resolves by downloading; its contents cannot be read directly,
// so any attempt to read or seek it before fetching returns an error.
type RemoteFile struct {
	url    string
	header http.Header
}

// NewRemoteFile creates a RemoteFile for the given URL and optional request
// headers.
func NewRemoteFile(url string, header http.Header) *RemoteFile {
	return &RemoteFile{
		url:    url,
		header: header,
	}
}

// URL returns the source URL of the remote file.
func (f *RemoteFile) URL() string {
	return f.url
}

// Header returns the request headers to use when fetching the remote file.
func (f *RemoteFile) Header() http.Header {
	return f.header
}

func (f *RemoteFile) Name() string {
	return path.Base(f.url)
}

func (f *RemoteFile) Read([]byte) (int, error) {
	return 0, fmt.Errorf("%w: a remote file must be fetched before it can be read", ErrSystemFailure)
}

func (f *RemoteFile) Seek(int64, int) (int64, error) {
	return 0, fmt.Errorf("%w: a remote file must be fetched before it can be seeked", ErrSystemFailure)
}

func (f *RemoteFile) Close() error {
	return nil
}

func (f *RemoteFile) Stat() (fs.FileInfo, error) {
	return f.Metadata(), nil
}

func (f *RemoteFile) Metadata() fs.FileInfo {
	return &BufferFileInfo{
		name:    f.Name(),
		size:    0,
		mode:    0644,
		modTime: time.Now(),
	}
}
