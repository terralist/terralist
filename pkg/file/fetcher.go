package file

type Fetcher interface {
	// Fetch downloads a file or a directory from a given URL
	// and returns it
	// If the file is a directory, it will be archived
	// If the file is an archive, it will be decompressed, and
	// then compressed back to zip
	Fetch(name string, url string) (File, error)

	// FetchFile downloads a file from a given URL and returns it
	FetchFile(name string, url string) (File, error)

	// FetchFileChecksum downloads a file from a given URL while
	// checking a given checksum and returns it
	FetchFileChecksum(name string, url string, checksum string) (File, error)

	// FetchDir downloads all files from a given URL and returns
	// them as an archive
	FetchDir(name string, url string) (File, error)

	// FetchDirChecksum downloads all files from a given URL while
	// checking a given checksum and returns them as an archive
	FetchDirChecksum(name string, url string, checksum string) (File, error)
}

const (
	file = iota
	dir
	any
)

type defaultFetcher struct{}

func NewFetcher() Fetcher {
	return &defaultFetcher{}
}

func (*defaultFetcher) Fetch(name string, url string) (File, error) {
	return fetch(name, url, "", any)
}

func (*defaultFetcher) FetchFile(name string, url string) (File, error) {
	return fetch(name, url, "", file)
}

func (*defaultFetcher) FetchFileChecksum(name string, url string, checksum string) (File, error) {
	return fetch(name, url, checksum, file)
}

func (*defaultFetcher) FetchDir(name string, url string) (File, error) {
	return fetch(name, url, "", dir)
}

func (*defaultFetcher) FetchDirChecksum(name string, url string, checksum string) (File, error) {
	return fetch(name, url, checksum, dir)
}
