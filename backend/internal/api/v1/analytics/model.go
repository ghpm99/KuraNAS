package analytics

import (
	"database/sql"
	"time"
)

type PeriodConfig struct {
	Label    string
	Interval string
}

type OverviewDataModel struct {
	StorageKpis      StorageKpisModel
	TimeSeries       []StorageTimeSeriesModel
	Types            []TypeDistributionModel
	Extensions       []ExtensionDistributionModel
	HotFolders       []FolderHotModel
	TopFolders       []FolderUsageModel
	RecentFiles      []RecentFileModel
	Duplicates       DuplicatesSummaryModel
	TopDuplicateSets []DuplicateGroupModel
	HealthStatus     sql.NullString
	LastScanStart    sql.NullTime
	LastScanEnd      sql.NullTime
	ErrorsLast24h    int64
	RecentErrors     []LogErrorModel
}

type StorageKpisModel struct {
	UsedBytes    int64
	GrowthBytes  int64
	FilesAdded   int64
	FilesTotal   int64
	FoldersTotal int64
}

type StorageTimeSeriesModel struct {
	Date      time.Time
	UsedBytes int64
}

type TypeDistributionModel struct {
	Category string
	Count    int64
	Bytes    int64
}

type ExtensionDistributionModel struct {
	Extension string
	Count     int64
	Bytes     int64
}

type FolderHotModel struct {
	ParentPath string
	NewFiles   int64
	AddedBytes int64
	LastEvent  sql.NullTime
}

type FolderUsageModel struct {
	ParentPath   string
	TotalFiles   int64
	TotalBytes   int64
	LastModified sql.NullTime
}

type RecentFileModel struct {
	ID         int
	Name       string
	Path       string
	ParentPath string
	Size       int64
	Format     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type DuplicatesSummaryModel struct {
	GroupsTotal      int64
	FilesTotal       int64
	ReclaimableBytes int64
}

type DuplicateGroupModel struct {
	Signature       string
	Copies          int64
	ItemSize        int64
	ReclaimableSize int64
	Paths           []string
}

type LogErrorModel struct {
	Name        string
	Description sql.NullString
	CreatedAt   time.Time
}
