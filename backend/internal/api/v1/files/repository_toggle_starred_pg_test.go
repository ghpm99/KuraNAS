package files

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"nas-go/api/internal/testutil"
)

func findFileByPath(t *testing.T, repo *Repository, path string) FileModel {
	t.Helper()
	files, err := repo.GetFilesByNameAndPath("a.txt", path, 1)
	if err != nil || len(files) != 1 {
		t.Fatalf("file %s not found: %v", path, err)
	}
	file := files[0]
	return file
}

func TestPostgres_ToggleFileStarredTwiceRestoresOriginal(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	insertFileRow(t, repo, "a.txt", "/srv/a.txt", "/srv", 10, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	original := findFileByPath(t, repo, "/srv/a.txt")

	firstValue, err := repo.ToggleFileStarred(original.ID)
	if err != nil || firstValue == original.Starred {
		t.Fatalf("first toggle: value=%v err=%v", firstValue, err)
	}
	secondValue, err := repo.ToggleFileStarred(original.ID)
	if err != nil || secondValue != original.Starred {
		t.Fatalf("second toggle: value=%v err=%v", secondValue, err)
	}
}

func TestPostgres_ToggleFileStarredLeavesOtherColumnsUntouched(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	insertFileRow(t, repo, "a.txt", "/srv/a.txt", "/srv", 10, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	before := findFileByPath(t, repo, "/srv/a.txt")

	if _, err := repo.ToggleFileStarred(before.ID); err != nil {
		t.Fatalf("toggle: %v", err)
	}

	after := findFileByPath(t, repo, "/srv/a.txt")
	if after.Name != before.Name || after.Path != before.Path || after.CheckSum != before.CheckSum || !after.UpdatedAt.Equal(before.UpdatedAt) {
		t.Fatalf("toggle changed other columns: before=%+v after=%+v", before, after)
	}
	if after.Starred == before.Starred {
		t.Fatalf("expected starred to change")
	}
}

func TestPostgres_ToggleFileStarredIgnoresSoftDeletedFile(t *testing.T) {
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

	if _, err := repo.ToggleFileStarred(file.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	var starred bool
	row := repo.GetDbContext().GetDatabase().QueryRow("SELECT starred FROM home_file WHERE id = $1", file.ID)
	if err := row.Scan(&starred); err != nil || starred != file.Starred {
		t.Fatalf("soft-deleted row changed: starred=%v err=%v", starred, err)
	}
}

func TestPostgres_UpdateFileNeverOverwritesStarred(t *testing.T) {
	repo := NewRepository(testutil.NewPostgresDB(t, "kuranas_files_it"))
	truncateHomeFile(t, repo)
	insertFileRow(t, repo, "a.txt", "/srv/a.txt", "/srv", 10, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	starredFile := findFileByPath(t, repo, "/srv/a.txt")
	if _, err := repo.ToggleFileStarred(starredFile.ID); err != nil {
		t.Fatalf("toggle: %v", err)
	}

	staleSnapshot := starredFile
	staleSnapshot.Starred = false
	staleSnapshot.Size = 999
	staleSnapshot.CheckSum = "changed-checksum"
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, updateErr := repo.UpdateFile(tx, staleSnapshot)
		return updateErr
	})
	if err != nil {
		t.Fatalf("update file: %v", err)
	}

	reloaded := findFileByPath(t, repo, "/srv/a.txt")
	if reloaded.Size != 999 || reloaded.CheckSum != "changed-checksum" {
		t.Fatalf("expected changed columns to persist, got %+v", reloaded)
	}
	if !reloaded.Starred {
		t.Fatalf("UpdateFile must not overwrite starred")
	}
}
