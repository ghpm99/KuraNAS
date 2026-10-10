package video

import (
	"math"

	"nas-go/api/internal/api/v1/video/playlist"
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

	continueWatching, err := s.continueWatchingCatalogItems(clientID, normalizedLimit)
	if err != nil {
		return VideoHomeCatalogDto{}, err
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

	return catalog, nil
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

func (s *Service) GetContinueWatching(clientID string, limit int) ([]ContinueWatchingItemDto, error) {
	models, err := s.Repository.GetContinueWatchingVideos(clientID, limit)
	if err != nil {
		return nil, err
	}

	items := make([]ContinueWatchingItemDto, 0, len(models))
	for _, model := range models {
		items = append(items, model.ToDto())
	}
	return items, nil
}

func (s *Service) continueWatchingCatalogItems(clientID string, limit int) ([]VideoCatalogItemDto, error) {
	models, err := s.Repository.GetContinueWatchingVideos(clientID, limit)
	if err != nil {
		return nil, err
	}

	items := make([]VideoCatalogItemDto, 0, len(models))
	for _, model := range models {
		items = append(items, VideoCatalogItemDto{
			Video:       model.VideoFileModel.ToDto(),
			Status:      "in_progress",
			ProgressPct: progressPercentage(model.PositionSeconds, model.DurationSeconds),
		})
	}
	return items, nil
}

func progressPercentage(positionSeconds float64, durationSeconds float64) float64 {
	if durationSeconds <= 0 {
		return 0
	}
	return math.Min(100, math.Max(0, positionSeconds/durationSeconds*100))
}
