package search

import (
	"context"
	"encoding/json"
	"log"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/prompts"
	"strings"
	"time"
)

const (
	aiQueryMinWords    = 2
	aiMaxKeywords      = 3
	aiExpansionTimeout = 10 * time.Second
)

type aiSearchExpansion struct {
	Keywords   []string `json:"keywords"`
	Suggestion string   `json:"suggestion"`
}

func (s *Service) SearchGlobalWithAI(query string, limit int) (GlobalSearchResponseDto, error) {
	response, err := s.SearchGlobal(query, limit)
	if err != nil || response.Query == "" {
		return response, err
	}

	expansion, isExpanded := s.expandQueryWithAI(response.Query)
	if !isExpanded {
		return response, nil
	}

	effectiveLimit := clampLimit(limit)
	response.Suggestion = expansion.Suggestion
	response = s.mergeAIResults(response, expansion.Keywords, effectiveLimit)
	return trimToLimit(response, effectiveLimit), nil
}

func (s *Service) expandQueryWithAI(query string) (aiSearchExpansion, bool) {
	if s.AIService == nil || len(strings.Fields(query)) < aiQueryMinWords {
		return aiSearchExpansion{}, false
	}

	if cached, isCached := s.expansionCache.get(query); isCached {
		return cached, true
	}

	expansion, err := s.requestAIExpansion(query)
	if err != nil {
		log.Printf("AI search expansion failed: %v\n", err)
		return aiSearchExpansion{}, false
	}

	s.expansionCache.put(query, expansion)
	return expansion, true
}

func (s *Service) requestAIExpansion(query string) (aiSearchExpansion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), aiExpansionTimeout)
	defer cancel()

	aiResponse, err := s.AIService.Execute(ctx, ai.Request{
		TaskType:     ai.TaskExtraction,
		SystemPrompt: prompts.SearchExtractionSystemPrompt(),
		Prompt:       prompts.SearchExtractionUserPrompt(query),
		MaxTokens:    150,
		Temperature:  0.1,
	})
	if err != nil {
		return aiSearchExpansion{}, err
	}

	var expansion aiSearchExpansion
	if err := json.Unmarshal([]byte(stripCodeFence(aiResponse.Content)), &expansion); err != nil {
		return aiSearchExpansion{}, err
	}

	expansion.Keywords = limitKeywords(expansion.Keywords)
	return expansion, nil
}

func stripCodeFence(content string) string {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "```") {
		return content
	}

	contentLines := make([]string, 0)
	for _, line := range strings.Split(content, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "```") {
			contentLines = append(contentLines, line)
		}
	}
	return strings.Join(contentLines, "\n")
}

func limitKeywords(keywords []string) []string {
	cleaned := make([]string, 0, aiMaxKeywords)
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			continue
		}
		cleaned = append(cleaned, keyword)
		if len(cleaned) == aiMaxKeywords {
			break
		}
	}
	return cleaned
}

func (s *Service) mergeAIResults(response GlobalSearchResponseDto, keywords []string, limit int) GlobalSearchResponseDto {
	existingFileIDs := make(map[int]bool)
	for _, file := range response.Files {
		existingFileIDs[file.ID] = true
	}
	existingFolderIDs := make(map[int]bool)
	for _, folder := range response.Folders {
		existingFolderIDs[folder.ID] = true
	}
	existingVideoIDs := make(map[int]bool)
	for _, video := range response.Videos {
		existingVideoIDs[video.ID] = true
	}
	existingImageIDs := make(map[int]bool)
	for _, image := range response.Images {
		existingImageIDs[image.ID] = true
	}

	for _, keyword := range keywords {
		if files, err := s.Repository.SearchFiles(keyword, limit); err == nil {
			for _, file := range files {
				if !existingFileIDs[file.ID] {
					existingFileIDs[file.ID] = true
					response.Files = append(response.Files, mapFiles([]FileResultModel{file})...)
				}
			}
		}

		if folders, err := s.Repository.SearchFolders(keyword, limit); err == nil {
			for _, folder := range folders {
				if !existingFolderIDs[folder.ID] {
					existingFolderIDs[folder.ID] = true
					response.Folders = append(response.Folders, mapFolders([]FolderResultModel{folder})...)
				}
			}
		}

		if videos, err := s.Repository.SearchVideos(keyword, limit); err == nil {
			for _, video := range videos {
				if !existingVideoIDs[video.ID] {
					existingVideoIDs[video.ID] = true
					response.Videos = append(response.Videos, mapVideos([]VideoResultModel{video})...)
				}
			}
		}

		if images, err := s.Repository.SearchImages(keyword, limit); err == nil {
			for _, image := range images {
				if !existingImageIDs[image.ID] {
					existingImageIDs[image.ID] = true
					response.Images = append(response.Images, mapImages([]ImageResultModel{image})...)
				}
			}
		}
	}

	return response
}

func trimToLimit(response GlobalSearchResponseDto, limit int) GlobalSearchResponseDto {
	response.Files = firstN(response.Files, limit)
	response.Folders = firstN(response.Folders, limit)
	response.Videos = firstN(response.Videos, limit)
	response.Images = firstN(response.Images, limit)
	response.Documents = firstN(response.Documents, limit)
	return response
}

func firstN[T any](items []T, limit int) []T {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}
