package image

import (
	"strings"
	"testing"
	"time"
)

func TestBuildLibraryListQueryKeysetBindsOnlyPlaceholders(t *testing.T) {
	takenAt := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	takenFrom := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	query, arguments, err := buildLibraryListQuery(LibraryListQuery{
		Filter: LibraryFilter{
			NameQuery:   "50%_off'; DROP TABLE home_file;--",
			Categories:  []ClassificationCategory{ClassificationCategoryPhoto},
			OnlyStarred: true,
			Formats:     []string{".jpg"},
			TakenFrom:   &takenFrom,
			Camera:      "Canon EOS",
			Folder:      "/photos",
		},
		Sort:   LibrarySortTakenAt,
		Order:  LibrarySortOrderDesc,
		Cursor: &LibraryCursor{TakenAt: &takenAt, FileID: 5},
		Limit:  61,
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if strings.Contains(query, "DROP TABLE") || strings.Contains(query, "@") {
		t.Fatalf("user text or unresolved placeholder leaked into SQL: %s", query)
	}
	for _, expected := range []string{"< ($8::timestamptz, $9)", "ORDER BY COALESCE(im.taken_at, '-infinity'::timestamptz) DESC, im.file_id DESC", "LIMIT $10", "hf.starred", "hf.deleted_at IS NULL"} {
		if !strings.Contains(query, expected) {
			t.Fatalf("expected %q in query:\n%s", expected, query)
		}
	}
	if strings.Contains(query, "OFFSET") {
		t.Fatalf("keyset query must not use OFFSET:\n%s", query)
	}
	if len(arguments) != 10 {
		t.Fatalf("expected 10 arguments, got %d: %v", len(arguments), arguments)
	}
	if arguments[1] != `%50\%\_off'; drop table home\_file;--%` {
		t.Fatalf("unexpected name pattern %v", arguments[1])
	}
}

func TestBuildLibraryListQueryOffsetOrderings(t *testing.T) {
	testCases := []struct {
		sort          LibrarySort
		order         LibrarySortOrder
		expectedOrder string
	}{
		{LibrarySortTakenAt, LibrarySortOrderAsc, "ORDER BY im.taken_at ASC NULLS LAST, im.file_id ASC"},
		{LibrarySortName, LibrarySortOrderAsc, "ORDER BY hf.name ASC, hf.id DESC"},
		{LibrarySortName, LibrarySortOrderDesc, "ORDER BY hf.name DESC, hf.id DESC"},
		{LibrarySortSize, LibrarySortOrderAsc, "ORDER BY hf.size ASC, hf.id DESC"},
		{LibrarySortSize, LibrarySortOrderDesc, "ORDER BY hf.size DESC, hf.id DESC"},
	}
	for _, testCase := range testCases {
		query, arguments, err := buildLibraryListQuery(LibraryListQuery{Sort: testCase.sort, Order: testCase.order, Limit: 10, Offset: 20})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		if !strings.Contains(query, testCase.expectedOrder) || !strings.Contains(query, "LIMIT $2 OFFSET $3") {
			t.Fatalf("unexpected query for %s/%s:\n%s", testCase.sort, testCase.order, query)
		}
		if len(arguments) != 3 || arguments[1] != 10 || arguments[2] != 20 {
			t.Fatalf("unexpected arguments %v", arguments)
		}
	}
}

func TestBuildLibraryListQueryUndatedCursorAndSeek(t *testing.T) {
	seekBefore := time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC)
	query, arguments, err := buildLibraryListQuery(LibraryListQuery{
		Sort:        LibrarySortTakenAt,
		Order:       LibrarySortOrderDesc,
		Cursor:      &LibraryCursor{FileID: 3},
		TakenBefore: &seekBefore,
		Limit:       5,
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, expected := range []string{"< $2::timestamptz)", "< ($3::timestamptz, $4)", "LIMIT $5"} {
		if !strings.Contains(query, expected) {
			t.Fatalf("expected %q in query:\n%s", expected, query)
		}
	}
	if len(arguments) != 5 || arguments[2] != "-infinity" {
		t.Fatalf("unexpected arguments %v", arguments)
	}
}

func TestBuildLibraryListQueryRejectsUnknownSort(t *testing.T) {
	if _, _, err := buildLibraryListQuery(LibraryListQuery{Sort: "bogus", Order: LibrarySortOrderAsc}); err == nil {
		t.Fatal("expected error for unsupported sort")
	}
}

func TestBuildLibraryCountAndTimelineQueries(t *testing.T) {
	takenTo := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	filter := LibraryFilter{TakenTo: &takenTo}

	countQuery, countArguments := buildLibraryCountQuery(filter)
	if !strings.Contains(countQuery, "count(*)") || !strings.Contains(countQuery, "im.taken_at <= $2") || len(countArguments) != 2 {
		t.Fatalf("unexpected count query: %s %v", countQuery, countArguments)
	}

	timelineQuery, timelineArguments := buildLibraryTimelineQuery(filter)
	for _, expected := range []string{"GROUP BY 1, 2", "ORDER BY 1 DESC, 2 DESC", "im.taken_at IS NOT NULL", "im.taken_at <= $2"} {
		if !strings.Contains(timelineQuery, expected) {
			t.Fatalf("expected %q in timeline query:\n%s", expected, timelineQuery)
		}
	}
	if len(timelineArguments) != 2 {
		t.Fatalf("unexpected timeline arguments %v", timelineArguments)
	}
}

func TestBuildLibraryListQueryNewerThanListsOldestFirstAfterThePivot(t *testing.T) {
	takenAt := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	query, arguments, err := buildLibraryListQuery(LibraryListQuery{
		Sort:      LibrarySortTakenAt,
		Order:     LibrarySortOrderDesc,
		NewerThan: &LibraryCursor{TakenAt: &takenAt, FileID: 9},
		Limit:     4,
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	for _, expected := range []string{"> ($2::timestamptz, $3)", "ASC, im.file_id ASC", "LIMIT $4"} {
		if !strings.Contains(query, expected) {
			t.Fatalf("expected %q in query:\n%s", expected, query)
		}
	}
	if strings.Contains(query, "DESC") || len(arguments) != 4 || arguments[2] != 9 {
		t.Fatalf("unexpected query or arguments %v:\n%s", arguments, query)
	}
}
