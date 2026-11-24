package video

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"nas-go/api/internal/testutil"
)

const (
	progressClientID  = "tv"
	episodesPerSeason = 8
)

func newTwoSeasonFixture(t *testing.T) (*seriesFixture, int) {
	t.Helper()
	fileNames := []string{}
	for season := 1; season <= 2; season++ {
		for episode := 1; episode <= episodesPerSeason; episode++ {
			fileNames = append(fileNames, fmt.Sprintf("Show S%02dE%02d.mkv", season, episode))
		}
	}
	fixture := newSeriesFixture(t, "/data/Show", fileNames)
	return fixture, fixture.seriesPlaylistID(t)
}

func (fixture *seriesFixture) saveProgress(t *testing.T, clientID string, fileName string, positionSeconds float64, durationSeconds float64, isCompleted bool) {
	t.Helper()
	err := fixture.repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		_, upsertErr := fixture.repository.UpsertVideoWatchProgress(tx, VideoWatchProgressModel{
			ClientID: clientID, VideoID: fixture.videoIDs[fileName],
			PositionSeconds: positionSeconds, DurationSeconds: durationSeconds, Completed: isCompleted,
		})
		return upsertErr
	})
	if err != nil {
		t.Fatalf("save progress %s: %v", fileName, err)
	}
}

func (fixture *seriesFixture) completedFileNames(t *testing.T, clientID string) []string {
	t.Helper()
	rows, err := fixture.repository.DbContext.GetDatabase().Query(
		`SELECT hf.name FROM video_watch_progress wp JOIN home_file hf ON hf.id = wp.video_id
		 WHERE wp.client_id = $1 AND wp.completed AND hf.name LIKE 'Show %' ORDER BY hf.name`, clientID)
	if err != nil {
		t.Fatalf("completed names: %v", err)
	}
	defer rows.Close()
	names := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		names = append(names, name)
	}
	return names
}

func TestPlaylistHeaderSeasonsCountEffectiveSeasonsAndWatched_Postgres(t *testing.T) {
	fixture, playlistID := newTwoSeasonFixture(t)
	fixture.saveProgress(t, progressClientID, "Show S01E01.mkv", 100, 100, true)
	fixture.saveProgress(t, "phone", "Show S01E02.mkv", 100, 100, true)
	fixture.overrideEpisode(t, "Show S01E03.mkv", 2, 9)

	header, err := fixture.service.GetPlaylistHeaderByID(progressClientID, playlistID)
	if err != nil {
		t.Fatalf("header: %v", err)
	}
	want := []PlaylistSeasonDto{{Season: 1, EpisodeCount: 7, WatchedCount: 1}, {Season: 2, EpisodeCount: 9, WatchedCount: 0}}
	if !reflect.DeepEqual(header.Seasons, want) {
		t.Fatalf("seasons = %+v, want %+v", header.Seasons, want)
	}
}

func TestPlaylistHeaderResumeModes_Postgres(t *testing.T) {
	fixture, playlistID := newTwoSeasonFixture(t)
	resumeOf := func() *PlaylistResumeDto {
		t.Helper()
		header, err := fixture.service.GetPlaylistHeaderByID(progressClientID, playlistID)
		if err != nil {
			t.Fatalf("header: %v", err)
		}
		return header.Resume
	}
	firstEpisodeID := fixture.videoIDs["Show S01E01.mkv"]
	firstSeason := 1

	startResume := resumeOf()
	if startResume == nil || startResume.Mode != "start" || startResume.VideoID != firstEpisodeID || *startResume.Season != 1 || *startResume.Episode != 1 || startResume.PositionSeconds != 0 {
		t.Fatalf("start resume = %+v", startResume)
	}

	fixture.saveProgress(t, progressClientID, "Show S01E01.mkv", 100, 100, true)
	fixture.saveProgress(t, progressClientID, "Show S02E01.mkv", 40, 100, false)
	fixture.saveProgress(t, progressClientID, "Show S01E03.mkv", 25, 100, false)
	got := resumeOf()
	if got == nil || got.Mode != "continue" || got.VideoID != fixture.videoIDs["Show S01E03.mkv"] || got.PositionSeconds != 25 || *got.Season != 1 || *got.Episode != 3 {
		t.Fatalf("continue resume = %+v", got)
	}

	if _, err := fixture.service.SetPlaylistWatched(progressClientID, playlistID, nil, true); err != nil {
		t.Fatalf("watch all: %v", err)
	}
	got = resumeOf()
	if got == nil || got.Mode != "rewatch" || got.VideoID != firstEpisodeID || got.PositionSeconds != 0 {
		t.Fatalf("rewatch resume = %+v", got)
	}

	if _, err := fixture.service.SetPlaylistWatched(progressClientID, playlistID, &firstSeason, false); err != nil {
		t.Fatalf("unwatch season 1: %v", err)
	}
	got = resumeOf()
	if got == nil || got.Mode != "start" || got.VideoID != firstEpisodeID {
		t.Fatalf("start after unwatch = %+v", got)
	}
}

func TestPlaylistHeaderResumeIsNullForEmptyPlaylist_Postgres(t *testing.T) {
	fixture, playlistID := newTwoSeasonFixture(t)
	if _, err := fixture.repository.DbContext.GetDatabase().Exec(`DELETE FROM video_playlist_item WHERE playlist_id = $1`, playlistID); err != nil {
		t.Fatalf("empty playlist: %v", err)
	}

	header, err := fixture.service.GetPlaylistHeaderByID(progressClientID, playlistID)
	if err != nil {
		t.Fatalf("header: %v", err)
	}
	if header.Resume != nil || len(header.Seasons) != 0 {
		t.Fatalf("empty playlist header = resume %+v seasons %+v", header.Resume, header.Seasons)
	}
}

func TestSetPlaylistWatchedMarksOnlyTheRequestedSeason_Postgres(t *testing.T) {
	fixture, playlistID := newTwoSeasonFixture(t)
	fixture.saveProgress(t, progressClientID, "Show S02E02.mkv", 10, 600, false)
	season := 2

	updatedCount, err := fixture.service.SetPlaylistWatched(progressClientID, playlistID, &season, true)
	if err != nil || updatedCount != 8 {
		t.Fatalf("updated = %d err = %v, want 8", updatedCount, err)
	}
	want := []string{"Show S02E01.mkv", "Show S02E02.mkv", "Show S02E03.mkv", "Show S02E04.mkv", "Show S02E05.mkv", "Show S02E06.mkv", "Show S02E07.mkv", "Show S02E08.mkv"}
	if got := fixture.completedFileNames(t, progressClientID); !reflect.DeepEqual(got, want) {
		t.Fatalf("completed = %v, want %v", got, want)
	}
	if got := fixture.completedFileNames(t, "phone"); len(got) != 0 {
		t.Fatalf("other client must be untouched, got %v", got)
	}

	storedProgress, err := fixture.repository.GetVideoWatchProgress(progressClientID, fixture.videoIDs["Show S02E02.mkv"])
	if err != nil || storedProgress.PositionSeconds != 600 || storedProgress.DurationSeconds != 600 {
		t.Fatalf("known duration must be kept as position, got %+v err=%v", storedProgress, err)
	}
}

func TestSetPlaylistWatchedHonorsOverriddenSeason_Postgres(t *testing.T) {
	fixture, playlistID := newTwoSeasonFixture(t)
	fixture.overrideEpisode(t, "Show S01E03.mkv", 2, 9)
	season := 2

	updatedCount, err := fixture.service.SetPlaylistWatched(progressClientID, playlistID, &season, true)
	if err != nil || updatedCount != 9 {
		t.Fatalf("updated = %d err = %v, want 9", updatedCount, err)
	}
	if got := fixture.completedFileNames(t, progressClientID); !reflect.DeepEqual(got, []string{"Show S01E03.mkv", "Show S02E01.mkv", "Show S02E02.mkv", "Show S02E03.mkv", "Show S02E04.mkv", "Show S02E05.mkv", "Show S02E06.mkv", "Show S02E07.mkv", "Show S02E08.mkv"}) {
		t.Fatalf("completed = %v", got)
	}
}

func TestSetPlaylistWatchedWholeSeriesAndUnwatch_Postgres(t *testing.T) {
	fixture, playlistID := newTwoSeasonFixture(t)

	updatedCount, err := fixture.service.SetPlaylistWatched(progressClientID, playlistID, nil, true)
	if err != nil || updatedCount != 16 {
		t.Fatalf("updated = %d err = %v, want 16", updatedCount, err)
	}
	if got := fixture.completedFileNames(t, progressClientID); len(got) != 16 {
		t.Fatalf("completed = %v, want 9", got)
	}

	updatedCount, err = fixture.service.SetPlaylistWatched(progressClientID, playlistID, nil, false)
	if err != nil || updatedCount != 16 {
		t.Fatalf("unwatch updated = %d err = %v, want 16", updatedCount, err)
	}
	if got := fixture.completedFileNames(t, progressClientID); len(got) != 0 {
		t.Fatalf("completed after unwatch = %v", got)
	}
	storedProgress, err := fixture.repository.GetVideoWatchProgress(progressClientID, fixture.videoIDs["Show S01E01.mkv"])
	if err != nil || storedProgress.Completed || storedProgress.PositionSeconds != 0 {
		t.Fatalf("unwatched progress = %+v err=%v", storedProgress, err)
	}
}

func TestSetPlaylistWatchedRejectsOwnerAndUnknownPlaylists_Postgres(t *testing.T) {
	repository := NewRepository(testutil.NewPostgresDB(t, "kuranas_video_it"))
	service := NewService(repository).(*Service)
	seedVideoFiles(t, repository, []seededVideoFile{{name: "a.mkv", folder: "/data/A"}})
	ownerPlaylistID := seedPlaylists(t, repository, []seededPlaylist{{sourcePath: "/mine", isAuto: false}})["/mine"]

	if _, err := service.SetPlaylistWatched(progressClientID, ownerPlaylistID, nil, true); !errors.Is(err, ErrPlaylistNotAutomaticSeries) {
		t.Fatalf("owner playlist err = %v, want ErrPlaylistNotAutomaticSeries", err)
	}
	if _, err := service.SetPlaylistWatched(progressClientID, 987654, nil, true); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unknown playlist err = %v, want sql.ErrNoRows", err)
	}
}
