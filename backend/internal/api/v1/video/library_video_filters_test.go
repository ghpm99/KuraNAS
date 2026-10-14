package video

import (
	"reflect"
	"testing"
)

func TestParseLibraryVideoFiltersAcceptsEmptyAsNoFilter(t *testing.T) {
	filters, isValid := ParseLibraryVideoFilters("", "", "")
	if !isValid {
		t.Fatal("expected valid")
	}
	if !reflect.DeepEqual(filters, LibraryVideoFilters{}) {
		t.Fatalf("unexpected filters: %+v", filters)
	}
}

func TestParseLibraryVideoFiltersNormalizesFormats(t *testing.T) {
	filters, isValid := ParseLibraryVideoFilters("MP4, .mkv,mp4", "", "")
	if !isValid {
		t.Fatal("expected valid")
	}
	if !reflect.DeepEqual(filters.Formats, []string{".mp4", ".mkv"}) {
		t.Fatalf("unexpected formats: %v", filters.Formats)
	}
}

func TestParseLibraryVideoFiltersMapsResolutionBuckets(t *testing.T) {
	testCases := []struct {
		resolution string
		minHeight  *int
		maxHeight  *int
	}{
		{"sd", nil, intPointer(720)},
		{"hd", intPointer(720), intPointer(1080)},
		{"fhd", intPointer(1080), intPointer(2160)},
		{"uhd", intPointer(2160), nil},
	}
	for _, testCase := range testCases {
		filters, isValid := ParseLibraryVideoFilters("", testCase.resolution, "")
		if !isValid {
			t.Fatalf("%s: expected valid", testCase.resolution)
		}
		if !reflect.DeepEqual(filters.MinHeight, testCase.minHeight) || !reflect.DeepEqual(filters.MaxHeightExclusive, testCase.maxHeight) {
			t.Fatalf("%s: unexpected bounds %+v", testCase.resolution, filters)
		}
	}
}

func TestParseLibraryVideoFiltersMapsDurationBuckets(t *testing.T) {
	testCases := []struct {
		duration   string
		minSeconds *float64
		maxSeconds *float64
	}{
		{"short", nil, floatPointer(1200)},
		{"medium", floatPointer(1200), floatPointer(3600)},
		{"long", floatPointer(3600), nil},
	}
	for _, testCase := range testCases {
		filters, isValid := ParseLibraryVideoFilters("", "", testCase.duration)
		if !isValid {
			t.Fatalf("%s: expected valid", testCase.duration)
		}
		if !reflect.DeepEqual(filters.MinDurationSeconds, testCase.minSeconds) || !reflect.DeepEqual(filters.MaxDurationSeconds, testCase.maxSeconds) {
			t.Fatalf("%s: unexpected bounds %+v", testCase.duration, filters)
		}
	}
}

func TestParseLibraryVideoFiltersRejectsInvalidValues(t *testing.T) {
	testCases := [][3]string{
		{"mp3", "", ""},
		{"mp4,exe", "", ""},
		{"mp4,", "", ""},
		{"", "4k", ""},
		{"", "HD", ""},
		{"", "", "forever"},
	}
	for _, testCase := range testCases {
		if _, isValid := ParseLibraryVideoFilters(testCase[0], testCase[1], testCase[2]); isValid {
			t.Fatalf("%v: expected invalid", testCase)
		}
	}
}
