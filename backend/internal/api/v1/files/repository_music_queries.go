package files

import (
	"database/sql"
	"fmt"

	queries "nas-go/api/pkg/database/queries/file"
	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

func (r *Repository) GetMusic(page int, pageSize int) (utils.PaginationResponse[FileModel], error) {

	paginationResponse := utils.PaginationResponse[FileModel]{
		Items: []FileModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		pq.Array(utils.AudioFormats),
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(
			queries.GetMusicQuery,
			args...,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var file FileModel
			var metadata AudioMetadataModel

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
				&metadata.Mime,
				&metadata.Length,
				&metadata.Bitrate,
				&metadata.SampleRate,
				&metadata.Channels,
				&metadata.BitrateMode,
				&metadata.EncoderInfo,
				&metadata.BitDepth,
				&metadata.Title,
				&metadata.Artist,
				&metadata.Album,
				&metadata.AlbumArtist,
				&metadata.TrackNumber,
				&metadata.Genre,
				&metadata.Composer,
				&metadata.Year,
				&metadata.RecordingDate,
				&metadata.Encoder,
				&metadata.Publisher,
				&metadata.OriginalReleaseDate,
				&metadata.OriginalArtist,
				&metadata.Lyricist,
				&metadata.Lyrics,
				&metadata.CreatedAt,
			); err != nil {
				return err
			}

			file.Metadata = metadata

			paginationResponse.Items = append(paginationResponse.Items, file)
		}

		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to query files: %w", err)
	}

	paginationResponse.UpdatePagination()

	return paginationResponse, nil
}

func (r *Repository) GetMusicArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistDto], error) {
	paginationResponse := utils.PaginationResponse[MusicArtistDto]{
		Items: []MusicArtistDto{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		pq.Array(utils.AudioFormats),
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetMusicArtistsQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var artist MusicArtistDto
			if err := rows.Scan(&artist.Artist, &artist.TrackCount, &artist.AlbumCount); err != nil {
				return err
			}
			paginationResponse.Items = append(paginationResponse.Items, artist)
		}
		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to query artists: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) GetMusicByArtist(artist string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	paginationResponse := utils.PaginationResponse[FileModel]{
		Items: []FileModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		pq.Array(utils.AudioFormats),
		artist,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetMusicByArtistQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var file FileModel
			var metadata AudioMetadataModel

			if err := rows.Scan(
				&file.ID, &file.Name, &file.Path, &file.ParentPath,
				&file.Format, &file.Size, &file.UpdatedAt, &file.CreatedAt,
				&file.LastInteraction, &file.LastBackup, &file.Type,
				&file.CheckSum, &file.DeletedAt, &file.Starred,
				&metadata.ID, &metadata.FileId, &metadata.Path,
				&metadata.Mime, &metadata.Length, &metadata.Bitrate,
				&metadata.SampleRate, &metadata.Channels,
				&metadata.BitrateMode, &metadata.EncoderInfo, &metadata.BitDepth,
				&metadata.Title, &metadata.Artist, &metadata.Album,
				&metadata.AlbumArtist, &metadata.TrackNumber, &metadata.Genre,
				&metadata.Composer, &metadata.Year, &metadata.RecordingDate,
				&metadata.Encoder, &metadata.Publisher, &metadata.OriginalReleaseDate,
				&metadata.OriginalArtist, &metadata.Lyricist, &metadata.Lyrics,
				&metadata.CreatedAt,
			); err != nil {
				return err
			}

			file.Metadata = metadata
			paginationResponse.Items = append(paginationResponse.Items, file)
		}
		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to query music by artist: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) GetMusicAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumDto], error) {
	paginationResponse := utils.PaginationResponse[MusicAlbumDto]{
		Items: []MusicAlbumDto{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		pq.Array(utils.AudioFormats),
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetMusicAlbumsQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var album MusicAlbumDto
			if err := rows.Scan(&album.Album, &album.Artist, &album.Year, &album.TrackCount); err != nil {
				return err
			}
			paginationResponse.Items = append(paginationResponse.Items, album)
		}
		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to query albums: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) GetMusicByAlbum(album string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	paginationResponse := utils.PaginationResponse[FileModel]{
		Items: []FileModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		pq.Array(utils.AudioFormats),
		album,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetMusicByAlbumQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var file FileModel
			var metadata AudioMetadataModel

			if err := rows.Scan(
				&file.ID, &file.Name, &file.Path, &file.ParentPath,
				&file.Format, &file.Size, &file.UpdatedAt, &file.CreatedAt,
				&file.LastInteraction, &file.LastBackup, &file.Type,
				&file.CheckSum, &file.DeletedAt, &file.Starred,
				&metadata.ID, &metadata.FileId, &metadata.Path,
				&metadata.Mime, &metadata.Length, &metadata.Bitrate,
				&metadata.SampleRate, &metadata.Channels,
				&metadata.BitrateMode, &metadata.EncoderInfo, &metadata.BitDepth,
				&metadata.Title, &metadata.Artist, &metadata.Album,
				&metadata.AlbumArtist, &metadata.TrackNumber, &metadata.Genre,
				&metadata.Composer, &metadata.Year, &metadata.RecordingDate,
				&metadata.Encoder, &metadata.Publisher, &metadata.OriginalReleaseDate,
				&metadata.OriginalArtist, &metadata.Lyricist, &metadata.Lyrics,
				&metadata.CreatedAt,
			); err != nil {
				return err
			}

			file.Metadata = metadata
			paginationResponse.Items = append(paginationResponse.Items, file)
		}
		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to query music by album: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) GetMusicGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreDto], error) {
	paginationResponse := utils.PaginationResponse[MusicGenreDto]{
		Items: []MusicGenreDto{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		pq.Array(utils.AudioFormats),
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetMusicGenresQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var genre MusicGenreDto
			if err := rows.Scan(&genre.Genre, &genre.TrackCount); err != nil {
				return err
			}
			paginationResponse.Items = append(paginationResponse.Items, genre)
		}
		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to query genres: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) GetMusicByGenre(genre string, page int, pageSize int) (utils.PaginationResponse[FileModel], error) {
	paginationResponse := utils.PaginationResponse[FileModel]{
		Items: []FileModel{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		pq.Array(utils.AudioFormats),
		genre,
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetMusicByGenreQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var file FileModel
			var metadata AudioMetadataModel

			if err := rows.Scan(
				&file.ID, &file.Name, &file.Path, &file.ParentPath,
				&file.Format, &file.Size, &file.UpdatedAt, &file.CreatedAt,
				&file.LastInteraction, &file.LastBackup, &file.Type,
				&file.CheckSum, &file.DeletedAt, &file.Starred,
				&metadata.ID, &metadata.FileId, &metadata.Path,
				&metadata.Mime, &metadata.Length, &metadata.Bitrate,
				&metadata.SampleRate, &metadata.Channels,
				&metadata.BitrateMode, &metadata.EncoderInfo, &metadata.BitDepth,
				&metadata.Title, &metadata.Artist, &metadata.Album,
				&metadata.AlbumArtist, &metadata.TrackNumber, &metadata.Genre,
				&metadata.Composer, &metadata.Year, &metadata.RecordingDate,
				&metadata.Encoder, &metadata.Publisher, &metadata.OriginalReleaseDate,
				&metadata.OriginalArtist, &metadata.Lyricist, &metadata.Lyrics,
				&metadata.CreatedAt,
			); err != nil {
				return err
			}

			file.Metadata = metadata
			paginationResponse.Items = append(paginationResponse.Items, file)
		}
		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to query music by genre: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}

func (r *Repository) GetMusicFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderDto], error) {
	paginationResponse := utils.PaginationResponse[MusicFolderDto]{
		Items: []MusicFolderDto{},
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  false,
			HasPrev:  false,
		},
	}

	args := []any{
		pq.Array(utils.AudioFormats),
		pageSize + 1,
		utils.CalculateOffset(page, pageSize),
	}

	err := r.DbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(queries.GetMusicFoldersQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var folder MusicFolderDto
			if err := rows.Scan(&folder.Folder, &folder.TrackCount); err != nil {
				return err
			}
			paginationResponse.Items = append(paginationResponse.Items, folder)
		}
		return nil
	})

	if err != nil {
		return paginationResponse, fmt.Errorf("failed to query music folders: %w", err)
	}

	paginationResponse.UpdatePagination()
	return paginationResponse, nil
}
