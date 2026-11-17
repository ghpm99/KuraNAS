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
				GenreKeys: []string{"electro pop"}, GenreLabels: []string{"Électro Pop"},
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

func TestBuildCatalogGroupingKeysSplitsGenresBySlashExceptKnownCompounds(t *testing.T) {
	testCases := map[string]struct {
		genre      string
		wantLabels []string
	}{
		"slash splits":             {"Rock/Pop", []string{"Rock", "Pop"}},
		"slash with spaces splits": {"Rock / Pop", []string{"Rock", "Pop"}},
		"known compound is kept":   {"R&B/Soul", []string{"R&B / Soul"}},
		"mixed separators":         {"Rock;Pop/Funk", []string{"Rock", "Pop", "Funk"}},
		"compound beside others":   {"Rock; R&B/Soul, Jazz/Blues", []string{"Rock", "R&B / Soul", "Jazz", "Blues"}},
	}
	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			got := BuildCatalogGroupingKeys("", "", "", testCase.genre)
			if !reflect.DeepEqual(got.GenreLabels, testCase.wantLabels) {
				t.Fatalf("labels = %v, want %v", got.GenreLabels, testCase.wantLabels)
			}
			if len(got.GenreKeys) != len(testCase.wantLabels) {
				t.Fatalf("keys = %v, want one per label", got.GenreKeys)
			}
		})
	}
}

func TestBuildCatalogGroupingKeysTreatsGenresDifferingOnlyInCaseAccentsOrSpacesAsOne(t *testing.T) {
	for _, genre := range []string{"Música", "musica", "MÚSICA ", "  música"} {
		got := BuildCatalogGroupingKeys("", "", "", genre)
		if !reflect.DeepEqual(got.GenreKeys, []string{"musica"}) {
			t.Fatalf("%q keys = %v, want [musica]", genre, got.GenreKeys)
		}
	}
	accented := BuildCatalogGroupingKeys("", "", "", "música")
	if !reflect.DeepEqual(accented.GenreLabels, []string{"Música"}) {
		t.Fatalf("label = %v, want accent preserved", accented.GenreLabels)
	}
}

func TestBuildCatalogGroupingKeysKeepsOneGenreWhenTheTagRepeatsItWithAndWithoutAccent(t *testing.T) {
	got := BuildCatalogGroupingKeys("", "", "", "Música; Musica")
	if !reflect.DeepEqual(got.GenreKeys, []string{"musica"}) || len(got.GenreLabels) != 1 {
		t.Fatalf("keys=%v labels=%v, want a single genre", got.GenreKeys, got.GenreLabels)
	}
}

func TestBuildCatalogGroupingKeysKeepsAccentsInArtistAndAlbumKeys(t *testing.T) {
	got := BuildCatalogGroupingKeys("Beyoncé", "", "Lemonade Ópera", "")
	if got.ArtistKey != "beyoncé" || got.AlbumKey != "beyoncé::lemonade ópera" {
		t.Fatalf("artist key %q album key %q must keep accents", got.ArtistKey, got.AlbumKey)
	}
}
