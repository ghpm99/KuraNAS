package files

import (
	"database/sql"
	"slices"
	"testing"

	"nas-go/api/internal/testutil"
)

func remainingPaths(t *testing.T, repo *Repository) []string {
	t.Helper()
	rows, err := repo.GetDbContext().GetDatabase().Query("SELECT path FROM home_file ORDER BY path")
	if err != nil {
		t.Fatalf("list paths: %v", err)
	}
	defer rows.Close()
	paths := []string{}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			t.Fatalf("scan path: %v", err)
		}
		paths = append(paths, path)
	}
	return paths
}

func deleteDeletedSubtree(t *testing.T, repo *Repository, path string) int64 {
	t.Helper()
	var removedRows int64
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		removed, deleteErr := repo.DeleteDeletedSubtree(tx, path)
		removedRows = removed
		return deleteErr
	})
	if err != nil {
		t.Fatalf("DeleteDeletedSubtree %q: %v", path, err)
	}
	return removedRows
}

func TestPostgres_DeleteDeletedSubtreeRemovesOnlySoftDeletedRowsOfTheSubtree(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	insertSearchRow(t, repo, "docs", "/srv/docs", "/srv", Directory)
	insertSearchRow(t, repo, "old.txt", "/srv/docs/old.txt", "/srv/docs", File)
	insertSearchRow(t, repo, "recreated.txt", "/srv/docs/recreated.txt", "/srv/docs", File)
	insertSearchRow(t, repo, "docs2", "/srv/docs2", "/srv", Directory)
	insertSearchRow(t, repo, "neighbor.txt", "/srv/docs2/neighbor.txt", "/srv/docs2", File)
	for _, deletedPath := range []string{"/srv/docs", "/srv/docs/old.txt", "/srv/docs2", "/srv/docs2/neighbor.txt"} {
		softDeleteRow(t, repo, deletedPath)
	}

	removedRows := deleteDeletedSubtree(t, repo, "/srv/docs")

	expectedPaths := []string{"/srv/docs/recreated.txt", "/srv/docs2", "/srv/docs2/neighbor.txt"}
	if removedRows != 2 || !slices.Equal(remainingPaths(t, repo), expectedPaths) {
		t.Fatalf("removed=%d remaining=%v, want 2 and %v", removedRows, remainingPaths(t, repo), expectedPaths)
	}
}

func TestPostgres_DeleteDeletedSubtreeCascadesToRecentAccess(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	insertSearchRow(t, repo, "a.txt", "/srv/a.txt", "/srv", File)
	file := findFileByPath(t, repo, "/srv/a.txt")
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, insertErr := tx.Exec("INSERT INTO recent_file (ip_address, file_id) VALUES ('10.0.0.1', $1)", file.ID)
		return insertErr
	})
	if err != nil {
		t.Fatalf("insert recent access: %v", err)
	}
	softDeleteRow(t, repo, "/srv/a.txt")

	deleteDeletedSubtree(t, repo, "/srv/a.txt")

	var recentAccessCount int
	row := repo.GetDbContext().GetDatabase().QueryRow("SELECT COUNT(*) FROM recent_file WHERE file_id = $1", file.ID)
	if err := row.Scan(&recentAccessCount); err != nil || recentAccessCount != 0 {
		t.Fatalf("recent access must cascade: count=%d err=%v", recentAccessCount, err)
	}
}
