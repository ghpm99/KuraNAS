package ingest

// RemoteFetchStepPayload is the worker step payload for a yt-dlp download. The
// preset and the resolved absolute output directory are computed at enqueue
// time so the worker never re-reads configuration or the roots registry.
type RemoteFetchStepPayload struct {
	URL       string `json:"url"`
	Preset    string `json:"preset"`
	OutputDir string `json:"output_dir"`
	Binary    string `json:"binary"`
}
