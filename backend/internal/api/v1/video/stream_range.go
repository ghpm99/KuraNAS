package video

import (
	"regexp"
	"strconv"
	"strings"
)

var byteRangeSpecPattern = regexp.MustCompile(`^(\d+-\d*|-\d+)$`)

type requestedRange struct {
	header          string
	isUnsatisfiable bool
}

func parseRequestedRange(rangeHeader string, fileSize int64) requestedRange {
	rangeSet, hasBytesUnit := strings.CutPrefix(strings.TrimSpace(rangeHeader), "bytes=")
	if !hasBytesUnit {
		return requestedRange{}
	}
	firstSpec, _, _ := strings.Cut(rangeSet, ",")
	firstSpec = strings.TrimSpace(firstSpec)
	if !byteRangeSpecPattern.MatchString(firstSpec) {
		return requestedRange{}
	}

	startText, endText, _ := strings.Cut(firstSpec, "-")
	if startText == "" {
		suffixLength, err := strconv.ParseInt(endText, 10, 64)
		if err != nil {
			return requestedRange{}
		}
		return requestedRange{header: "bytes=" + firstSpec, isUnsatisfiable: suffixLength == 0}
	}

	start, err := strconv.ParseInt(startText, 10, 64)
	if err != nil {
		return requestedRange{}
	}
	if endText != "" {
		end, err := strconv.ParseInt(endText, 10, 64)
		if err != nil || end < start {
			return requestedRange{}
		}
	}
	return requestedRange{header: "bytes=" + firstSpec, isUnsatisfiable: start >= fileSize}
}
