package music

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	queries "nas-go/api/pkg/database/queries/music"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAddPlaylistTrackReportsDuplicateAsAlreadyInPlaylist(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	expectPlaylistLock(mock, 10)
	mock.ExpectQuery(regexp.QuoteMeta(queries.AddPlaylistTrackQuery)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "position", "added_at"}))
	mock.ExpectRollback()

	tx, _ := db.Begin()
	_, err := repo.AddPlaylistTrack(tx, 10, 20)
	_ = tx.Rollback()

	if !errors.Is(err, ErrTrackAlreadyInPlaylist) {
		t.Fatalf("expected ErrTrackAlreadyInPlaylist, got %v", err)
	}
}

func TestPlaylistWritesFailAsNotFoundWhenPlaylistCannotBeLocked(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()

	operations := map[string]func(tx *sql.Tx) error{
		"add": func(tx *sql.Tx) error {
			_, err := repo.AddPlaylistTrack(tx, 10, 20)
			return err
		},
		"remove":  func(tx *sql.Tx) error { return repo.RemovePlaylistTrack(tx, 10, 20) },
		"reorder": func(tx *sql.Tx) error { return repo.ReorderPlaylistTrack(tx, 10, 20, 1) },
	}

	for name, operation := range operations {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(queries.LockPlaylistQuery)).
			WithArgs(10).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectRollback()

		tx, _ := db.Begin()
		err := operation(tx)
		_ = tx.Rollback()

		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("%s: expected sql.ErrNoRows, got %v", name, err)
		}
	}
}

func TestReorderPlaylistTrackFailsWhenTrackIsNotInPlaylist(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	expectPlaylistLock(mock, 10)
	mock.ExpectExec(regexp.QuoteMeta(queries.ReorderPlaylistTrackQuery)).
		WithArgs(2, 10, 20).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	tx, _ := db.Begin()
	err := repo.ReorderPlaylistTrack(tx, 10, 20, 2)
	_ = tx.Rollback()

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestRemovePlaylistTrackFailsWhenCompactionFails(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	expectPlaylistLock(mock, 10)
	mock.ExpectExec(regexp.QuoteMeta(queries.RemovePlaylistTrackQuery)).
		WithArgs(10, 20).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.CompactPlaylistPositionsQuery)).
		WithArgs(10).
		WillReturnError(errors.New("compact failed"))
	mock.ExpectRollback()

	tx, _ := db.Begin()
	err := repo.RemovePlaylistTrack(tx, 10, 20)
	_ = tx.Rollback()

	if err == nil {
		t.Fatalf("expected compaction error")
	}
}

func TestGetPlaylistsPassesTrimmedNameSearch(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlaylistsQuery)).
		WithArgs(11, 0, "road").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_system", "created_at", "updated_at", "track_count", "is_ai_generated"}))
	mock.ExpectRollback()

	if _, err := repo.GetPlaylists(1, 10, "  road "); err != nil {
		t.Fatalf("GetPlaylists failed: %v", err)
	}
}
