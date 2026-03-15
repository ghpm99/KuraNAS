package analytics

import "time"

type OverviewDto struct {
	Period      string               `json:"period"`
	GeneratedAt time.Time            `json:"generated_at"`
	Storage     StorageDto           `json:"storage"`
	Counts      CountsDto            `json:"counts"`
	TimeSeries  []TimeSeriesPointDto `json:"time_series"`
	Types       []TypeBreakdownDto   `json:"types"`
	Extensions  []ExtensionDto       `json:"extensions"`
	HotFolders  []HotFolderDto       `json:"hot_folders"`
	TopFolders  []FolderUsageDto     `json:"top_folders"`
	RecentFiles []RecentFileDto      `json:"recent_files"`
	Duplicates  DuplicatesDto        `json:"duplicates"`
	Library     LibraryDto           `json:"library"`
	Processing  ProcessingDto        `json:"processing"`
	Health      HealthDto            `json:"health"`
}

type StorageDto struct {
	TotalBytes  int64 `json:"total_bytes"`
	UsedBytes   int64 `json:"used_bytes"`
	FreeBytes   int64 `json:"free_bytes"`
	GrowthBytes int64 `json:"growth_bytes"`
}

type CountsDto struct {
	FilesTotal int64 `json:"files_total"`
	FilesAdded int64 `json:"files_added"`
	Folders    int64 `json:"folders"`
}

type TimeSeriesPointDto struct {
	Date      string `json:"date"`
	UsedBytes int64  `json:"used_bytes"`
}

type TypeBreakdownDto struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
	Bytes int64  `json:"bytes"`
}

type ExtensionDto struct {
	Ext   string `json:"ext"`
	Count int64  `json:"count"`
	Bytes int64  `json:"bytes"`
}

type HotFolderDto struct {
	Path       string `json:"path"`
	NewFiles   int64  `json:"new_files"`
	AddedBytes int64  `json:"added_bytes"`
	LastEvent  string `json:"last_event"`
}

type FolderUsageDto struct {
	Path         string `json:"path"`
	Files        int64  `json:"files"`
	Bytes        int64  `json:"bytes"`
	LastModified string `json:"last_modified"`
}

type RecentFileDto struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	ParentPath string `json:"parent_path"`
	Format     string `json:"format"`
	SizeBytes  int64  `json:"size_bytes"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type DuplicatesDto struct {
	Groups          int64               `json:"groups"`
	Files           int64               `json:"files"`
	ReclaimableSize int64               `json:"reclaimable_size"`
	TopGroups       []DuplicateGroupDto `json:"top_groups"`
}

type DuplicateGroupDto struct {
	Signature       string   `json:"signature"`
	Copies          int64    `json:"copies"`
	SizeBytes       int64    `json:"size_bytes"`
	ReclaimableSize int64    `json:"reclaimable_size"`
	Paths           []string `json:"paths"`
}

type LibraryDto struct {
	CategorizedMedia  int64 `json:"categorized_media"`
	AudioWithMetadata int64 `json:"audio_with_metadata"`
	VideoWithMetadata int64 `json:"video_with_metadata"`
	ImageWithMetadata int64 `json:"image_with_metadata"`
	ImageClassified   int64 `json:"image_classified"`
}

type ProcessingDto struct {
	MetadataPending  int64 `json:"metadata_pending"`
	MetadataFailed   int64 `json:"metadata_failed"`
	ThumbnailPending int64 `json:"thumbnail_pending"`
	ThumbnailFailed  int64 `json:"thumbnail_failed"`
}

type HealthDto struct {
	Status          string   `json:"status"`
	LastScanAt      string   `json:"last_scan_at"`
	LastScanSeconds int64    `json:"last_scan_seconds"`
	IndexedFiles    int64    `json:"indexed_files"`
	ErrorsLast24h   int64    `json:"errors_last_24h"`
	RecentErrors    []string `json:"recent_errors"`
}
