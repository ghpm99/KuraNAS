package files

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

func readStarredAt(t *testing.T, repo *Repository, id int) sql.NullTime {
	t.Helper()
	var starredAt sql.NullTime
	row := repo.GetDbContext().GetDatabase().QueryRow("SELECT starred_at FROM home_file WHERE id = $1", id)
	if err := row.Scan(&starredAt); err != nil {
		t.Fatalf("read starred_at: %v", err)
	}
	return starredAt
}

func setStarredAt(t *testing.T, repo *Repository, id int, starredAt time.Time) {
	t.Helper()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec("UPDATE home_file SET starred = true, starred_at = $2 WHERE id = $1", id, starredAt)
		return execErr
	})
	if err != nil {
		t.Fatalf("set starred_at: %v", err)
	}
}

func assertStarredAtRecent(t *testing.T, starredAt sql.NullTime) {
	t.Helper()
	if !starredAt.Valid || time.Since(starredAt.Time) > time.Minute || time.Until(starredAt.Time) > time.Minute {
		t.Fatalf("starred_at must be about now, got %v", starredAt)
	}
}

func TestPostgres_ToggleFileStarredMaintainsStarredAt(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	insertFilesForBatchStarring(t, repo, "/srv/1")
	fileID := findFileByPath(t, repo, "/srv/1").ID

	if _, err := repo.ToggleFileStarred(fileID); err != nil {
		t.Fatalf("toggle to true: %v", err)
	}
	assertStarredAtRecent(t, readStarredAt(t, repo, fileID))

	if _, err := repo.ToggleFileStarred(fileID); err != nil {
		t.Fatalf("toggle to false: %v", err)
	}
	if readStarredAt(t, repo, fileID).Valid {
		t.Fatal("starred_at must be cleared when unstarred")
	}
}

func TestPostgres_SetFileStarredPreservesFirstStarredAt(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	fileID := insertFilesForBatchStarring(t, repo, "/srv/1")[0]

	if _, err := repo.SetFileStarred(fileID, true); err != nil {
		t.Fatalf("first set true: %v", err)
	}
	assertStarredAtRecent(t, readStarredAt(t, repo, fileID))

	pastStarredAt := time.Date(2020, 5, 6, 7, 8, 9, 0, time.UTC)
	setStarredAt(t, repo, fileID, pastStarredAt)
	if _, err := repo.SetFileStarred(fileID, true); err != nil {
		t.Fatalf("second set true: %v", err)
	}
	if starredAt := readStarredAt(t, repo, fileID); !starredAt.Valid || !starredAt.Time.Equal(pastStarredAt) {
		t.Fatalf("starred_at must be preserved, got %v", starredAt)
	}

	if _, err := repo.SetFileStarred(fileID, false); err != nil {
		t.Fatalf("set false: %v", err)
	}
	if readStarredAt(t, repo, fileID).Valid {
		t.Fatal("starred_at must be cleared when unstarred")
	}
}

func TestPostgres_SetFilesStarredPreservesFirstStarredAt(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	fileIDs := insertFilesForBatchStarring(t, repo, "/srv/1", "/srv/2")
	pastStarredAt := time.Date(2020, 5, 6, 7, 8, 9, 0, time.UTC)
	setStarredAt(t, repo, fileIDs[0], pastStarredAt)

	if _, err := repo.SetFilesStarred(fileIDs, true); err != nil {
		t.Fatalf("set true: %v", err)
	}
	if starredAt := readStarredAt(t, repo, fileIDs[0]); !starredAt.Valid || !starredAt.Time.Equal(pastStarredAt) {
		t.Fatalf("already starred file must keep starred_at, got %v", starredAt)
	}
	assertStarredAtRecent(t, readStarredAt(t, repo, fileIDs[1]))

	if _, err := repo.SetFilesStarred(fileIDs, false); err != nil {
		t.Fatalf("set false: %v", err)
	}
	if readStarredAt(t, repo, fileIDs[0]).Valid || readStarredAt(t, repo, fileIDs[1]).Valid {
		t.Fatal("starred_at must be cleared when unstarred")
	}
}
