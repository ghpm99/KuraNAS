package music

import (
	"database/sql"
	"errors"
	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"
	"strings"
	"time"
)

const (
	defaultMusicHomeLimit      = 4
	defaultAutomaticTrackLimit = 50
)

var ErrAutoPlaylistReadOnly = errors.New("automatic playlists are read-only")

func normalizePagination(page int, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	return page, pageSize
}

func paginateItems[T any](items []T, page int, pageSize int) utils.PaginationResponse[T] {
	page, pageSize = normalizePagination(page, pageSize)

	offset := utils.CalculateOffset(page, pageSize)
	if offset >= len(items) {
		return utils.PaginationResponse[T]{
			Items: []T{},
			Pagination: utils.Pagination{
				Page:     page,
				PageSize: pageSize,
				HasNext:  false,
				HasPrev:  page > 1,
			},
		}
	}

	end := offset + pageSize
	hasNext := end < len(items)
	if end > len(items) {
		end = len(items)
	}

	return utils.PaginationResponse[T]{
		Items: items[offset:end],
		Pagination: utils.Pagination{
			Page:     page,
			PageSize: pageSize,
			HasNext:  hasNext,
			HasPrev:  page > 1,
		},
	}
}

func fileModelToPlaylistTrackDto(fileModel files.FileModel, position int) (PlaylistTrackDto, error) {
	fileDto, err := fileModel.ToDto()
	if err != nil {
		return PlaylistTrackDto{}, err
	}
	fileDto.Metadata = fileModel.Metadata

	return PlaylistTrackDto{
		ID:       fileModel.ID,
		Position: position,
		AddedAt:  fileModel.CreatedAt,
		File:     fileDto,
	}, nil
}

func (s *Service) loadPlaylistTracksByIDs(fileIDs []int, page int, pageSize int) (utils.PaginationResponse[PlaylistTrackDto], error) {
	paginatedIDs := paginateItems(fileIDs, page, pageSize)
	filesByID := map[int]files.FileModel{}

	fileModels, err := s.Repository.GetLibraryFilesByIDs(paginatedIDs.Items)
	if err != nil {
		return utils.PaginationResponse[PlaylistTrackDto]{}, err
	}

	for _, fileModel := range fileModels {
		filesByID[fileModel.ID] = fileModel
	}

	items := make([]PlaylistTrackDto, 0, len(paginatedIDs.Items))
	offset := utils.CalculateOffset(page, pageSize)
	for index, fileID := range paginatedIDs.Items {
		fileModel, exists := filesByID[fileID]
		if !exists {
			continue
		}

		trackDto, err := fileModelToPlaylistTrackDto(fileModel, offset+index+1)
		if err != nil {
			return utils.PaginationResponse[PlaylistTrackDto]{}, err
		}
		items = append(items, trackDto)
	}

	return utils.PaginationResponse[PlaylistTrackDto]{
		Items: items,
		Pagination: utils.Pagination{
			Page:     paginatedIDs.Pagination.Page,
			PageSize: paginatedIDs.Pagination.PageSize,
			HasNext:  paginatedIDs.Pagination.HasNext,
			HasPrev:  paginatedIDs.Pagination.HasPrev,
		},
	}, nil
}

func (s *Service) GetHomeCatalog(clientID string, limit int) (MusicHomeCatalogDto, error) {
	if limit <= 0 {
		limit = defaultMusicHomeLimit
	}

	summary, err := s.Repository.GetLibrarySummary()
	if err != nil {
		return MusicHomeCatalogDto{}, err
	}

	playlists, err := s.GetAutomaticPlaylists(clientID)
	if err != nil {
		return MusicHomeCatalogDto{}, err
	}
	if len(playlists) > limit {
		playlists = playlists[:limit]
	}

	artists, err := s.Repository.GetLibraryArtistGroups(1, limit)
	if err != nil {
		return MusicHomeCatalogDto{}, err
	}

	albums, err := s.Repository.GetLibraryAlbumGroups(1, limit)
	if err != nil {
		return MusicHomeCatalogDto{}, err
	}

	return MusicHomeCatalogDto{
		Summary:   summary,
		Playlists: playlists,
		Artists:   artists.Items,
		Albums:    albums.Items,
	}, nil
}

func (s *Service) GetLibraryTracks(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	tracks, err := s.Repository.GetLibraryTracks(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}
	return files.ParsePaginationToDto(&tracks)
}

func (s *Service) SearchLibraryTracks(searchText string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	tracks, err := s.Repository.SearchLibraryTracks(searchText, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}
	return files.ParsePaginationToDto(&tracks)
}

func (s *Service) GetLibraryArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistGroupDto], error) {
	page, pageSize = normalizePagination(page, pageSize)
	return s.Repository.GetLibraryArtistGroups(page, pageSize)
}

func (s *Service) GetLibraryAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumGroupDto], error) {
	page, pageSize = normalizePagination(page, pageSize)
	return s.Repository.GetLibraryAlbumGroups(page, pageSize)
}

func (s *Service) GetLibraryGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreGroupDto], error) {
	page, pageSize = normalizePagination(page, pageSize)
	return s.Repository.GetLibraryGenreGroups(page, pageSize)
}

func (s *Service) GetLibraryFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderGroupDto], error) {
	page, pageSize = normalizePagination(page, pageSize)
	return s.Repository.GetLibraryFolderGroups(page, pageSize)
}

func (s *Service) loadLibraryTracksOfIDPage(paginatedIDs utils.PaginationResponse[int]) (utils.PaginationResponse[files.FileDto], error) {
	fileModels, err := s.Repository.GetLibraryFilesByIDs(paginatedIDs.Items)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}

	filesByID := map[int]files.FileDto{}
	for _, fileModel := range fileModels {
		fileDto, dtoErr := fileModel.ToDto()
		if dtoErr != nil {
			return utils.PaginationResponse[files.FileDto]{}, dtoErr
		}
		fileDto.Metadata = fileModel.Metadata
		filesByID[fileModel.ID] = fileDto
	}

	items := make([]files.FileDto, 0, len(paginatedIDs.Items))
	for _, fileID := range paginatedIDs.Items {
		fileDto, exists := filesByID[fileID]
		if exists {
			items = append(items, fileDto)
		}
	}

	return utils.PaginationResponse[files.FileDto]{
		Items: items,
		Pagination: utils.Pagination{
			Page:     paginatedIDs.Pagination.Page,
			PageSize: paginatedIDs.Pagination.PageSize,
			HasNext:  paginatedIDs.Pagination.HasNext,
			HasPrev:  paginatedIDs.Pagination.HasPrev,
		},
	}, nil
}

func (s *Service) GetLibraryTracksByArtist(artistKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	page, pageSize = normalizePagination(page, pageSize)
	paginatedIDs, err := s.Repository.GetLibraryTrackIDsByArtist(artistKey, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}
	return s.loadLibraryTracksOfIDPage(paginatedIDs)
}

func (s *Service) GetLibraryTracksByAlbum(albumKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	page, pageSize = normalizePagination(page, pageSize)
	paginatedIDs, err := s.Repository.GetLibraryTrackIDsByAlbum(albumKey, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}
	return s.loadLibraryTracksOfIDPage(paginatedIDs)
}

func (s *Service) GetLibraryTracksByGenre(genreKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	page, pageSize = normalizePagination(page, pageSize)
	paginatedIDs, err := s.Repository.GetLibraryTrackIDsByGenre(genreKey, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}
	return s.loadLibraryTracksOfIDPage(paginatedIDs)
}

func (s *Service) GetLibraryTracksByFolder(folderPath string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	trimmedFolder := strings.TrimSpace(folderPath)
	if trimmedFolder == "" {
		return utils.PaginationResponse[files.FileDto]{}, nil
	}

	page, pageSize = normalizePagination(page, pageSize)
	paginatedIDs, err := s.Repository.GetLibraryTrackIDsByFolder(trimmedFolder, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}
	return s.loadLibraryTracksOfIDPage(paginatedIDs)
}

func buildAutomaticPlaylistDto(id int, nameKey string, descriptionKey string, sourceKey string, trackCount int) PlaylistDto {
	now := time.Now()
	return PlaylistDto{
		ID:          id,
		Name:        i18n.GetMessage(nameKey),
		Description: i18n.GetMessage(descriptionKey),
		IsSystem:    true,
		IsAuto:      true,
		Kind:        PlaylistKindAutomatic,
		SourceKey:   sourceKey,
		CreatedAt:   now,
		UpdatedAt:   now,
		TrackCount:  trackCount,
	}
}

func (s *Service) getOptionalPlayerState(clientID string) *PlayerStateModel {
	state, err := s.Repository.GetPlayerState(clientID)
	if err != nil {
		return nil
	}
	return &state
}

func (s *Service) getContinueListeningSourceTracks(state *PlayerStateModel) []PlaylistTrackModel {
	if state == nil || !state.PlaylistID.Valid || state.PlaylistID.Int64 <= 0 {
		return nil
	}

	tracks, err := s.Repository.GetPlaylistTracks(int(state.PlaylistID.Int64), 1, defaultAutomaticTrackLimit)
	if err != nil {
		return nil
	}

	return tracks.Items
}

func buildContinueListeningTrackIDs(recentFileIDs []int, state *PlayerStateModel, playlistTracks []PlaylistTrackModel) []int {
	if state == nil || !state.CurrentFileID.Valid {
		return []int{}
	}

	currentID := int(state.CurrentFileID.Int64)
	ids := []int{currentID}
	seen := map[int]bool{currentID: true}

	for _, track := range playlistTracks {
		if seen[track.FileID] {
			continue
		}
		seen[track.FileID] = true
		ids = append(ids, track.FileID)
		if len(ids) == defaultAutomaticTrackLimit {
			return ids
		}
	}

	for _, fileID := range recentFileIDs {
		if seen[fileID] {
			continue
		}
		seen[fileID] = true
		ids = append(ids, fileID)
		if len(ids) == defaultAutomaticTrackLimit {
			break
		}
	}

	return ids
}

func (s *Service) loadContinueListeningTrackIDs(clientID string) ([]int, error) {
	recentFileIDs, err := s.Repository.GetRecentLibraryFileIDs(defaultAutomaticTrackLimit * 2)
	if err != nil {
		return nil, err
	}

	state := s.getOptionalPlayerState(clientID)
	playlistTracks := s.getContinueListeningSourceTracks(state)

	return buildContinueListeningTrackIDs(recentFileIDs, state, playlistTracks), nil
}

func (s *Service) automaticPlaylistTrackIDs(clientID string, playlistID int) ([]int, error) {
	switch playlistID {
	case AutoPlaylistContinueListeningID:
		return s.loadContinueListeningTrackIDs(clientID)
	case AutoPlaylistRecentlyAddedID:
		return s.Repository.GetRecentLibraryFileIDs(defaultAutomaticTrackLimit)
	case AutoPlaylistFavoritesID:
		return s.Repository.GetFavoriteLibraryFileIDs(defaultAutomaticTrackLimit)
	default:
		return nil, sql.ErrNoRows
	}
}

func (s *Service) GetAutomaticPlaylists(clientID string) ([]PlaylistDto, error) {
	continueTracks, err := s.automaticPlaylistTrackIDs(clientID, AutoPlaylistContinueListeningID)
	if err != nil {
		return nil, err
	}

	recentTracks, err := s.automaticPlaylistTrackIDs(clientID, AutoPlaylistRecentlyAddedID)
	if err != nil {
		return nil, err
	}

	favoriteTracks, err := s.automaticPlaylistTrackIDs(clientID, AutoPlaylistFavoritesID)
	if err != nil {
		return nil, err
	}

	return []PlaylistDto{
		buildAutomaticPlaylistDto(
			AutoPlaylistContinueListeningID,
			"MUSIC_AUTO_PLAYLIST_CONTINUE_NAME",
			"MUSIC_AUTO_PLAYLIST_CONTINUE_DESCRIPTION",
			autoPlaylistContinueListeningKey,
			len(continueTracks),
		),
		buildAutomaticPlaylistDto(
			AutoPlaylistRecentlyAddedID,
			"MUSIC_AUTO_PLAYLIST_RECENT_NAME",
			"MUSIC_AUTO_PLAYLIST_RECENT_DESCRIPTION",
			autoPlaylistRecentlyAddedKey,
			len(recentTracks),
		),
		buildAutomaticPlaylistDto(
			AutoPlaylistFavoritesID,
			"MUSIC_AUTO_PLAYLIST_FAVORITES_NAME",
			"MUSIC_AUTO_PLAYLIST_FAVORITES_DESCRIPTION",
			autoPlaylistFavoritesKey,
			len(favoriteTracks),
		),
	}, nil
}
