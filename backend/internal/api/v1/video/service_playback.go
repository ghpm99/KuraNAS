package video

import (
	"database/sql"
	"fmt"

	"nas-go/api/internal/api/v1/video/playlist"
)

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
