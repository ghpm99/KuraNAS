package search

import (
	"database/sql"
	"fmt"
	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/search"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

type Repository struct {
	DbContext *database.DbContext
}

func NewRepository(database *database.DbContext) *Repository {
	return &Repository{DbContext: database}
}

func (r *Repository) scanRows(query string, scanFn func(*sql.Rows) error, args ...any) error {
	return r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			if err := scanFn(rows); err != nil {
				return err
			}
		}

		return rows.Err()
	})
}

func (r *Repository) SearchFiles(query string, limit int) ([]FileResultModel, error) {
	results := []FileResultModel{}
	err := r.scanRows(queries.SearchFilesQuery, func(rows *sql.Rows) error {
		var item FileResultModel
		if err := rows.Scan(&item.ID, &item.Name, &item.Path, &item.ParentPath, &item.Format, &item.Starred); err != nil {
			return err
		}
		results = append(results, item)
		return nil
	}, query, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar arquivos: %w", err)
	}
	return results, nil
}

func (r *Repository) SearchFolders(query string, limit int) ([]FolderResultModel, error) {
	results := []FolderResultModel{}
	err := r.scanRows(queries.SearchFoldersQuery, func(rows *sql.Rows) error {
		var item FolderResultModel
		if err := rows.Scan(&item.ID, &item.Name, &item.Path, &item.ParentPath, &item.Starred); err != nil {
			return err
		}
		results = append(results, item)
		return nil
	}, query, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar pastas: %w", err)
	}
	return results, nil
}

func (r *Repository) SearchArtists(query string, limit int) ([]ArtistResultModel, error) {
	results := []ArtistResultModel{}
	err := r.scanRows(queries.SearchArtistsQuery, func(rows *sql.Rows) error {
		var item ArtistResultModel
		if err := rows.Scan(&item.Artist, &item.TrackCount, &item.AlbumCount); err != nil {
			return err
		}
		results = append(results, item)
		return nil
	}, query, pq.Array(utils.AudioFormats), limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar artistas: %w", err)
	}
	return results, nil
}

func (r *Repository) SearchAlbums(query string, limit int) ([]AlbumResultModel, error) {
	results := []AlbumResultModel{}
	err := r.scanRows(queries.SearchAlbumsQuery, func(rows *sql.Rows) error {
		var item AlbumResultModel
		if err := rows.Scan(&item.Artist, &item.Album, &item.Year, &item.TrackCount); err != nil {
			return err
		}
		results = append(results, item)
		return nil
	}, query, pq.Array(utils.AudioFormats), limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar albuns: %w", err)
	}
	return results, nil
}

func (r *Repository) SearchMusicPlaylists(query string, limit int) ([]MusicPlaylistResultModel, error) {
	results := []MusicPlaylistResultModel{}
	err := r.scanRows(queries.SearchMusicPlaylistsQuery, func(rows *sql.Rows) error {
		var item MusicPlaylistResultModel
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.IsSystem, &item.UpdatedAt, &item.TrackCount); err != nil {
			return err
		}
		results = append(results, item)
		return nil
	}, query, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar playlists de musica: %w", err)
	}
	return results, nil
}

func (r *Repository) SearchVideoPlaylists(query string, limit int) ([]VideoPlaylistResultModel, error) {
	results := []VideoPlaylistResultModel{}
	err := r.scanRows(queries.SearchVideoPlaylistsQuery, func(rows *sql.Rows) error {
		var item VideoPlaylistResultModel
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Type,
			&item.Classification,
			&item.SourcePath,
			&item.IsAuto,
			&item.UpdatedAt,
			&item.ItemCount,
		); err != nil {
			return err
		}
		results = append(results, item)
		return nil
	}, query, limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar playlists de video: %w", err)
	}
	return results, nil
}

func (r *Repository) SearchVideos(query string, limit int) ([]VideoResultModel, error) {
	results := []VideoResultModel{}
	err := r.scanRows(queries.SearchVideosQuery, func(rows *sql.Rows) error {
		var item VideoResultModel
		if err := rows.Scan(&item.ID, &item.Name, &item.Path, &item.ParentPath, &item.Format); err != nil {
			return err
		}
		results = append(results, item)
		return nil
	}, query, pq.Array(utils.VideoFormats), limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar videos: %w", err)
	}
	return results, nil
}

func (r *Repository) SearchImages(query string, limit int) ([]ImageResultModel, error) {
	results := []ImageResultModel{}
	err := r.scanRows(queries.SearchImagesQuery, func(rows *sql.Rows) error {
		var item ImageResultModel
		if err := rows.Scan(&item.ID, &item.Name, &item.Path, &item.ParentPath, &item.Format, &item.Category, &item.Context); err != nil {
			return err
		}
		results = append(results, item)
		return nil
	}, query, pq.Array(utils.ImageFormats), limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar imagens: %w", err)
	}
	return results, nil
}
