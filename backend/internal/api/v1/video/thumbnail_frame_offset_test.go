package video

import (
	"database/sql"
	"math"
	"testing"
)

func knownDuration(seconds float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: seconds, Valid: true}
}

func TestThumbnailFrameOffsetSeconds(t *testing.T) {
	testCases := []struct {
		name     string
		duration sql.NullFloat64
		expected float64
	}{
		{"unknown duration keeps legacy offset", sql.NullFloat64{}, 3},
		{"not a number is unknown", knownDuration(math.NaN()), 3},
		{"infinite is unknown", knownDuration(math.Inf(1)), 3},
		{"zero duration", knownDuration(0), 0},
		{"below one second", knownDuration(0.9), 0},
		{"exactly one second", knownDuration(1), 1},
		{"five seconds is raised to minimum", knownDuration(5), 1},
		{"ten seconds", knownDuration(10), 1},
		{"2999 seconds", knownDuration(2999), 299.9},
		{"3000 seconds hits ceiling", knownDuration(3000), 300},
		{"3001 seconds stays at ceiling", knownDuration(3001), 300},
		{"one hour stays at ceiling", knownDuration(3600), 300},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := thumbnailFrameOffsetSeconds(testCase.duration)
			if math.Abs(actual-testCase.expected) > 1e-9 {
				t.Fatalf("expected %v, got %v", testCase.expected, actual)
			}
		})
	}
}

func TestPreviewFrameOffsetSeconds(t *testing.T) {
	testCases := []struct {
		name     string
		duration sql.NullFloat64
		expected float64
	}{
		{"unknown duration keeps legacy offset", sql.NullFloat64{}, 3},
		{"one hour", knownDuration(3600), 300},
		{"long video away from the end", knownDuration(100), 10},
		{"clip would pass the end so offset recedes", knownDuration(3), 0.5},
		{"clip ends exactly at the end", knownDuration(3.5), 1},
		{"shorter than clip uses zero", knownDuration(2), 0},
		{"exactly clip length uses zero", knownDuration(2.5), 0},
		{"below one second uses zero", knownDuration(0.5), 0},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := previewFrameOffsetSeconds(testCase.duration)
			if math.Abs(actual-testCase.expected) > 1e-9 {
				t.Fatalf("expected %v, got %v", testCase.expected, actual)
			}
			if testCase.duration.Valid && actual+previewClipSeconds > testCase.duration.Float64 && actual != 0 {
				t.Fatalf("preview clip passes the end: offset %v duration %v", actual, testCase.duration.Float64)
			}
		})
	}
}
