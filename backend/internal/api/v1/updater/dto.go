package updater

type UpdateStatusDto struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
	ReleaseURL      string `json:"release_url"`
	ReleaseDate     string `json:"release_date"`
	ReleaseNotes    string `json:"release_notes"`
	AssetName       string `json:"asset_name"`
	AssetSize       int64  `json:"asset_size"`
}

type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	HTMLURL     string        `json:"html_url"`
	PublishedAt string        `json:"published_at"`
	Body        string        `json:"body"`
	Assets      []GitHubAsset `json:"assets"`
}

type GitHubAsset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}
