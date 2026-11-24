package music

import (
	"database/sql"
	"errors"
	"reflect"
	"sort"
	"testing"
)

func createRockRulePlaylist(t *testing.T, environment *catalogTestEnvironment) PlaylistDto {
	t.Helper()
	rules := PlaylistRules{Match: "any", Genres: []string{"Rock"}}
	playlist, err := environment.service.CreatePlaylist(CreatePlaylistRequest{Name: "Rock", Rules: &rules})
	if err != nil {
		t.Fatalf("create rule playlist: %v", err)
	}
	return playlist
}

func (environment *catalogTestEnvironment) playlistTrackFileNames(t *testing.T, playlistID int, pageSize int) []string {
	t.Helper()
	fileNames := []string{}
	for page := 1; ; page++ {
		pagination, err := environment.service.Repository.GetPlaylistTracks(playlistID, page, pageSize)
		if err != nil {
			t.Fatalf("tracks page %d: %v", page, err)
		}
		for _, track := range pagination.Items {
			fileNames = append(fileNames, environment.namesByFileID[track.FileID])
		}
		if !pagination.Pagination.HasNext {
			return fileNames
		}
	}
}

func (environment *catalogTestEnvironment) playlistQueueFileNames(t *testing.T, playlistID int) []string {
	t.Helper()
	entries, err := environment.service.Repository.GetPlaylistQueue(playlistID, maxQueueEntries+1)
	if err != nil {
		t.Fatalf("queue: %v", err)
	}
	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		fileNames = append(fileNames, environment.namesByFileID[entry.FileID])
	}
	return fileNames
}

func (environment *catalogTestEnvironment) removeTrackFromPlaylist(t *testing.T, playlistID int, fileName string) {
	t.Helper()
	if err := environment.service.RemovePlaylistTrack(playlistID, environment.fileIDsByName[fileName]); err != nil {
		t.Fatalf("remove %s: %v", fileName, err)
	}
}

func TestRulePlaylistTracksPaginateOrderedByTitleThenID_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	playlist := createRockRulePlaylist(t, environment)

	firstPage, err := environment.service.Repository.GetPlaylistTracks(playlist.ID, 1, 2)
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	secondPage, err := environment.service.Repository.GetPlaylistTracks(playlist.ID, 2, 2)
	if err != nil {
		t.Fatalf("second page: %v", err)
	}

	if len(firstPage.Items) != 2 || !firstPage.Pagination.HasNext {
		t.Fatalf("first page must hold 2 items and have next: %+v", firstPage.Pagination)
	}
	if len(secondPage.Items) != 1 || secondPage.Pagination.HasNext || !secondPage.Pagination.HasPrev {
		t.Fatalf("second page must hold the last item without next: %+v", secondPage.Pagination)
	}
	allNames := environment.playlistTrackFileNames(t, playlist.ID, 2)
	if !reflect.DeepEqual(allNames, []string{"rock-a.mp3", "rockpop-c.mp3", "pop-rock-e.mp3"}) {
		t.Fatalf("tracks must come ordered by title A, C, E, got %v", allNames)
	}
	firstTrack := firstPage.Items[0]
	if firstTrack.Position != 1 || firstTrack.PlaylistID != playlist.ID || firstTrack.MetadataTitle != "A" || firstTrack.AddedAt.IsZero() {
		t.Fatalf("unexpected first rule track: %+v", firstTrack)
	}
}

func TestRulePlaylistTracksWithEqualTitlesFallBackToFileID_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	playlist := createRockRulePlaylist(t, environment)
	environment.seedTracks(t, []catalogTrackFixture{
		{name: "twin-1.mp3", parentPath: "/Music/Twins", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "Twin", Genre: "Rock"}},
		{name: "twin-2.mp3", parentPath: "/Music/Twins", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "Twin", Genre: "Rock"}},
	})

	fileNames := environment.playlistTrackFileNames(t, playlist.ID, 10)

	if !reflect.DeepEqual(fileNames[3:], []string{"twin-1.mp3", "twin-2.mp3"}) {
		t.Fatalf("equal titles must be ordered by file id, got %v", fileNames)
	}
}

func TestRulePlaylistQueueHoldsEveryRuleTrackOnceInTrackOrder_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	playlist := createRockRulePlaylist(t, environment)

	queueNames := environment.playlistQueueFileNames(t, playlist.ID)

	if !reflect.DeepEqual(queueNames, environment.playlistTrackFileNames(t, playlist.ID, 50)) {
		t.Fatalf("queue must match the track listing, got %v", queueNames)
	}
	distinctNames := map[string]bool{}
	for _, queueName := range queueNames {
		distinctNames[queueName] = true
	}
	if len(distinctNames) != len(queueNames) || len(queueNames) != 3 {
		t.Fatalf("queue must hold the 3 rule tracks without duplicates, got %v", queueNames)
	}
}

func TestRulePlaylistQueueIsLimitedLikeOtherQueues_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	playlist := createRockRulePlaylist(t, environment)

	entries, err := environment.service.Repository.GetPlaylistQueue(playlist.ID, 2)
	if err != nil || len(entries) != 2 {
		t.Fatalf("limit must cap the queue at 2 entries, got %d err %v", len(entries), err)
	}
}

func TestRemovingTrackFromRulePlaylistExcludesItFromTracksQueueAndCount_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	playlist := createRockRulePlaylist(t, environment)

	environment.removeTrackFromPlaylist(t, playlist.ID, "rockpop-c.mp3")
	environment.removeTrackFromPlaylist(t, playlist.ID, "rockpop-c.mp3")

	expectedNames := []string{"rock-a.mp3", "pop-rock-e.mp3"}
	if tracks := environment.playlistTrackFileNames(t, playlist.ID, 10); !reflect.DeepEqual(tracks, expectedNames) {
		t.Fatalf("tracks after exclusion: %v", tracks)
	}
	if queue := environment.playlistQueueFileNames(t, playlist.ID); !reflect.DeepEqual(queue, expectedNames) {
		t.Fatalf("queue after exclusion: %v", queue)
	}
	fetched, err := environment.service.GetPlaylistByID(playlist.ID)
	if err != nil || fetched.TrackCount != 2 {
		t.Fatalf("count after exclusion: %+v err %v", fetched, err)
	}
}

func TestExcludedTrackStaysOutAfterNewMatchingTracksAndReindex_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	playlist := createRockRulePlaylist(t, environment)
	environment.removeTrackFromPlaylist(t, playlist.ID, "rock-a.mp3")

	environment.seedTracks(t, []catalogTrackFixture{
		{name: "new-rock.mp3", parentPath: "/Music/New", updatedAt: environment.catalogBaseMoment,
			metadata: &AudioMetadataModel{Title: "New", Genre: "Rock"}},
	})
	environment.reindex(t, "rock-a.mp3", AudioMetadataModel{Title: "A", Artist: "Band A", Album: "Alb1", Genre: "Rock", Year: "1995", Length: 1})

	fileNames := environment.playlistTrackFileNames(t, playlist.ID, 10)
	sort.Strings(fileNames)
	if !reflect.DeepEqual(fileNames, []string{"new-rock.mp3", "pop-rock-e.mp3", "rockpop-c.mp3"}) {
		t.Fatalf("excluded track must stay out and the new one come in, got %v", fileNames)
	}
}

func TestExclusionOfOneRulePlaylistDoesNotAffectAnother_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	first := createRockRulePlaylist(t, environment)
	second := createRockRulePlaylist(t, environment)

	environment.removeTrackFromPlaylist(t, first.ID, "rock-a.mp3")

	if len(environment.playlistTrackFileNames(t, second.ID, 10)) != 3 {
		t.Fatalf("second playlist must keep its 3 tracks")
	}
}

func TestAddAndReorderOnRulePlaylistAreRefusedAndLeaveNoTrackRows_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	playlist := createRockRulePlaylist(t, environment)

	_, addErr := environment.service.AddPlaylistTrack(playlist.ID, environment.fileIDsByName["pop-d.mp3"])
	reorderErr := environment.service.ReorderPlaylistTracks(playlist.ID, []ReorderTrackItem{{FileID: environment.fileIDsByName["rock-a.mp3"], Position: 2}})

	if !isRuleReadOnly(addErr) || !isRuleReadOnly(reorderErr) {
		t.Fatalf("add and reorder must be refused, got %v and %v", addErr, reorderErr)
	}
	if storedRows := environment.playlistTrackRowCount(t, playlist.ID); storedRows != 0 {
		t.Fatalf("rule playlist must not store track rows, got %d", storedRows)
	}
}

func TestReplacingTracksOfRulePlaylistIsRefused_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	playlist := createRockRulePlaylist(t, environment)

	err := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		return environment.service.Repository.ReplacePlaylistTracks(tx, playlist.ID, []int{environment.fileIDsByName["pop-d.mp3"]})
	})

	if !isRuleReadOnly(err) || environment.playlistTrackRowCount(t, playlist.ID) != 0 {
		t.Fatalf("replace must be refused without writing rows, got %v", err)
	}
}

func TestManualPlaylistKeepsAddReorderRemoveBehavior_Postgres(t *testing.T) {
	environment := newPlaylistRuleEnvironment(t)
	manual, err := environment.service.CreatePlaylist(CreatePlaylistRequest{Name: "Manual"})
	if err != nil {
		t.Fatalf("create manual: %v", err)
	}
	for _, fileName := range []string{"pop-d.mp3", "rock-a.mp3", "metal-b.mp3"} {
		if _, err := environment.service.AddPlaylistTrack(manual.ID, environment.fileIDsByName[fileName]); err != nil {
			t.Fatalf("add %s: %v", fileName, err)
		}
	}
	if _, err := environment.service.AddPlaylistTrack(manual.ID, environment.fileIDsByName["pop-d.mp3"]); err == nil {
		t.Fatalf("duplicate add must still be refused")
	}
	if err := environment.service.ReorderPlaylistTracks(manual.ID, []ReorderTrackItem{{FileID: environment.fileIDsByName["metal-b.mp3"], Position: 1}}); err != nil {
		t.Fatalf("reorder: %v", err)
	}
	environment.removeTrackFromPlaylist(t, manual.ID, "rock-a.mp3")

	remaining := environment.playlistTrackFileNames(t, manual.ID, 10)
	if len(remaining) != 2 || environment.manualPlaylistExclusionCount(t, manual.ID) != 0 {
		t.Fatalf("manual removal must delete the row and record no exclusion, got %v", remaining)
	}
	if queue := environment.playlistQueueFileNames(t, manual.ID); !reflect.DeepEqual(queue, remaining) {
		t.Fatalf("manual queue %v must follow track order %v", queue, remaining)
	}
	err = environment.service.RemovePlaylistTrack(manual.ID, environment.fileIDsByName["rock-a.mp3"])
	if err == nil {
		t.Fatalf("removing an absent track from a manual playlist must still fail")
	}
}

func isRuleReadOnly(err error) bool {
	return errors.Is(err, ErrPlaylistRuleReadOnly)
}

func (environment *catalogTestEnvironment) playlistTrackRowCount(t *testing.T, playlistID int) int {
	t.Helper()
	return environment.countRows(t, `SELECT COUNT(*) FROM playlist_track WHERE playlist_id = $1`, playlistID)
}

func (environment *catalogTestEnvironment) manualPlaylistExclusionCount(t *testing.T, playlistID int) int {
	t.Helper()
	return environment.countRows(t, `SELECT COUNT(*) FROM playlist_rule_exclusion WHERE playlist_id = $1`, playlistID)
}

func (environment *catalogTestEnvironment) countRows(t *testing.T, query string, playlistID int) int {
	t.Helper()
	var rowCount int
	err := environment.dbContext.QueryTx(func(tx *sql.Tx) error {
		return tx.QueryRow(query, playlistID).Scan(&rowCount)
	})
	if err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return rowCount
}
