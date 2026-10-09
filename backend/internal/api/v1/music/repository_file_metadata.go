package music

import (
	"database/sql"
	"fmt"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/music"
)

type AudioSummaryRepositoryInterface interface {
	GetAudioSummaryByFileID(fileID int) (AudioSummaryDto, error)
}

type AudioSummaryRepository struct {
	Db *database.DbContext
}

func NewAudioSummaryRepository(db *database.DbContext) *AudioSummaryRepository {
	return &AudioSummaryRepository{Db: db}
}

func (r *AudioSummaryRepository) GetAudioSummaryByFileID(fileID int) (AudioSummaryDto, error) {
	var summary AudioSummaryDto

	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		row := tx.QueryRow(queries.GetAudioSummaryByFileIDQuery, fileID)
		return row.Scan(
			&summary.Title,
			&summary.Artist,
			&summary.Album,
			&summary.Genre,
			&summary.Year,
			&summary.TrackNumber,
			&summary.LengthSeconds,
			&summary.Bitrate,
			&summary.SampleRate,
			&summary.Channels,
		)
	})
	if err != nil {
		return AudioSummaryDto{}, fmt.Errorf("GetAudioSummaryByFileID: %w", err)
	}
	return summary, nil
}
