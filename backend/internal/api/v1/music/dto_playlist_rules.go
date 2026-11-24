package music

import (
	"errors"
	"strings"
)

const (
	PlaylistRuleMatchAny = "any"
	PlaylistRuleMatchAll = "all"

	playlistRuleMinYear = 1000
	playlistRuleMaxYear = 9999
)

var ErrPlaylistRulesNotAllowed = errors.New("rules are only allowed on rule playlists")

// PlaylistRules is the criteria of a rule playlist. Match is "any" (one
// criterion is enough) or "all" (every present criterion must hold); inside a
// list criterion any entry matches.
type PlaylistRules struct {
	Match         string   `json:"match"`
	Genres        []string `json:"genres"`
	Artists       []string `json:"artists"`
	Albums        []string `json:"albums"`
	Folders       []string `json:"folders"`
	YearFrom      *int     `json:"year_from"`
	YearTo        *int     `json:"year_to"`
	FavoritesOnly bool     `json:"favorites_only"`
}

type PlaylistRulePreviewDto struct {
	TrackCount int `json:"track_count"`
}

type PlaylistRuleTotals struct {
	TrackCount         int
	TotalLengthSeconds float64
}

// Sanitized returns a copy with entries trimmed, blanks and duplicates
// dropped and no nil lists, which is the form that is validated and stored.
func (rules PlaylistRules) Sanitized() PlaylistRules {
	return PlaylistRules{
		Match:         strings.TrimSpace(rules.Match),
		Genres:        distinctTrimmedEntries(rules.Genres, strings.TrimSpace),
		Artists:       distinctTrimmedEntries(rules.Artists, strings.TrimSpace),
		Albums:        distinctTrimmedEntries(rules.Albums, strings.TrimSpace),
		Folders:       distinctTrimmedEntries(rules.Folders, trimFolderEntry),
		YearFrom:      rules.YearFrom,
		YearTo:        rules.YearTo,
		FavoritesOnly: rules.FavoritesOnly,
	}
}

// Validate refuses rules with an unknown match, an invalid year range or no
// criterion at all. It expects sanitized rules.
func (rules PlaylistRules) Validate() *fieldValidationError {
	if rules.Match != PlaylistRuleMatchAny && rules.Match != PlaylistRuleMatchAll {
		return &fieldValidationError{Field: "match", MessageKey: "ERROR_MUSIC_PLAYLIST_RULE_MATCH_INVALID"}
	}
	if isPlaylistRuleYearOutOfRange(rules.YearFrom) || isPlaylistRuleYearOutOfRange(rules.YearTo) {
		return &fieldValidationError{Field: "year_from", MessageKey: "ERROR_MUSIC_PLAYLIST_RULE_YEAR_INVALID"}
	}
	if rules.YearFrom != nil && rules.YearTo != nil && *rules.YearFrom > *rules.YearTo {
		return &fieldValidationError{Field: "year_to", MessageKey: "ERROR_MUSIC_PLAYLIST_RULE_YEAR_INVALID"}
	}
	if !rules.hasAnyCriterion() {
		return &fieldValidationError{Field: "rules", MessageKey: "ERROR_MUSIC_PLAYLIST_RULE_EMPTY"}
	}
	return nil
}

func (rules PlaylistRules) hasAnyCriterion() bool {
	return len(rules.Genres) > 0 ||
		len(rules.Artists) > 0 ||
		len(rules.Albums) > 0 ||
		len(rules.Folders) > 0 ||
		rules.YearFrom != nil ||
		rules.YearTo != nil ||
		rules.FavoritesOnly
}

func isPlaylistRuleYearOutOfRange(year *int) bool {
	return year != nil && (*year < playlistRuleMinYear || *year > playlistRuleMaxYear)
}

func validatedPlaylistRules(rules PlaylistRules) (PlaylistRules, *fieldValidationError) {
	sanitizedRules := rules.Sanitized()
	return sanitizedRules, sanitizedRules.Validate()
}

func trimFolderEntry(folder string) string {
	trimmed := strings.TrimSpace(folder)
	if trimmed == "/" {
		return trimmed
	}
	return strings.TrimRight(trimmed, "/")
}

func distinctTrimmedEntries(entries []string, normalizeEntry func(string) string) []string {
	distinctEntries := make([]string, 0, len(entries))
	seenEntries := make(map[string]bool, len(entries))
	for _, entry := range entries {
		normalizedEntry := normalizeEntry(entry)
		if normalizedEntry == "" || seenEntries[normalizedEntry] {
			continue
		}
		seenEntries[normalizedEntry] = true
		distinctEntries = append(distinctEntries, normalizedEntry)
	}
	return distinctEntries
}
