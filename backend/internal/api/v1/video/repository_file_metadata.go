package video

import (
	"database/sql"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/video"
)

type VideoSummaryRepositoryInterface interface {
	GetVideoSummaryByFileID(fileID int) (VideoSummaryDto, error)
}

type VideoSummaryRepository struct {
	Db *database.DbContext
}

func NewVideoSummaryRepository(db *database.DbContext) *VideoSummaryRepository {
	return &VideoSummaryRepository{Db: db}
}

func (r *VideoSummaryRepository) GetVideoSummaryByFileID(fileID int) (VideoSummaryDto, error) {
	var summary VideoSummaryDto

	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.GetVideoSummaryByFileIDQuery, fileID)
		return row.Scan(
			&summary.Duration,
			&summary.Width,
			&summary.Height,
			&summary.CodecName,
			&summary.FrameRate,
			&summary.BitRate,
			&summary.AudioCodec,
			&summary.FormatName,
		)
	})
	if err != nil {
		return VideoSummaryDto{}, fmt.Errorf("GetVideoSummaryByFileID: %w", err)
	}
	return summary, nil
}
