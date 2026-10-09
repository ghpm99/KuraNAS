package music

import (
	"regexp"
	"strconv"
)

const CurrentAudioTagsExtractedVersion = 2

var numberWithOptionalTotalPattern = regexp.MustCompile(`^\s*(\d{1,9})(?:\s*/\s*(\d{1,9}))?`)

// ParseNumberWithTotal reads values such as "3" or "3/12" and returns the
// number and, when present, the total. Unparseable input yields nil, nil.
func ParseNumberWithTotal(rawNumbering string) (number *int, total *int) {
	numberingMatch := numberWithOptionalTotalPattern.FindStringSubmatch(rawNumbering)
	if numberingMatch == nil {
		return nil, nil
	}
	return parseDigits(numberingMatch[1]), parseDigits(numberingMatch[2])
}

func parseDigits(digits string) *int {
	if digits == "" {
		return nil
	}
	parsedNumber, err := strconv.Atoi(digits)
	if err != nil {
		return nil
	}
	return &parsedNumber
}

// ResolveNumbering derives the numeric track/disc columns from the raw
// "N" or "N/TOTAL" tag text.
func (metadata *AudioMetadataModel) ResolveNumbering() {
	metadata.TrackNo, metadata.TrackTotal = ParseNumberWithTotal(metadata.TrackNumber)
	metadata.DiscNumber, metadata.DiscTotal = ParseNumberWithTotal(metadata.DiscNumberText)
}
