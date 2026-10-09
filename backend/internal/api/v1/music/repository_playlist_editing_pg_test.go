package music

import (
	"database/sql"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func seedPlaylistWithTracks(t *testing.T, environment *catalogTestEnvironment, trackNames []string) (int, []int) {
	t.Helper()
	fixtures := make([]catalogTrackFixture, 0, len(trackNames))
	for _, name := range trackNames {
		fixtures = append(fixtures, catalogTrackFixture{name: name, parentPath: "/music/p", updatedAt: environment.catalogBaseMoment})
	}
	environment.seedTracks(t, fixtures)

	var playlistID int
	err := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		playlist, err := environment.service.Repository.CreatePlaylist(tx, "Mix", "", false)
		playlistID = playlist.ID
		return err
	})
	if err != nil {
		t.Fatalf("create playlist: %v", err)
	}

	fileIDs := make([]int, 0, len(trackNames))
	for _, name := range trackNames {
		fileID := environment.fileIDsByName[name]
		if _, err := environment.service.AddPlaylistTrack(playlistID, fileID); err != nil {
			t.Fatalf("add %s: %v", name, err)
		}
		fileIDs = append(fileIDs, fileID)
	}
	return playlistID, fileIDs
}

func playlistPositionsByFileID(t *testing.T, environment *catalogTestEnvironment, playlistID int) map[int]int {
	t.Helper()
	positionsByFileID := map[int]int{}
	err := environment.dbContext.QueryTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(`SELECT file_id, position FROM playlist_track WHERE playlist_id = $1`, playlistID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var fileID, position int
			if err := rows.Scan(&fileID, &position); err != nil {
				return err
			}
			positionsByFileID[fileID] = position
		}
		return rows.Err()
	})
	if err != nil {
		t.Fatalf("read positions: %v", err)
	}
	return positionsByFileID
}

func fileIDsInPlaylistOrder(positionsByFileID map[int]int) []int {
	orderedFileIDs := make([]int, len(positionsByFileID))
	for fileID, position := range positionsByFileID {
		orderedFileIDs[position-1] = fileID
	}
	return orderedFileIDs
}

func TestReorderPlaylistTrackShiftsNeighboursIntoContiguousPositions_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	playlistID, fileIDs := seedPlaylistWithTracks(t, environment, []string{"a.mp3", "b.mp3", "c.mp3", "d.mp3"})
	a, b, c, d := fileIDs[0], fileIDs[1], fileIDs[2], fileIDs[3]

	steps := []struct {
		name     string
		fileID   int
		position int
		expected []int
	}{
		{"move down", a, 3, []int{b, c, a, d}},
		{"move up", d, 1, []int{d, b, c, a}},
		{"clamp past the end", d, 99, []int{b, c, a, d}},
		{"clamp before the start", a, -5, []int{a, b, c, d}},
		{"same position", b, 2, []int{a, b, c, d}},
	}
	for _, step := range steps {
		err := environment.service.ReorderPlaylistTracks(playlistID, []ReorderTrackItem{{FileID: step.fileID, Position: step.position}})
		if err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}
		actual := fileIDsInPlaylistOrder(playlistPositionsByFileID(t, environment, playlistID))
		if !reflect.DeepEqual(actual, step.expected) {
			t.Fatalf("%s: expected %v, got %v", step.name, step.expected, actual)
		}
	}
}

func TestReorderPlaylistTrackRejectsFileOutsideThePlaylistWithoutTouchingOthers_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	playlistID, fileIDs := seedPlaylistWithTracks(t, environment, []string{"a.mp3", "b.mp3"})

	err := environment.service.ReorderPlaylistTracks(playlistID, []ReorderTrackItem{{FileID: 987654, Position: 1}})

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	expected := map[int]int{fileIDs[0]: 1, fileIDs[1]: 2}
	if actual := playlistPositionsByFileID(t, environment, playlistID); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("positions changed: %v", actual)
	}
}

func TestRemovePlaylistTrackCompactsRemainingPositions_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	playlistID, fileIDs := seedPlaylistWithTracks(t, environment, []string{"a.mp3", "b.mp3", "c.mp3", "d.mp3"})

	if err := environment.service.RemovePlaylistTrack(playlistID, fileIDs[1]); err != nil {
		t.Fatalf("remove: %v", err)
	}

	expected := map[int]int{fileIDs[0]: 1, fileIDs[2]: 2, fileIDs[3]: 3}
	if actual := playlistPositionsByFileID(t, environment, playlistID); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func TestAddPlaylistTrackRejectsDuplicateWithoutChangingPositions_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	playlistID, fileIDs := seedPlaylistWithTracks(t, environment, []string{"a.mp3", "b.mp3"})

	_, err := environment.service.AddPlaylistTrack(playlistID, fileIDs[0])

	if !errors.Is(err, ErrTrackAlreadyInPlaylist) {
		t.Fatalf("expected ErrTrackAlreadyInPlaylist, got %v", err)
	}
	expected := map[int]int{fileIDs[0]: 1, fileIDs[1]: 2}
	if actual := playlistPositionsByFileID(t, environment, playlistID); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("positions changed: %v", actual)
	}
}

func TestConcurrentAddsNeverDuplicatePositions_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	const concurrentTrackCount = 12
	trackNames := make([]string, 0, concurrentTrackCount)
	for index := 0; index < concurrentTrackCount; index++ {
		trackNames = append(trackNames, string(rune('a'+index))+".mp3")
	}
	playlistID, _ := seedPlaylistWithTracks(t, environment, nil)
	environment.seedTracks(t, tracksNamed(trackNames, environment))

	var waitGroup sync.WaitGroup
	failures := make(chan error, concurrentTrackCount)
	for _, name := range trackNames {
		waitGroup.Add(1)
		go func(fileID int) {
			defer waitGroup.Done()
			_, err := environment.service.AddPlaylistTrack(playlistID, fileID)
			failures <- err
		}(environment.fileIDsByName[name])
	}
	waitGroup.Wait()
	close(failures)

	for err := range failures {
		if err != nil {
			t.Fatalf("concurrent add failed: %v", err)
		}
	}
	positionsByFileID := playlistPositionsByFileID(t, environment, playlistID)
	seenPositions := map[int]bool{}
	for _, position := range positionsByFileID {
		seenPositions[position] = true
	}
	if len(positionsByFileID) != concurrentTrackCount || len(seenPositions) != concurrentTrackCount {
		t.Fatalf("expected %d distinct positions, got %v", concurrentTrackCount, positionsByFileID)
	}
	for position := 1; position <= concurrentTrackCount; position++ {
		if !seenPositions[position] {
			t.Fatalf("position %d missing: %v", position, positionsByFileID)
		}
	}
}

func tracksNamed(trackNames []string, environment *catalogTestEnvironment) []catalogTrackFixture {
	fixtures := make([]catalogTrackFixture, 0, len(trackNames))
	for _, name := range trackNames {
		fixtures = append(fixtures, catalogTrackFixture{name: name, parentPath: "/music/c", updatedAt: environment.catalogBaseMoment})
	}
	return fixtures
}

func TestGetPlaylistsFiltersByNameIgnoringCase_Postgres(t *testing.T) {
	environment := newCatalogTestEnvironment(t)
	err := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		for _, name := range []string{"Road Trip", "Workout", "roadhouse"} {
			if _, err := environment.service.Repository.CreatePlaylist(tx, name, "", false); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed playlists: %v", err)
	}

	page, err := environment.service.Repository.GetPlaylists(1, 10, "ROAD")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	names := map[string]bool{}
	for _, playlist := range page.Items {
		names[playlist.Name] = true
	}
	if len(names) != 2 || !names["Road Trip"] || !names["roadhouse"] {
		t.Fatalf("unexpected search result: %v", names)
	}

	everything, err := environment.service.Repository.GetPlaylists(1, 10, "")
	if err != nil || len(everything.Items) != 3 {
		t.Fatalf("expected all playlists without search, got %d (%v)", len(everything.Items), err)
	}
}
