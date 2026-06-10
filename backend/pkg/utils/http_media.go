package utils

import (
	"mime"
	"strconv"
	"strings"
)

// ParseHTTPRange parses a single HTTP Range header value ("bytes=start-end")
// against the given file size, returning the resolved [start, end] interval.
// Only the first range of a multi-range header is honored.
func ParseHTTPRange(rangeHeader string, fileSize int64) (int64, int64, bool) {
	if fileSize <= 0 {
		return 0, 0, false
	}

	parts := strings.SplitN(strings.TrimSpace(rangeHeader), "=", 2)
	if len(parts) != 2 || parts[0] != "bytes" {
		return 0, 0, false
	}

	rangeValue := strings.TrimSpace(parts[1])
	if rangeValue == "" {
		return 0, 0, false
	}

	// Only first range is supported.
	if commaIndex := strings.Index(rangeValue, ","); commaIndex >= 0 {
		rangeValue = strings.TrimSpace(rangeValue[:commaIndex])
	}

	bounds := strings.SplitN(rangeValue, "-", 2)
	if len(bounds) != 2 {
		return 0, 0, false
	}

	startText := strings.TrimSpace(bounds[0])
	endText := strings.TrimSpace(bounds[1])

	var start int64
	var end int64
	var err error

	if startText == "" {
		// Suffix byte range: bytes=-500
		suffixLength, parseErr := strconv.ParseInt(endText, 10, 64)
		if parseErr != nil || suffixLength <= 0 {
			return 0, 0, false
		}
		if suffixLength > fileSize {
			suffixLength = fileSize
		}
		start = fileSize - suffixLength
		end = fileSize - 1
	} else {
		start, err = strconv.ParseInt(startText, 10, 64)
		if err != nil || start < 0 || start >= fileSize {
			return 0, 0, false
		}

		if endText == "" {
			// Open ended range: bytes=500-
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(endText, 10, 64)
			if err != nil {
				return 0, 0, false
			}
			if end >= fileSize {
				end = fileSize - 1
			}
		}
	}

	if end < start {
		return 0, 0, false
	}

	return start, end, true
}

// ContentTypeByFormat resolves a MIME content type from a file extension
// (with or without the leading dot), falling back when unknown.
func ContentTypeByFormat(format string, fallback string) string {
	ext := strings.TrimSpace(format)
	if ext == "" {
		return fallback
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	contentType := mime.TypeByExtension(strings.ToLower(ext))
	if contentType == "" {
		return fallback
	}
	return contentType
}
