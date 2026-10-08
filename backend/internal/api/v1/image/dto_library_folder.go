package image

type LibraryFolderDto struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	ImageCount  int    `json:"image_count"`
	CoverFileID int    `json:"cover_file_id"`
}
