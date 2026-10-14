package video

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/video"
)

func (r *Repository) CreateCustomPlaylist(tx *sql.Tx, sourcePath string, name string) (VideoPlaylistModel, error) {
	var playlist VideoPlaylistModel
	err := tx.QueryRow(queries.CreateCustomPlaylistQuery, sourcePath, name).Scan(
		&playlist.ID,
		&playlist.Type,
		&playlist.SourcePath,
		&playlist.Name,
		&playlist.IsHidden,
		&playlist.IsAuto,
		&playlist.GroupMode,
		&playlist.Classification,
		&playlist.CreatedAt,
		&playlist.UpdatedAt,
		&playlist.LastPlayedAt,
	)
	if err != nil {
		return playlist, fmt.Errorf("falha ao criar playlist personalizada: %w", err)
	}
	return playlist, nil
}

func (r *Repository) DeleteCustomPlaylist(tx *sql.Tx, playlistID int) error {
	if _, err := tx.Exec(queries.DeleteCustomPlaylistQuery, playlistID); err != nil {
		return fmt.Errorf("falha ao excluir playlist personalizada: %w", err)
	}
	return nil
}
