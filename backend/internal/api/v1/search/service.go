package search

import (
	"strings"
)

const (
	defaultSearchLimit = 6
	maxSearchLimit     = 12
)

type Service struct {
	Repository RepositoryInterface
}

func NewService(repository RepositoryInterface) ServiceInterface {
	return &Service{Repository: repository}
}

func (s *Service) SearchGlobal(query string, limit int) (GlobalSearchResponseDto, error) {
	normalizedQuery := strings.TrimSpace(query)
	response := GlobalSearchResponseDto{
		Query:     normalizedQuery,
		Files:     []FileResultDto{},
		Folders:   []FolderResultDto{},
		Artists:   []ArtistResultDto{},
		Albums:    []AlbumResultDto{},
		Playlists: []PlaylistResultDto{},
		Videos:    []VideoResultDto{},
		Images:    []ImageResultDto{},
	}

	if normalizedQuery == "" {
		return response, nil
	}

	effectiveLimit := clampLimit(limit)

	files, err := s.Repository.SearchFiles(normalizedQuery, effectiveLimit)
	if err != nil {
		return response, err
	}

	folders, err := s.Repository.SearchFolders(normalizedQuery, effectiveLimit)
	if err != nil {
		return response, err
	}

	artists, err := s.Repository.SearchArtists(normalizedQuery, effectiveLimit)
	if err != nil {
		return response, err
	}

	albums, err := s.Repository.SearchAlbums(normalizedQuery, effectiveLimit)
	if err != nil {
		return response, err
	}

	musicPlaylists, err := s.Repository.SearchMusicPlaylists(normalizedQuery, effectiveLimit)
	if err != nil {
		return response, err
	}

	videoPlaylists, err := s.Repository.SearchVideoPlaylists(normalizedQuery, effectiveLimit)
	if err != nil {
		return response, err
	}

	videos, err := s.Repository.SearchVideos(normalizedQuery, effectiveLimit)
	if err != nil {
		return response, err
	}

	images, err := s.Repository.SearchImages(normalizedQuery, effectiveLimit)
	if err != nil {
		return response, err
	}

	response.Files = mapFiles(files)
	response.Folders = mapFolders(folders)
	response.Artists = mapArtists(artists)
	response.Albums = mapAlbums(albums)
	response.Playlists = append(mapMusicPlaylists(musicPlaylists), mapVideoPlaylists(videoPlaylists)...)
	response.Videos = mapVideos(videos)
	response.Images = mapImages(images)

	return response, nil
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultSearchLimit
	}
	if limit > maxSearchLimit {
		return maxSearchLimit
	}
	return limit
}

func mapFiles(items []FileResultModel) []FileResultDto {
	results := make([]FileResultDto, 0, len(items))
	for _, item := range items {
		results = append(results, FileResultDto{
			ID:         item.ID,
			Name:       item.Name,
			Path:       item.Path,
			ParentPath: item.ParentPath,
			Format:     item.Format,
			Starred:    item.Starred,
		})
	}
	return results
}

func mapFolders(items []FolderResultModel) []FolderResultDto {
	results := make([]FolderResultDto, 0, len(items))
	for _, item := range items {
		results = append(results, FolderResultDto{
			ID:         item.ID,
			Name:       item.Name,
			Path:       item.Path,
			ParentPath: item.ParentPath,
			Starred:    item.Starred,
		})
	}
	return results
}

func mapArtists(items []ArtistResultModel) []ArtistResultDto {
	results := make([]ArtistResultDto, 0, len(items))
	for _, item := range items {
		results = append(results, ArtistResultDto{
			Key:        normalizeLookupKey(item.Artist),
			Artist:     item.Artist,
			TrackCount: item.TrackCount,
			AlbumCount: item.AlbumCount,
		})
	}
	return results
}

func mapAlbums(items []AlbumResultModel) []AlbumResultDto {
	results := make([]AlbumResultDto, 0, len(items))
	for _, item := range items {
		results = append(results, AlbumResultDto{
			Key:        normalizeLookupKey(item.Artist + "::" + item.Album),
			Artist:     item.Artist,
			Album:      item.Album,
			Year:       item.Year,
			TrackCount: item.TrackCount,
		})
	}
	return results
}

func mapMusicPlaylists(items []MusicPlaylistResultModel) []PlaylistResultDto {
	results := make([]PlaylistResultDto, 0, len(items))
	for _, item := range items {
		description := strings.TrimSpace(item.Description)
		results = append(results, PlaylistResultDto{
			Scope:       "music",
			ID:          item.ID,
			Name:        item.Name,
			Description: description,
			Count:       item.TrackCount,
			IsAuto:      item.IsSystem,
		})
	}
	return results
}

func mapVideoPlaylists(items []VideoPlaylistResultModel) []PlaylistResultDto {
	results := make([]PlaylistResultDto, 0, len(items))
	for _, item := range items {
		results = append(results, PlaylistResultDto{
			Scope:          "video",
			ID:             item.ID,
			Name:           item.Name,
			Description:    item.Type,
			Count:          item.ItemCount,
			Classification: item.Classification,
			SourcePath:     item.SourcePath,
			IsAuto:         item.IsAuto,
		})
	}
	return results
}

func mapVideos(items []VideoResultModel) []VideoResultDto {
	results := make([]VideoResultDto, 0, len(items))
	for _, item := range items {
		results = append(results, VideoResultDto{
			ID:         item.ID,
			Name:       item.Name,
			Path:       item.Path,
			ParentPath: item.ParentPath,
			Format:     item.Format,
		})
	}
	return results
}

func mapImages(items []ImageResultModel) []ImageResultDto {
	results := make([]ImageResultDto, 0, len(items))
	for _, item := range items {
		results = append(results, ImageResultDto{
			ID:         item.ID,
			Name:       item.Name,
			Path:       item.Path,
			ParentPath: item.ParentPath,
			Format:     item.Format,
			Category:   item.Category,
			Context:    item.Context,
		})
	}
	return results
}

func normalizeLookupKey(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.NewReplacer("_", " ", "-", " ", ".", " ").Replace(normalized)
	return strings.Join(strings.Fields(normalized), " ")
}
