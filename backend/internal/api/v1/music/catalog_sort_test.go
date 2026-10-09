package music

import "testing"

func TestParseCatalogSort(t *testing.T) {
	cases := []struct {
		name          string
		rawField      string
		rawOrder      string
		isYearAllowed bool
		want          CatalogSort
		isValid       bool
	}{
		{"defaults to tracks descending", "", "", false, CatalogSort{Field: CatalogSortByTracks, IsDescending: true}, true},
		{"name defaults to ascending", "name", "", false, CatalogSort{Field: CatalogSortByName}, true},
		{"recent defaults to descending", "recent", "", false, CatalogSort{Field: CatalogSortByRecent, IsDescending: true}, true},
		{"explicit ascending", "tracks", "asc", false, CatalogSort{Field: CatalogSortByTracks}, true},
		{"explicit descending", "name", "desc", false, CatalogSort{Field: CatalogSortByName, IsDescending: true}, true},
		{"year allowed", "year", "", true, CatalogSort{Field: CatalogSortByYear, IsDescending: true}, true},
		{"year rejected", "year", "", false, CatalogSort{}, false},
		{"unknown field", "size", "", true, CatalogSort{}, false},
		{"unknown order", "name", "up", true, CatalogSort{}, false},
	}
	for _, testCase := range cases {
		got, isValid := ParseCatalogSort(testCase.rawField, testCase.rawOrder, testCase.isYearAllowed)
		if isValid != testCase.isValid || got != testCase.want {
			t.Fatalf("%s: got %+v valid=%v, want %+v valid=%v", testCase.name, got, isValid, testCase.want, testCase.isValid)
		}
	}
}
