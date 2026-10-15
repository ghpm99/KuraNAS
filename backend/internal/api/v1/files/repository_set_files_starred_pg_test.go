package files

import (
	"database/sql"
	"reflect"
	"sort"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

func insertFilesForBatchStarring(t *testing.T, repo *Repository, paths ...string) []int {
	t.Helper()
	fileIDs := make([]int, 0, len(paths))
	for _, path := range paths {
		insertFileRow(t, repo, "a.txt", path, "/srv", 10, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
		fileIDs = append(fileIDs, findFileByPath(t, repo, path).ID)
	}
	return fileIDs
}

func readStarred(t *testing.T, repo *Repository, id int) bool {
	t.Helper()
	var isStarred bool
	row := repo.GetDbContext().GetDatabase().QueryRow("SELECT starred FROM home_file WHERE id = $1", id)
	if err := row.Scan(&isStarred); err != nil {
		t.Fatalf("read starred: %v", err)
	}
	return isStarred
}

func TestPostgres_SetFilesStarredMarksSeveralIgnoringSoftDeletedAndMissing(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	fileIDs := insertFilesForBatchStarring(t, repo, "/srv/1", "/srv/2", "/srv/3")
	softDeleteErr := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("UPDATE home_file SET deleted_at = now(), starred = false WHERE id = $1", fileIDs[2])
		return err
	})
	if softDeleteErr != nil {
		t.Fatalf("soft delete: %v", softDeleteErr)
	}
	missingID := fileIDs[2] + 1000

	updatedIDs, err := repo.SetFilesStarred([]int{fileIDs[0], fileIDs[1], fileIDs[2], missingID}, true)
	if err != nil {
		t.Fatalf("set starred: %v", err)
	}

	sort.Ints(updatedIDs)
	if !reflect.DeepEqual(updatedIDs, []int{fileIDs[0], fileIDs[1]}) {
		t.Fatalf("unexpected updated ids %v", updatedIDs)
	}
	if !readStarred(t, repo, fileIDs[0]) || !readStarred(t, repo, fileIDs[1]) {
		t.Fatal("active files must be starred")
	}
	if readStarred(t, repo, fileIDs[2]) {
		t.Fatal("soft-deleted file must not change")
	}
}

func TestPostgres_SetFilesStarredIsIdempotent(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	fileIDs := insertFilesForBatchStarring(t, repo, "/srv/1", "/srv/2")

	for attempt := 0; attempt < 2; attempt++ {
		updatedIDs, err := repo.SetFilesStarred(fileIDs, true)
		if err != nil || len(updatedIDs) != 2 {
			t.Fatalf("set true attempt %d: ids=%v err=%v", attempt, updatedIDs, err)
		}
	}
	if _, err := repo.SetFilesStarred(fileIDs, false); err != nil {
		t.Fatalf("set false: %v", err)
	}
	if readStarred(t, repo, fileIDs[0]) || readStarred(t, repo, fileIDs[1]) {
		t.Fatal("files must be unstarred")
	}
}
