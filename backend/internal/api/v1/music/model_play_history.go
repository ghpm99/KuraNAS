package music

import "time"

type PlayedTrackModel struct {
	FileID       int
	PlayCount    int
	LastPlayedAt time.Time
}
