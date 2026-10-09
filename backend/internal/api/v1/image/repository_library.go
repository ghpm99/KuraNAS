package image

import (
	"database/sql"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/image"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

type LibraryRepository struct {
	Db *database.DbContext
}

func NewLibraryRepository(db *database.DbContext) *LibraryRepository {
	return &LibraryRepository{Db: db}
}

func (r *LibraryRepository) ListLibraryImages(query LibraryListQuery) ([]LibraryItemModel, error) {
	listQuery, arguments, err := buildLibraryListQuery(query)
	if err != nil {
		return nil, fmt.Errorf("ListLibraryImages: %w", err)
	}

	items := []LibraryItemModel{}
	err = r.Db.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(listQuery, arguments...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var item LibraryItemModel
			var takenAt sql.NullTime
			if err := rows.Scan(
				&item.FileID,
				&item.Name,
				&item.Path,
				&item.ParentPath,
				&item.Format,
				&item.Size,
				&item.Width,
				&item.Height,
				&takenAt,
				&item.Category,
				&item.Starred,
				&item.IsCold,
				&item.UpdatedAt,
			); err != nil {
				return err
			}
			if takenAt.Valid {
				item.TakenAt = &takenAt.Time
			}
			items = append(items, item)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("ListLibraryImages: %w", err)
	}
	return items, nil
}

func (r *LibraryRepository) GetLibraryItemCursor(fileID int) (LibraryCursor, error) {
	var takenAt sql.NullTime
	var cursor LibraryCursor
	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.LibraryItemCursorQuery, pq.Array(utils.ImageFormats), fileID).Scan(&takenAt, &cursor.FileID)
	})
	if err != nil {
		return LibraryCursor{}, fmt.Errorf("GetLibraryItemCursor: %w", err)
	}
	if takenAt.Valid {
		cursor.TakenAt = &takenAt.Time
	}
	return cursor, nil
}

func (r *LibraryRepository) CountLibraryImages(filter LibraryFilter) (int, error) {
	countQuery, arguments := buildLibraryCountQuery(filter)

	var total int
	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(countQuery, arguments...).Scan(&total)
	})
	if err != nil {
		return 0, fmt.Errorf("CountLibraryImages: %w", err)
	}
	return total, nil
}

func (r *LibraryRepository) ListLibraryTimeline(filter LibraryFilter) ([]LibraryTimelineBucketModel, error) {
	timelineQuery, arguments := buildLibraryTimelineQuery(filter)

	buckets := []LibraryTimelineBucketModel{}
	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(timelineQuery, arguments...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var bucket LibraryTimelineBucketModel
			if err := rows.Scan(&bucket.Year, &bucket.Month, &bucket.Count); err != nil {
				return err
			}
			buckets = append(buckets, bucket)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("ListLibraryTimeline: %w", err)
	}
	return buckets, nil
}
