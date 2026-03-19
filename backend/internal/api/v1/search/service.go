package search

import (
	"context"
	"encoding/json"
	"log"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/prompts"
	"strings"
	"time"
)

const (
	defaultSearchLimit = 6
	maxSearchLimit     = 12
)

type Service struct {
	Repository RepositoryInterface
	AIService  ai.ServiceInterface
}

func NewService(repository RepositoryInterface, aiService ai.ServiceInterface) ServiceInterface {
	return &Service{Repository: repository, AIService: aiService}
}

const aiQueryMinWords = 2

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

	response, err := s.executeSearch(normalizedQuery, effectiveLimit, response)
	if err != nil {
		return response, err
	}

	aiKeywords, suggestion := s.expandQueryWithAI(normalizedQuery)
	if suggestion != "" {
		response.Suggestion = suggestion
	}

	if len(aiKeywords) > 0 {
		response = s.mergeAIResults(response, aiKeywords, effectiveLimit)
	}

	return response, nil
}

func (s *Service) executeSearch(query string, limit int, response GlobalSearchResponseDto) (GlobalSearchResponseDto, error) {
	files, err := s.Repository.SearchFiles(query, limit)
	if err != nil {
		return response, err
	}

	folders, err := s.Repository.SearchFolders(query, limit)
	if err != nil {
		return response, err
	}

	artists, err := s.Repository.SearchArtists(query, limit)
	if err != nil {
		return response, err
	}

	albums, err := s.Repository.SearchAlbums(query, limit)
	if err != nil {
		return response, err
	}

	musicPlaylists, err := s.Repository.SearchMusicPlaylists(query, limit)
	if err != nil {
		return response, err
	}

	videoPlaylists, err := s.Repository.SearchVideoPlaylists(query, limit)
	if err != nil {
		return response, err
	}

	videos, err := s.Repository.SearchVideos(query, limit)
	if err != nil {
		return response, err
	}

	images, err := s.Repository.SearchImages(query, limit)
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

type aiSearchExpansion struct {
	Keywords   []string `json:"keywords"`
	Suggestion string   `json:"suggestion"`
}

func (s *Service) expandQueryWithAI(query string) ([]string, string) {
	if s.AIService == nil {
		return nil, ""
	}

	words := strings.Fields(query)
	if len(words) < aiQueryMinWords {
		return nil, ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	prompt := prompts.SearchExtractionUserPrompt(query)

	resp, err := s.AIService.Execute(ctx, ai.Request{
		TaskType:     ai.TaskExtraction,
		SystemPrompt: prompts.SearchExtractionSystemPrompt(),
		Prompt:       prompt,
		MaxTokens:    150,
		Temperature:  0.1,
	})
	if err != nil {
		log.Printf("AI search expansion failed: %v\n", err)
		return nil, ""
	}

	content := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		filtered := make([]string, 0, len(lines))
		for _, line := range lines {
			if !strings.HasPrefix(strings.TrimSpace(line), "```") {
				filtered = append(filtered, line)
			}
		}
		content = strings.Join(filtered, "\n")
	}

	var expansion aiSearchExpansion
	if err := json.Unmarshal([]byte(content), &expansion); err != nil {
		log.Printf("AI search expansion parse error: %v\n", err)
		return nil, ""
	}

	return expansion.Keywords, expansion.Suggestion
}

func (s *Service) mergeAIResults(response GlobalSearchResponseDto, keywords []string, limit int) GlobalSearchResponseDto {
	existingFileIDs := make(map[int]bool)
	for _, f := range response.Files {
		existingFileIDs[f.ID] = true
	}
	existingFolderIDs := make(map[int]bool)
	for _, f := range response.Folders {
		existingFolderIDs[f.ID] = true
	}
	existingVideoIDs := make(map[int]bool)
	for _, v := range response.Videos {
		existingVideoIDs[v.ID] = true
	}
	existingImageIDs := make(map[int]bool)
	for _, i := range response.Images {
		existingImageIDs[i.ID] = true
	}

	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			continue
		}

		if files, err := s.Repository.SearchFiles(keyword, limit); err == nil {
			for _, f := range files {
				if !existingFileIDs[f.ID] {
					existingFileIDs[f.ID] = true
					response.Files = append(response.Files, mapFiles([]FileResultModel{f})...)
				}
			}
		}

		if folders, err := s.Repository.SearchFolders(keyword, limit); err == nil {
			for _, f := range folders {
				if !existingFolderIDs[f.ID] {
					existingFolderIDs[f.ID] = true
					response.Folders = append(response.Folders, mapFolders([]FolderResultModel{f})...)
				}
			}
		}

		if videos, err := s.Repository.SearchVideos(keyword, limit); err == nil {
			for _, v := range videos {
				if !existingVideoIDs[v.ID] {
					existingVideoIDs[v.ID] = true
					response.Videos = append(response.Videos, mapVideos([]VideoResultModel{v})...)
				}
			}
		}

		if images, err := s.Repository.SearchImages(keyword, limit); err == nil {
			for _, i := range images {
				if !existingImageIDs[i.ID] {
					existingImageIDs[i.ID] = true
					response.Images = append(response.Images, mapImages([]ImageResultModel{i})...)
				}
			}
		}
	}

	return response
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
			Path:       config.ToRelativePath(item.Path),
			ParentPath: config.ToRelativePath(item.ParentPath),
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
			Path:       config.ToRelativePath(item.Path),
			ParentPath: config.ToRelativePath(item.ParentPath),
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
			Path:       config.ToRelativePath(item.Path),
			ParentPath: config.ToRelativePath(item.ParentPath),
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
			Path:       config.ToRelativePath(item.Path),
			ParentPath: config.ToRelativePath(item.ParentPath),
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
