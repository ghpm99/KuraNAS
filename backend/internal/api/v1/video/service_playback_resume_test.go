package video

import (
	"database/sql"
	"errors"
	"testing"
)

func TestStartPlaybackResumesEachVideoFromItsOwnProgress(t *testing.T) {
	playbackState := VideoPlaybackStateModel{}
	progressByVideoID := map[int]VideoWatchProgressModel{}

	repo := &videoRepoMock{
		getVideoFileByIDFn: func(id int) (VideoFileModel, error) {
			return VideoFileModel{ID: id, ParentPath: "/series"}, nil
		},
		getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
			return VideoPlaylistModel{ID: id}, nil
		},
		checkVideoInPlaylistFn: func(playlistID int, videoID int) (bool, error) { return true, nil },
		getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
			if playbackState.ClientID == "" {
				return VideoPlaybackStateModel{}, sql.ErrNoRows
			}
			return playbackState, nil
		},
		upsertPlaybackStateFn: func(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error) {
			playbackState = state
			return state, nil
		},
		getVideoWatchProgressFn: func(clientID string, videoID int) (VideoWatchProgressModel, error) {
			progress, isStored := progressByVideoID[videoID]
			if !isStored {
				return VideoWatchProgressModel{}, sql.ErrNoRows
			}
			return progress, nil
		},
		upsertVideoWatchProgressFn: func(tx *sql.Tx, progress VideoWatchProgressModel) (VideoWatchProgressModel, error) {
			progressByVideoID[progress.VideoID] = progress
			return progress, nil
		},
	}
	service := newVideoServiceForTest(t, repo)
	playlistID := 7

	if _, err := service.StartPlayback("tv", 1, &playlistID); err != nil {
		t.Fatalf("start video 1: %v", err)
	}
	if _, err := service.UpdatePlaybackState("tv", UpdatePlaybackStateRequest{CurrentTime: ptrFloat(120), Duration: ptrFloat(600)}); err != nil {
		t.Fatalf("update video 1: %v", err)
	}

	if _, err := service.StartPlayback("tv", 2, &playlistID); err != nil {
		t.Fatalf("start video 2: %v", err)
	}
	if _, err := service.UpdatePlaybackState("tv", UpdatePlaybackStateRequest{CurrentTime: ptrFloat(45), Duration: ptrFloat(300)}); err != nil {
		t.Fatalf("update video 2: %v", err)
	}

	resumedFirst, err := service.StartPlayback("tv", 1, &playlistID)
	if err != nil {
		t.Fatalf("resume video 1: %v", err)
	}
	if resumedFirst.PlaybackState.CurrentTime != 120 || resumedFirst.PlaybackState.Duration != 600 {
		t.Fatalf("expected video 1 to resume at 120/600, got %v/%v", resumedFirst.PlaybackState.CurrentTime, resumedFirst.PlaybackState.Duration)
	}

	resumedSecond, err := service.StartPlayback("tv", 2, &playlistID)
	if err != nil {
		t.Fatalf("resume video 2: %v", err)
	}
	if resumedSecond.PlaybackState.CurrentTime != 45 || resumedSecond.PlaybackState.Duration != 300 {
		t.Fatalf("expected video 2 to resume at 45/300, got %v/%v", resumedSecond.PlaybackState.CurrentTime, resumedSecond.PlaybackState.Duration)
	}

	neverWatched, err := service.StartPlayback("tv", 3, &playlistID)
	if err != nil {
		t.Fatalf("start video 3: %v", err)
	}
	if neverWatched.PlaybackState.CurrentTime != 0 {
		t.Fatalf("expected unwatched video to start at 0, got %v", neverWatched.PlaybackState.CurrentTime)
	}
}

func newShiftPlaybackRepoForTest(progressByVideoID map[int]VideoWatchProgressModel, progressErr error, currentVideoID int) *videoRepoMock {
	playbackState := VideoPlaybackStateModel{
		ID:          1,
		ClientID:    "tv",
		PlaylistID:  sql.NullInt64{Int64: 7, Valid: true},
		VideoID:     sql.NullInt64{Int64: int64(currentVideoID), Valid: true},
		CurrentTime: 99,
		Duration:    100,
	}
	return &videoRepoMock{
		getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) { return playbackState, nil },
		getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) {
			return VideoPlaylistModel{ID: id}, nil
		},
		getVideoPlaylistItemsFn: func(playlistID int) ([]VideoPlaylistItemModel, error) {
			return []VideoPlaylistItemModel{{VideoID: 1}, {VideoID: 2}, {VideoID: 3}}, nil
		},
		upsertPlaybackStateFn: func(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error) {
			playbackState = state
			return state, nil
		},
		touchPlaylistFn: func(tx *sql.Tx, playlistID int) error { return nil },
		getVideoWatchProgressFn: func(clientID string, videoID int) (VideoWatchProgressModel, error) {
			if progressErr != nil {
				return VideoWatchProgressModel{}, progressErr
			}
			progress, isStored := progressByVideoID[videoID]
			if !isStored {
				return VideoWatchProgressModel{}, sql.ErrNoRows
			}
			return progress, nil
		},
	}
}

func TestShiftPlaybackResumesDestinationVideoFromStoredProgress(t *testing.T) {
	storedProgress := map[int]VideoWatchProgressModel{
		1: {ClientID: "tv", VideoID: 1, PositionSeconds: 30, DurationSeconds: 90},
		3: {ClientID: "tv", VideoID: 3, PositionSeconds: 60, DurationSeconds: 120},
	}

	nextSession, err := newVideoServiceForTest(t, newShiftPlaybackRepoForTest(storedProgress, nil, 2)).NextVideo("tv")
	if err != nil {
		t.Fatalf("next video: %v", err)
	}
	if nextSession.PlaybackState.CurrentTime != 60 || nextSession.PlaybackState.Duration != 120 {
		t.Fatalf("expected next video to resume at 60/120, got %v/%v", nextSession.PlaybackState.CurrentTime, nextSession.PlaybackState.Duration)
	}

	previousSession, err := newVideoServiceForTest(t, newShiftPlaybackRepoForTest(storedProgress, nil, 2)).PreviousVideo("tv")
	if err != nil {
		t.Fatalf("previous video: %v", err)
	}
	if previousSession.PlaybackState.CurrentTime != 30 || previousSession.PlaybackState.Duration != 90 {
		t.Fatalf("expected previous video to resume at 30/90, got %v/%v", previousSession.PlaybackState.CurrentTime, previousSession.PlaybackState.Duration)
	}
}

func TestShiftPlaybackStartsAtZeroWhenDestinationHasNoStoredProgress(t *testing.T) {
	session, err := newVideoServiceForTest(t, newShiftPlaybackRepoForTest(map[int]VideoWatchProgressModel{}, nil, 1)).NextVideo("tv")
	if err != nil {
		t.Fatalf("next video: %v", err)
	}
	if session.PlaybackState.CurrentTime != 0 || session.PlaybackState.Duration != 0 {
		t.Fatalf("expected 0/0, got %v/%v", session.PlaybackState.CurrentTime, session.PlaybackState.Duration)
	}
}

func TestShiftPlaybackPropagatesStoredProgressError(t *testing.T) {
	progressErr := errors.New("progress lookup failed")
	_, err := newVideoServiceForTest(t, newShiftPlaybackRepoForTest(nil, progressErr, 1)).NextVideo("tv")
	if !errors.Is(err, progressErr) {
		t.Fatalf("expected progress error, got %v", err)
	}
}

func TestStateRestartedWhenFinished(t *testing.T) {
	cases := []struct {
		name         string
		state        VideoPlaybackStateModel
		expectedTime float64
	}{
		{"completed flag", VideoPlaybackStateModel{CurrentTime: 100, Duration: 6000, Completed: true}, 0},
		{"last 5 percent", VideoPlaybackStateModel{CurrentTime: 5800, Duration: 6000}, 0},
		{"under 15 seconds from end", VideoPlaybackStateModel{CurrentTime: 590, Duration: 600}, 0},
		{"mid video", VideoPlaybackStateModel{CurrentTime: 3000, Duration: 6000}, 3000},
		{"unknown duration", VideoPlaybackStateModel{CurrentTime: 50}, 50},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			restarted := stateRestartedWhenFinished(testCase.state)
			if restarted.CurrentTime != testCase.expectedTime {
				t.Fatalf("expected position %v, got %v", testCase.expectedTime, restarted.CurrentTime)
			}
			if testCase.expectedTime == 0 && restarted.Completed {
				t.Fatalf("expected completed=false after restart")
			}
			if restarted.Duration != testCase.state.Duration {
				t.Fatalf("duration must be preserved")
			}
		})
	}
}

func TestShiftPlaybackRestartsFinishedDestinationAtZero(t *testing.T) {
	storedProgress := map[int]VideoWatchProgressModel{
		3: {ClientID: "tv", VideoID: 3, PositionSeconds: 120, DurationSeconds: 120, Completed: true},
		1: {ClientID: "tv", VideoID: 1, PositionSeconds: 118, DurationSeconds: 120},
	}

	nextSession, err := newVideoServiceForTest(t, newShiftPlaybackRepoForTest(storedProgress, nil, 2)).NextVideo("tv")
	if err != nil {
		t.Fatalf("next video: %v", err)
	}
	if nextSession.PlaybackState.CurrentTime != 0 || nextSession.PlaybackState.Completed {
		t.Fatalf("expected next to restart at 0 uncompleted, got %v/%v", nextSession.PlaybackState.CurrentTime, nextSession.PlaybackState.Completed)
	}

	previousSession, err := newVideoServiceForTest(t, newShiftPlaybackRepoForTest(storedProgress, nil, 2)).PreviousVideo("tv")
	if err != nil {
		t.Fatalf("previous video: %v", err)
	}
	if previousSession.PlaybackState.CurrentTime != 0 {
		t.Fatalf("expected previous near end to restart at 0, got %v", previousSession.PlaybackState.CurrentTime)
	}
}

func TestStartPlaybackSameVideoRestartsWhenCurrentStateIsFinished(t *testing.T) {
	playbackState := VideoPlaybackStateModel{
		ID:          1,
		ClientID:    "tv",
		PlaylistID:  sql.NullInt64{Int64: 7, Valid: true},
		VideoID:     sql.NullInt64{Int64: 1, Valid: true},
		CurrentTime: 600,
		Duration:    600,
		Completed:   true,
	}
	repo := &videoRepoMock{
		getVideoFileByIDFn:     func(id int) (VideoFileModel, error) { return VideoFileModel{ID: id}, nil },
		getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) { return VideoPlaylistModel{ID: id}, nil },
		checkVideoInPlaylistFn: func(playlistID int, videoID int) (bool, error) { return true, nil },
		getPlaybackStateFn:     func(clientID string) (VideoPlaybackStateModel, error) { return playbackState, nil },
		upsertPlaybackStateFn: func(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error) {
			return state, nil
		},
	}
	playlistID := 7

	session, err := newVideoServiceForTest(t, repo).StartPlayback("tv", 1, &playlistID)
	if err != nil {
		t.Fatalf("start playback: %v", err)
	}
	if session.PlaybackState.CurrentTime != 0 || session.PlaybackState.Completed {
		t.Fatalf("expected restart at 0 uncompleted, got %v/%v", session.PlaybackState.CurrentTime, session.PlaybackState.Completed)
	}
}

func TestStartPlaybackRestartsCompletedStoredProgressAtZero(t *testing.T) {
	progressByVideoID := map[int]VideoWatchProgressModel{
		1: {ClientID: "tv", VideoID: 1, PositionSeconds: 600, DurationSeconds: 600, Completed: true},
	}
	repo := &videoRepoMock{
		getVideoFileByIDFn:     func(id int) (VideoFileModel, error) { return VideoFileModel{ID: id}, nil },
		getVideoPlaylistByIDFn: func(id int) (VideoPlaylistModel, error) { return VideoPlaylistModel{ID: id}, nil },
		checkVideoInPlaylistFn: func(playlistID int, videoID int) (bool, error) { return true, nil },
		getPlaybackStateFn: func(clientID string) (VideoPlaybackStateModel, error) {
			return VideoPlaybackStateModel{}, sql.ErrNoRows
		},
		upsertPlaybackStateFn: func(tx *sql.Tx, state VideoPlaybackStateModel) (VideoPlaybackStateModel, error) {
			return state, nil
		},
		getVideoWatchProgressFn: func(clientID string, videoID int) (VideoWatchProgressModel, error) {
			return progressByVideoID[videoID], nil
		},
	}
	playlistID := 7

	session, err := newVideoServiceForTest(t, repo).StartPlayback("tv", 1, &playlistID)
	if err != nil {
		t.Fatalf("start playback: %v", err)
	}
	if session.PlaybackState.CurrentTime != 0 || session.PlaybackState.Completed || session.PlaybackState.Duration != 600 {
		t.Fatalf("unexpected state %+v", session.PlaybackState)
	}
}
