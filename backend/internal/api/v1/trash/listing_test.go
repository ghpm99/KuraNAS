package trash

import "testing"

func TestParseListing(t *testing.T) {
	testCases := []struct {
		name            string
		rawSortKey      string
		rawDirection    string
		rawSearchText   string
		expectedListing Listing
		isValid         bool
	}{
		{"absent parameters use the default", "", "", "", DefaultListing, true},
		{"known key and direction", "size", "asc", "", Listing{SortKey: SortBySize, Direction: SortAscending}, true},
		{"key alone keeps the default direction", "name", "", "", Listing{SortKey: SortByName, Direction: SortDescending}, true},
		{"search text is trimmed and lowercased", "", "", "  Fotos ", Listing{SortKey: SortByDeletedAt, Direction: SortDescending, SearchText: "fotos"}, true},
		{"unknown key is rejected", "color", "asc", "", Listing{SortKey: "color", Direction: SortAscending}, false},
		{"unknown direction is rejected", "name", "sideways", "", Listing{SortKey: SortByName, Direction: "sideways"}, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			listing, isValid := ParseListing(testCase.rawSortKey, testCase.rawDirection, testCase.rawSearchText)
			if isValid != testCase.isValid || listing != testCase.expectedListing {
				t.Fatalf("got %+v valid=%v, want %+v valid=%v", listing, isValid, testCase.expectedListing, testCase.isValid)
			}
		})
	}
}
