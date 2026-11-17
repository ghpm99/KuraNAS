package music

type RenameGenreRequest struct {
	Name string `json:"name"`
}

type GenreRenameResultDto struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	TrackCount int    `json:"track_count"`
	Merged     bool   `json:"merged"`
}

type GenreTrackSource struct {
	FileID int
	Genre  string
}
