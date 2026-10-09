package music

import (
	"database/sql"
	"fmt"
	"strings"

	queries "nas-go/api/pkg/database/queries/music"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

func (r *Repository) queryQueueEntries(query string, args ...any) ([]MusicQueueEntryModel, error) {
	entries := []MusicQueueEntryModel{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var entry MusicQueueEntryModel
			if scanErr := rows.Scan(
				&entry.FileID,
				&entry.Name,
				&entry.Path,
				&entry.Format,
				&entry.Title,
				&entry.Artist,
				&entry.Album,
				&entry.LengthSeconds,
			); scanErr != nil {
				return scanErr
			}
			entries = append(entries, entry)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao montar a fila de reproducao: %w", err)
	}

	return entries, nil
}

func (r *Repository) GetLibraryQueueByArtist(artistKey string, limit int) ([]MusicQueueEntryModel, error) {
	return r.queryQueueEntries(queries.GetLibraryQueueByArtistQuery, pq.Array(utils.AudioFormats), artistKey, limit)
}

func (r *Repository) GetLibraryQueueByAlbum(albumKey string, limit int) ([]MusicQueueEntryModel, error) {
	return r.queryQueueEntries(queries.GetLibraryQueueByAlbumQuery, pq.Array(utils.AudioFormats), albumKey, limit)
}

func (r *Repository) GetLibraryQueueByGenre(genreKey string, limit int) ([]MusicQueueEntryModel, error) {
	return r.queryQueueEntries(queries.GetLibraryQueueByGenreQuery, pq.Array(utils.AudioFormats), genreKey, limit)
}

func (r *Repository) GetLibraryQueueByFolder(folderPath string, limit int) ([]MusicQueueEntryModel, error) {
	subfolderPrefix := folderPath
	if !strings.HasSuffix(subfolderPrefix, "/") {
		subfolderPrefix += "/"
	}
	subfolderPattern := likePatternEscaper.Replace(subfolderPrefix) + "%"

	return r.queryQueueEntries(queries.GetLibraryQueueByFolderQuery, pq.Array(utils.AudioFormats), folderPath, subfolderPattern, limit)
}

func (r *Repository) GetPlaylistQueue(playlistID int, limit int) ([]MusicQueueEntryModel, error) {
	return r.queryQueueEntries(queries.GetPlaylistQueueQuery, playlistID, limit)
}

func (r *Repository) GetLibraryQueueByFileIDs(fileIDs []int) ([]MusicQueueEntryModel, error) {
	return r.queryQueueEntries(queries.GetLibraryQueueByFileIDsQuery, pq.Array(utils.AudioFormats), pq.Array(fileIDs))
}
