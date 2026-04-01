package video

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"nas-go/api/internal/api/v1/video/playlist"
	"nas-go/api/pkg/utils"
)

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
