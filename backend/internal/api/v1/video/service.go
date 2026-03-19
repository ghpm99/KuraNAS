package video

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/video/playlist"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/prompts"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	Repository     RepositoryInterface
	PlaylistEngine *playlist.PlaylistEngine
	AIService      ai.ServiceInterface
}

type videoItemProgress struct {
	Status      string
	ProgressPct float64
}

var (
	ErrVideoNotInPlaylist      = errors.New("video not in selected playlist")
	ErrPlaybackStateNotFound   = errors.New("playback state not found")
	ErrInvalidBehaviorEvent    = errors.New("invalid behavior event")
	ErrPlaylistNameRequired    = errors.New("playlist name is required")
	ErrPlaylistReorderRequired = errors.New("playlist reorder items are required")
	ErrNoVideosForContext      = errors.New("no videos found for context")
	ErrPlaybackNavigation      = errors.New("playback navigation unavailable")
	ErrPlaylistWithoutItems    = errors.New("playlist has no items")
)

func NewService(repository RepositoryInterface, aiService ai.ServiceInterface) ServiceInterface {
	return &Service{
		Repository:     repository,
		PlaylistEngine: playlist.NewPlaylistEngine(),
		AIService:      aiService,
	}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	return database.ExecOptionalTx(s.Repository.GetDbContext(), fn)
}

// ---------------------------------------------------------------------------
// Playback
// ---------------------------------------------------------------------------

func (s *Service) StartPlayback(clientID string, videoID int, playlistID *int) (PlaybackSessionDto, error) {
	videoFile, err := s.Repository.GetVideoFileByID(videoID)
	if err != nil {
		return PlaybackSessionDto{}, err
	}

	var pl VideoPlaylistModel
	if playlistID != nil {
		pl, err = s.Repository.GetVideoPlaylistByID(*playlistID)
		if err != nil {
			return PlaybackSessionDto{}, err
		}
		inPlaylist, checkErr := s.Repository.CheckVideoInPlaylist(pl.ID, videoID)
		if checkErr != nil {
			return PlaybackSessionDto{}, checkErr
		}
		if !inPlaylist {
			return PlaybackSessionDto{}, ErrVideoNotInPlaylist
		}
	} else {
		pl, err = s.ensureContextPlaylist(string(ContextFolder), videoFile.ParentPath)
		if err != nil {
			return PlaybackSessionDto{}, err
		}
	}

	state, _ := s.Repository.GetPlaybackState(clientID)
	if !state.PlaylistID.Valid || int(state.PlaylistID.Int64) != pl.ID || !state.VideoID.Valid || int(state.VideoID.Int64) != videoID {
		state = VideoPlaybackStateModel{
			ClientID:    clientID,
			CurrentTime: 0,
			Duration:    0,
			IsPaused:    false,
			Completed:   false,
		}
	}
	state.PlaylistID = sql.NullInt64{Int64: int64(pl.ID), Valid: true}
	state.VideoID = sql.NullInt64{Int64: int64(videoID), Valid: true}

	if err := s.withTransaction(func(tx *sql.Tx) error {
		updatedState, upsertErr := s.Repository.UpsertPlaybackState(tx, state)
		if upsertErr != nil {
			return upsertErr
		}
		state = updatedState

		// Emitir evento de comportamento "started"
		_, _ = s.Repository.InsertBehaviorEvent(tx, VideoBehaviorEventModel{
			ClientID:   clientID,
			VideoID:    videoID,
			PlaylistID: pl.ID,
			EventType:  string(playlist.EventStarted),
			Position:   0,
			Duration:   state.Duration,
		})

		return s.Repository.TouchPlaylist(tx, pl.ID)
	}); err != nil {
		return PlaybackSessionDto{}, err
	}

	return s.buildSession(clientID, pl, state)
}

func (s *Service) GetPlaybackState(clientID string) (PlaybackSessionDto, error) {
	state, err := s.Repository.GetPlaybackState(clientID)
	if err != nil {
		return PlaybackSessionDto{}, err
	}
	if !state.PlaylistID.Valid {
		return PlaybackSessionDto{}, ErrPlaybackStateNotFound
	}

	pl, err := s.playlistByIDAndClient(state)
	if err != nil {
		return PlaybackSessionDto{}, err
	}

	return s.buildSession(clientID, pl, state)
}

func (s *Service) UpdatePlaybackState(clientID string, req UpdatePlaybackStateRequest) (VideoPlaybackStateDto, error) {
	state, err := s.Repository.GetPlaybackState(clientID)
	if err != nil {
		state = VideoPlaybackStateModel{ClientID: clientID, IsPaused: true}
	}

	state.ClientID = clientID
	if req.PlaylistID != nil {
		state.PlaylistID = sql.NullInt64{Int64: int64(*req.PlaylistID), Valid: true}
	}
	if req.VideoID != nil {
		state.VideoID = sql.NullInt64{Int64: int64(*req.VideoID), Valid: true}
	}
	if req.CurrentTime != nil {
		state.CurrentTime = *req.CurrentTime
	}
	if req.Duration != nil {
		state.Duration = *req.Duration
	}
	if req.IsPaused != nil {
		state.IsPaused = *req.IsPaused
	}
	if req.Completed != nil {
		state.Completed = *req.Completed
	}

	if err := s.withTransaction(func(tx *sql.Tx) error {
		updatedState, upsertErr := s.Repository.UpsertPlaybackState(tx, state)
		if upsertErr != nil {
			return upsertErr
		}
		state = updatedState

		// Emitir evento de comportamento baseado no estado
		if req.Completed != nil && *req.Completed {
			videoID := 0
			if state.VideoID.Valid {
				videoID = int(state.VideoID.Int64)
			}
			playlistID := 0
			if state.PlaylistID.Valid {
				playlistID = int(state.PlaylistID.Int64)
			}
			watchedPct := 0.0
			if state.Duration > 0 {
				watchedPct = (state.CurrentTime / state.Duration) * 100
			}
			_, _ = s.Repository.InsertBehaviorEvent(tx, VideoBehaviorEventModel{
				ClientID:   clientID,
				VideoID:    videoID,
				PlaylistID: playlistID,
				EventType:  string(playlist.EventCompleted),
				Position:   state.CurrentTime,
				Duration:   state.Duration,
				WatchedPct: watchedPct,
			})
		}

		if state.PlaylistID.Valid {
			return s.Repository.TouchPlaylist(tx, int(state.PlaylistID.Int64))
		}
		return nil
	}); err != nil {
		return VideoPlaybackStateDto{}, fmt.Errorf("erro ao atualizar estado do player de video: %w", err)
	}

	state.ClientID = clientID
	return state.ToDto(), nil
}

func (s *Service) NextVideo(clientID string) (PlaybackSessionDto, error) {
	return s.shiftPlayback(clientID, 1)
}

func (s *Service) PreviousVideo(clientID string) (PlaybackSessionDto, error) {
	return s.shiftPlayback(clientID, -1)
}

// ---------------------------------------------------------------------------
// Behavior tracking
// ---------------------------------------------------------------------------

func (s *Service) TrackBehaviorEvent(clientID string, req TrackBehaviorEventRequest) error {
	validEvents := map[string]bool{
		string(playlist.EventStarted):   true,
		string(playlist.EventPaused):    true,
		string(playlist.EventResumed):   true,
		string(playlist.EventCompleted): true,
		string(playlist.EventSkipped):   true,
		string(playlist.EventAbandoned): true,
	}
	if !validEvents[req.EventType] {
		return fmt.Errorf("%w: %s", ErrInvalidBehaviorEvent, req.EventType)
	}

	watchedPct := 0.0
	if req.Duration > 0 {
		watchedPct = (req.Position / req.Duration) * 100
	}

	playlistID := 0
	if req.PlaylistID != nil {
		playlistID = *req.PlaylistID
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		_, err := s.Repository.InsertBehaviorEvent(tx, VideoBehaviorEventModel{
			ClientID:   clientID,
			VideoID:    req.VideoID,
			PlaylistID: playlistID,
			EventType:  req.EventType,
			Position:   req.Position,
			Duration:   req.Duration,
			WatchedPct: watchedPct,
		})
		return err
	})
}

// ---------------------------------------------------------------------------
// Smart playlists (powered by playlist engine)
// ---------------------------------------------------------------------------

func (s *Service) RebuildSmartPlaylists() error {
	// Buscar videos com metadados enriquecidos
	videosWithMeta, err := s.Repository.GetAllVideosWithMetadata()
	if err != nil {
		return err
	}

	// Converter para o formato do engine
	entries := make([]playlist.VideoEntry, 0, len(videosWithMeta))
	for _, v := range videosWithMeta {
		entry := playlist.VideoEntry{
			ID:         v.ID,
			Name:       v.Name,
			Path:       v.Path,
			ParentPath: v.ParentPath,
			Format:     v.Format,
			Size:       v.Size,
			CreatedAt:  v.CreatedAt,
			UpdatedAt:  v.UpdatedAt,
		}

		// Enricher com metadados se disponiveis
		if v.MetaWidth.Valid || v.MetaDuration.Valid {
			meta := &playlist.VideoMeta{}
			if v.MetaDuration.Valid {
				meta.Duration = parseDurationSeconds(v.MetaDuration.String)
			}
			if v.MetaWidth.Valid {
				meta.Width = int(v.MetaWidth.Int64)
			}
			if v.MetaHeight.Valid {
				meta.Height = int(v.MetaHeight.Int64)
			}
			if v.MetaFrameRate.Valid {
				meta.FrameRate = v.MetaFrameRate.Float64
			}
			if v.MetaCodecName.Valid {
				meta.CodecName = v.MetaCodecName.String
			}
			if v.MetaAspectRatio.Valid {
				meta.AspectRatio = v.MetaAspectRatio.String
			}
			if v.MetaAudioChannels.Valid {
				meta.AudioChannels = int(v.MetaAudioChannels.Int64)
			}
			if v.MetaAudioCodec.Valid {
				meta.AudioCodec = v.MetaAudioCodec.String
			}
			if v.MetaAudioSampleRate.Valid {
				meta.AudioSampleRate = v.MetaAudioSampleRate.String
			}
			entry.Meta = meta
		}

		entries = append(entries, entry)
	}

	// Buscar eventos de comportamento para alimentar o engine
	behaviorEvents, _ := s.Repository.GetAllBehaviorEvents(500)
	engineBehavior := make([]playlist.BehaviorEvent, 0, len(behaviorEvents))
	for _, e := range behaviorEvents {
		engineBehavior = append(engineBehavior, playlist.BehaviorEvent{
			ClientID:       e.ClientID,
			VideoID:        e.VideoID,
			PlaylistID:     e.PlaylistID,
			EventType:      playlist.BehaviorEventType(e.EventType),
			Position:       e.Position,
			Duration:       e.Duration,
			WatchedPercent: e.WatchedPct,
			Timestamp:      e.CreatedAt,
		})
	}

	// Executar o engine
	result := s.PlaylistEngine.Build(playlist.BuildInput{
		Videos:         entries,
		BehaviorEvents: engineBehavior,
	})

	// Converter para smart groups e persistir
	groups := result.ToSmartGroups()

	return s.withTransaction(func(tx *sql.Tx) error {
		for _, group := range groups {
			pl, upsertErr := s.Repository.UpsertAutoPlaylist(
				tx,
				group.PlaylistType,
				group.SourceKey,
				group.Name,
				group.GroupMode,
				group.Classification,
			)
			if upsertErr != nil {
				return upsertErr
			}

			exclusions, exclusionsErr := s.Repository.GetPlaylistExclusions(pl.ID)
			if exclusionsErr != nil {
				return exclusionsErr
			}

			filtered := make([]int, 0, len(group.VideoIDs))
			for _, id := range group.VideoIDs {
				if !exclusions[id] {
					filtered = append(filtered, id)
				}
			}

			if err := s.Repository.DeleteAutoPlaylistItems(tx, pl.ID); err != nil {
				return err
			}
			if err := s.Repository.InsertPlaylistItemsWithSource(tx, pl.ID, filtered, "auto"); err != nil {
				return err
			}
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// Home catalog (uses classifier for better categorization)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Playlist CRUD
// ---------------------------------------------------------------------------

func (s *Service) GetPlaylists(includeHidden bool) ([]VideoPlaylistDto, error) {
	models, err := s.Repository.GetVideoPlaylists(includeHidden)
	if err != nil {
		return nil, err
	}

	result := make([]VideoPlaylistDto, 0, len(models))
	for _, model := range models {
		result = append(result, model.ToDto(nil))
	}
	return result, nil
}

func (s *Service) GetPlaylistMemberships(includeHidden bool) ([]VideoPlaylistMembershipDto, error) {
	models, err := s.Repository.GetVideoPlaylistMemberships(includeHidden)
	if err != nil {
		return nil, err
	}

	result := make([]VideoPlaylistMembershipDto, 0, len(models))
	for _, model := range models {
		result = append(result, VideoPlaylistMembershipDto{
			PlaylistID: model.PlaylistID,
			VideoID:    model.VideoID,
		})
	}

	return result, nil
}

func (s *Service) GetPlaylistByID(clientID string, id int) (VideoPlaylistDto, error) {
	pl, err := s.Repository.GetVideoPlaylistByID(id)
	if err != nil {
		return VideoPlaylistDto{}, err
	}

	items, err := s.Repository.GetVideoPlaylistItemsDetailed(id)
	if err != nil {
		return VideoPlaylistDto{}, err
	}

	progressByVideo := s.buildPlaylistProgress(clientID, items)

	itemDtos := make([]VideoPlaylistItemDto, 0, len(items))
	for _, item := range items {
		progress := progressByVideo[item.VideoID]
		itemDtos = append(itemDtos, item.ToDto(progress.Status, progress.ProgressPct))
	}
	pl.ItemCount = len(itemDtos)
	return pl.ToDto(itemDtos), nil
}

func (s *Service) SetPlaylistHidden(playlistID int, hidden bool) error {
	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.SetPlaylistHidden(tx, playlistID, hidden)
	})
}

func (s *Service) AddVideoToPlaylist(playlistID int, videoID int) error {
	return s.withTransaction(func(tx *sql.Tx) error {
		if err := s.Repository.AddPlaylistVideoManual(tx, playlistID, videoID); err != nil {
			return err
		}
		return s.Repository.DeletePlaylistExclusion(tx, playlistID, videoID)
	})
}

func (s *Service) RemoveVideoFromPlaylist(playlistID int, videoID int) error {
	pl, err := s.Repository.GetVideoPlaylistByID(playlistID)
	if err != nil {
		return err
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		if err := s.Repository.RemovePlaylistVideo(tx, playlistID, videoID); err != nil {
			return err
		}
		if pl.IsAuto {
			return s.Repository.UpsertPlaylistExclusion(tx, playlistID, videoID)
		}
		return nil
	})
}

func (s *Service) GetUnassignedVideos(limit int) ([]VideoFileDto, error) {
	if limit <= 0 {
		limit = 2000
	}
	models, err := s.Repository.GetUnassignedVideos(limit)
	if err != nil {
		return nil, err
	}
	result := make([]VideoFileDto, 0, len(models))
	for _, model := range models {
		result = append(result, model.ToDto())
	}
	return result, nil
}

func (s *Service) ListLibraryVideos(page int, pageSize int, searchQuery string) (utils.PaginationResponse[VideoFileDto], error) {
	models, err := s.Repository.ListLibraryVideos(page, pageSize, searchQuery)
	if err != nil {
		return utils.PaginationResponse[VideoFileDto]{}, err
	}

	items := make([]VideoFileDto, 0, len(models.Items))
	for _, model := range models.Items {
		items = append(items, model.ToDto())
	}

	return utils.PaginationResponse[VideoFileDto]{
		Items: items,
		Pagination: utils.Pagination{
			Page:     models.Pagination.Page,
			PageSize: models.Pagination.PageSize,
			HasNext:  models.Pagination.HasNext,
			HasPrev:  models.Pagination.HasPrev,
		},
	}, nil
}

func (s *Service) UpdatePlaylistName(playlistID int, name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ErrPlaylistNameRequired
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.UpdatePlaylistName(tx, playlistID, trimmed)
	})
}

func (s *Service) ReorderPlaylistItems(playlistID int, items []ReorderPlaylistItemRequest) error {
	if len(items) == 0 {
		return ErrPlaylistReorderRequired
	}

	seenVideo := map[int]bool{}
	seenOrder := map[int]bool{}
	for _, item := range items {
		if seenVideo[item.VideoID] {
			return fmt.Errorf("video_id duplicado na reordenacao: %d", item.VideoID)
		}
		if seenOrder[item.OrderIndex] {
			return fmt.Errorf("order_index duplicado na reordenacao: %d", item.OrderIndex)
		}
		seenVideo[item.VideoID] = true
		seenOrder[item.OrderIndex] = true
	}

	videoIDs := make([]int, len(items))
	orderIndices := make([]int, len(items))
	for i, item := range items {
		videoIDs[i] = item.VideoID
		orderIndices[i] = item.OrderIndex
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.ReorderPlaylistItems(tx, playlistID, videoIDs, orderIndices)
	})
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (s *Service) ensureContextPlaylist(contextType string, sourcePath string) (VideoPlaylistModel, error) {
	pl, err := s.Repository.GetPlaylistByContext(contextType, sourcePath)
	if err == nil {
		items, itemsErr := s.Repository.GetPlaylistItems(pl.ID)
		if itemsErr == nil && len(items) > 0 {
			return pl, nil
		}
	}

	videos, err := s.Repository.GetVideosByParentPath(sourcePath)
	if err != nil {
		return VideoPlaylistModel{}, err
	}
	if len(videos) == 0 {
		return VideoPlaylistModel{}, ErrNoVideosForContext
	}

	if pl.ID == 0 {
		err = s.withTransaction(func(tx *sql.Tx) error {
			created, createErr := s.Repository.CreatePlaylist(tx, contextType, sourcePath)
			if createErr != nil {
				return createErr
			}
			pl = created

			videoIDs := make([]int, 0, len(videos))
			for _, video := range videos {
				videoIDs = append(videoIDs, video.ID)
			}
			return s.Repository.ReplacePlaylistItems(tx, pl.ID, videoIDs)
		})
		if err != nil {
			return VideoPlaylistModel{}, err
		}
		return pl, nil
	}

	err = s.withTransaction(func(tx *sql.Tx) error {
		videoIDs := make([]int, 0, len(videos))
		for _, video := range videos {
			videoIDs = append(videoIDs, video.ID)
		}
		return s.Repository.ReplacePlaylistItems(tx, pl.ID, videoIDs)
	})
	if err != nil {
		return VideoPlaylistModel{}, err
	}
	return pl, nil
}

func (s *Service) buildSession(clientID string, pl VideoPlaylistModel, state VideoPlaybackStateModel) (PlaybackSessionDto, error) {
	items, err := s.Repository.GetPlaylistItems(pl.ID)
	if err != nil {
		return PlaybackSessionDto{}, err
	}

	itemDtos := make([]VideoPlaylistItemDto, 0, len(items))
	currentVideoID := -1
	if state.VideoID.Valid {
		currentVideoID = int(state.VideoID.Int64)
	}

	for _, item := range items {
		progress := videoItemProgress{Status: "not_started", ProgressPct: 0}
		if item.VideoID == currentVideoID {
			progress = playlistProgressFromState(state)
		}
		itemDtos = append(itemDtos, item.ToDto(progress.Status, progress.ProgressPct))
	}

	state.ClientID = clientID
	return PlaybackSessionDto{
		Playlist:      pl.ToDto(itemDtos),
		PlaybackState: state.ToDto(),
	}, nil
}

func (s *Service) shiftPlayback(clientID string, direction int) (PlaybackSessionDto, error) {
	state, err := s.Repository.GetPlaybackState(clientID)
	if err != nil {
		return PlaybackSessionDto{}, err
	}
	if !state.PlaylistID.Valid || !state.VideoID.Valid {
		return PlaybackSessionDto{}, ErrPlaybackNavigation
	}

	pl, err := s.playlistByIDAndClient(state)
	if err != nil {
		return PlaybackSessionDto{}, err
	}

	items, err := s.Repository.GetPlaylistItems(pl.ID)
	if err != nil {
		return PlaybackSessionDto{}, err
	}
	if len(items) == 0 {
		return PlaybackSessionDto{}, ErrPlaylistWithoutItems
	}

	currentIndex := -1
	for i, item := range items {
		if item.VideoID == int(state.VideoID.Int64) {
			currentIndex = i
			break
		}
	}
	if currentIndex < 0 {
		currentIndex = 0
	}

	nextIndex := currentIndex + direction
	if nextIndex < 0 {
		nextIndex = 0
	}
	if nextIndex >= len(items) {
		nextIndex = len(items) - 1
	}

	// Emitir evento de skip se avancou
	prevVideoID := int(state.VideoID.Int64)

	state.VideoID = sql.NullInt64{Int64: int64(items[nextIndex].VideoID), Valid: true}
	state.CurrentTime = 0
	state.Duration = 0
	state.IsPaused = false
	state.Completed = false

	if err := s.withTransaction(func(tx *sql.Tx) error {
		updatedState, upsertErr := s.Repository.UpsertPlaybackState(tx, state)
		if upsertErr != nil {
			return upsertErr
		}
		state = updatedState

		// Registrar skip do video anterior
		if direction > 0 {
			_, _ = s.Repository.InsertBehaviorEvent(tx, VideoBehaviorEventModel{
				ClientID:   clientID,
				VideoID:    prevVideoID,
				PlaylistID: pl.ID,
				EventType:  string(playlist.EventSkipped),
			})
		}

		return s.Repository.TouchPlaylist(tx, pl.ID)
	}); err != nil {
		return PlaybackSessionDto{}, err
	}

	return s.buildSession(clientID, pl, state)
}

func (s *Service) playlistByIDAndClient(state VideoPlaybackStateModel) (VideoPlaylistModel, error) {
	playlistID := int(state.PlaylistID.Int64)
	pl, err := s.Repository.GetVideoPlaylistByID(playlistID)
	if err != nil {
		return VideoPlaylistModel{}, err
	}
	return pl, nil
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

func (s *Service) buildPlaylistProgress(clientID string, items []VideoPlaylistItemModel) map[int]videoItemProgress {
	progressByVideo := make(map[int]videoItemProgress, len(items))

	for _, item := range items {
		progressByVideo[item.VideoID] = videoItemProgress{
			Status:      "not_started",
			ProgressPct: 0,
		}
	}

	state, err := s.Repository.GetPlaybackState(clientID)
	if err == nil && state.VideoID.Valid {
		videoID := int(state.VideoID.Int64)
		if _, ok := progressByVideo[videoID]; ok {
			progressByVideo[videoID] = playlistProgressFromState(state)
		}
	}

	events, err := s.Repository.GetBehaviorEvents(clientID, len(items)*4+8)
	if err != nil {
		return progressByVideo
	}

	for _, event := range events {
		current, exists := progressByVideo[event.VideoID]
		if !exists || current.Status != "not_started" {
			continue
		}

		progressByVideo[event.VideoID] = playlistProgressFromEvent(event)
	}

	return progressByVideo
}

func playlistProgressFromState(state VideoPlaybackStateModel) videoItemProgress {
	progress := 0.0
	status := "not_started"

	if state.Completed {
		status = "completed"
		progress = 100
	} else if state.CurrentTime > 0 {
		status = "in_progress"
		if state.Duration > 0 {
			progress = (state.CurrentTime / state.Duration) * 100
		}
	}

	return videoItemProgress{
		Status:      status,
		ProgressPct: clampProgressPct(progress),
	}
}

func playlistProgressFromEvent(event VideoBehaviorEventModel) videoItemProgress {
	progress := event.WatchedPct
	if progress <= 0 && event.Duration > 0 {
		progress = (event.Position / event.Duration) * 100
	}

	switch event.EventType {
	case string(playlist.EventCompleted):
		return videoItemProgress{Status: "completed", ProgressPct: 100}
	case string(playlist.EventStarted), string(playlist.EventPaused), string(playlist.EventResumed),
		string(playlist.EventSkipped), string(playlist.EventAbandoned):
		if progress > 0 {
			return videoItemProgress{
				Status:      "in_progress",
				ProgressPct: clampProgressPct(progress),
			}
		}
	}

	return videoItemProgress{Status: "not_started", ProgressPct: 0}
}

func clampProgressPct(progress float64) float64 {
	if progress < 0 {
		return 0
	}
	if progress > 100 {
		return 100
	}
	return progress
}

func videoModelToEntry(v VideoFileModel) playlist.VideoEntry {
	return playlist.VideoEntry{
		ID:         v.ID,
		Name:       v.Name,
		Path:       v.Path,
		ParentPath: v.ParentPath,
		Format:     v.Format,
		Size:       v.Size,
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
	}
}

// parseDurationSeconds converte duration string (ex: "3600.000000") para float64 em segundos.
func parseDurationSeconds(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}
