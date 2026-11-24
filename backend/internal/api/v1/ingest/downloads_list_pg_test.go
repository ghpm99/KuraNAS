package ingest

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/database"
)

type seededDownload struct {
	state     string
	createdAt time.Time
	finalPath string
}

func seedDownloadRow(t *testing.T, dbContext *database.DbContext, download seededDownload) int {
	t.Helper()
	var downloadID int
	err := dbContext.GetDatabase().QueryRow(
		`INSERT INTO ingest_download (url, preset, target_root_id, target_root_name, subfolder, state, final_path, created_at)
		 VALUES ('https://x.test/v', 'video_1080', 2, 'Videos', 'Shows', $1, $2, $3) RETURNING id`,
		download.state, download.finalPath, download.createdAt,
	).Scan(&downloadID)
	if err != nil {
		t.Fatalf("seed download: %v", err)
	}
	return downloadID
}

func seedLibraryFile(t *testing.T, dbContext *database.DbContext, path string, deletedAt sql.NullTime) int {
	t.Helper()
	var fileID int
	err := dbContext.GetDatabase().QueryRow(
		`INSERT INTO home_file (name, path, parent_path, format, size, updated_at, created_at, deleted_at)
		 VALUES ('f', $1, '/media', '.mp4', 1, now(), now(), $2) RETURNING id`, path, deletedAt,
	).Scan(&fileID)
	if err != nil {
		t.Fatalf("seed file: %v", err)
	}
	return fileID
}

func clearDownloadsAndFiles(t *testing.T, dbContext *database.DbContext) {
	t.Helper()
	for _, statement := range []string{`DELETE FROM ingest_download`, `DELETE FROM home_file`} {
		if _, err := dbContext.GetDatabase().Exec(statement); err != nil {
			t.Fatalf("clear tables: %v", err)
		}
	}
}

func downloadIDs(downloads []DownloadDto) []int {
	ids := make([]int, 0, len(downloads))
	for _, download := range downloads {
		ids = append(ids, download.ID)
	}
	return ids
}

func assertIDs(t *testing.T, got []int, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("ids = %v, want %v", got, want)
		}
	}
}

func TestDownloadRepositoryListDownloads_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_ingest_download_list_it")
	repository := NewDownloadRepository(dbContext)
	clearDownloadsAndFiles(t, dbContext)
	now := time.Now().UTC().Truncate(time.Second)

	oldestRunning := seedDownloadRow(t, dbContext, seededDownload{state: "running", createdAt: now.Add(-3 * time.Hour)})
	tiedFirst := seedDownloadRow(t, dbContext, seededDownload{state: "queued", createdAt: now.Add(-time.Hour)})
	tiedSecond := seedDownloadRow(t, dbContext, seededDownload{state: "failed", createdAt: now.Add(-time.Hour)})
	newestCompleted := seedDownloadRow(t, dbContext, seededDownload{state: "completed", createdAt: now})

	t.Run("newest first with id tiebreak", func(t *testing.T) {
		downloads, err := repository.ListDownloads("", 1, 10)
		if err != nil {
			t.Fatalf("ListDownloads: %v", err)
		}
		assertIDs(t, downloadIDs(downloads), []int{newestCompleted, tiedSecond, tiedFirst, oldestRunning})
	})

	t.Run("active filter", func(t *testing.T) {
		downloads, err := repository.ListDownloads(DownloadStateFilterActive, 1, 10)
		if err != nil {
			t.Fatalf("ListDownloads: %v", err)
		}
		assertIDs(t, downloadIDs(downloads), []int{tiedFirst, oldestRunning})
	})

	t.Run("finished filter", func(t *testing.T) {
		downloads, err := repository.ListDownloads(DownloadStateFilterFinished, 1, 10)
		if err != nil {
			t.Fatalf("ListDownloads: %v", err)
		}
		assertIDs(t, downloadIDs(downloads), []int{newestCompleted, tiedSecond})
	})

	t.Run("pagination returns one extra row", func(t *testing.T) {
		firstPage, err := repository.ListDownloads("", 1, 3)
		if err != nil {
			t.Fatalf("ListDownloads: %v", err)
		}
		assertIDs(t, downloadIDs(firstPage), []int{newestCompleted, tiedSecond, tiedFirst, oldestRunning})
		secondPage, err := repository.ListDownloads("", 2, 3)
		if err != nil {
			t.Fatalf("ListDownloads: %v", err)
		}
		assertIDs(t, downloadIDs(secondPage), []int{oldestRunning})
	})
}

func TestDownloadRepositoryListDownloadsResolvesFileID_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_ingest_download_file_id_it")
	repository := NewDownloadRepository(dbContext)
	clearDownloadsAndFiles(t, dbContext)
	now := time.Now().UTC()

	indexedPath := "/media/Videos/Shows/indexed.mp4"
	deletedPath := "/media/Videos/Shows/deleted.mp4"
	notIndexedPath := "/media/Videos/Shows/not-indexed.mp4"
	indexedFileID := seedLibraryFile(t, dbContext, indexedPath, sql.NullTime{})
	seedLibraryFile(t, dbContext, deletedPath, sql.NullTime{Time: now, Valid: true})

	indexed := seedDownloadRow(t, dbContext, seededDownload{state: "completed", createdAt: now.Add(-3 * time.Minute), finalPath: indexedPath})
	deleted := seedDownloadRow(t, dbContext, seededDownload{state: "completed", createdAt: now.Add(-2 * time.Minute), finalPath: deletedPath})
	notIndexed := seedDownloadRow(t, dbContext, seededDownload{state: "completed", createdAt: now.Add(-time.Minute), finalPath: notIndexedPath})
	running := seedDownloadRow(t, dbContext, seededDownload{state: "running", createdAt: now})

	downloads, err := repository.ListDownloads("", 1, 10)
	if err != nil {
		t.Fatalf("ListDownloads: %v", err)
	}
	fileIDsByDownloadID := map[int]*int{}
	for _, download := range downloads {
		fileIDsByDownloadID[download.ID] = download.FileID
	}
	if fileID := fileIDsByDownloadID[indexed]; fileID == nil || *fileID != indexedFileID {
		t.Fatalf("indexed file_id = %v, want %d", fileID, indexedFileID)
	}
	for name, downloadID := range map[string]int{"deleted": deleted, "not indexed": notIndexed, "running": running} {
		if fileIDsByDownloadID[downloadID] != nil {
			t.Fatalf("%s file_id = %d, want nil", name, *fileIDsByDownloadID[downloadID])
		}
	}
}
