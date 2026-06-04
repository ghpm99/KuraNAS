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

func (r *Repository) GetStorageKpis(period PeriodConfig) (StorageKpisModel, error) {
	var result StorageKpisModel
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.StorageKPIsQuery, period.Interval).Scan(
			&result.UsedBytes,
			&result.GrowthBytes,
			&result.FilesAdded,
			&result.FilesTotal,
			&result.FoldersTotal,
		)
	})
	if err != nil {
		return result, fmt.Errorf("analytics storage kpis failed: %w", err)
	}
	return result, nil
}

func (r *Repository) GetStorageTimeSeries(period PeriodConfig) ([]StorageTimeSeriesModel, error) {
	result := []StorageTimeSeriesModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.StorageTimeSeriesQuery, period.Interval)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var point StorageTimeSeriesModel
			if err := rows.Scan(&point.Date, &point.UsedBytes); err != nil {
				return err
			}
			result = append(result, point)
		}
		return rows.Err()
	})
	if err != nil {
		return result, fmt.Errorf("analytics storage timeseries failed: %w", err)
	}
	return result, nil
}

func (r *Repository) GetTypeDistribution() ([]TypeDistributionModel, error) {
	result := []TypeDistributionModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.TypeDistributionQuery)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item TypeDistributionModel
			if err := rows.Scan(&item.Category, &item.Count, &item.Bytes); err != nil {
				return err
			}
			result = append(result, item)
		}
		return rows.Err()
	})
	if err != nil {
		return result, fmt.Errorf("analytics type distribution failed: %w", err)
	}
	return result, nil
}

func (r *Repository) GetExtensionDistribution(limit int) ([]ExtensionDistributionModel, error) {
	result := []ExtensionDistributionModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.ExtensionDistributionQuery, limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item ExtensionDistributionModel
			if err := rows.Scan(&item.Extension, &item.Count, &item.Bytes); err != nil {
				return err
			}
			if strings.EqualFold(item.Extension, "<none>") {
				item.Extension = "unknown"
			}
			result = append(result, item)
		}
		return rows.Err()
	})
	if err != nil {
		return result, fmt.Errorf("analytics extension distribution failed: %w", err)
	}
	return result, nil
}

func (r *Repository) GetRecentFiles(limit int) ([]RecentFileModel, error) {
	result := []RecentFileModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.RecentFilesQuery, limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item RecentFileModel
			if err := rows.Scan(&item.ID, &item.Name, &item.Path, &item.ParentPath, &item.Size, &item.Format, &item.CreatedAt, &item.UpdatedAt); err != nil {
				return err
			}
			result = append(result, item)
		}
		return rows.Err()
	})
	if err != nil {
		return result, fmt.Errorf("analytics recent files failed: %w", err)
	}
	return result, nil
}

func (r *Repository) GetTopFolders(limit int) ([]FolderUsageModel, error) {
	result := []FolderUsageModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.FolderSizeRankQuery, limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item FolderUsageModel
			if err := rows.Scan(&item.ParentPath, &item.TotalFiles, &item.TotalBytes, &item.LastModified); err != nil {
				return err
			}
			result = append(result, item)
		}
		return rows.Err()
	})
	if err != nil {
		return result, fmt.Errorf("analytics top folders failed: %w", err)
	}
	return result, nil
}

func (r *Repository) GetHotFolders(period PeriodConfig, limit int) ([]FolderHotModel, error) {
	result := []FolderHotModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.FolderHotRankQuery, period.Interval, limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item FolderHotModel
			if err := rows.Scan(&item.ParentPath, &item.NewFiles, &item.AddedBytes, &item.LastEvent); err != nil {
				return err
			}
			result = append(result, item)
		}
		return rows.Err()
	})
	if err != nil {
		return result, fmt.Errorf("analytics hot folders failed: %w", err)
	}
	return result, nil
}

func (r *Repository) GetDuplicatesSummary() (DuplicatesSummaryModel, error) {
	var result DuplicatesSummaryModel
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.DuplicatesSummaryQuery).Scan(
			&result.GroupsTotal,
			&result.FilesTotal,
			&result.ReclaimableBytes,
		)
	})
	if err != nil {
		return result, fmt.Errorf("analytics duplicates summary failed: %w", err)
	}
	return result, nil
}

func (r *Repository) GetDuplicateGroups(limit int) ([]DuplicateGroupModel, error) {
	result := []DuplicateGroupModel{}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.DuplicatesTopGroupsQuery, limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item DuplicateGroupModel
			if err := rows.Scan(&item.Signature, &item.Copies, &item.ItemSize, &item.ReclaimableSize, pq.Array(&item.Paths)); err != nil {
				return err
			}
			result = append(result, item)
		}
		return rows.Err()
	})
	if err != nil {
		return result, fmt.Errorf("analytics duplicate groups failed: %w", err)
	}
	return result, nil
}

func (r *Repository) GetLibrarySummary() (LibrarySummaryModel, error) {
	var result LibrarySummaryModel
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.LibraryMetadataSummaryQuery).Scan(
			&result.CategorizedMedia,
			&result.AudioWithMetadata,
			&result.VideoWithMetadata,
			&result.ImageWithMetadata,
			&result.ImageClassified,
		)
	})
	if err != nil {
		return result, fmt.Errorf("analytics library summary failed: %w", err)
	}
	return result, nil
}

func (r *Repository) GetProcessingSummary() (ProcessingSummaryModel, error) {
	var result ProcessingSummaryModel
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.ProcessingQueueSummaryQuery).Scan(
			&result.MetadataPending,
			&result.MetadataFailed,
			&result.ThumbnailPending,
			&result.ThumbnailFailed,
			&result.RecurringTimeouts,
		)
	})
	if err != nil {
		return result, fmt.Errorf("analytics processing summary failed: %w", err)
	}
	return result, nil
}

func (r *Repository) GetHealth() (HealthModel, error) {
	result := HealthModel{RecentErrors: []LogErrorModel{}}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		if err := tx.QueryRow(queries.IndexHealthStatusQuery).Scan(
			&result.Status, &result.LastScanStart, &result.LastScanEnd,
		); err != nil && err != sql.ErrNoRows {
			return err
		}

		if err := tx.QueryRow(queries.IndexFilesTotalQuery).Scan(&result.IndexedFiles); err != nil {
			return err
		}

		if err := tx.QueryRow(queries.IndexErrorsRecentQuery).Scan(&result.ErrorsLast24h); err != nil {
			return err
		}

		errRows, err := tx.Query(queries.IndexErrorsLatestQuery, 5)
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
		return errRows.Err()
	})
	if err != nil {
		return result, fmt.Errorf("analytics health failed: %w", err)
	}
	return result, nil
}
