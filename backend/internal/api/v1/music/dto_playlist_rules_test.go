package music

import "testing"

func TestPlaylistRulesValidateRefusesInvalidRules(t *testing.T) {
	refusals := map[string]struct {
		rules      PlaylistRules
		messageKey string
		field      string
	}{
		"missing match":       {PlaylistRules{Genres: []string{"Rock"}}, "ERROR_MUSIC_PLAYLIST_RULE_MATCH_INVALID", "match"},
		"unknown match":       {PlaylistRules{Match: "some", Genres: []string{"Rock"}}, "ERROR_MUSIC_PLAYLIST_RULE_MATCH_INVALID", "match"},
		"no criterion":        {PlaylistRules{Match: "any"}, "ERROR_MUSIC_PLAYLIST_RULE_EMPTY", "rules"},
		"blank entries":       {PlaylistRules{Match: "all", Genres: []string{"  "}}, "ERROR_MUSIC_PLAYLIST_RULE_EMPTY", "rules"},
		"year below range":    {PlaylistRules{Match: "any", YearFrom: intPointer(999)}, "ERROR_MUSIC_PLAYLIST_RULE_YEAR_INVALID", "year_from"},
		"year above range":    {PlaylistRules{Match: "any", YearTo: intPointer(10000)}, "ERROR_MUSIC_PLAYLIST_RULE_YEAR_INVALID", "year_from"},
		"inverted year range": {PlaylistRules{Match: "any", YearFrom: intPointer(2010), YearTo: intPointer(2000)}, "ERROR_MUSIC_PLAYLIST_RULE_YEAR_INVALID", "year_to"},
	}

	for name, refusal := range refusals {
		_, validationError := validatedPlaylistRules(refusal.rules)
		if validationError == nil || validationError.MessageKey != refusal.messageKey || validationError.Field != refusal.field {
			t.Fatalf("%s: expected %s on %s, got %+v", name, refusal.messageKey, refusal.field, validationError)
		}
	}
}

func TestPlaylistRulesValidateAcceptsAnySingleCriterion(t *testing.T) {
	accepted := []PlaylistRules{
		{Match: "any", Genres: []string{"Rock"}},
		{Match: "all", Artists: []string{"Band"}},
		{Match: "all", Albums: []string{"a::b"}},
		{Match: "any", Folders: []string{"/Music"}},
		{Match: "any", YearFrom: intPointer(1000)},
		{Match: "any", YearTo: intPointer(9999)},
		{Match: "any", FavoritesOnly: true},
	}
	for _, rules := range accepted {
		if _, validationError := validatedPlaylistRules(rules); validationError != nil {
			t.Fatalf("%+v must be accepted, got %+v", rules, validationError)
		}
	}
}

func TestPlaylistRulesSanitizedTrimsDeduplicatesAndNeverReturnsNilLists(t *testing.T) {
	sanitized := PlaylistRules{Match: " any ", Genres: []string{" Rock ", "Rock", ""}, Folders: []string{"/Music/Rock//", "/"}}.Sanitized()

	if sanitized.Match != "any" || len(sanitized.Genres) != 1 || sanitized.Genres[0] != "Rock" {
		t.Fatalf("unexpected sanitized rules: %+v", sanitized)
	}
	if len(sanitized.Folders) != 2 || sanitized.Folders[0] != "/Music/Rock" || sanitized.Folders[1] != "/" {
		t.Fatalf("unexpected folders: %+v", sanitized.Folders)
	}
	if sanitized.Artists == nil || sanitized.Albums == nil {
		t.Fatal("lists must be empty, never nil, so they encode as []")
	}
}
