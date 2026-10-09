package music

import (
	"database/sql"
	"fmt"
	"strings"

	queries "nas-go/api/pkg/database/queries/music"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

func (r *Repository) GetLibraryAlbumSummary(albumKey string) (MusicAlbumSummaryDto, error) {
	var summary MusicAlbumSummaryDto
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetLibraryAlbumSummaryQuery, pq.Array(utils.AudioFormats), albumKey).Scan(
			&summary.Key, &summary.Name, &summary.Artist, &summary.Year,
			&summary.TrackCount, &summary.TotalLengthSeconds, &summary.DiscCount,
		)
	})
	if err != nil {
		return MusicAlbumSummaryDto{}, fmt.Errorf("falha ao carregar resumo do album: %w", err)
	}
	return summary, nil
}

func (r *Repository) GetLibraryArtistSummary(artistKey string) (MusicArtistSummaryDto, error) {
	var summary MusicArtistSummaryDto
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetLibraryArtistSummaryQuery, pq.Array(utils.AudioFormats), artistKey).Scan(
			&summary.Key, &summary.Name, &summary.TrackCount, &summary.AlbumCount, &summary.TotalLengthSeconds,
		)
	})
	if err != nil {
		return MusicArtistSummaryDto{}, fmt.Errorf("falha ao carregar resumo do artista: %w", err)
	}
	return summary, nil
}

func (r *Repository) GetLibraryGenreSummary(genreKey string) (MusicGroupSummaryDto, error) {
	var summary MusicGroupSummaryDto
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetLibraryGenreSummaryQuery, pq.Array(utils.AudioFormats), genreKey).Scan(
			&summary.Key, &summary.Name, &summary.TrackCount, &summary.TotalLengthSeconds,
		)
	})
	if err != nil {
		return MusicGroupSummaryDto{}, fmt.Errorf("falha ao carregar resumo do genero: %w", err)
	}
	return summary, nil
}

func (r *Repository) GetLibraryFolderSummary(folderPath string) (MusicGroupSummaryDto, error) {
	summary := MusicGroupSummaryDto{Key: folderPath, Name: folderPath}
	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(queries.GetLibraryFolderSummaryQuery,
			pq.Array(utils.AudioFormats), folderPath, subfolderLikePattern(folderPath)).Scan(
			&summary.TrackCount, &summary.TotalLengthSeconds,
		)
	})
	if err != nil {
		return MusicGroupSummaryDto{}, fmt.Errorf("falha ao carregar resumo da pasta: %w", err)
	}
	return summary, nil
}

func (r *Repository) GetLibraryAlbumGroupsByArtist(artistKey string, page int, pageSize int) (utils.PaginationResponse[MusicAlbumGroupDto], error) {
	paginatedGroups, err := queryCatalogPage(r.DbContext, queries.GetLibraryAlbumGroupsByArtistQuery,
		[]any{pq.Array(utils.AudioFormats), artistKey}, page, pageSize, scanAlbumGroup)
	if err != nil {
		return paginatedGroups, fmt.Errorf("falha ao listar albuns do artista: %w", err)
	}
	return paginatedGroups, nil
}

func scanAlbumGroup(rows *sql.Rows) (MusicAlbumGroupDto, error) {
	var group MusicAlbumGroupDto
	err := rows.Scan(&group.Key, &group.Album, &group.Artist, &group.Year, &group.TrackCount)
	return group, err
}

func subfolderLikePattern(folderPath string) string {
	subfolderPrefix := folderPath
	if !strings.HasSuffix(subfolderPrefix, "/") {
		subfolderPrefix += "/"
	}
	return likePatternEscaper.Replace(subfolderPrefix) + "%"
}
