package video

import (
	"errors"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/video/playlist"
)

func TestLoadBehaviorEventsDerivesEventsFromProgress(t *testing.T) {
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	var progressLimit, skipLimit int
	var skipType string
	repo := &videoRepoMock{
		listRecentVideoWatchProgressFn: func(limit int) ([]VideoWatchProgressModel, error) {
			progressLimit = limit
			return []VideoWatchProgressModel{
				{ClientID: "tv", VideoID: 1, PositionSeconds: 100, DurationSeconds: 100, Completed: true, UpdatedAt: base.Add(time.Hour)},
			}, nil
		},
		listRecentBehaviorEventsByTypeFn: func(eventType string, limit int) ([]VideoBehaviorEventModel, error) {
			skipType, skipLimit = eventType, limit
			return []VideoBehaviorEventModel{{ClientID: "tv", VideoID: 2, EventType: "skipped", CreatedAt: base}}, nil
		},
	}
	svc := newVideoServiceForTest(t, repo)

	events := svc.loadBehaviorEvents()

	if progressLimit != 500 || skipLimit != 500 || skipType != "skipped" {
		t.Fatalf("unexpected read arguments progress=%d skip=%d type=%s", progressLimit, skipLimit, skipType)
	}
	expectedTypes := []playlist.BehaviorEventType{playlist.EventSkipped, playlist.EventStarted, playlist.EventCompleted}
	if len(events) != len(expectedTypes) {
		t.Fatalf("expected %d events, got %+v", len(expectedTypes), events)
	}
	for index, eventType := range expectedTypes {
		if events[index].EventType != eventType {
			t.Fatalf("position %d: expected %s, got %+v", index, eventType, events)
		}
	}
}

func TestLoadBehaviorEventsIgnoresReadErrors(t *testing.T) {
	progressFails := newVideoServiceForTest(t, &videoRepoMock{
		listRecentVideoWatchProgressFn: func(int) ([]VideoWatchProgressModel, error) { return nil, errors.New("boom") },
	})
	if events := progressFails.loadBehaviorEvents(); events != nil {
		t.Fatalf("expected no events on progress error, got %+v", events)
	}

	skipsFail := newVideoServiceForTest(t, &videoRepoMock{
		listRecentBehaviorEventsByTypeFn: func(string, int) ([]VideoBehaviorEventModel, error) { return nil, errors.New("boom") },
	})
	if events := skipsFail.loadBehaviorEvents(); events != nil {
		t.Fatalf("expected no events on skip error, got %+v", events)
	}
}

func TestRebuildSmartPlaylistsReadsProgressForBehavior(t *testing.T) {
	progressRead := false
	repo := &videoRepoMock{
		getAllVideosWithMetadataFn: func() ([]VideoWithMetadataModel, error) { return nil, nil },
		listRecentVideoWatchProgressFn: func(int) ([]VideoWatchProgressModel, error) {
			progressRead = true
			return nil, nil
		},
	}
	svc := newVideoServiceForTest(t, repo)

	if err := svc.RebuildSmartPlaylists(); err != nil {
		t.Fatalf("rebuild failed: %v", err)
	}
	if !progressRead {
		t.Fatalf("expected rebuild to read recent watch progress")
	}
}
