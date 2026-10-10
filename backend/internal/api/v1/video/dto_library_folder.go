package video

type LibraryFolderDto struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	VideoCount  int    `json:"video_count"`
	CoverFileID int    `json:"cover_file_id"`
}
