package files

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

func TestPostgres_SetFileStarredIsIdempotent(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	insertFileRow(t, repo, "a.txt", "/srv/a.txt", "/srv", 10, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	file := findFileByPath(t, repo, "/srv/a.txt")

	for attempt := 0; attempt < 2; attempt++ {
		isStarred, err := repo.SetFileStarred(file.ID, true)
		if err != nil || !isStarred {
			t.Fatalf("set true attempt %d: value=%v err=%v", attempt, isStarred, err)
		}
	}
	isStarred, err := repo.SetFileStarred(file.ID, false)
	if err != nil || isStarred {
		t.Fatalf("set false: value=%v err=%v", isStarred, err)
	}
}

func TestPostgres_SetFileStarredIgnoresSoftDeletedFile(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	insertFileRow(t, repo, "a.txt", "/srv/a.txt", "/srv", 10, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	file := findFileByPath(t, repo, "/srv/a.txt")
	softDeleteErr := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("UPDATE home_file SET deleted_at = now() WHERE id = $1", file.ID)
		return err
	})
	if softDeleteErr != nil {
		t.Fatalf("soft delete: %v", softDeleteErr)
	}

	if _, err := repo.SetFileStarred(file.ID, !file.Starred); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	var starred bool
	row := repo.GetDbContext().GetDatabase().QueryRow("SELECT starred FROM home_file WHERE id = $1", file.ID)
	if err := row.Scan(&starred); err != nil || starred != file.Starred {
		t.Fatalf("soft-deleted row changed: starred=%v err=%v", starred, err)
	}
}
