package music

import (
	"reflect"
	"testing"
)

func TestBuildCatalogGroupingKeys(t *testing.T) {
	testCases := []struct {
		name        string
		artist      string
		albumArtist string
		album       string
		genre       string
		expected    CatalogGroupingKeys
	}{
		{
			name:        "album artist wins over artist",
			artist:      "Guest",
			albumArtist: " The  Band ",
			album:       "Live_Album",
			genre:       "rnb; Hip-Hop | rnb",
			expected: CatalogGroupingKeys{
				ArtistKey: "the band", ArtistLabel: "The Band",
				AlbumKey: "the band::live album", AlbumLabel: "Live_Album",
				GenreKeys: []string{"r&b", "hip hop"}, GenreLabels: []string{"R&B", "Hip-Hop"},
			},
		},
		{
			name:   "artist without album has no album key",
			artist: "Solo",
			expected: CatalogGroupingKeys{
				ArtistKey: "solo", ArtistLabel: "Solo",
				GenreKeys: []string{}, GenreLabels: []string{},
			},
		},
		{
			name:  "album without artist has no album key",
			album: "Orphan",
			expected: CatalogGroupingKeys{
				AlbumLabel: "Orphan",
				GenreKeys:  []string{}, GenreLabels: []string{},
			},
		},
		{
			name:     "untagged track still yields non-nil genre slices",
			expected: CatalogGroupingKeys{GenreKeys: []string{}, GenreLabels: []string{}},
		},
		{
			name:   "accented genre keeps valid utf-8 capitalization",
			artist: "A",
			genre:  "électro pop",
			expected: CatalogGroupingKeys{
				ArtistKey: "a", ArtistLabel: "A",
				GenreKeys: []string{"électro pop"}, GenreLabels: []string{"Électro Pop"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := BuildCatalogGroupingKeys(testCase.artist, testCase.albumArtist, testCase.album, testCase.genre)
			if !reflect.DeepEqual(got, testCase.expected) {
				t.Fatalf("BuildCatalogGroupingKeys = %+v, want %+v", got, testCase.expected)
			}
		})
	}
}
