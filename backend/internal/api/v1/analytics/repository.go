package analytics

import (
	"database/sql"
	"fmt"
	"strings"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/analytics"

	"github.com/lib/pq"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(database *database.DbContext) *Repository {
	return &Repository{DbContext: database}
}

func (r *Repository) GetDbContext() *database.DbContext {
	return r.DbContext
}

func (r *Repository) GetOverviewData(period PeriodConfig, limits OverviewLimits) (OverviewDataModel, error) {
	result := OverviewDataModel{
		TimeSeries:       []StorageTimeSeriesModel{},
		Types:            []TypeDistributionModel{},
		Extensions:       []ExtensionDistributionModel{},
		HotFolders:       []FolderHotModel{},
		TopFolders:       []FolderUsageModel{},
		RecentFiles:      []RecentFileModel{},
		TopDuplicateSets: []DuplicateGroupModel{},
		RecentErrors:     []LogErrorModel{},
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		if err := tx.QueryRow(queries.StorageKPIsQuery, period.Interval).Scan(
			&result.StorageKpis.UsedBytes,
			&result.StorageKpis.GrowthBytes,
			&result.StorageKpis.FilesAdded,
			&result.StorageKpis.FilesTotal,
			&result.StorageKpis.FoldersTotal,
		); err != nil {
			return err
		}

		timeSeriesRows, err := tx.Query(queries.StorageTimeSeriesQuery, period.Interval)
		if err != nil {
			return err
		}
		defer timeSeriesRows.Close()
		for timeSeriesRows.Next() {
			var point StorageTimeSeriesModel
			if err := timeSeriesRows.Scan(&point.Date, &point.UsedBytes); err != nil {
				return err
			}
			result.TimeSeries = append(result.TimeSeries, point)
		}

		typesRows, err := tx.Query(queries.TypeDistributionQuery)
		if err != nil {
			return err
		}
		defer typesRows.Close()
		for typesRows.Next() {
			var item TypeDistributionModel
			if err := typesRows.Scan(&item.Category, &item.Count, &item.Bytes); err != nil {
				return err
			}
			result.Types = append(result.Types, item)
		}

		extRows, err := tx.Query(queries.ExtensionDistributionQuery, limits.TopExtensions)
		if err != nil {
			return err
		}
		defer extRows.Close()
		for extRows.Next() {
			var item ExtensionDistributionModel
			if err := extRows.Scan(&item.Extension, &item.Count, &item.Bytes); err != nil {
				return err
			}
			result.Extensions = append(result.Extensions, item)
		}

		recentRows, err := tx.Query(queries.RecentFilesQuery, limits.RecentFiles)
		if err != nil {
			return err
		}
		defer recentRows.Close()
		for recentRows.Next() {
			var item RecentFileModel
			if err := recentRows.Scan(&item.ID, &item.Name, &item.Path, &item.ParentPath, &item.Size, &item.Format, &item.CreatedAt, &item.UpdatedAt); err != nil {
				return err
			}
			result.RecentFiles = append(result.RecentFiles, item)
		}

		topFolderRows, err := tx.Query(queries.FolderSizeRankQuery, limits.TopFolders)
		if err != nil {
			return err
		}
		defer topFolderRows.Close()
		for topFolderRows.Next() {
			var item FolderUsageModel
			if err := topFolderRows.Scan(&item.ParentPath, &item.TotalFiles, &item.TotalBytes, &item.LastModified); err != nil {
				return err
			}
			result.TopFolders = append(result.TopFolders, item)
		}

		hotRows, err := tx.Query(queries.FolderHotRankQuery, period.Interval, limits.TopHotFolders)
		if err != nil {
			return err
		}
		defer hotRows.Close()
		for hotRows.Next() {
			var item FolderHotModel
			if err := hotRows.Scan(&item.ParentPath, &item.NewFiles, &item.AddedBytes, &item.LastEvent); err != nil {
				return err
			}
			result.HotFolders = append(result.HotFolders, item)
		}

		if err := tx.QueryRow(queries.DuplicatesSummaryQuery).Scan(
			&result.Duplicates.GroupsTotal,
			&result.Duplicates.FilesTotal,
			&result.Duplicates.ReclaimableBytes,
		); err != nil {
			return err
		}

		if err := tx.QueryRow(queries.LibraryMetadataSummaryQuery).Scan(
			&result.LibrarySummary.CategorizedMedia,
			&result.LibrarySummary.AudioWithMetadata,
			&result.LibrarySummary.VideoWithMetadata,
			&result.LibrarySummary.ImageWithMetadata,
			&result.LibrarySummary.ImageClassified,
		); err != nil {
			return err
		}

		if err := tx.QueryRow(queries.ProcessingQueueSummaryQuery).Scan(
			&result.Processing.MetadataPending,
			&result.Processing.MetadataFailed,
			&result.Processing.ThumbnailPending,
			&result.Processing.ThumbnailFailed,
		); err != nil {
			return err
		}

		dupRows, err := tx.Query(queries.DuplicatesTopGroupsQuery, limits.TopDuplicates)
		if err != nil {
			return err
		}
		defer dupRows.Close()
		for dupRows.Next() {
			var item DuplicateGroupModel
			if err := dupRows.Scan(&item.Signature, &item.Copies, &item.ItemSize, &item.ReclaimableSize, pq.Array(&item.Paths)); err != nil {
				return err
			}
			result.TopDuplicateSets = append(result.TopDuplicateSets, item)
		}

		healthRow := tx.QueryRow(queries.IndexHealthStatusQuery)
		if err := healthRow.Scan(&result.HealthStatus, &result.LastScanStart, &result.LastScanEnd); err != nil && err != sql.ErrNoRows {
			return err
		}

		if err := tx.QueryRow(queries.IndexErrorsRecentQuery).Scan(&result.ErrorsLast24h); err != nil {
			return err
		}

		errRows, err := tx.Query(queries.IndexErrorsLatestQuery, limits.RecentLogError)
		if err != nil {
			return err
		}
		defer errRows.Close()
		for errRows.Next() {
			var item LogErrorModel
			if err := errRows.Scan(&item.Name, &item.Description, &item.CreatedAt); err != nil {
				return err
			}
			result.RecentErrors = append(result.RecentErrors, item)
		}

		return nil
	})

	if err != nil {
		return result, fmt.Errorf("analytics repository failed: %w", err)
	}

	for index := range result.Extensions {
		if strings.EqualFold(result.Extensions[index].Extension, "<none>") {
			result.Extensions[index].Extension = "unknown"
		}
	}

	return result, nil
}
