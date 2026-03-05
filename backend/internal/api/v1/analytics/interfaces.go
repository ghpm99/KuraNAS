package analytics

import "nas-go/api/pkg/database"

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetOverviewData(period PeriodConfig, limits OverviewLimits) (OverviewDataModel, error)
}

type ServiceInterface interface {
	GetOverview(period string) (OverviewDto, error)
}

type OverviewLimits struct {
	RecentFiles    int
	TopExtensions  int
	TopFolders     int
	TopHotFolders  int
	TopDuplicates  int
	RecentLogError int
}
