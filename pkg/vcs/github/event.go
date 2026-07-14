package github

type ReleaseAsset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Release struct {
	TagName    string         `json:"tag_name"`
	Draft      bool           `json:"draft"`
	Prerelease bool           `json:"prerelease"`
	ZipballURL string         `json:"zipball_url"`
	TarballURL string         `json:"tarball_url"`
	Assets     []ReleaseAsset `json:"assets"`
}

type Repository struct {
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}

type ReleasePayload struct {
	Action     string     `json:"action"`
	Release    Release    `json:"release"`
	Repository Repository `json:"repository"`
}
