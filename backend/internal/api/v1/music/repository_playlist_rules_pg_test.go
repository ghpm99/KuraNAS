package music

import (
	"database/sql"
	"errors"
	"testing"
)

func playlistRuleFixture(environment *catalogTestEnvironment) []catalogTrackFixture {
	base := environment.catalogBaseMoment
	return []catalogTrackFixture{
		{name: "rock-a.mp3", parentPath: "/Music/Rock", updatedAt: base,
			metadata: &AudioMetadataModel{Title: "A", Artist: "Band A", Album: "Alb1", Genre: "Rock", Year: "1995", Length: 1}},
		{name: "metal-b.mp3", parentPath: "/Music/Rock/Live", updatedAt: base,
			metadata: &AudioMetadataModel{Title: "B", Artist: "Band B", Album: "Alb2", Genre: "Metal", Year: "2005-03-01", Length: 2}},
		{name: "rockpop-c.mp3", parentPath: "/Music/Rock2", updatedAt: base, isStarred: true,
			metadata: &AudioMetadataModel{Title: "C", Artist: "Band A", Album: "Alb1", Genre: "Rock; Pop", Year: "2010", Length: 4}},
		{name: "pop-d.mp3", parentPath: "/Music/Pop", updatedAt: base, isStarred: true,
			metadata: &AudioMetadataModel{Title: "D", Artist: "Band C", Album: "Alb3", Genre: "Pop", Year: "1980", Length: 8}},
		{name: "deleted.mp3", parentPath: "/Music/Rock", updatedAt: base, isDeleted: true,
			metadata: &AudioMetadataModel{Title: "Gone", Artist: "Band A", Album: "Alb1", Genre: "Rock", Year: "1995", Length: 16}},
		{name: "cover.jpg", parentPath: "/Music/Rock", format: ".jpg", updatedAt: base,
			metadata: &AudioMetadataModel{Title: "Cover", Artist: "Band A", Album: "Alb1", Genre: "Rock", Year: "1995", Length: 32}},
		{name: "pop-rock-e.mp3", parentPath: "/Music/Pop", updatedAt: base,
			metadata: &AudioMetadataModel{Title: "E", Artist: "Band D", Album: "Alb4", Genre: "rock", Year: "1999", Length: 64}},
		{name: "untagged.mp3", parentPath: "/Music/Rock", updatedAt: base},
	}
}

func (environment *catalogTestEnvironment) ruleTotals(t *testing.T, rules PlaylistRules, playlistID *int) PlaylistRuleTotals {
	t.Helper()
	totals, err := environment.service.Repository.CountPlaylistRuleTracks(rules.Sanitized(), playlistID)
	if err != nil {
		t.Fatalf("count rule tracks: %v", err)
	}
	return totals
}

func (environment *catalogTestEnvironment) expectRuleMatches(t *testing.T, rules PlaylistRules, expectedNames ...string) {
	t.Helper()
	totals := environment.ruleTotals(t, rules, nil)
	expectedLength := 0.0
	for _, expectedName := range expectedNames {
		expectedLength += environment.lengthOf(t, expectedName)
	}
	if totals.TrackCount != len(expectedNames) || totals.TotalLengthSeconds != expectedLength {
		t.Fatalf("rule %+v: expected %d tracks %v (total length %v), got %d tracks with total length %v",
			rules, len(expectedNames), expectedNames, expectedLength, totals.TrackCount, totals.TotalLengthSeconds)
	}
}

func (environment *catalogTestEnvironment) lengthOf(t *testing.T, fileName string) float64 {
	t.Helper()
	var length sql.NullFloat64
	err := environment.dbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(`SELECT length FROM audio_metadata WHERE file_id = $1`, environment.fileIDsByName[fileName]).Scan(&length)
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("length of %s: %v", fileName, err)
	}
	return length.Float64
}

func newPlaylistRuleEnvironment(t *testing.T) *catalogTestEnvironment {
	t.Helper()
	environment := newCatalogTestEnvironment(t)
	environment.seedTracks(t, playlistRuleFixture(environment))
	return environment
}

func TestPlaylistRuleAnyMatchesTracksOfAnyListedGenre_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)

	environment.expectRuleMatches(t,
		PlaylistRules{Match: "any", Genres: []string{"Rock", "Metal"}},
		"rock-a.mp3", "metal-b.mp3", "rockpop-c.mp3", "pop-rock-e.mp3")
}

func TestPlaylistRuleGenreEntryIsNormalizedLikeCatalogKeys_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)

	environment.expectRuleMatches(t,
		PlaylistRules{Match: "all", Genres: []string{"  ROCK "}},
		"rock-a.mp3", "rockpop-c.mp3", "pop-rock-e.mp3")
}

func TestPlaylistRuleAnyUnionsDifferentCriteria_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)

	environment.expectRuleMatches(t,
		PlaylistRules{Match: "any", Genres: []string{"Rock"}, FavoritesOnly: true},
		"rock-a.mp3", "rockpop-c.mp3", "pop-d.mp3", "pop-rock-e.mp3")
}

func TestPlaylistRuleAllIntersectsOnlyThePresentCriteria_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)

	environment.expectRuleMatches(t,
		PlaylistRules{Match: "all", Genres: []string{"Rock"}, FavoritesOnly: true},
		"rockpop-c.mp3")
	environment.expectRuleMatches(t,
		PlaylistRules{Match: "all", Artists: []string{"Band A"}, Genres: []string{"Pop"}},
		"rockpop-c.mp3")
	environment.expectRuleMatches(t,
		PlaylistRules{Match: "all", Genres: []string{"Rock", "Pop"}},
		"rock-a.mp3", "rockpop-c.mp3", "pop-d.mp3", "pop-rock-e.mp3")
}

func TestPlaylistRuleMatchesArtistAndAlbumKeys_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)

	environment.expectRuleMatches(t, PlaylistRules{Match: "any", Artists: []string{"band a"}}, "rock-a.mp3", "rockpop-c.mp3")
	environment.expectRuleMatches(t, PlaylistRules{Match: "any", Albums: []string{"Band B::Alb2"}}, "metal-b.mp3")
	environment.expectRuleMatches(t, PlaylistRules{Match: "any", Artists: []string{"Band C"}, Albums: []string{"Band B::Alb2"}}, "pop-d.mp3", "metal-b.mp3")
}

func TestPlaylistRuleFolderIncludesSubfoldersButNotSiblingsWithSamePrefix_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)

	environment.expectRuleMatches(t, PlaylistRules{Match: "any", Folders: []string{"/Music/Rock"}},
		"rock-a.mp3", "metal-b.mp3", "untagged.mp3")
	environment.expectRuleMatches(t, PlaylistRules{Match: "any", Folders: []string{"/Music/Rock/"}},
		"rock-a.mp3", "metal-b.mp3", "untagged.mp3")
	environment.expectRuleMatches(t, PlaylistRules{Match: "any", Folders: []string{"/Music/Rock", "/Music/Pop"}},
		"rock-a.mp3", "metal-b.mp3", "untagged.mp3", "pop-d.mp3", "pop-rock-e.mp3")
	environment.expectRuleMatches(t, PlaylistRules{Match: "any", Folders: []string{"/Music/R_ck"}})
	environment.expectRuleMatches(t, PlaylistRules{Match: "any", Folders: []string{"/Music/R%"}})
}

func TestPlaylistRuleYearRangeUsesLeadingFourDigits_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)

	environment.expectRuleMatches(t, PlaylistRules{Match: "any", YearFrom: intPointer(1990), YearTo: intPointer(2005)},
		"rock-a.mp3", "metal-b.mp3", "pop-rock-e.mp3")
	environment.expectRuleMatches(t, PlaylistRules{Match: "any", YearFrom: intPointer(2005)}, "metal-b.mp3", "rockpop-c.mp3")
	environment.expectRuleMatches(t, PlaylistRules{Match: "any", YearTo: intPointer(1990)}, "pop-d.mp3")
}

func TestPlaylistRuleFavoritesOnlyKeepsStarredTracks_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)

	environment.expectRuleMatches(t, PlaylistRules{Match: "any", FavoritesOnly: true}, "rockpop-c.mp3", "pop-d.mp3")
}

func TestPlaylistRuleHonorsCorrectedMetadata_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	environment.setOverride(t, "pop-d.mp3", `genre = $2`, "Rock")
	environment.setOverride(t, "rock-a.mp3", `genre = $2`, "Jazz")
	environment.recomputeKeys(t, "pop-d.mp3", "rock-a.mp3")

	environment.expectRuleMatches(t, PlaylistRules{Match: "any", Genres: []string{"Rock"}},
		"rockpop-c.mp3", "pop-d.mp3", "pop-rock-e.mp3")
	environment.expectRuleMatches(t, PlaylistRules{Match: "any", Genres: []string{"Jazz"}}, "rock-a.mp3")
}

func TestPlaylistRuleDiscountsExplicitExclusionsOfThatPlaylistOnly_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	rules := PlaylistRules{Match: "any", Genres: []string{"Rock"}}
	service := environment.service

	first, err := service.CreatePlaylist(CreatePlaylistRequest{Name: "First", Rules: &rules})
	if err != nil {
		t.Fatalf("create first: %v", err)
	}
	second, err := service.CreatePlaylist(CreatePlaylistRequest{Name: "Second", Rules: &rules})
	if err != nil {
		t.Fatalf("create second: %v", err)
	}

	err = environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`INSERT INTO playlist_rule_exclusion (playlist_id, file_id) VALUES ($1, $2)`, first.ID, environment.fileIDsByName["rock-a.mp3"])
		return execErr
	})
	if err != nil {
		t.Fatalf("insert exclusion: %v", err)
	}

	firstTotals := environment.ruleTotals(t, rules, &first.ID)
	secondTotals := environment.ruleTotals(t, rules, &second.ID)
	previewTotals := environment.ruleTotals(t, rules, nil)
	if firstTotals.TrackCount != 2 || firstTotals.TotalLengthSeconds != 68 {
		t.Fatalf("excluded track must leave the first playlist, got %+v", firstTotals)
	}
	if secondTotals.TrackCount != 3 || previewTotals.TrackCount != 3 {
		t.Fatalf("exclusion must not leak: second %+v preview %+v", secondTotals, previewTotals)
	}

	listed, err := service.GetPlaylists(1, 10, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	trackCountByName := map[string]int{}
	for _, playlist := range listed.Items {
		trackCountByName[playlist.Name] = playlist.TrackCount
	}
	if trackCountByName["First"] != 2 || trackCountByName["Second"] != 3 {
		t.Fatalf("list must compute each rule playlist on its own, got %+v", trackCountByName)
	}
}

func TestPlaylistRulePreviewMatchesCreatedPlaylistTrackCount_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	service := environment.service
	rules := PlaylistRules{Match: "any", Genres: []string{"Rock", "Metal"}, Folders: []string{"/Music/Pop"}}

	preview, err := service.PreviewPlaylistRules(rules)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	created, err := service.CreatePlaylist(CreatePlaylistRequest{Name: "Heavy", Description: "d", Rules: &rules})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if created.Kind != PlaylistKindRule || created.Rules == nil || created.Rules.Match != "any" {
		t.Fatalf("created playlist must be a rule playlist carrying its rules, got %+v", created)
	}
	if preview.TrackCount != created.TrackCount || preview.TrackCount != 5 {
		t.Fatalf("preview %d must equal created track_count %d (expected 5)", preview.TrackCount, created.TrackCount)
	}
	if created.TotalLengthSeconds != 1+2+4+8+64 {
		t.Fatalf("unexpected total length %v", created.TotalLengthSeconds)
	}

	fetched, err := service.GetPlaylistByID(created.ID)
	if err != nil || fetched.TrackCount != created.TrackCount || fetched.Kind != PlaylistKindRule {
		t.Fatalf("get by id: %+v err %v", fetched, err)
	}
}

func TestPlaylistRuleFollowsLibraryWithoutManualAction_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	rules := PlaylistRules{Match: "any", Genres: []string{"Rock"}}
	created, err := environment.service.CreatePlaylist(CreatePlaylistRequest{Name: "Rock", Rules: &rules})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	environment.seedTracks(t, []catalogTrackFixture{
		{name: "new-rock.mp3", parentPath: "/Music/New", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "New", Artist: "Band N", Album: "AlbN", Genre: "Rock"}},
	})

	fetched, err := environment.service.GetPlaylistByID(created.ID)
	if err != nil || fetched.TrackCount != created.TrackCount+1 {
		t.Fatalf("new rock track must join the playlist: before %d after %+v err %v", created.TrackCount, fetched, err)
	}
}

func TestUpdatePlaylistChangesRulesOnlyOfRulePlaylists_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	service := environment.service
	rules := PlaylistRules{Match: "any", Genres: []string{"Metal"}}
	rulePlaylist, err := service.CreatePlaylist(CreatePlaylistRequest{Name: "Rule", Rules: &rules})
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}
	manualPlaylist, err := service.CreatePlaylist(CreatePlaylistRequest{Name: "Manual"})
	if err != nil || manualPlaylist.Kind != PlaylistKindManual || manualPlaylist.Rules != nil {
		t.Fatalf("manual playlist: %+v err %v", manualPlaylist, err)
	}

	widenedRules := PlaylistRules{Match: "all", Genres: []string{"Rock"}, FavoritesOnly: true}
	updated, err := service.UpdatePlaylist(rulePlaylist.ID, UpdatePlaylistRequest{Name: "Renamed", Rules: &widenedRules})
	if err != nil {
		t.Fatalf("update rule: %v", err)
	}
	if updated.Name != "Renamed" || updated.Kind != PlaylistKindRule || updated.Rules.Match != "all" || updated.TrackCount != 1 {
		t.Fatalf("rules must be replaced and recounted, got %+v", updated)
	}

	renamedOnly, err := service.UpdatePlaylist(rulePlaylist.ID, UpdatePlaylistRequest{Name: "Again"})
	if err != nil || renamedOnly.Rules == nil || renamedOnly.Rules.Match != "all" {
		t.Fatalf("omitting rules must keep them: %+v err %v", renamedOnly, err)
	}

	_, err = service.UpdatePlaylist(manualPlaylist.ID, UpdatePlaylistRequest{Name: "Manual", Rules: &rules})
	if !errors.Is(err, ErrPlaylistRulesNotAllowed) {
		t.Fatalf("rules on a manual playlist must be refused, got %v", err)
	}

	emptyRules := PlaylistRules{Match: "any"}
	_, err = service.UpdatePlaylist(rulePlaylist.ID, UpdatePlaylistRequest{Name: "Again", Rules: &emptyRules})
	var validationError *fieldValidationError
	if !errors.As(err, &validationError) || validationError.MessageKey != "ERROR_MUSIC_PLAYLIST_RULE_EMPTY" {
		t.Fatalf("empty rules must be refused, got %v", err)
	}
}

func TestPlaylistTotalLengthOfManualPlaylistSumsActiveTracks_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	playlist, err := environment.service.CreatePlaylist(CreatePlaylistRequest{Name: "Manual"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	for _, trackName := range []string{"rock-a.mp3", "pop-d.mp3", "deleted.mp3"} {
		if _, err := environment.service.AddPlaylistTrack(playlist.ID, environment.fileIDsByName[trackName]); err != nil {
			t.Fatalf("add %s: %v", trackName, err)
		}
	}

	fetched, err := environment.service.GetPlaylistByID(playlist.ID)
	if err != nil || fetched.TotalLengthSeconds != 9 {
		t.Fatalf("expected total length 9, got %+v err %v", fetched, err)
	}
}

func TestPlaylistKindIsPersistedForSystemAndAIPlaylists_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	nowPlaying, err := environment.service.GetOrCreateNowPlaying()
	if err != nil || nowPlaying.Kind != PlaylistKindSystem {
		t.Fatalf("now playing: %+v err %v", nowPlaying, err)
	}

	err = environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		_, createErr := environment.service.Repository.CreateAIPlaylist(tx, "AI", "")
		return createErr
	})
	if err != nil {
		t.Fatalf("create ai: %v", err)
	}
	aiPlaylists, err := environment.service.Repository.GetAIPlaylists()
	if err != nil || len(aiPlaylists) != 1 || aiPlaylists[0].Kind != PlaylistKindAI {
		t.Fatalf("ai playlists: %+v err %v", aiPlaylists, err)
	}
}
