package image

import (
	"strings"
	"time"
	"unicode"
)

const (
	exifTimestampLayout = "2006:01:02 15:04:05"
	earliestTakenAt     = 1900
)

// ParseExifTimestamp parses an EXIF "YYYY:MM:DD HH:MM:SS" text as UTC. It
// returns false for blank, malformed, out-of-range or pre-1900 values.
func ParseExifTimestamp(rawTimestamp string) (time.Time, bool) {
	trimmedTimestamp := strings.TrimFunc(rawTimestamp, isBlankOrNul)
	parsedTimestamp, err := time.ParseInLocation(exifTimestampLayout, trimmedTimestamp, time.UTC)
	if err != nil || parsedTimestamp.Year() < earliestTakenAt {
		return time.Time{}, false
	}
	return parsedTimestamp, true
}

// ResolveTakenAt picks the capture instant from EXIF DateTimeOriginal, then
// DateTime. It returns nil when neither is valid; the persistence query then
// falls back to the file modification time.
func ResolveTakenAt(metadata MetadataModel) *time.Time {
	for _, rawTimestamp := range []string{metadata.DateTimeOriginal, metadata.DateTime} {
		if parsedTimestamp, isValid := ParseExifTimestamp(rawTimestamp); isValid {
			return &parsedTimestamp
		}
	}
	return nil
}

func isBlankOrNul(character rune) bool {
	return character == 0 || unicode.IsSpace(character)
}
