package queries

import _ "embed"

//go:embed storage_kpis.sql
var StorageKPIsQuery string

//go:embed storage_timeseries.sql
var StorageTimeSeriesQuery string

//go:embed type_distribution.sql
var TypeDistributionQuery string

//go:embed extension_distribution.sql
var ExtensionDistributionQuery string

//go:embed recent_files.sql
var RecentFilesQuery string

//go:embed folder_size_rank.sql
var FolderSizeRankQuery string

//go:embed folder_hot_rank.sql
var FolderHotRankQuery string

//go:embed duplicates_summary.sql
var DuplicatesSummaryQuery string

//go:embed library_metadata_summary.sql
var LibraryMetadataSummaryQuery string

//go:embed processing_queue_summary.sql
var ProcessingQueueSummaryQuery string

//go:embed duplicates_top_groups.sql
var DuplicatesTopGroupsQuery string

//go:embed index_health_status.sql
var IndexHealthStatusQuery string

//go:embed index_errors_recent.sql
var IndexErrorsRecentQuery string

//go:embed index_errors_latest.sql
var IndexErrorsLatestQuery string
