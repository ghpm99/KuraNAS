package files

import (
	"strings"
	"testing"
	"time"

	queries "nas-go/api/pkg/database/queries/files"
)

func buildForTest(t *testing.T, query FileSearchQuery) (string, []any) {
	t.Helper()
	statement, arguments, err := buildSearchFilesQuery(query)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return statement, arguments
}

func TestBuildSearchFilesQueryWithoutFiltersUsesOnlyNameTermsAndRelevance(t *testing.T) {
	statement, arguments := buildForTest(t, FileSearchQuery{Query: "Foo bar", Page: 1, PageSize: 10})

	if !strings.Contains(statement, "lower(hf.name) LIKE $1") || !strings.Contains(statement, "LIKE ALL ($2::text[])") {
		t.Fatalf("expected driving + all-terms name match, got %s", statement)
	}
	if !strings.Contains(statement, "WHEN lower(hf.name) = $3 THEN 0") {
		t.Fatalf("expected relevance ordering, got %s", statement)
	}
	if strings.Contains(statement, "@") {
		t.Fatalf("unbound placeholder left in %s", statement)
	}
	if len(arguments) != 7 || arguments[0] != "%foo%" {
		t.Fatalf("unexpected arguments %v", arguments)
	}
	if arguments[2] != "foo bar" || arguments[3] != "foo bar%" || arguments[4] != "%foo bar%" {
		t.Fatalf("unexpected relevance arguments %v", arguments[2:5])
	}
	if arguments[5] != 11 || arguments[6] != 0 {
		t.Fatalf("unexpected page arguments %v", arguments[5:])
	}
}

func TestBuildSearchFilesQueryAddsScopeFragments(t *testing.T) {
	descendants, _ := buildForTest(t, FileSearchQuery{Query: "ab", Scope: SearchScopeDescendants, ScopePath: "/srv/", Page: 1, PageSize: 10})
	if !strings.Contains(descendants, "hf.path ^@ $3") {
		t.Fatalf("expected descendants fragment, got %s", descendants)
	}
	children, _ := buildForTest(t, FileSearchQuery{Query: "ab", Scope: SearchScopeChildren, ScopePath: "/srv", Page: 1, PageSize: 10})
	if !strings.Contains(children, "hf.parent_path = $3") {
		t.Fatalf("expected children fragment, got %s", children)
	}
}

func TestBuildSearchFilesQueryAddsEachFilterFragment(t *testing.T) {
	modifiedFrom := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	modifiedTo := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	minSize, maxSize := int64(10), int64(20)
	tests := []struct {
		name     string
		filter   FileSearchFilter
		fragment string
	}{
		{"folder kind", FileSearchFilter{Kinds: []FileSearchKind{SearchKindFolder}}, queries.SearchFilterKindFolderQuery},
		{"format kind", FileSearchFilter{Kinds: []FileSearchKind{SearchKindImage}}, "hf.type = 2 AND hf.format = ANY($3)"},
		{"other kind", FileSearchFilter{Kinds: []FileSearchKind{SearchKindOther}}, "hf.type = 2 AND NOT (hf.format = ANY($3))"},
		{"modified from", FileSearchFilter{ModifiedFrom: &modifiedFrom}, "hf.updated_at >= $3"},
		{"modified to", FileSearchFilter{ModifiedTo: &modifiedTo}, "hf.updated_at < $3"},
		{"min size", FileSearchFilter{MinSize: &minSize}, "hf.size >= $3"},
		{"max size", FileSearchFilter{MaxSize: &maxSize}, "hf.size <= $3"},
		{"hot tier", FileSearchFilter{Tier: TierHot}, "hf.physical_path IS NULL"},
		{"cold tier", FileSearchFilter{Tier: TierCold}, "hf.physical_path IS NOT NULL"},
		{"starred", FileSearchFilter{OnlyStarred: true}, "AND hf.starred\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			statement, _ := buildForTest(t, FileSearchQuery{Query: "ab", Filter: tc.filter, Page: 1, PageSize: 10})
			if !strings.Contains(statement, strings.TrimSpace(tc.fragment)) {
				t.Fatalf("expected %q in %s", tc.fragment, statement)
			}
		})
	}
}

func TestBuildSearchFilesQueryJoinsMultipleKindsWithOr(t *testing.T) {
	statement, arguments := buildForTest(t, FileSearchQuery{
		Query:    "ab",
		Filter:   FileSearchFilter{Kinds: []FileSearchKind{SearchKindFolder, SearchKindImage, SearchKindOther}},
		Page:     1,
		PageSize: 10,
	})
	expected := "(hf.type = 1 OR hf.type = 2 AND hf.format = ANY($3) OR hf.type = 2 AND NOT (hf.format = ANY($4)))"
	if !strings.Contains(statement, expected) {
		t.Fatalf("expected %q in %s", expected, statement)
	}
	if len(arguments) != 9 {
		t.Fatalf("expected 9 arguments, got %d", len(arguments))
	}
}

func TestBuildSearchFilesQueryPicksOrderFragment(t *testing.T) {
	tests := []struct {
		sort     FileSearchSort
		order    FileSearchOrder
		fragment string
	}{
		{SearchSortName, "", queries.SearchOrderNameAscQuery},
		{SearchSortName, SearchOrderDescending, queries.SearchOrderNameDescQuery},
		{SearchSortSize, "", queries.SearchOrderSizeDescQuery},
		{SearchSortSize, SearchOrderAscending, queries.SearchOrderSizeAscQuery},
		{SearchSortModified, "", queries.SearchOrderModifiedDescQuery},
		{SearchSortModified, SearchOrderAscending, queries.SearchOrderModifiedAscQuery},
	}
	for _, tc := range tests {
		statement, _ := buildForTest(t, FileSearchQuery{Query: "ab", Filter: FileSearchFilter{Sort: tc.sort, Order: tc.order}, Page: 1, PageSize: 10})
		if !strings.Contains(statement, strings.TrimSpace(tc.fragment)) {
			t.Fatalf("sort %s/%s: expected %q in %s", tc.sort, tc.order, tc.fragment, statement)
		}
	}
}

func TestBuildSearchFilesQueryRejectsUnknownSortAndEmptyTerms(t *testing.T) {
	if _, _, err := buildSearchFilesQuery(FileSearchQuery{Query: "ab", Filter: FileSearchFilter{Sort: "color"}}); err == nil {
		t.Fatalf("expected an error for unknown sort")
	}
	if _, _, err := buildSearchFilesQuery(FileSearchQuery{Query: "  "}); err == nil {
		t.Fatalf("expected an error for empty terms")
	}
}
