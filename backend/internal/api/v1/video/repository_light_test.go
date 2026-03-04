package video

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/video"

	"github.com/DATA-DOG/go-sqlmock"
)

func newVideoRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewRepository(database.NewDbContext(db)), mock, db
}

func TestVideoRepositoryReadPaths(t *testing.T) {
	repo, mock, db := newVideoRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	if repo.GetDbContext() == nil {
		t.Fatalf("expected db context")
	}

	videoCols := []string{"id", "name", "path", "parent_path", "format", "size", "created_at", "updated_at"}
	playlistCols := []string{"id", "type", "source_path", "name", "is_hidden", "is_auto", "group_mode", "classification", "created_at", "updated_at", "last_played_at"}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideoFileByIDQuery)).
		WillReturnRows(sqlmock.NewRows(videoCols).AddRow(1, "v", "/v", "/", ".mp4", 100, now, now))
	mock.ExpectRollback()
	if v, err := repo.GetVideoFileByID(1); err != nil || v.ID != 1 {
		t.Fatalf("GetVideoFileByID failed v=%+v err=%v", v, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideosByParentPathQuery)).
		WillReturnRows(sqlmock.NewRows(videoCols).AddRow(2, "v2", "/p/v2", "/p", ".mp4", 120, now, now))
	mock.ExpectRollback()
	if out, err := repo.GetVideosByParentPath("/p"); err != nil || len(out) != 1 {
		t.Fatalf("GetVideosByParentPath failed len=%d err=%v", len(out), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlaylistByContextQuery)).
		WillReturnRows(sqlmock.NewRows(playlistCols).AddRow(1, "folder", "/p", "name", false, false, "", "", now, now, nil))
	mock.ExpectRollback()
	if p, err := repo.GetPlaylistByContext("folder", "/p"); err != nil || p.ID != 1 {
		t.Fatalf("GetPlaylistByContext failed p=%+v err=%v", p, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlaylistItemsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "playlist_id", "video_id", "order_index", "source_kind", "name", "path", "parent_path", "format", "size", "created_at", "updated_at"}).
			AddRow(1, 1, 2, 0, "auto", "v2", "/p/v2", "/p", ".mp4", 120, now, now))
	mock.ExpectRollback()
	if items, err := repo.GetPlaylistItems(1); err != nil || len(items) != 1 {
		t.Fatalf("GetPlaylistItems failed len=%d err=%v", len(items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlaybackStateQuery)).
		WithArgs("c1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "client_id", "playlist_id", "video_id", "current_time", "duration", "is_paused", "completed", "last_update"}).
			AddRow(1, "c1", 1, 2, 1.0, 10.0, false, false, now))
	mock.ExpectRollback()
	if st, err := repo.GetPlaybackState("c1"); err != nil || st.ID != 1 {
		t.Fatalf("GetPlaybackState failed st=%+v err=%v", st, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetCatalogVideosQuery)).
		WillReturnRows(sqlmock.NewRows(videoCols).AddRow(1, "v", "/v", "/", ".mp4", 100, now, now))
	mock.ExpectRollback()
	if out, err := repo.GetCatalogVideos(10); err != nil || len(out) != 1 {
		t.Fatalf("GetCatalogVideos failed len=%d err=%v", len(out), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetRecentVideosQuery)).
		WillReturnRows(sqlmock.NewRows(videoCols).AddRow(1, "v", "/v", "/", ".mp4", 100, now, now))
	mock.ExpectRollback()
	if out, err := repo.GetRecentVideos(10); err != nil || len(out) != 1 {
		t.Fatalf("GetRecentVideos failed len=%d err=%v", len(out), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAllVideosForGroupingQuery)).
		WillReturnRows(sqlmock.NewRows(videoCols).AddRow(1, "v", "/v", "/", ".mp4", 100, now, now))
	mock.ExpectRollback()
	if out, err := repo.GetAllVideosForGrouping(); err != nil || len(out) != 1 {
		t.Fatalf("GetAllVideosForGrouping failed len=%d err=%v", len(out), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlaylistExclusionsQuery)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"video_id"}).AddRow(2))
	mock.ExpectRollback()
	if ex, err := repo.GetPlaylistExclusions(1); err != nil || !ex[2] {
		t.Fatalf("GetPlaylistExclusions failed ex=%v err=%v", ex, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideoPlaylistsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "source_path", "name", "is_hidden", "is_auto", "group_mode", "classification", "created_at", "updated_at", "last_played_at", "item_count", "cover_video_id"}).
			AddRow(1, "folder", "/p", "name", false, false, "", "", now, now, nil, 2, nil))
	mock.ExpectRollback()
	if out, err := repo.GetVideoPlaylists(true); err != nil || len(out) != 1 {
		t.Fatalf("GetVideoPlaylists failed len=%d err=%v", len(out), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideoPlaylistByIDQuery)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(playlistCols).AddRow(1, "folder", "/p", "name", false, false, "", "", now, now, nil))
	mock.ExpectRollback()
	if p, err := repo.GetVideoPlaylistByID(1); err != nil || p.ID != 1 {
		t.Fatalf("GetVideoPlaylistByID failed p=%+v err=%v", p, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideoPlaylistItemsDetailedQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "playlist_id", "video_id", "order_index", "source_kind", "name", "path", "parent_path", "format", "size", "created_at", "updated_at"}).
			AddRow(1, 1, 2, 0, "manual", "v2", "/p/v2", "/p", ".mp4", 120, now, now))
	mock.ExpectRollback()
	if out, err := repo.GetVideoPlaylistItemsDetailed(1); err != nil || len(out) != 1 {
		t.Fatalf("GetVideoPlaylistItemsDetailed failed len=%d err=%v", len(out), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetUnassignedVideosQuery)).
		WillReturnRows(sqlmock.NewRows(videoCols).AddRow(3, "u", "/u", "/", ".mp4", 33, now, now))
	mock.ExpectRollback()
	if out, err := repo.GetUnassignedVideos(10); err != nil || len(out) != 1 {
		t.Fatalf("GetUnassignedVideos failed len=%d err=%v", len(out), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CheckVideoInPlaylistQuery)).
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()
	if ok, err := repo.CheckVideoInPlaylist(1, 2); err != nil || !ok {
		t.Fatalf("CheckVideoInPlaylist failed ok=%v err=%v", ok, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestVideoRepositoryWritePaths(t *testing.T) {
	repo, mock, db := newVideoRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CreatePlaylistQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "source_path", "name", "is_hidden", "is_auto", "group_mode", "classification", "created_at", "updated_at", "last_played_at"}).
			AddRow(10, "folder", "/p", "p", false, false, "", "", now, now, nil))
	mock.ExpectCommit()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, err := repo.CreatePlaylist(tx, "folder", "/p")
		return err
	})
	if err != nil {
		t.Fatalf("CreatePlaylist failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeletePlaylistItemsQuery)).
		WithArgs(10).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.InsertPlaylistItemsQuery)).
		WithArgs(10, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.ReplacePlaylistItems(tx, 10, []int{1, 2})
	})
	if err != nil {
		t.Fatalf("ReplacePlaylistItems failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertPlaybackStateQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "last_update"}).AddRow(1, now))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, err := repo.UpsertPlaybackState(tx, VideoPlaybackStateModel{ClientID: "c1"})
		return err
	})
	if err != nil {
		t.Fatalf("UpsertPlaybackState failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.TouchPlaylistQuery)).
		WithArgs(10).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.TouchPlaylist(tx, 10)
	})
	if err != nil {
		t.Fatalf("TouchPlaylist failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertAutoPlaylistQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type", "source_path", "name", "is_hidden", "is_auto", "group_mode", "classification", "created_at", "updated_at", "last_played_at"}).
			AddRow(11, "smart", "/s", "n", false, true, "gm", "c", now, now, nil))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, err := repo.UpsertAutoPlaylist(tx, "smart", "/s", "n", "gm", "c")
		return err
	})
	if err != nil {
		t.Fatalf("UpsertAutoPlaylist failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteAutoPlaylistItemsQuery)).
		WithArgs(11).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.DeleteAutoPlaylistItems(tx, 11)
	})
	if err != nil {
		t.Fatalf("DeleteAutoPlaylistItems failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.InsertPlaylistItemsWithSourceQuery)).
		WithArgs(11, sqlmock.AnyArg(), "auto").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.InsertPlaylistItemsWithSource(tx, 11, []int{1}, "auto")
	})
	if err != nil {
		t.Fatalf("InsertPlaylistItemsWithSource failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.SetPlaylistHiddenQuery)).
		WithArgs(11, true).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.AddPlaylistVideoManualQuery)).
		WithArgs(11, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.RemovePlaylistVideoQuery)).
		WithArgs(11, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.UpsertPlaylistExclusionQuery)).
		WithArgs(11, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.DeletePlaylistExclusionQuery)).
		WithArgs(11, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdatePlaylistNameQuery)).
		WithArgs(11, "new").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.ReorderPlaylistItemQuery)).
		WithArgs(11, 2, 0).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		if err := repo.SetPlaylistHidden(tx, 11, true); err != nil {
			return err
		}
		if err := repo.AddPlaylistVideoManual(tx, 11, 2); err != nil {
			return err
		}
		if err := repo.RemovePlaylistVideo(tx, 11, 2); err != nil {
			return err
		}
		if err := repo.UpsertPlaylistExclusion(tx, 11, 2); err != nil {
			return err
		}
		if err := repo.DeletePlaylistExclusion(tx, 11, 2); err != nil {
			return err
		}
		if err := repo.UpdatePlaylistName(tx, 11, "new"); err != nil {
			return err
		}
		return repo.ReorderPlaylistItem(tx, 11, 2, 0)
	})
	if err != nil {
		t.Fatalf("write operations failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
