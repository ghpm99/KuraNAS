package analytics

import (
	"nas-go/api/internal/api/v1/trash"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetStorageKpis(period PeriodConfig) (StorageKpisModel, error)
	GetStorageTimeSeries(period PeriodConfig) ([]StorageTimeSeriesModel, error)
	GetTypeDistribution() ([]TypeDistributionModel, error)
	GetExtensionDistribution(limit int) ([]ExtensionDistributionModel, error)
	GetRecentFiles(limit int) ([]RecentFileModel, error)
	GetTopFolders(limit int) ([]FolderUsageModel, error)
	GetHotFolders(period PeriodConfig, limit int) ([]FolderHotModel, error)
	GetDuplicatesSummary() (DuplicatesSummaryModel, error)
	GetDuplicateGroups(limit int) ([]DuplicateGroupModel, error)
	GetImageDuplicatesSummary() (DuplicatesSummaryModel, error)
	GetImageDuplicateGroups(limit int) ([]DuplicateGroupModel, error)
	GetDuplicateGroupsPage(category DuplicateCategory, page int, pageSize int) (utils.PaginationResponse[DuplicateGroupDetailModel], error)
	GetLibrarySummary() (LibrarySummaryModel, error)
	GetProcessingSummary() (ProcessingSummaryModel, error)
	GetHealth() (HealthModel, error)
}

type TrashSummaryReader interface {
	GetSummary() (trash.TrashSummaryDto, error)
}

type ServiceInterface interface {
	GetStorage(period string) (StorageStatsDto, error)
	GetTimeSeries(period string) ([]TimeSeriesPointDto, error)
	GetTypes() ([]TypeBreakdownDto, error)
	GetExtensions(limit int) ([]ExtensionDto, error)
	GetRecentFiles(limit int) ([]RecentFileDto, error)
	GetTopFolders(limit int) ([]FolderUsageDto, error)
	GetHotFolders(period string, limit int) ([]HotFolderDto, error)
	GetDuplicatesSummary() (DuplicatesSummaryDto, error)
	GetDuplicateGroups(limit int) ([]DuplicateGroupDto, error)
	GetImageDuplicatesSummary() (DuplicatesSummaryDto, error)
	GetImageDuplicateGroups(limit int) ([]DuplicateGroupDto, error)
	GetDuplicateGroupsPage(category DuplicateCategory, page int, pageSize int) (utils.PaginationResponse[DuplicateGroupDetailDto], error)
	GetLibrary() (LibraryDto, error)
	GetProcessing() (ProcessingDto, error)
	GetHealth() (HealthDto, error)
	GetAIUsage() (AIUsageDto, error)
	GetTrash() (TrashDto, error)
	GetInsights(period string) ([]string, error)
}
