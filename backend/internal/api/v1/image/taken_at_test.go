package image

import (
	"testing"
	"time"
)

func TestParseExifTimestamp(t *testing.T) {
	testCases := []struct {
		name         string
		rawTimestamp string
		isValid      bool
		expected     time.Time
	}{
		{"regular", "2021:07:15 18:30:05", true, time.Date(2021, 7, 15, 18, 30, 5, 0, time.UTC)},
		{"trailing nul and spaces", " 2021:07:15 18:30:05\x00 ", true, time.Date(2021, 7, 15, 18, 30, 5, 0, time.UTC)},
		{"blank", "", false, time.Time{}},
		{"zero date", "0000:00:00 00:00:00", false, time.Time{}},
		{"month out of range", "2021:13:15 18:30:05", false, time.Time{}},
		{"day out of range", "2021:02:31 18:30:05", false, time.Time{}},
		{"dash separated", "2021-07-15 18:30:05", false, time.Time{}},
		{"before 1900", "1800:01:01 00:00:00", false, time.Time{}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			parsedTimestamp, isValid := ParseExifTimestamp(testCase.rawTimestamp)
			if isValid != testCase.isValid {
				t.Fatalf("isValid = %v, want %v", isValid, testCase.isValid)
			}
			if isValid && !parsedTimestamp.Equal(testCase.expected) {
				t.Fatalf("parsed = %v, want %v", parsedTimestamp, testCase.expected)
			}
		})
	}
}

func TestResolveTakenAtPrefersOriginalThenDateTimeThenNil(t *testing.T) {
	original := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	fallback := time.Date(2021, 1, 2, 3, 4, 5, 0, time.UTC)

	withBoth := ResolveTakenAt(MetadataModel{DateTimeOriginal: "2020:01:02 03:04:05", DateTime: "2021:01:02 03:04:05"})
	if withBoth == nil || !withBoth.Equal(original) {
		t.Fatalf("expected original %v, got %v", original, withBoth)
	}

	withInvalidOriginal := ResolveTakenAt(MetadataModel{DateTimeOriginal: "0000:00:00 00:00:00", DateTime: "2021:01:02 03:04:05"})
	if withInvalidOriginal == nil || !withInvalidOriginal.Equal(fallback) {
		t.Fatalf("expected datetime fallback %v, got %v", fallback, withInvalidOriginal)
	}

	if ResolveTakenAt(MetadataModel{}) != nil {
		t.Fatal("expected nil when no EXIF timestamp is valid")
	}
}
