package video

import (
	"database/sql"
	"errors"
)

func (s *Service) SetVideoWatched(clientID string, videoID int, isWatched bool) error {
	if _, err := s.Repository.GetVideoFileByID(videoID); err != nil {
		return err
	}

	knownDurationSeconds, err := s.knownDurationSeconds(clientID, videoID)
	if err != nil {
		return err
	}

	progress := VideoWatchProgressModel{
		ClientID:        clientID,
		VideoID:         videoID,
		DurationSeconds: knownDurationSeconds,
		Completed:       isWatched,
	}
	if isWatched {
		progress.PositionSeconds = knownDurationSeconds
	}

	return s.withTransaction(func(tx *sql.Tx) error {
		_, upsertErr := s.Repository.UpsertVideoWatchProgress(tx, progress)
		return upsertErr
	})
}

func (s *Service) SetPlaylistWatched(clientID string, playlistID int, season *int, isWatched bool) (int, error) {
	targetPlaylist, err := s.Repository.GetVideoPlaylistByID(playlistID)
	if err != nil {
		return 0, err
	}
	if !isAutomaticSeries(targetPlaylist) {
		return 0, ErrPlaylistNotAutomaticSeries
	}

	updatedCount := 0
	err = s.withTransaction(func(tx *sql.Tx) error {
		var updateErr error
		updatedCount, updateErr = s.Repository.SetPlaylistWatched(tx, clientID, playlistID, season, isWatched)
		return updateErr
	})
	return updatedCount, err
}

func (s *Service) knownDurationSeconds(clientID string, videoID int) (float64, error) {
	storedProgress, err := s.Repository.GetVideoWatchProgress(clientID, videoID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return storedProgress.DurationSeconds, nil
}

func (s *Service) overrideWithStoredWatchProgress(clientID string, items []VideoPlaylistItemModel, progressByVideo map[int]videoItemProgress) {
	videoIDs := make([]int, 0, len(items))
	for _, item := range items {
		videoIDs = append(videoIDs, item.VideoID)
	}

	storedProgressList, err := s.Repository.GetVideoWatchProgressByVideos(clientID, videoIDs)
	if err != nil {
		return
	}
	for _, storedProgress := range storedProgressList {
		progressByVideo[storedProgress.VideoID] = progressFromStoredWatchProgress(storedProgress)
	}
}

func progressFromStoredWatchProgress(storedProgress VideoWatchProgressModel) videoItemProgress {
	return playlistProgressFromState(VideoPlaybackStateModel{
		CurrentTime: storedProgress.PositionSeconds,
		Duration:    storedProgress.DurationSeconds,
		Completed:   storedProgress.Completed,
	})
}
