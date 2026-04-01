package video

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"nas-go/api/internal/api/v1/video/playlist"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/prompts"
)

func (s *Service) GetHomeCatalog(clientID string, limit int) (VideoHomeCatalogDto, error) {
	const defaultLimit = 24
	const maxLimit = 100

	normalizedLimit := limit
	if normalizedLimit <= 0 {
		normalizedLimit = defaultLimit
	} else if normalizedLimit > maxLimit {
		normalizedLimit = maxLimit
	}

	fetchLimit := normalizedLimit * 4

	allVideos, err := s.Repository.GetCatalogVideos(fetchLimit)
	if err != nil {
		return VideoHomeCatalogDto{}, err
	}
	recentVideos, err := s.Repository.GetRecentVideos(normalizedLimit)
	if err != nil {
		return VideoHomeCatalogDto{}, err
	}

	state, _ := s.Repository.GetPlaybackState(clientID)

	// Usar o classifier do engine para categorizar
	classifier := s.PlaylistEngine.Classifier
	series := make([]VideoCatalogItemDto, 0, normalizedLimit)
	movies := make([]VideoCatalogItemDto, 0, normalizedLimit)
	personal := make([]VideoCatalogItemDto, 0, normalizedLimit)

	for _, video := range allVideos {
		item := s.toCatalogItem(video, state)
		entry := videoModelToEntry(video)
		classified := classifier.Classify(entry)

		switch classified.Classification {
		case playlist.ClassSeries, playlist.ClassAnime:
			if len(series) < normalizedLimit {
				series = append(series, item)
			}
		case playlist.ClassMovie:
			if len(movies) < normalizedLimit {
				movies = append(movies, item)
			}
		default:
			if classified.Classification != playlist.ClassProgram && len(personal) < normalizedLimit {
				personal = append(personal, item)
			}
		}
	}

	recent := make([]VideoCatalogItemDto, 0, len(recentVideos))
	for _, video := range recentVideos {
		recent = append(recent, s.toCatalogItem(video, state))
	}

	continueWatching := []VideoCatalogItemDto{}
	if state.VideoID.Valid {
		for _, video := range allVideos {
			if video.ID == int(state.VideoID.Int64) {
				item := s.toCatalogItem(video, state)
				if item.Status == "in_progress" {
					continueWatching = append(continueWatching, item)
				}
				break
			}
		}
	}

	catalog := VideoHomeCatalogDto{
		Sections: []VideoCatalogSectionDto{
			{Key: "continue", Title: "Continue assistindo", Items: continueWatching},
			{Key: "series", Title: "Series", Items: series},
			{Key: "movies", Title: "Filmes", Items: movies},
			{Key: "personal", Title: "Videos pessoais", Items: personal},
			{Key: "recent", Title: "Adicionados recentemente", Items: recent},
		},
	}

	s.enrichCatalogDescriptions(&catalog)

	return catalog, nil
}

func (s *Service) enrichCatalogDescriptions(catalog *VideoHomeCatalogDto) {
	if s.AIService == nil {
		return
	}

	var parts []string
	for _, section := range catalog.Sections {
		if len(section.Items) == 0 {
			continue
		}
		names := make([]string, 0, min(len(section.Items), 5))
		for i, item := range section.Items {
			if i >= 5 {
				break
			}
			names = append(names, item.Video.Name)
		}
		parts = append(parts, fmt.Sprintf("Section '%s' (%d items): %s", section.Title, len(section.Items), strings.Join(names, ", ")))
	}

	if len(parts) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	prompt := prompts.VideoCatalogDescriptionsUserPrompt(strings.Join(parts, "\n"))

	resp, err := s.AIService.Execute(ctx, ai.Request{
		TaskType:     ai.TaskGeneration,
		SystemPrompt: prompts.VideoCatalogDescriptionsSystemPrompt(),
		Prompt:       prompt,
		MaxTokens:    300,
		Temperature:  0.3,
	})
	if err != nil {
		log.Printf("AI catalog descriptions failed: %v\n", err)
		return
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

	var descriptions map[string]string
	if err := json.Unmarshal([]byte(content), &descriptions); err != nil {
		log.Printf("AI catalog descriptions parse error: %v\n", err)
		return
	}

	for i := range catalog.Sections {
		if desc, ok := descriptions[catalog.Sections[i].Key]; ok {
			catalog.Sections[i].Description = desc
		}
	}
}
func (s *Service) toCatalogItem(video VideoFileModel, state VideoPlaybackStateModel) VideoCatalogItemDto {
	status := "not_started"
	progressPct := 0.0
	if state.VideoID.Valid && int(state.VideoID.Int64) == video.ID {
		if state.Completed {
			status = "completed"
			progressPct = 100
		} else if state.CurrentTime > 0 {
			status = "in_progress"
			if state.Duration > 0 {
				progressPct = (state.CurrentTime / state.Duration) * 100
			}
		}
	}

	if progressPct < 0 {
		progressPct = 0
	}
	if progressPct > 100 {
		progressPct = 100
	}

	return VideoCatalogItemDto{
		Video:       video.ToDto(),
		Status:      status,
		ProgressPct: progressPct,
	}
}
