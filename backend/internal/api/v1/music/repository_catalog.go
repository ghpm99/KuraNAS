package music

import (
	"database/sql"
	"fmt"
	"strings"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/music"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

var likePatternEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func queryCatalogPage[Row any](
	dbContext *database.DbContext,
	query string,
	leadingArgs []any,
	page int,
	pageSize int,
	scanRow func(rows *sql.Rows) (Row, error),
) (utils.PaginationResponse[Row], error) {
	paginationResponse := utils.PaginationResponse[Row]{
		Items:      []Row{},
		Pagination: utils.Pagination{Page: page, PageSize: pageSize},
	}

	args := append(append([]any{}, leadingArgs...), pageSize+1, utils.CalculateOffset(page, pageSize))

	err := dbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			row, scanErr := scanRow(rows)
			if scanErr != nil {
				return scanErr
			}
			paginationResponse.Items = append(paginationResponse.Items, row)
		}
		return rows.Err()
	})
	if err != nil {
		return paginationResponse, err
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func queryFileIDs(dbContext *database.DbContext, query string, args ...any) ([]int, error) {
	fileIDs := []int{}

	err := dbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var fileID int
			if scanErr := rows.Scan(&fileID); scanErr != nil {
				return scanErr
			}
			fileIDs = append(fileIDs, fileID)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}

	return fileIDs, nil
}

func scanFileID(rows *sql.Rows) (int, error) {
	var fileID int
	err := rows.Scan(&fileID)
	return fileID, err
}

func (r *Repository) GetLibrarySummary() (MusicLibrarySummaryDto, error) {
	var summary MusicLibrarySummaryDto

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetLibrarySummaryQuery, pq.Array(utils.AudioFormats)).Scan(
			&summary.TotalTracks,
			&summary.TotalArtists,
			&summary.TotalAlbums,
			&summary.TotalGenres,
			&summary.TotalFolders,
		)
	})
	if err != nil {
		return summary, fmt.Errorf("falha ao contar a biblioteca de musica: %w", err)
	}

	return summary, nil
}

func (r *Repository) GetLibraryArtistGroups(page int, pageSize int) (utils.PaginationResponse[MusicArtistGroupDto], error) {
	paginatedGroups, err := queryCatalogPage(r.DbContext, queries.GetLibraryArtistGroupsQuery,
		[]any{pq.Array(utils.AudioFormats)}, page, pageSize,
		func(rows *sql.Rows) (MusicArtistGroupDto, error) {
			var group MusicArtistGroupDto
			err := rows.Scan(&group.Key, &group.Artist, &group.TrackCount, &group.AlbumCount)
			return group, err
		})
	if err != nil {
		return paginatedGroups, fmt.Errorf("falha ao listar artistas da biblioteca: %w", err)
	}
	return paginatedGroups, nil
}

func (r *Repository) GetLibraryAlbumGroups(page int, pageSize int) (utils.PaginationResponse[MusicAlbumGroupDto], error) {
	paginatedGroups, err := queryCatalogPage(r.DbContext, queries.GetLibraryAlbumGroupsQuery,
		[]any{pq.Array(utils.AudioFormats)}, page, pageSize,
		func(rows *sql.Rows) (MusicAlbumGroupDto, error) {
			var group MusicAlbumGroupDto
			err := rows.Scan(&group.Key, &group.Album, &group.Artist, &group.Year, &group.TrackCount)
			return group, err
		})
	if err != nil {
		return paginatedGroups, fmt.Errorf("falha ao listar albuns da biblioteca: %w", err)
	}
	return paginatedGroups, nil
}

func (r *Repository) GetLibraryGenreGroups(page int, pageSize int) (utils.PaginationResponse[MusicGenreGroupDto], error) {
	paginatedGroups, err := queryCatalogPage(r.DbContext, queries.GetLibraryGenreGroupsQuery,
		[]any{pq.Array(utils.AudioFormats)}, page, pageSize,
		func(rows *sql.Rows) (MusicGenreGroupDto, error) {
			var group MusicGenreGroupDto
			err := rows.Scan(&group.Key, &group.Genre, &group.TrackCount)
			return group, err
		})
	if err != nil {
		return paginatedGroups, fmt.Errorf("falha ao listar generos da biblioteca: %w", err)
	}
	return paginatedGroups, nil
}

func (r *Repository) GetLibraryFolderGroups(page int, pageSize int) (utils.PaginationResponse[MusicFolderGroupDto], error) {
	paginatedGroups, err := queryCatalogPage(r.DbContext, queries.GetLibraryFolderGroupsQuery,
		[]any{pq.Array(utils.AudioFormats)}, page, pageSize,
		func(rows *sql.Rows) (MusicFolderGroupDto, error) {
			var group MusicFolderGroupDto
			err := rows.Scan(&group.Folder, &group.TrackCount)
			return group, err
		})
	if err != nil {
		return paginatedGroups, fmt.Errorf("falha ao listar pastas da biblioteca: %w", err)
	}
	return paginatedGroups, nil
}

func (r *Repository) GetLibraryTrackIDsByArtist(artistKey string, page int, pageSize int) (utils.PaginationResponse[int], error) {
	return r.queryTrackIDsByKey(queries.GetLibraryTrackIDsByArtistQuery, artistKey, page, pageSize)
}

func (r *Repository) GetLibraryTrackIDsByAlbum(albumKey string, page int, pageSize int) (utils.PaginationResponse[int], error) {
	return r.queryTrackIDsByKey(queries.GetLibraryTrackIDsByAlbumQuery, albumKey, page, pageSize)
}

func (r *Repository) GetLibraryTrackIDsByGenre(genreKey string, page int, pageSize int) (utils.PaginationResponse[int], error) {
	return r.queryTrackIDsByKey(queries.GetLibraryTrackIDsByGenreQuery, genreKey, page, pageSize)
}

func (r *Repository) GetLibraryTrackIDsByFolder(folderPath string, page int, pageSize int) (utils.PaginationResponse[int], error) {
	subfolderPrefix := folderPath
	if !strings.HasSuffix(subfolderPrefix, "/") {
		subfolderPrefix += "/"
	}
	subfolderPattern := likePatternEscaper.Replace(subfolderPrefix) + "%"

	paginatedIDs, err := queryCatalogPage(r.DbContext, queries.GetLibraryTrackIDsByFolderQuery,
		[]any{pq.Array(utils.AudioFormats), folderPath, subfolderPattern}, page, pageSize, scanFileID)
	if err != nil {
		return paginatedIDs, fmt.Errorf("falha ao listar faixas da pasta: %w", err)
	}
	return paginatedIDs, nil
}

func (r *Repository) queryTrackIDsByKey(query string, groupKey string, page int, pageSize int) (utils.PaginationResponse[int], error) {
	paginatedIDs, err := queryCatalogPage(r.DbContext, query,
		[]any{pq.Array(utils.AudioFormats), groupKey}, page, pageSize, scanFileID)
	if err != nil {
		return paginatedIDs, fmt.Errorf("falha ao listar faixas da biblioteca: %w", err)
	}
	return paginatedIDs, nil
}

func (r *Repository) GetRecentLibraryFileIDs(limit int) ([]int, error) {
	fileIDs, err := queryFileIDs(r.DbContext, queries.GetRecentLibraryFileIDsQuery, pq.Array(utils.AudioFormats), limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar faixas recentes: %w", err)
	}
	return fileIDs, nil
}

func (r *Repository) GetFavoriteLibraryFileIDs(limit int) ([]int, error) {
	fileIDs, err := queryFileIDs(r.DbContext, queries.GetFavoriteLibraryFileIDsQuery, pq.Array(utils.AudioFormats), limit)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar faixas favoritas: %w", err)
	}
	return fileIDs, nil
}

func (r *Repository) GetLibraryFileIDsByArtistKeys(artistKeys []string) ([]int, error) {
	if len(artistKeys) == 0 {
		return []int{}, nil
	}

	fileIDs, err := queryFileIDs(r.DbContext, queries.GetLibraryFileIDsByArtistKeysQuery,
		pq.Array(utils.AudioFormats), pq.Array(artistKeys))
	if err != nil {
		return nil, fmt.Errorf("falha ao listar faixas dos artistas: %w", err)
	}
	return fileIDs, nil
}

func (r *Repository) GetArtistClusterInputs() ([]artistClusterInput, error) {
	inputs := []artistClusterInput{}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetArtistClusterInputsQuery, pq.Array(utils.AudioFormats))
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var input artistClusterInput
			if scanErr := rows.Scan(&input.Key, &input.Artist, &input.GenreHint, &input.TrackCount); scanErr != nil {
				return scanErr
			}
			inputs = append(inputs, input)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao listar artistas para agrupamento: %w", err)
	}

	return inputs, nil
}

func (r *AudioMetadataRepository) ListAudioWithoutCatalogKeys(afterAudioMetadataID int, limit int) ([]AudioCatalogKeySource, error) {
	pendingRows := []AudioCatalogKeySource{}

	err := r.Db.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.SelectAudioWithoutCatalogKeysQuery, afterAudioMetadataID, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var pendingRow AudioCatalogKeySource
			if scanErr := rows.Scan(&pendingRow.AudioMetadataID, &pendingRow.Artist, &pendingRow.AlbumArtist, &pendingRow.Album, &pendingRow.Genre); scanErr != nil {
				return scanErr
			}
			pendingRows = append(pendingRows, pendingRow)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao listar audios sem chaves de catalogo: %w", err)
	}

	return pendingRows, nil
}

func (r *AudioMetadataRepository) UpdateAudioCatalogKeys(tx *sql.Tx, audioMetadataID int, groupingKeys CatalogGroupingKeys) error {
	_, err := tx.Exec(queries.UpdateAudioCatalogKeysQuery,
		audioMetadataID,
		groupingKeys.ArtistKey,
		groupingKeys.ArtistLabel,
		groupingKeys.AlbumKey,
		groupingKeys.AlbumLabel,
		pq.Array(groupingKeys.GenreKeys),
		pq.Array(groupingKeys.GenreLabels),
	)
	if err != nil {
		return fmt.Errorf("falha ao atualizar chaves de catalogo do audio: %w", err)
	}
	return nil
}
