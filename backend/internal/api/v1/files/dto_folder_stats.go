package files

type FolderStatsDto struct {
	FileCount      int   `json:"file_count"`
	FolderCount    int   `json:"folder_count"`
	TotalSizeBytes int64 `json:"total_size_bytes"`
}
