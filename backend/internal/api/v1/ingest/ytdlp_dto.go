package ingest

// YtDlpStatusDto reports the installed yt-dlp version against the latest GitHub
// release. It is intentionally small — one concern, the update status.
type YtDlpStatusDto struct {
	Installed       bool   `json:"installed"`
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
	ReleaseURL      string `json:"release_url"`
	ReleaseDate     string `json:"release_date"`
}
