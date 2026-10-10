package video

import (
	"database/sql"
	"errors"
	"testing"
)

func newWatchedServiceForTest(t *testing.T, storedProgress *VideoWatchProgressModel, saved *VideoWatchProgressModel) *Service {
	t.Helper()
	return newVideoServiceForTest(t, &videoRepoMock{
		getVideoFileByIDFn: func(id int) (VideoFileModel, error) {
			if id != 9 {
				return VideoFileModel{}, sql.ErrNoRows
			}
			return VideoFileModel{ID: 9}, nil
		},
		getVideoWatchProgressFn: func(string, int) (VideoWatchProgressModel, error) {
			if storedProgress == nil {
				return VideoWatchProgressModel{}, sql.ErrNoRows
			}
			return *storedProgress, nil
		},
		upsertVideoWatchProgressFn: func(_ *sql.Tx, progress VideoWatchProgressModel) (VideoWatchProgressModel, error) {
			*saved = progress
			return progress, nil
		},
	})
}

func TestSetVideoWatchedMarksCompletedAtKnownDuration(t *testing.T) {
	var saved VideoWatchProgressModel
	service := newWatchedServiceForTest(t, &VideoWatchProgressModel{PositionSeconds: 10, DurationSeconds: 120}, &saved)
	if err := service.SetVideoWatched("tv", 9, true); err != nil {
		t.Fatalf("SetVideoWatched: %v", err)
	}
	if !saved.Completed || saved.PositionSeconds != 120 || saved.DurationSeconds != 120 || saved.ClientID != "tv" || saved.VideoID != 9 {
		t.Fatalf("unexpected saved progress %+v", saved)
	}
}

func TestSetVideoWatchedWithoutKnownDurationStoresZeroPosition(t *testing.T) {
	var saved VideoWatchProgressModel
	service := newWatchedServiceForTest(t, nil, &saved)
	if err := service.SetVideoWatched("tv", 9, true); err != nil {
		t.Fatalf("SetVideoWatched: %v", err)
	}
	if !saved.Completed || saved.PositionSeconds != 0 || saved.DurationSeconds != 0 {
		t.Fatalf("unexpected saved progress %+v", saved)
	}
}

func TestSetVideoWatchedFalseResetsPositionAndCompleted(t *testing.T) {
	var saved VideoWatchProgressModel
	service := newWatchedServiceForTest(t, &VideoWatchProgressModel{PositionSeconds: 120, DurationSeconds: 120, Completed: true}, &saved)
	if err := service.SetVideoWatched("tv", 9, false); err != nil {
		t.Fatalf("SetVideoWatched: %v", err)
	}
	if saved.Completed || saved.PositionSeconds != 0 || saved.DurationSeconds != 120 {
		t.Fatalf("unexpected saved progress %+v", saved)
	}
}

func TestSetVideoWatchedReportsUnknownVideoAndStorageFailure(t *testing.T) {
	var saved VideoWatchProgressModel
	service := newWatchedServiceForTest(t, nil, &saved)
	if err := service.SetVideoWatched("tv", 404, true); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	failing := newVideoServiceForTest(t, &videoRepoMock{
		getVideoFileByIDFn: func(int) (VideoFileModel, error) { return VideoFileModel{ID: 9}, nil },
		getVideoWatchProgressFn: func(string, int) (VideoWatchProgressModel, error) {
			return VideoWatchProgressModel{}, errors.New("boom")
		},
	})
	if err := failing.SetVideoWatched("tv", 9, true); err == nil {
		t.Fatalf("expected storage error")
	}
}

func TestPlaylistProgressReflectsStoredWatchProgress(t *testing.T) {
	service := newVideoServiceForTest(t, &videoRepoMock{
		getVideoWatchProgressByVideosFn: func(clientID string, videoIDs []int) ([]VideoWatchProgressModel, error) {
			return []VideoWatchProgressModel{
				{VideoID: 1, PositionSeconds: 100, DurationSeconds: 100, Completed: true},
				{VideoID: 2, PositionSeconds: 25, DurationSeconds: 100},
			}, nil
		},
	})
	items := []VideoPlaylistItemModel{{VideoID: 1}, {VideoID: 2}, {VideoID: 3}}
	progressByVideo := service.buildPlaylistProgress("tv", items)
	if progressByVideo[1].Status != "completed" || progressByVideo[2].Status != "in_progress" || progressByVideo[2].ProgressPct != 25 || progressByVideo[3].Status != "not_started" {
		t.Fatalf("unexpected progress %+v", progressByVideo)
	}
}
