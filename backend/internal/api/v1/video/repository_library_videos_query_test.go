package video

import (
	"os"
	"strings"
	"testing"
)

func TestLibraryVideosQueryForDefaultOrderingMatchesOriginalQueryText(t *testing.T) {
	originalQuery, err := os.ReadFile("testdata/library_videos_recent_desc.sql")
	if err != nil {
		t.Fatalf("read golden query: %v", err)
	}

	composedQuery, err := libraryVideosQueryFor(LibraryVideoOrdering{Sort: LibraryVideoSortRecent, Direction: SortDirectionDescending})
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	if composedQuery != string(originalQuery) {
		t.Fatalf("default query drifted from the original:\n%s", composedQuery)
	}
}

func TestLibraryVideosQueryForComposesEveryWhitelistedOrderingWithoutPlaceholdersBeyondPaging(t *testing.T) {
	for ordering := range libraryVideosOrderFragments {
		composedQuery, err := libraryVideosQueryFor(ordering)
		if err != nil {
			t.Fatalf("%+v: %v", ordering, err)
		}
		if strings.Contains(composedQuery, "$5") {
			t.Fatalf("%+v: unexpected extra placeholder", ordering)
		}
		if strings.Count(composedQuery, "ORDER BY") != 1 {
			t.Fatalf("%+v: expected exactly one ORDER BY", ordering)
		}
	}
}

func TestLibraryVideosQueryForRejectsOrderingOutsideTheWhitelist(t *testing.T) {
	if _, err := libraryVideosQueryFor(LibraryVideoOrdering{Sort: "name; DROP TABLE home_file", Direction: SortDirectionAscending}); err == nil {
		t.Fatal("expected an error for a non-whitelisted ordering")
	}
}
