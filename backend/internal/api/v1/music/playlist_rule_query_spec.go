package music

import (
	"encoding/json"
	"strings"
)

type playlistRuleQuerySpec struct {
	PlaylistID     *int     `json:"playlist_id"`
	MatchAll       bool     `json:"match_all"`
	GenreKeys      []string `json:"genre_keys"`
	ArtistKeys     []string `json:"artist_keys"`
	AlbumKeys      []string `json:"album_keys"`
	Folders        []string `json:"folders"`
	FolderPatterns []string `json:"folder_patterns"`
	YearFrom       *int     `json:"year_from"`
	YearTo         *int     `json:"year_to"`
	FavoritesOnly  bool     `json:"favorites_only"`
}

func buildPlaylistRuleQuerySpec(rules PlaylistRules, playlistID *int) playlistRuleQuerySpec {
	folders := distinctTrimmedEntries(rules.Folders, trimFolderEntry)
	return playlistRuleQuerySpec{
		PlaylistID:     playlistID,
		MatchAll:       rules.Match == PlaylistRuleMatchAll,
		GenreKeys:      distinctTrimmedEntries(rules.Genres, genreKeyOfRuleEntry),
		ArtistKeys:     distinctTrimmedEntries(rules.Artists, artistKeyOfRuleEntry),
		AlbumKeys:      distinctTrimmedEntries(rules.Albums, albumKeyOfRuleEntry),
		Folders:        folders,
		FolderPatterns: subfolderPatternsOf(folders),
		YearFrom:       rules.YearFrom,
		YearTo:         rules.YearTo,
		FavoritesOnly:  rules.FavoritesOnly,
	}
}

func encodePlaylistRuleQuerySpecs(specs []playlistRuleQuerySpec) (string, error) {
	encodedSpecs, err := json.Marshal(specs)
	if err != nil {
		return "", err
	}
	return string(encodedSpecs), nil
}

func genreKeyOfRuleEntry(genreEntry string) string {
	return genreLookupKey(normalizeGenreLabel(genreEntry))
}

func artistKeyOfRuleEntry(artistEntry string) string {
	return artistGroupKey(normalizeText(artistEntry))
}

func albumKeyOfRuleEntry(albumEntry string) string {
	return normalizeLookupKey(albumEntry)
}

func subfolderPatternsOf(folders []string) []string {
	patterns := make([]string, 0, len(folders))
	for _, folder := range folders {
		subfolderPrefix := folder
		if !strings.HasSuffix(subfolderPrefix, "/") {
			subfolderPrefix += "/"
		}
		patterns = append(patterns, likePatternEscaper.Replace(subfolderPrefix)+"%")
	}
	return patterns
}
