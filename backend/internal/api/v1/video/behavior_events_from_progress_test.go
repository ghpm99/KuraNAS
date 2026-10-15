package video

import (
	"testing"
	"time"

	"nas-go/api/internal/api/v1/video/playlist"
)

func TestBuildBehaviorEventsFromProgress(t *testing.T) {
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	t.Run("completed progress yields started and completed", func(t *testing.T) {
		events := buildBehaviorEventsFromProgress([]VideoWatchProgressModel{
			{ClientID: "tv", VideoID: 1, PositionSeconds: 95, DurationSeconds: 100, Completed: true, UpdatedAt: base},
		}, nil)

		if len(events) != 2 || events[0].EventType != playlist.EventStarted || events[1].EventType != playlist.EventCompleted {
			t.Fatalf("expected started+completed, got %+v", events)
		}
		if events[1].Duration != 100 || events[1].Position != 95 || events[1].WatchedPercent != 95 {
			t.Fatalf("unexpected completed event %+v", events[1])
		}
		if events[0].Duration != 100 || events[0].Position != 95 || !events[0].Timestamp.Equal(base) {
			t.Fatalf("unexpected started event %+v", events[0])
		}
	})

	t.Run("partial progress yields started and abandoned", func(t *testing.T) {
		events := buildBehaviorEventsFromProgress([]VideoWatchProgressModel{
			{ClientID: "tv", VideoID: 2, PositionSeconds: 25, DurationSeconds: 100, UpdatedAt: base},
		}, nil)

		if len(events) != 2 || events[0].EventType != playlist.EventStarted || events[1].EventType != playlist.EventAbandoned {
			t.Fatalf("expected started+abandoned, got %+v", events)
		}
		if events[1].Position != 25 || events[1].Duration != 100 || events[1].WatchedPercent != 25 {
			t.Fatalf("unexpected abandoned event %+v", events[1])
		}
	})

	t.Run("unstarted progress yields nothing", func(t *testing.T) {
		events := buildBehaviorEventsFromProgress([]VideoWatchProgressModel{
			{ClientID: "tv", VideoID: 3, PositionSeconds: 0, DurationSeconds: 100, UpdatedAt: base},
		}, nil)

		if len(events) != 0 {
			t.Fatalf("expected no events, got %+v", events)
		}
	})

	t.Run("watched percent is capped and safe without duration", func(t *testing.T) {
		events := buildBehaviorEventsFromProgress([]VideoWatchProgressModel{
			{VideoID: 4, PositionSeconds: 150, DurationSeconds: 100, Completed: true, UpdatedAt: base},
			{VideoID: 5, PositionSeconds: 10, DurationSeconds: 0, UpdatedAt: base},
		}, nil)

		if events[1].WatchedPercent != 100 || events[3].WatchedPercent != 0 {
			t.Fatalf("unexpected percents %+v", events)
		}
	})

	t.Run("skip event becomes skipped", func(t *testing.T) {
		events := buildBehaviorEventsFromProgress(nil, []VideoBehaviorEventModel{
			{ClientID: "tv", VideoID: 7, PlaylistID: 9, EventType: "skipped", Position: 5, Duration: 50, WatchedPct: 10, CreatedAt: base},
		})

		if len(events) != 1 {
			t.Fatalf("expected one event, got %+v", events)
		}
		expected := playlist.BehaviorEvent{ClientID: "tv", VideoID: 7, PlaylistID: 9, EventType: playlist.EventSkipped, Position: 5, Duration: 50, WatchedPercent: 10, Timestamp: base}
		if events[0] != expected {
			t.Fatalf("expected %+v, got %+v", expected, events[0])
		}
	})

	t.Run("progress and skips are merged in ascending timestamp order", func(t *testing.T) {
		events := buildBehaviorEventsFromProgress(
			[]VideoWatchProgressModel{
				{VideoID: 1, PositionSeconds: 10, DurationSeconds: 100, UpdatedAt: base.Add(3 * time.Hour)},
				{VideoID: 2, PositionSeconds: 10, DurationSeconds: 100, Completed: true, UpdatedAt: base.Add(1 * time.Hour)},
			},
			[]VideoBehaviorEventModel{
				{VideoID: 3, EventType: "skipped", CreatedAt: base.Add(2 * time.Hour)},
				{VideoID: 4, EventType: "skipped", CreatedAt: base},
			},
		)

		expectedVideoOrder := []int{4, 2, 2, 3, 1, 1}
		if len(events) != len(expectedVideoOrder) {
			t.Fatalf("expected %d events, got %+v", len(expectedVideoOrder), events)
		}
		for index, videoID := range expectedVideoOrder {
			if events[index].VideoID != videoID {
				t.Fatalf("position %d: expected video %d, got %+v", index, videoID, events)
			}
		}
		if events[1].EventType != playlist.EventStarted || events[2].EventType != playlist.EventCompleted {
			t.Fatalf("expected started before completed for same row, got %+v", events)
		}
	})
}
