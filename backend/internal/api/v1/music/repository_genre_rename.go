package music

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/music"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

// ListGenreTrackSources returns, ordered by file id, the effective genre text
// of every active indexed track whose catalog genre keys contain genreKey.
func (r *AudioMetadataRepository) ListGenreTrackSources(genreKey string) ([]GenreTrackSource, error) {
	sources := []GenreTrackSource{}
	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetGenreTrackSourcesQuery, genreKey, pq.Array(utils.AudioFormats))
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var source GenreTrackSource
			if err := rows.Scan(&source.FileID, &source.Genre); err != nil {
				return err
			}
			sources = append(sources, source)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao listar faixas do genero: %w", err)
	}
	return sources, nil
}

// CountActiveTracksByGenreKey returns how many active indexed tracks have
// genreKey among their catalog genre keys.
func (r *AudioMetadataRepository) CountActiveTracksByGenreKey(genreKey string) (int, error) {
	var trackCount int
	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.CountActiveTracksByGenreKeyQuery, genreKey, pq.Array(utils.AudioFormats)).Scan(&trackCount)
	})
	if err != nil {
		return 0, fmt.Errorf("falha ao contar faixas do genero: %w", err)
	}
	return trackCount, nil
}
