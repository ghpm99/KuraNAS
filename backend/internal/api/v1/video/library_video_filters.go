package video

import (
	"slices"
	"strings"

	"nas-go/api/pkg/utils"
)

type LibraryVideoResolution string

type LibraryVideoDuration string

const (
	LibraryVideoResolutionSD  LibraryVideoResolution = "sd"
	LibraryVideoResolutionHD  LibraryVideoResolution = "hd"
	LibraryVideoResolutionFHD LibraryVideoResolution = "fhd"
	LibraryVideoResolutionUHD LibraryVideoResolution = "uhd"
)

const (
	LibraryVideoDurationShort  LibraryVideoDuration = "short"
	LibraryVideoDurationMedium LibraryVideoDuration = "medium"
	LibraryVideoDurationLong   LibraryVideoDuration = "long"
)

const (
	hdMinimumHeight  = 720
	fhdMinimumHeight = 1080
	uhdMinimumHeight = 2160

	mediumMinimumDurationSeconds = 20 * 60
	longMinimumDurationSeconds   = 60 * 60
)

type LibraryVideoFilters struct {
	Formats            []string
	Resolution         LibraryVideoResolution
	Duration           LibraryVideoDuration
	MinHeight          *int
	MaxHeightExclusive *int
	MinDurationSeconds *float64
	MaxDurationSeconds *float64
}

func ParseLibraryVideoFilters(rawFormats string, rawResolution string, rawDuration string) (LibraryVideoFilters, bool) {
	formats, areFormatsValid := parseLibraryVideoFormats(rawFormats)
	if !areFormatsValid {
		return LibraryVideoFilters{}, false
	}

	filters := LibraryVideoFilters{Formats: formats}
	if !applyResolutionBucket(&filters, LibraryVideoResolution(rawResolution)) {
		return LibraryVideoFilters{}, false
	}
	if !applyDurationBucket(&filters, LibraryVideoDuration(rawDuration)) {
		return LibraryVideoFilters{}, false
	}
	return filters, true
}

func parseLibraryVideoFormats(rawFormats string) ([]string, bool) {
	if strings.TrimSpace(rawFormats) == "" {
		return nil, true
	}

	formats := []string{}
	for _, rawFormat := range strings.Split(rawFormats, ",") {
		format := strings.ToLower(strings.TrimSpace(rawFormat))
		if !strings.HasPrefix(format, ".") {
			format = "." + format
		}
		if !slices.Contains(utils.VideoFormats, format) {
			return nil, false
		}
		if !slices.Contains(formats, format) {
			formats = append(formats, format)
		}
	}
	return formats, true
}

func applyResolutionBucket(filters *LibraryVideoFilters, resolution LibraryVideoResolution) bool {
	switch resolution {
	case "":
		return true
	case LibraryVideoResolutionSD:
		filters.MaxHeightExclusive = intPointer(hdMinimumHeight)
	case LibraryVideoResolutionHD:
		filters.MinHeight = intPointer(hdMinimumHeight)
		filters.MaxHeightExclusive = intPointer(fhdMinimumHeight)
	case LibraryVideoResolutionFHD:
		filters.MinHeight = intPointer(fhdMinimumHeight)
		filters.MaxHeightExclusive = intPointer(uhdMinimumHeight)
	case LibraryVideoResolutionUHD:
		filters.MinHeight = intPointer(uhdMinimumHeight)
	default:
		return false
	}
	filters.Resolution = resolution
	return true
}

func applyDurationBucket(filters *LibraryVideoFilters, duration LibraryVideoDuration) bool {
	switch duration {
	case "":
		return true
	case LibraryVideoDurationShort:
		filters.MaxDurationSeconds = floatPointer(mediumMinimumDurationSeconds)
	case LibraryVideoDurationMedium:
		filters.MinDurationSeconds = floatPointer(mediumMinimumDurationSeconds)
		filters.MaxDurationSeconds = floatPointer(longMinimumDurationSeconds)
	case LibraryVideoDurationLong:
		filters.MinDurationSeconds = floatPointer(longMinimumDurationSeconds)
	default:
		return false
	}
	filters.Duration = duration
	return true
}

func intPointer(number int) *int {
	return &number
}

func floatPointer(number float64) *float64 {
	return &number
}
