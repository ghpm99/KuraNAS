package video

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"nas-go/api/internal/testutil"
)

const (
	starredVideoName   = "starred.mp4"
	unstarredVideoName = "plain.mp4"
	starredVideoFolder = "/data/starred"
)

func seedStarredAndUnstarredVideos(t *testing.T, repository *Repository) (starredID int, unstarredID int) {
	t.Helper()
	idsByName := seedVideoFiles(t, repository, []seededVideoFile{
		{name: starredVideoName, folder: starredVideoFolder},
		{name: unstarredVideoName, folder: starredVideoFolder},
	})
	starredID = idsByName[starredVideoName]
	unstarredID = idsByName[unstarredVideoName]

	seedErr := repository.DbContext.ExecTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`UPDATE home_file SET starred = TRUE WHERE id = $1`, starredID); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE home_file SET starred = NULL WHERE id = $1`, unstarredID); err != nil {
			return err
		}
		if _, err := tx.Exec(`TRUNCATE video_metadata RESTART IDENTITY`); err != nil {
			return err
		}
		for _, fileID := range []int{starredID, unstarredID} {
			if _, err := tx.Exec(`INSERT INTO video_metadata (file_id, path, classification) VALUES ($1, $2, 'movie')`, fileID, fmt.Sprintf("/m/%d", fileID)); err != nil {
				return err
			}
		}
		return nil
	})
	if seedErr != nil {
		t.Fatalf("seed starred: %v", seedErr)
	}
	return starredID, unstarredID
}

func assertStarredByVideoID(t *testing.T, source string, starredByVideoID map[int]bool, starredID int, unstarredID int) {
	t.Helper()
	if len(starredByVideoID) != 2 {
		t.Fatalf("%s: expected 2 videos, got %v", source, starredByVideoID)
	}
	if !starredByVideoID[starredID] {
		t.Fatalf("%s: starred video reported as not starred", source)
	}
	if starredByVideoID[unstarredID] {
		t.Fatalf("%s: unstarred video reported as starred", source)
	}
}

func starredByVideoIDOf(models []VideoFileModel) map[int]bool {
	starredByVideoID := map[int]bool{}
	for _, model := range models {
		starredByVideoID[model.ID] = model.Starred
	}
	return starredByVideoID
}

func TestVideoQueriesExposeStarred_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	starredID, unstarredID := seedStarredAndUnstarredVideos(t, repository)

	libraryPage, err := repository.ListLibraryVideos(LibraryVideosRequest{
		Ordering: LibraryVideoOrdering{Sort: LibraryVideoSortName, Direction: SortDirectionAscending},
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListLibraryVideos: %v", err)
	}
	assertStarredByVideoID(t, "library", starredByVideoIDOf(libraryPage.Items), starredID, unstarredID)

	folderVideos, err := repository.ListLibraryFolderVideos(starredVideoFolder, 10, 0)
	if err != nil {
		t.Fatalf("ListLibraryFolderVideos: %v", err)
	}
	assertStarredByVideoID(t, "folder videos", starredByVideoIDOf(folderVideos), starredID, unstarredID)

	catalogVideos, err := repository.GetCatalogVideos(10)
	if err != nil {
		t.Fatalf("GetCatalogVideos: %v", err)
	}
	assertStarredByVideoID(t, "catalog", starredByVideoIDOf(catalogVideos), starredID, unstarredID)

	recentVideos, err := repository.GetRecentVideos(10)
	if err != nil {
		t.Fatalf("GetRecentVideos: %v", err)
	}
	assertStarredByVideoID(t, "recent", starredByVideoIDOf(recentVideos), starredID, unstarredID)

	unassignedVideos, err := repository.GetUnassignedVideos(10)
	if err != nil {
		t.Fatalf("GetUnassignedVideos: %v", err)
	}
	assertStarredByVideoID(t, "unassigned", starredByVideoIDOf(unassignedVideos), starredID, unstarredID)

	videosInParent, err := repository.GetVideosByParentPath(starredVideoFolder)
	if err != nil {
		t.Fatalf("GetVideosByParentPath: %v", err)
	}
	assertStarredByVideoID(t, "parent path", starredByVideoIDOf(videosInParent), starredID, unstarredID)

	for _, movieSort := range []LibraryMovieSort{LibraryMovieSortName, LibraryMovieSortRecent} {
		movies, err := repository.ListLibraryMovies(movieSort, 10, 0)
		if err != nil {
			t.Fatalf("ListLibraryMovies %s: %v", movieSort, err)
		}
		assertStarredByVideoID(t, "movies "+string(movieSort), starredByVideoIDOf(movies), starredID, unstarredID)
	}

	starredVideo, err := repository.GetVideoFileByID(starredID)
	if err != nil || !starredVideo.Starred {
		t.Fatalf("GetVideoFileByID starred: %+v err=%v", starredVideo, err)
	}
	unstarredVideo, err := repository.GetVideoFileByID(unstarredID)
	if err != nil || unstarredVideo.Starred {
		t.Fatalf("GetVideoFileByID unstarred: %+v err=%v", unstarredVideo, err)
	}
}

func TestPlaylistItemsAndContinueWatchingExposeStarred_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_video_it")
	repository := NewRepository(dbContext)
	starredID, unstarredID := seedStarredAndUnstarredVideos(t, repository)
	playlistID := seedPlaylists(t, repository, []seededPlaylist{{sourcePath: "/starred", isAuto: false}})["/starred"]

	setupErr := dbContext.ExecTx(func(tx *sql.Tx) error {
		if err := repository.InsertPlaylistItemsWithSource(tx, playlistID, []int{starredID, unstarredID}, "manual"); err != nil {
			return err
		}
		for _, fileID := range []int{starredID, unstarredID} {
			_, err := repository.UpsertVideoWatchProgress(tx, VideoWatchProgressModel{ClientID: "tv", VideoID: fileID, PositionSeconds: 10, DurationSeconds: 100})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if setupErr != nil {
		t.Fatalf("setup: %v", setupErr)
	}

	playlistItemReaders := map[string]func() ([]VideoPlaylistItemModel, error){
		"playlist items":          func() ([]VideoPlaylistItemModel, error) { return repository.GetPlaylistItems(playlistID) },
		"playlist items detailed": func() ([]VideoPlaylistItemModel, error) { return repository.GetVideoPlaylistItemsDetailed(playlistID) },
		"playlist items page": func() ([]VideoPlaylistItemModel, error) {
			return repository.GetVideoPlaylistItemsPage(playlistID, 10, 0)
		},
	}
	for source, readItems := range playlistItemReaders {
		items, err := readItems()
		if err != nil {
			t.Fatalf("%s: %v", source, err)
		}
		starredByVideoID := map[int]bool{}
		for _, item := range items {
			starredByVideoID[item.VideoID] = item.Video.Starred
		}
		assertStarredByVideoID(t, source, starredByVideoID, starredID, unstarredID)
	}

	continueWatching, err := repository.GetContinueWatchingVideos("tv", 10)
	if err != nil {
		t.Fatalf("GetContinueWatchingVideos: %v", err)
	}
	starredByVideoID := map[int]bool{}
	for _, item := range continueWatching {
		starredByVideoID[item.ID] = item.Starred
	}
	assertStarredByVideoID(t, "continue watching", starredByVideoID, starredID, unstarredID)
}

func TestVideoFileDtoSerializesStarred(t *testing.T) {
	for _, isStarred := range []bool{true, false} {
		model := VideoFileModel{ID: 1, Starred: isStarred}
		dto := model.ToDto()
		payload, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		expectedFragment := `"starred":false`
		if isStarred {
			expectedFragment = `"starred":true`
		}
		if !strings.Contains(string(payload), expectedFragment) {
			t.Fatalf("expected %s in %s", expectedFragment, payload)
		}
	}
}
