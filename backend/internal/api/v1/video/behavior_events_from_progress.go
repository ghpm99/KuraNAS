package video

import (
	"sort"

	"nas-go/api/internal/api/v1/video/playlist"
)

const behaviorEventsReadLimit = 500

func buildBehaviorEventsFromProgress(
	watchProgress []VideoWatchProgressModel,
	skipEvents []VideoBehaviorEventModel,
) []playlist.BehaviorEvent {
	events := make([]playlist.BehaviorEvent, 0, len(watchProgress)*2+len(skipEvents))

	for _, progress := range watchProgress {
		events = append(events, eventsFromProgress(progress)...)
	}
	for _, skip := range skipEvents {
		events = append(events, playlist.BehaviorEvent{
			ClientID:       skip.ClientID,
			VideoID:        skip.VideoID,
			PlaylistID:     skip.PlaylistID,
			EventType:      playlist.EventSkipped,
			Position:       skip.Position,
			Duration:       skip.Duration,
			WatchedPercent: skip.WatchedPct,
			Timestamp:      skip.CreatedAt,
		})
	}

	sort.SliceStable(events, func(left, right int) bool {
		return events[left].Timestamp.Before(events[right].Timestamp)
	})
	return events
}

func eventsFromProgress(progress VideoWatchProgressModel) []playlist.BehaviorEvent {
	hasStarted := progress.PositionSeconds > 0 || progress.Completed
	if !hasStarted {
		return nil
	}

	finalEventType := playlist.EventAbandoned
	if progress.Completed {
		finalEventType = playlist.EventCompleted
	}

	return []playlist.BehaviorEvent{
		progressEvent(progress, playlist.EventStarted, 0),
		progressEvent(progress, finalEventType, watchedPercentOf(progress)),
	}
}

func progressEvent(progress VideoWatchProgressModel, eventType playlist.BehaviorEventType, watchedPercent float64) playlist.BehaviorEvent {
	return playlist.BehaviorEvent{
		ClientID:       progress.ClientID,
		VideoID:        progress.VideoID,
		EventType:      eventType,
		Position:       progress.PositionSeconds,
		Duration:       progress.DurationSeconds,
		WatchedPercent: watchedPercent,
		Timestamp:      progress.UpdatedAt,
	}
}

func watchedPercentOf(progress VideoWatchProgressModel) float64 {
	if progress.DurationSeconds <= 0 {
		return 0
	}
	percent := progress.PositionSeconds / progress.DurationSeconds * 100
	if percent > 100 {
		return 100
	}
	return percent
}
