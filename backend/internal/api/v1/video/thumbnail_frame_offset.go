package video

import (
	"database/sql"
	"math"
)

const (
	unknownDurationFrameOffsetSeconds = 3.0
	minimumFrameOffsetSeconds         = 1.0
	maximumFrameOffsetSeconds         = 300.0
	frameOffsetDurationFraction       = 0.1
	previewClipSeconds                = 2.5
)

func isUsableDuration(duration sql.NullFloat64) bool {
	return duration.Valid && !math.IsNaN(duration.Float64) && !math.IsInf(duration.Float64, 0)
}

func thumbnailFrameOffsetSeconds(duration sql.NullFloat64) float64 {
	if !isUsableDuration(duration) {
		return unknownDurationFrameOffsetSeconds
	}
	if duration.Float64 < minimumFrameOffsetSeconds {
		return 0
	}
	proportionalOffset := duration.Float64 * frameOffsetDurationFraction
	return math.Min(math.Max(proportionalOffset, minimumFrameOffsetSeconds), maximumFrameOffsetSeconds)
}

func previewFrameOffsetSeconds(duration sql.NullFloat64) float64 {
	frameOffset := thumbnailFrameOffsetSeconds(duration)
	if !isUsableDuration(duration) {
		return frameOffset
	}
	latestOffsetKeepingClipInside := duration.Float64 - previewClipSeconds
	if latestOffsetKeepingClipInside <= 0 {
		return 0
	}
	return math.Min(frameOffset, latestOffsetKeepingClipInside)
}
