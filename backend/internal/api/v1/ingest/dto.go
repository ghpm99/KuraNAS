package ingest

// FetchRequestDto is the body of POST /downloads/fetch: a media URL to pull
// server-side with yt-dlp, the chosen quality preset, and where it should land.
type FetchRequestDto struct {
	URL        string `json:"url"`
	Preset     string `json:"preset"`
	TargetRoot string `json:"target_root"`
	Subfolder  string `json:"subfolder"`
}

// FetchResponseDto is returned by POST /downloads/fetch. The download runs in a
// background job; the caller tracks progress through the jobs API.
type FetchResponseDto struct {
	JobID int `json:"job_id"`
}

// TargetDto is one enabled storage root the download can be saved into.
type TargetDto struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

// PresetDto is one selectable download quality/format option.
type PresetDto struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}
