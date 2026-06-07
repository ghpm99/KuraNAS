package distribution

// DownloadItemDto is the API shape of a distributable client app. The display
// name and description are already resolved to the active locale by the service
// (i18n happens at the source), so clients render them verbatim.
type DownloadItemDto struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Platform    string `json:"platform"`
	Version     string `json:"version"`
	MinOS       string `json:"min_os"`
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256"`
	DownloadURL string `json:"download_url"`
}
