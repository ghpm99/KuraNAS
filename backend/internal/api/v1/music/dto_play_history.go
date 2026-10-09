package music

import (
	"errors"
	"time"

	"nas-go/api/internal/api/v1/files"
)

const maxPlayedSeconds = 24 * 60 * 60

const mostPlayedRecentPeriod = 30 * 24 * time.Hour

var ErrInvalidPlayRequest = errors.New("invalid play request")

type PlayPeriod string

const (
	PlayPeriodAll        PlayPeriod = "all"
	PlayPeriodLast30Days PlayPeriod = "30d"
)

type RecordPlayRequest struct {
	FileID        int `json:"file_id"`
	PlayedSeconds int `json:"played_seconds"`
}

type MusicPlayedTrackDto struct {
	Track        files.FileDto `json:"track"`
	PlayCount    int           `json:"play_count"`
	LastPlayedAt time.Time     `json:"last_played_at"`
}

func (request RecordPlayRequest) isValid() bool {
	return request.FileID > 0 && request.PlayedSeconds >= 0 && request.PlayedSeconds <= maxPlayedSeconds
}

func (period PlayPeriod) isValid() bool {
	return period == PlayPeriodAll || period == PlayPeriodLast30Days
}

func (period PlayPeriod) earliestPlayedAt(now time.Time) *time.Time {
	if period != PlayPeriodLast30Days {
		return nil
	}
	earliest := now.Add(-mostPlayedRecentPeriod)
	return &earliest
}
