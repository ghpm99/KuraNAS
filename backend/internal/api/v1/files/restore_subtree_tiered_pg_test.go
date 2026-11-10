package files

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

var tieredRestoreSeedMoment = time.Date(2026, 5, 6, 7, 8, 9, 0, time.UTC)

func markRowAsColdAndDeleted(t *testing.T, repo *Repository, path string, coldPath string) {
	t.Helper()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec("UPDATE home_file SET physical_path = $2, deleted_at = now() WHERE path = $1", path, coldPath)
		return execErr
	})
	if err != nil {
		t.Fatalf("mark %q cold and deleted: %v", path, err)
	}
}

func softDeleteRow(t *testing.T, repo *Repository, path string) {
	t.Helper()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, execErr := tx.Exec("UPDATE home_file SET deleted_at = now() WHERE path = $1", path)
		return execErr
	})
	if err != nil {
		t.Fatalf("soft delete %q: %v", path, err)
	}
}

func restoreSubtree(t *testing.T, repo *Repository, path string) {
	t.Helper()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, restoreErr := repo.RestoreSubtree(tx, path)
		return restoreErr
	})
	if err != nil {
		t.Fatalf("RestoreSubtree %q: %v", path, err)
	}
}

func physicalPathOf(t *testing.T, repo *Repository, path string) sql.NullString {
	t.Helper()
	var physicalPath sql.NullString
	row := repo.GetDbContext().GetDatabase().QueryRow(
		"SELECT physical_path FROM home_file WHERE path = $1 AND deleted_at IS NULL", path,
	)
	if err := row.Scan(&physicalPath); err != nil {
		t.Fatalf("read physical_path of %q: %v", path, err)
	}
	return physicalPath
}

func TestPostgres_RestoreSubtreeClearsColdLocationOfTheRestoredFile(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	hotPath := filepath.Join(string(filepath.Separator), "srv", "movie.mkv")
	insertFileRow(t, repo, "movie.mkv", hotPath, filepath.Dir(hotPath), 10, tieredRestoreSeedMoment)
	markRowAsColdAndDeleted(t, repo, hotPath, "/cold/srv/movie.mkv")

	restoreSubtree(t, repo, hotPath)

	if physicalPath := physicalPathOf(t, repo, hotPath); physicalPath.Valid {
		t.Fatalf("restored file still points at the cold copy %q", physicalPath.String)
	}
}

func TestPostgres_RestoreSubtreeKeepsColdLocationOfDescendants(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	folderPath := filepath.Join(string(filepath.Separator), "srv", "shows")
	episodePath := filepath.Join(folderPath, "episode.mkv")
	insertSearchRow(t, repo, "shows", folderPath, filepath.Dir(folderPath), Directory)
	insertFileRow(t, repo, "episode.mkv", episodePath, folderPath, 10, tieredRestoreSeedMoment)
	softDeleteRow(t, repo, folderPath)
	markRowAsColdAndDeleted(t, repo, episodePath, "/cold/srv/shows/episode.mkv")

	restoreSubtree(t, repo, folderPath)

	physicalPath := physicalPathOf(t, repo, episodePath)
	if !physicalPath.Valid || physicalPath.String != "/cold/srv/shows/episode.mkv" {
		t.Fatalf("descendant lost its cold location: %+v", physicalPath)
	}
}
