package video

import (
	"database/sql"
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
