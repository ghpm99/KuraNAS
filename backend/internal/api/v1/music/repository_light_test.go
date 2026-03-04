package music

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/music"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMusicRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewRepository(database.NewDbContext(db)), mock, db
}

func TestMusicRepositoryBasicsAndReads(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	if repo.GetDbContext() == nil {
		t.Fatalf("expected db context")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlaylistsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_system", "created_at", "updated_at", "track_count"}).
			AddRow(1, "p", "d", false, now, now, 3))
	mock.ExpectRollback()
	playlists, err := repo.GetPlaylists(1, 10)
	if err != nil || len(playlists.Items) != 1 {
		t.Fatalf("GetPlaylists failed len=%d err=%v", len(playlists.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlaylistByIDQuery)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_system", "created_at", "updated_at", "track_count"}).
			AddRow(1, "p", "d", false, now, now, 3))
	mock.ExpectRollback()
	if p, err := repo.GetPlaylistByID(1); err != nil || p.ID != 1 {
		t.Fatalf("GetPlaylistByID failed p=%+v err=%v", p, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetNowPlayingQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_system", "created_at", "updated_at", "track_count"}).
			AddRow(2, "now", "queue", true, now, now, 1))
	mock.ExpectRollback()
	if p, err := repo.GetNowPlaying(); err != nil || p.ID != 2 {
		t.Fatalf("GetNowPlaying failed p=%+v err=%v", p, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlayerStateQuery)).
		WithArgs("c1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "client_id", "playlist_id", "current_file_id", "current_position", "volume", "shuffle", "repeat_mode", "updated_at"}).
			AddRow(1, "c1", nil, nil, 1.5, 0.8, false, "off", now))
	mock.ExpectRollback()
	if s, err := repo.GetPlayerState("c1"); err != nil || s.ClientID != "c1" {
		t.Fatalf("GetPlayerState failed s=%+v err=%v", s, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestMusicRepositoryWritePaths(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CreatePlaylistQuery)).
		WithArgs("p", "d", false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(10, now, now))
	mock.ExpectCommit()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		created, err := repo.CreatePlaylist(tx, "p", "d", false)
		if err != nil {
			return err
		}
		if created.ID != 10 {
			t.Fatalf("expected created ID 10")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("CreatePlaylist failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpdatePlaylistQuery)).
		WithArgs("n", "d2", 10).
		WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		updated, err := repo.UpdatePlaylist(tx, 10, "n", "d2")
		if err != nil {
			return err
		}
		if updated.ID != 10 {
			t.Fatalf("expected updated ID 10")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("UpdatePlaylist failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeletePlaylistQuery)).
		WithArgs(10).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.DeletePlaylist(tx, 10)
	})
	if err != nil {
		t.Fatalf("DeletePlaylist failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.AddPlaylistTrackQuery)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "position", "added_at"}).AddRow(1, 0, now))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		track, err := repo.AddPlaylistTrack(tx, 10, 20)
		if err != nil {
			return err
		}
		if track.ID != 1 {
			t.Fatalf("expected track ID 1")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("AddPlaylistTrack failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.RemovePlaylistTrackQuery)).
		WithArgs(10, 20).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.RemovePlaylistTrack(tx, 10, 20)
	})
	if err != nil {
		t.Fatalf("RemovePlaylistTrack failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.ReorderPlaylistTrackQuery)).
		WithArgs(2, 10, 20).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.ReorderPlaylistTrack(tx, 10, 20, 2)
	})
	if err != nil {
		t.Fatalf("ReorderPlaylistTrack failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertPlayerStateQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "updated_at"}).AddRow(1, now))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, err := repo.UpsertPlayerState(tx, PlayerStateModel{ClientID: "c1"})
		return err
	})
	if err != nil {
		t.Fatalf("UpsertPlayerState failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestMusicRepositoryGetPlaylistTracksScanError(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlaylistTracksQuery)).
		WithArgs(1, 11, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()

	_, err := repo.GetPlaylistTracks(1, 1, 10)
	if err == nil {
		t.Fatalf("expected scan error for partial row payload")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestMusicRepositoryWriteErrorBranches(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeletePlaylistQuery)).
		WithArgs(10).
		WillReturnResult(sqlmock.NewResult(0, 0))
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := repo.DeletePlaylist(tx, 10); err == nil {
		t.Fatalf("expected DeletePlaylist not-found error")
	}
	_ = tx.Rollback()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeletePlaylistQuery)).
		WithArgs(10).
		WillReturnError(errors.New("delete failed"))
	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := repo.DeletePlaylist(tx, 10); err == nil {
		t.Fatalf("expected DeletePlaylist exec error")
	}
	_ = tx.Rollback()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.RemovePlaylistTrackQuery)).
		WithArgs(10, 20).
		WillReturnResult(sqlmock.NewResult(0, 0))
	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := repo.RemovePlaylistTrack(tx, 10, 20); err == nil {
		t.Fatalf("expected RemovePlaylistTrack not-found error")
	}
	_ = tx.Rollback()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.RemovePlaylistTrackQuery)).
		WithArgs(10, 20).
		WillReturnError(errors.New("remove failed"))
	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := repo.RemovePlaylistTrack(tx, 10, 20); err == nil {
		t.Fatalf("expected RemovePlaylistTrack exec error")
	}
	_ = tx.Rollback()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.ReorderPlaylistTrackQuery)).
		WithArgs(2, 10, 20).
		WillReturnError(errors.New("reorder failed"))
	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := repo.ReorderPlaylistTrack(tx, 10, 20, 2); err == nil {
		t.Fatalf("expected ReorderPlaylistTrack error")
	}
	_ = tx.Rollback()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertPlayerStateQuery)).
		WillReturnError(errors.New("upsert state failed"))
	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if _, err := repo.UpsertPlayerState(tx, PlayerStateModel{ClientID: "c1"}); err == nil {
		t.Fatalf("expected UpsertPlayerState error")
	}
	_ = tx.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
