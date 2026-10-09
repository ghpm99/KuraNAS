package music

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"
)

func seedPlayHistoryLibrary(t *testing.T) (*catalogTestEnvironment, map[string]int) {
	t.Helper()
	environment := newCatalogTestEnvironment(t)
	base := environment.catalogBaseMoment
	environment.seedTracks(t, []catalogTrackFixture{
		trackFixture("one.mp3", "/m", base, AudioMetadataModel{Title: "One", Artist: "X", Album: "Y"}),
		trackFixture("two.mp3", "/m", base, AudioMetadataModel{Title: "Two", Artist: "X", Album: "Y"}),
		trackFixture("three.mp3", "/m", base, AudioMetadataModel{Title: "Three", Artist: "X", Album: "Y"}),
		{name: "gone.mp3", parentPath: "/m", updatedAt: base, isDeleted: true, metadata: &AudioMetadataModel{Title: "Gone"}},
		{name: "cover.jpg", parentPath: "/m", format: ".jpg", updatedAt: base},
	})
	return environment, environment.fileIDsByName
}

func backdatePlayEvents(t *testing.T, environment *catalogTestEnvironment, fileID int, playedAt time.Time) {
	t.Helper()
	err := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`UPDATE music_play_event SET played_at = $1 WHERE file_id = $2`, playedAt, fileID)
		return execErr
	})
	if err != nil {
		t.Fatalf("backdate: %v", err)
	}
}

func playedTrackFileIDs(playedTracks []MusicPlayedTrackDto) []int {
	fileIDs := make([]int, 0, len(playedTracks))
	for _, playedTrack := range playedTracks {
		fileIDs = append(fileIDs, playedTrack.Track.ID)
	}
	return fileIDs
}

func TestRecordPlayRejectsUnknownDeletedAndNonAudioFiles_Postgres(t *testing.T) {
	environment, ids := seedPlayHistoryLibrary(t)

	for _, fileID := range []int{999999, ids["gone.mp3"], ids["cover.jpg"]} {
		err := environment.service.RecordPlay("device-aaaaaaaa", RecordPlayRequest{FileID: fileID, PlayedSeconds: 30})
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("file %d error = %v, want sql.ErrNoRows", fileID, err)
		}
	}
}

func TestMostPlayedOrdersByPlayCountAndFiltersByPeriod_Postgres(t *testing.T) {
	environment, ids := seedPlayHistoryLibrary(t)
	service := environment.service

	plays := map[string]int{"one.mp3": 1, "two.mp3": 3, "three.mp3": 2}
	for name, count := range plays {
		for range count {
			if err := service.RecordPlay("device-aaaaaaaa", RecordPlayRequest{FileID: ids[name], PlayedSeconds: 35}); err != nil {
				t.Fatalf("record %s: %v", name, err)
			}
		}
	}

	allTime, err := service.GetMostPlayedTracks(PlayPeriodAll, 1, 10)
	if err != nil {
		t.Fatalf("most played: %v", err)
	}
	if got := playedTrackFileIDs(allTime.Items); !reflect.DeepEqual(got, []int{ids["two.mp3"], ids["three.mp3"], ids["one.mp3"]}) {
		t.Fatalf("all-time order = %v", got)
	}
	if allTime.Items[0].PlayCount != 3 || allTime.Items[0].Track.Name != "two.mp3" || allTime.Items[0].LastPlayedAt.IsZero() {
		t.Fatalf("first = %+v", allTime.Items[0])
	}

	backdatePlayEvents(t, environment, ids["two.mp3"], time.Now().Add(-60*24*time.Hour))
	lastThirtyDays, err := service.GetMostPlayedTracks(PlayPeriodLast30Days, 1, 10)
	if err != nil {
		t.Fatalf("most played 30d: %v", err)
	}
	if got := playedTrackFileIDs(lastThirtyDays.Items); !reflect.DeepEqual(got, []int{ids["three.mp3"], ids["one.mp3"]}) {
		t.Fatalf("30d order = %v", got)
	}
}

func TestMostPlayedPaginatesAndSkipsDeletedFiles_Postgres(t *testing.T) {
	environment, ids := seedPlayHistoryLibrary(t)
	service := environment.service

	for _, name := range []string{"one.mp3", "two.mp3", "three.mp3"} {
		_ = service.RecordPlay("device-aaaaaaaa", RecordPlayRequest{FileID: ids[name], PlayedSeconds: 30})
	}
	err := environment.dbContext.ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec(`INSERT INTO music_play_event (file_id, client_id, played_seconds) VALUES ($1, 'device-aaaaaaaa', 30)`, ids["gone.mp3"])
		return execErr
	})
	if err != nil {
		t.Fatalf("seed deleted play: %v", err)
	}

	firstPage, _ := service.GetMostPlayedTracks(PlayPeriodAll, 1, 2)
	secondPage, _ := service.GetMostPlayedTracks(PlayPeriodAll, 2, 2)
	if len(firstPage.Items) != 2 || !firstPage.Pagination.HasNext {
		t.Fatalf("first page = %+v", firstPage)
	}
	if len(secondPage.Items) != 1 || secondPage.Pagination.HasNext || secondPage.Pagination.HasPrev != true {
		t.Fatalf("second page = %+v", secondPage)
	}
}

func TestRecentPlaysAreDistinctTracksOrderedByLastPlay_Postgres(t *testing.T) {
	environment, ids := seedPlayHistoryLibrary(t)
	service := environment.service

	for _, name := range []string{"one.mp3", "two.mp3", "one.mp3", "three.mp3"} {
		if err := service.RecordPlay("device-aaaaaaaa", RecordPlayRequest{FileID: ids[name], PlayedSeconds: 30}); err != nil {
			t.Fatalf("record %s: %v", name, err)
		}
	}
	now := time.Now()
	backdatePlayEvents(t, environment, ids["two.mp3"], now.Add(-3*time.Hour))
	backdatePlayEvents(t, environment, ids["one.mp3"], now.Add(-2*time.Hour))
	backdatePlayEvents(t, environment, ids["three.mp3"], now.Add(-1*time.Hour))

	recentPlays, err := service.GetRecentlyPlayedTracks(1, 10)
	if err != nil {
		t.Fatalf("recent plays: %v", err)
	}
	if got := playedTrackFileIDs(recentPlays.Items); !reflect.DeepEqual(got, []int{ids["three.mp3"], ids["one.mp3"], ids["two.mp3"]}) {
		t.Fatalf("recent order = %v", got)
	}
	if recentPlays.Items[1].PlayCount != 2 {
		t.Fatalf("one.mp3 play count = %d, want 2", recentPlays.Items[1].PlayCount)
	}
}
