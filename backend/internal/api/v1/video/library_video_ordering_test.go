package video

import "testing"

func TestParseLibraryVideoOrderingAppliesDefaultsAndRejectsUnknownValues(t *testing.T) {
	testCases := []struct {
		rawSort      string
		rawDirection string
		expected     LibraryVideoOrdering
		isValid      bool
	}{
		{"", "", LibraryVideoOrdering{Sort: LibraryVideoSortRecent, Direction: SortDirectionDescending}, true},
		{"name", "", LibraryVideoOrdering{Sort: LibraryVideoSortName, Direction: SortDirectionAscending}, true},
		{"size", "", LibraryVideoOrdering{Sort: LibraryVideoSortSize, Direction: SortDirectionDescending}, true},
		{"duration", "", LibraryVideoOrdering{Sort: LibraryVideoSortDuration, Direction: SortDirectionDescending}, true},
		{"recent", "asc", LibraryVideoOrdering{Sort: LibraryVideoSortRecent, Direction: SortDirectionAscending}, true},
		{"name", "desc", LibraryVideoOrdering{Sort: LibraryVideoSortName, Direction: SortDirectionDescending}, true},
		{"bogus", "", LibraryVideoOrdering{}, false},
		{"name", "sideways", LibraryVideoOrdering{}, false},
	}
	for _, testCase := range testCases {
		ordering, isValid := ParseLibraryVideoOrdering(testCase.rawSort, testCase.rawDirection)
		if isValid != testCase.isValid {
			t.Fatalf("sort=%q order=%q: want valid=%v, got %v", testCase.rawSort, testCase.rawDirection, testCase.isValid, isValid)
		}
		if isValid && ordering != testCase.expected {
			t.Fatalf("sort=%q order=%q: want %+v, got %+v", testCase.rawSort, testCase.rawDirection, testCase.expected, ordering)
		}
	}
}
