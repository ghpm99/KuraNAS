package music

import (
	"database/sql"
	"errors"
	"fmt"
	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMusicHomeLimit      = 4
	defaultAutomaticTrackLimit = 50
)

var (
	ErrAutoPlaylistReadOnly = errors.New("automatic playlists are read-only")
	spaceNormalizerRegexp   = regexp.MustCompile(`\s+`)
)

type musicArtistAccumulator struct {
	Key             string
	Artist          string
	TrackCount      int
	Albums          map[string]bool
	LatestTimestamp time.Time
}

type musicAlbumAccumulator struct {
	Key             string
	Album           string
	Artist          string
	Year            string
	TrackCount      int
	LatestTimestamp time.Time
}

type musicGenreAccumulator struct {
	Key             string
	Genre           string
	TrackCount      int
	LatestTimestamp time.Time
}

type musicFolderAccumulator struct {
	Folder          string
	TrackCount      int
	LatestTimestamp time.Time
}

func normalizeText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	return spaceNormalizerRegexp.ReplaceAllString(trimmed, " ")
}

func normalizeLookupKey(value string) string {
	normalized := normalizeText(strings.ToLower(value))
	normalized = strings.NewReplacer("_", " ", "-", " ", ".", " ").Replace(normalized)
	return spaceNormalizerRegexp.ReplaceAllString(normalized, " ")
}

func normalizeGenreLabel(value string) string {
	normalized := normalizeLookupKey(value)

	switch normalized {
	case "r&b", "r & b", "rnb", "rhythm and blues":
		return "R&B"
	case "r&b/soul", "r&b / soul", "rnb/soul", "rnb / soul", "soul/r&b", "soul / r&b":
		return "R&B / Soul"
	case "hip hop", "hiphop", "hip-hop":
		return "Hip-Hop"
	case "lo fi", "lofi", "lo-fi":
		return "Lo-Fi"
	case "soundtrack", "ost":
		return "Soundtrack"
	}

	if normalized == "" {
		return ""
	}

	parts := strings.Fields(normalized)
	for index, part := range parts {
		if part == "&" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}

	return strings.Join(parts, " ")
}

func normalizeGenreLabels(value string) []string {
	normalized := normalizeText(value)
	if normalized == "" {
		return nil
	}

	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == ';' || r == ',' || r == '|'
	})

	if len(parts) == 0 {
		parts = []string{normalized}
	}

	labels := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		label := normalizeGenreLabel(part)
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		labels = append(labels, label)
	}

	return labels
}

func preferredArtist(entry MusicLibraryIndexEntryModel) string {
	if artist := normalizeText(entry.AlbumArtist); artist != "" {
		return artist
	}
	return normalizeText(entry.Artist)
}

func entryTimestamp(entry MusicLibraryIndexEntryModel) time.Time {
	if entry.LastInteraction.Valid {
		return entry.LastInteraction.Time
	}
	if !entry.UpdatedAt.IsZero() {
		return entry.UpdatedAt
	}
	return entry.CreatedAt
}

func parseTrackNumber(value string) int {
	normalized := normalizeText(value)
	if normalized == "" {
		return 0
	}

	if slashIndex := strings.Index(normalized, "/"); slashIndex >= 0 {
		normalized = normalized[:slashIndex]
	}

	number, err := strconv.Atoi(normalized)
	if err != nil {
		return 0
	}
	return number
}

func paginateItems[T any](items []T, page int, pageSize int) utils.PaginationResponse[T] {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}

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

func buildArtistGroups(indexEntries []MusicLibraryIndexEntryModel) []MusicArtistGroupDto {
	accumulator := map[string]*musicArtistAccumulator{}

	for _, entry := range indexEntries {
		artist := preferredArtist(entry)
		if artist == "" {
			continue
		}

		key := normalizeLookupKey(artist)
		group := accumulator[key]
		if group == nil {
			group = &musicArtistAccumulator{
				Key:    key,
				Artist: artist,
				Albums: map[string]bool{},
			}
			accumulator[key] = group
		}

		group.TrackCount++
		if album := normalizeLookupKey(entry.Album); album != "" {
			group.Albums[album] = true
		}
		if timestamp := entryTimestamp(entry); timestamp.After(group.LatestTimestamp) {
			group.LatestTimestamp = timestamp
		}
	}

	results := make([]MusicArtistGroupDto, 0, len(accumulator))
	for _, item := range accumulator {
		results = append(results, MusicArtistGroupDto{
			Key:        item.Key,
			Artist:     item.Artist,
			TrackCount: item.TrackCount,
			AlbumCount: len(item.Albums),
		})
	}

	sort.Slice(results, func(left, right int) bool {
		if results[left].TrackCount != results[right].TrackCount {
			return results[left].TrackCount > results[right].TrackCount
		}
		return results[left].Artist < results[right].Artist
	})

	return results
}

func buildAlbumGroups(indexEntries []MusicLibraryIndexEntryModel) []MusicAlbumGroupDto {
	accumulator := map[string]*musicAlbumAccumulator{}

	for _, entry := range indexEntries {
		album := normalizeText(entry.Album)
		artist := preferredArtist(entry)
		if album == "" || artist == "" {
			continue
		}

		key := normalizeLookupKey(fmt.Sprintf("%s::%s", artist, album))
		group := accumulator[key]
		if group == nil {
			group = &musicAlbumAccumulator{
				Key:    key,
				Album:  album,
				Artist: artist,
				Year:   normalizeText(entry.Year),
			}
			accumulator[key] = group
		}

		group.TrackCount++
		if group.Year == "" {
			group.Year = normalizeText(entry.Year)
		}
		if timestamp := entryTimestamp(entry); timestamp.After(group.LatestTimestamp) {
			group.LatestTimestamp = timestamp
		}
	}

	results := make([]MusicAlbumGroupDto, 0, len(accumulator))
	for _, item := range accumulator {
		results = append(results, MusicAlbumGroupDto{
			Key:        item.Key,
			Album:      item.Album,
			Artist:     item.Artist,
			Year:       item.Year,
			TrackCount: item.TrackCount,
		})
	}

	sort.Slice(results, func(left, right int) bool {
		if results[left].TrackCount != results[right].TrackCount {
			return results[left].TrackCount > results[right].TrackCount
		}
		if results[left].Artist != results[right].Artist {
			return results[left].Artist < results[right].Artist
		}
		return results[left].Album < results[right].Album
	})

	return results
}

func buildGenreGroups(indexEntries []MusicLibraryIndexEntryModel) []MusicGenreGroupDto {
	accumulator := map[string]*musicGenreAccumulator{}

	for _, entry := range indexEntries {
		for _, genre := range normalizeGenreLabels(entry.Genre) {
			key := normalizeLookupKey(genre)
			group := accumulator[key]
			if group == nil {
				group = &musicGenreAccumulator{
					Key:   key,
					Genre: genre,
				}
				accumulator[key] = group
			}
			group.TrackCount++
			if timestamp := entryTimestamp(entry); timestamp.After(group.LatestTimestamp) {
				group.LatestTimestamp = timestamp
			}
		}
	}

	results := make([]MusicGenreGroupDto, 0, len(accumulator))
	for _, item := range accumulator {
		results = append(results, MusicGenreGroupDto{
			Key:        item.Key,
			Genre:      item.Genre,
			TrackCount: item.TrackCount,
		})
	}

	sort.Slice(results, func(left, right int) bool {
		if results[left].TrackCount != results[right].TrackCount {
			return results[left].TrackCount > results[right].TrackCount
		}
		return results[left].Genre < results[right].Genre
	})

	return results
}

func buildFolderGroups(indexEntries []MusicLibraryIndexEntryModel) []MusicFolderGroupDto {
	accumulator := map[string]*musicFolderAccumulator{}

	for _, entry := range indexEntries {
		folder := normalizeText(entry.ParentPath)
		if folder == "" {
			folder = "/"
		}

		group := accumulator[folder]
		if group == nil {
			group = &musicFolderAccumulator{Folder: folder}
			accumulator[folder] = group
		}
		group.TrackCount++
		if timestamp := entryTimestamp(entry); timestamp.After(group.LatestTimestamp) {
			group.LatestTimestamp = timestamp
		}
	}

	results := make([]MusicFolderGroupDto, 0, len(accumulator))
	for _, item := range accumulator {
		results = append(results, MusicFolderGroupDto{
			Folder:     item.Folder,
			TrackCount: item.TrackCount,
		})
	}

	sort.Slice(results, func(left, right int) bool {
		if results[left].TrackCount != results[right].TrackCount {
			return results[left].TrackCount > results[right].TrackCount
		}
		return results[left].Folder < results[right].Folder
	})

	return results
}

func sortArtistTracks(indexEntries []MusicLibraryIndexEntryModel) {
	sort.Slice(indexEntries, func(left, right int) bool {
		leftAlbum := normalizeText(indexEntries[left].Album)
		rightAlbum := normalizeText(indexEntries[right].Album)
		if leftAlbum != rightAlbum {
			return leftAlbum < rightAlbum
		}
		leftTrack := parseTrackNumber(indexEntries[left].TrackNumber)
		rightTrack := parseTrackNumber(indexEntries[right].TrackNumber)
		if leftTrack != rightTrack {
			return leftTrack < rightTrack
		}
		return normalizeText(indexEntries[left].Title) < normalizeText(indexEntries[right].Title)
	})
}

func sortAlbumTracks(indexEntries []MusicLibraryIndexEntryModel) {
	sort.Slice(indexEntries, func(left, right int) bool {
		leftTrack := parseTrackNumber(indexEntries[left].TrackNumber)
		rightTrack := parseTrackNumber(indexEntries[right].TrackNumber)
		if leftTrack != rightTrack {
			return leftTrack < rightTrack
		}
		return normalizeText(indexEntries[left].Title) < normalizeText(indexEntries[right].Title)
	})
}

func sortGenreTracks(indexEntries []MusicLibraryIndexEntryModel) {
	sort.Slice(indexEntries, func(left, right int) bool {
		leftArtist := preferredArtist(indexEntries[left])
		rightArtist := preferredArtist(indexEntries[right])
		if leftArtist != rightArtist {
			return leftArtist < rightArtist
		}
		leftAlbum := normalizeText(indexEntries[left].Album)
		rightAlbum := normalizeText(indexEntries[right].Album)
		if leftAlbum != rightAlbum {
			return leftAlbum < rightAlbum
		}
		leftTrack := parseTrackNumber(indexEntries[left].TrackNumber)
		rightTrack := parseTrackNumber(indexEntries[right].TrackNumber)
		if leftTrack != rightTrack {
			return leftTrack < rightTrack
		}
		return normalizeText(indexEntries[left].Title) < normalizeText(indexEntries[right].Title)
	})
}

func limitFileIDs(indexEntries []MusicLibraryIndexEntryModel, limit int) []int {
	if limit <= 0 {
		limit = defaultAutomaticTrackLimit
	}

	ids := make([]int, 0, limit)
	seen := map[int]bool{}
	for _, entry := range indexEntries {
		if seen[entry.FileID] {
			continue
		}
		seen[entry.FileID] = true
		ids = append(ids, entry.FileID)
		if len(ids) == limit {
			break
		}
	}
	return ids
}

func buildRecentPlaylistTrackIDs(indexEntries []MusicLibraryIndexEntryModel) []int {
	sortedEntries := append([]MusicLibraryIndexEntryModel(nil), indexEntries...)
	sort.Slice(sortedEntries, func(left, right int) bool {
		return entryTimestamp(sortedEntries[left]).After(entryTimestamp(sortedEntries[right]))
	})

	return limitFileIDs(sortedEntries, defaultAutomaticTrackLimit)
}

func buildFavoritePlaylistTrackIDs(indexEntries []MusicLibraryIndexEntryModel) []int {
	favorites := make([]MusicLibraryIndexEntryModel, 0, len(indexEntries))
	for _, entry := range indexEntries {
		if entry.Starred {
			favorites = append(favorites, entry)
		}
	}

	sort.Slice(favorites, func(left, right int) bool {
		return entryTimestamp(favorites[left]).After(entryTimestamp(favorites[right]))
	})

	return limitFileIDs(favorites, defaultAutomaticTrackLimit)
}

func buildContinueListeningTrackIDs(indexEntries []MusicLibraryIndexEntryModel, state *PlayerStateModel, playlistTracks []PlaylistTrackModel) []int {
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

	for _, entry := range indexEntries {
		if seen[entry.FileID] || !entry.LastInteraction.Valid {
			continue
		}
		seen[entry.FileID] = true
		ids = append(ids, entry.FileID)
		if len(ids) == defaultAutomaticTrackLimit {
			break
		}
	}

	return ids
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

func (s *Service) buildAutomaticPlaylists(clientID string, indexEntries []MusicLibraryIndexEntryModel) []PlaylistDto {
	state := s.getOptionalPlayerState(clientID)
	playlistTracks := s.getContinueListeningSourceTracks(state)

	continueTracks := buildContinueListeningTrackIDs(indexEntries, state, playlistTracks)
	recentTracks := buildRecentPlaylistTrackIDs(indexEntries)
	favoriteTracks := buildFavoritePlaylistTrackIDs(indexEntries)

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
	}
}

func (s *Service) automaticPlaylistTrackIDs(clientID string, playlistID int, indexEntries []MusicLibraryIndexEntryModel) ([]int, error) {
	state := s.getOptionalPlayerState(clientID)
	playlistTracks := s.getContinueListeningSourceTracks(state)

	switch playlistID {
	case AutoPlaylistContinueListeningID:
		return buildContinueListeningTrackIDs(indexEntries, state, playlistTracks), nil
	case AutoPlaylistRecentlyAddedID:
		return buildRecentPlaylistTrackIDs(indexEntries), nil
	case AutoPlaylistFavoritesID:
		return buildFavoritePlaylistTrackIDs(indexEntries), nil
	default:
		return nil, sql.ErrNoRows
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

func (s *Service) GetAutomaticPlaylists(clientID string) ([]PlaylistDto, error) {
	indexEntries, err := s.Repository.GetLibraryIndexEntries()
	if err != nil {
		return nil, err
	}

	return s.buildAutomaticPlaylists(clientID, indexEntries), nil
}

func (s *Service) GetHomeCatalog(clientID string, limit int) (MusicHomeCatalogDto, error) {
	if limit <= 0 {
		limit = defaultMusicHomeLimit
	}

	indexEntries, err := s.Repository.GetLibraryIndexEntries()
	if err != nil {
		return MusicHomeCatalogDto{}, err
	}

	artists := buildArtistGroups(indexEntries)
	albums := buildAlbumGroups(indexEntries)
	genres := buildGenreGroups(indexEntries)
	folders := buildFolderGroups(indexEntries)
	playlists := s.buildAutomaticPlaylists(clientID, indexEntries)

	if len(playlists) > limit {
		playlists = playlists[:limit]
	}
	if len(artists) > limit {
		artists = artists[:limit]
	}
	if len(albums) > limit {
		albums = albums[:limit]
	}

	return MusicHomeCatalogDto{
		Summary: MusicLibrarySummaryDto{
			TotalTracks:  len(indexEntries),
			TotalArtists: len(buildArtistGroups(indexEntries)),
			TotalAlbums:  len(buildAlbumGroups(indexEntries)),
			TotalGenres:  len(genres),
			TotalFolders: len(folders),
		},
		Playlists: playlists,
		Artists:   artists,
		Albums:    albums,
	}, nil
}

func (s *Service) GetLibraryTracks(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	tracks, err := s.Repository.GetLibraryTracks(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}
	return files.ParsePaginationToDto(&tracks)
}

func (s *Service) GetLibraryArtists(page int, pageSize int) (utils.PaginationResponse[MusicArtistGroupDto], error) {
	indexEntries, err := s.Repository.GetLibraryIndexEntries()
	if err != nil {
		return utils.PaginationResponse[MusicArtistGroupDto]{}, err
	}
	return paginateItems(buildArtistGroups(indexEntries), page, pageSize), nil
}

func (s *Service) GetLibraryAlbums(page int, pageSize int) (utils.PaginationResponse[MusicAlbumGroupDto], error) {
	indexEntries, err := s.Repository.GetLibraryIndexEntries()
	if err != nil {
		return utils.PaginationResponse[MusicAlbumGroupDto]{}, err
	}
	return paginateItems(buildAlbumGroups(indexEntries), page, pageSize), nil
}

func (s *Service) GetLibraryGenres(page int, pageSize int) (utils.PaginationResponse[MusicGenreGroupDto], error) {
	indexEntries, err := s.Repository.GetLibraryIndexEntries()
	if err != nil {
		return utils.PaginationResponse[MusicGenreGroupDto]{}, err
	}
	return paginateItems(buildGenreGroups(indexEntries), page, pageSize), nil
}

func (s *Service) GetLibraryFolders(page int, pageSize int) (utils.PaginationResponse[MusicFolderGroupDto], error) {
	indexEntries, err := s.Repository.GetLibraryIndexEntries()
	if err != nil {
		return utils.PaginationResponse[MusicFolderGroupDto]{}, err
	}
	return paginateItems(buildFolderGroups(indexEntries), page, pageSize), nil
}

func (s *Service) getLibraryTracksByEntries(entries []MusicLibraryIndexEntryModel, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	fileIDs := make([]int, 0, len(entries))
	for _, entry := range entries {
		fileIDs = append(fileIDs, entry.FileID)
	}

	paginatedIDs := paginateItems(fileIDs, page, pageSize)
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
	indexEntries, err := s.Repository.GetLibraryIndexEntries()
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}

	filtered := make([]MusicLibraryIndexEntryModel, 0)
	for _, entry := range indexEntries {
		if normalizeLookupKey(preferredArtist(entry)) == artistKey {
			filtered = append(filtered, entry)
		}
	}

	sortArtistTracks(filtered)
	return s.getLibraryTracksByEntries(filtered, page, pageSize)
}

func (s *Service) GetLibraryTracksByAlbum(albumKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	indexEntries, err := s.Repository.GetLibraryIndexEntries()
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}

	filtered := make([]MusicLibraryIndexEntryModel, 0)
	for _, entry := range indexEntries {
		key := normalizeLookupKey(fmt.Sprintf("%s::%s", preferredArtist(entry), normalizeText(entry.Album)))
		if key == albumKey {
			filtered = append(filtered, entry)
		}
	}

	sortAlbumTracks(filtered)
	return s.getLibraryTracksByEntries(filtered, page, pageSize)
}

func (s *Service) GetLibraryTracksByGenre(genreKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	indexEntries, err := s.Repository.GetLibraryIndexEntries()
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}

	filtered := make([]MusicLibraryIndexEntryModel, 0)
	for _, entry := range indexEntries {
		matchesGenre := false
		for _, genre := range normalizeGenreLabels(entry.Genre) {
			if normalizeLookupKey(genre) == genreKey {
				matchesGenre = true
				break
			}
		}
		if matchesGenre {
			filtered = append(filtered, entry)
		}
	}

	sortGenreTracks(filtered)
	return s.getLibraryTracksByEntries(filtered, page, pageSize)
}

func (s *Service) GetLibraryTracksByFolder(folderPath string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	indexEntries, err := s.Repository.GetLibraryIndexEntries()
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}

	normalizedFolder := strings.TrimSpace(folderPath)
	if normalizedFolder == "" {
		return utils.PaginationResponse[files.FileDto]{}, nil
	}

	prefix := normalizedFolder
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	filtered := make([]MusicLibraryIndexEntryModel, 0)
	for _, entry := range indexEntries {
		if entry.ParentPath == normalizedFolder || strings.HasPrefix(entry.ParentPath, prefix) {
			filtered = append(filtered, entry)
		}
	}

	sort.SliceStable(filtered, func(left int, right int) bool {
		if filtered[left].ParentPath != filtered[right].ParentPath {
			return filtered[left].ParentPath < filtered[right].ParentPath
		}
		if filtered[left].Album != filtered[right].Album {
			return filtered[left].Album < filtered[right].Album
		}
		if filtered[left].TrackNumber != filtered[right].TrackNumber {
			return filtered[left].TrackNumber < filtered[right].TrackNumber
		}

		return filtered[left].FileName < filtered[right].FileName
	})

	return s.getLibraryTracksByEntries(filtered, page, pageSize)
}
