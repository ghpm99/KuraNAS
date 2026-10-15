package scan

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/testutil"
	"nas-go/api/pkg/utils"
)

func TestUpdateFileRecordKeepsStarAndIdWhenSnapshotIsStale_Postgres(t *testing.T) {
	dbContext := testutil.NewPostgresDB(t, "kuranas_worker_it")
	repo := files.NewRepository(dbContext)
	service := &files.Service{Repository: repo, Tasks: make(chan utils.Task, 4)}

	err := dbContext.ExecTx(func(tx *sql.Tx) error {
		_, truncateErr := tx.Exec("TRUNCATE home_file RESTART IDENTITY CASCADE")
		return truncateErr
	})
	if err != nil {
		t.Fatalf("truncate home_file: %v", err)
	}

	var createdID int
	err = dbContext.ExecTx(func(tx *sql.Tx) error {
		created, createErr := repo.CreateFile(tx, files.FileModel{
			Name: "clip.txt", Path: "/srv/clip.txt", ParentPath: "/srv", Format: ".txt", Size: 1,
			UpdatedAt: time.Now(), CreatedAt: time.Now(), Type: files.File,
		})
		createdID = created.ID
		return createErr
	})
	if err != nil {
		t.Fatalf("create row: %v", err)
	}

	staleSnapshot, err := service.GetFileById(createdID)
	if err != nil {
		t.Fatalf("GetFileById: %v", err)
	}
	if staleSnapshot.Starred {
		t.Fatal("snapshot must be taken before starring")
	}

	if _, err := repo.SetFileStarred(createdID, true); err != nil {
		t.Fatalf("SetFileStarred: %v", err)
	}
	var starredAtBefore sql.NullTime
	readStarredAt := func() sql.NullTime {
		var starredAt sql.NullTime
		row := dbContext.GetDatabase().QueryRow("SELECT starred_at FROM home_file WHERE id = $1", createdID)
		if scanErr := row.Scan(&starredAt); scanErr != nil {
			t.Fatalf("read starred_at: %v", scanErr)
		}
		return starredAt
	}
	starredAtBefore = readStarredAt()

	finalizedFile := files.FileDto{
		Format: ".txt", Size: 2048, UpdatedAt: time.Now(), CreatedAt: staleSnapshot.CreatedAt, CheckSum: "abc123",
	}
	if _, err := UpdateFileRecord(service, finalizedFile, staleSnapshot); err != nil {
		t.Fatalf("UpdateFileRecord: %v", err)
	}

	reloaded, err := service.GetFileById(createdID)
	if err != nil {
		t.Fatalf("GetFileById after update: %v", err)
	}
	if reloaded.ID != createdID || reloaded.Size != 2048 || reloaded.CheckSum != "abc123" {
		t.Fatalf("update not applied or id changed: %+v", reloaded)
	}
	if !reloaded.Starred {
		t.Fatal("worker update with stale Starred=false snapshot cleared the star")
	}
	if starredAtAfter := readStarredAt(); !starredAtAfter.Valid || !starredAtAfter.Time.Equal(starredAtBefore.Time) {
		t.Fatalf("starred_at changed: before=%v after=%v", starredAtBefore, starredAtAfter)
	}
}
