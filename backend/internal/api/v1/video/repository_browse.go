package video

import (
	"database/sql"
	"errors"
	"fmt"

	files "nas-go/api/internal/api/v1/files"
	queries "nas-go/api/pkg/database/queries/video"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

// GetVideos lists video files joined with their metadata complement.
// Moved from the files core; serves GET /files/videos with the same shape.
func (r *Repository) GetVideos(page int, pageSize int) (utils.PaginationResponse[files.FileModel], error) {
	paginationResponse := utils.PaginationResponse[files.FileModel]{
		Items: []files.FileModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		pq.Array(utils.VideoFormats),
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetVideosQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var file files.FileModel
			var metadata VideoMetadataModel

			if err := rows.Scan(
				&file.ID,
				&file.Name,
				&file.Path,
				&file.ParentPath,
				&file.Format,
				&file.Size,
				&file.UpdatedAt,
				&file.CreatedAt,
				&file.LastInteraction,
				&file.LastBackup,
				&file.Type,
				&file.CheckSum,
				&file.DeletedAt,
				&file.Starred,
				&metadata.ID,
				&metadata.FileId,
				&metadata.Path,
				&metadata.FormatName,
				&metadata.Size,
				&metadata.Duration,
				&metadata.Width,
				&metadata.Height,
				&metadata.FrameRate,
				&metadata.NbFrames,
				&metadata.BitRate,
				&metadata.CodecName,
				&metadata.CodecLongName,
				&metadata.PixFmt,
				&metadata.Level,
				&metadata.Profile,
				&metadata.AspectRatio,
				&metadata.AudioCodec,
				&metadata.AudioChannels,
				&metadata.AudioSampleRate,
				&metadata.AudioBitRate,
				&metadata.CreatedAt,
			); err != nil {
				return err
			}

			file.Metadata = metadata

			paginationResponse.Items = append(paginationResponse.Items, file)
		}

		return rows.Err()
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to query files: %w", err)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
}

func (r *Repository) GetVideoDurationSecondsByFileID(fileID int) (sql.NullFloat64, error) {
	var durationSeconds sql.NullFloat64
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetVideoDurationSecondsByFileIDQuery, fileID).Scan(&durationSeconds)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return sql.NullFloat64{}, nil
	}
	if err != nil {
		return sql.NullFloat64{}, fmt.Errorf("GetVideoDurationSecondsByFileID: %w", err)
	}
	return durationSeconds, nil
}
