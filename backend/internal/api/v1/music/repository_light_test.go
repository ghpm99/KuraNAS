package music

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/music"

	"github.com/DATA-DOG/go-sqlmock"
)

func sequentialCols(total int) []string {
	cols := make([]string, 0, total)
	for i := 1; i <= total; i++ {
		cols = append(cols, fmt.Sprintf("c%d", i))
	}
	return cols
}

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
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_system", "created_at", "updated_at", "track_count", "is_ai_generated"}).
			AddRow(1, "p", "d", false, now, now, 3, false))
	mock.ExpectRollback()
	playlists, err := repo.GetPlaylists(1, 10)
	if err != nil || len(playlists.Items) != 1 {
		t.Fatalf("GetPlaylists failed len=%d err=%v", len(playlists.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlaylistByIDQuery)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_system", "created_at", "updated_at", "track_count", "is_ai_generated"}).
			AddRow(1, "p", "d", false, now, now, 3, false))
	mock.ExpectRollback()
	if p, err := repo.GetPlaylistByID(1); err != nil || p.ID != 1 {
		t.Fatalf("GetPlaylistByID failed p=%+v err=%v", p, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetNowPlayingQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_system", "created_at", "updated_at", "track_count", "is_ai_generated"}).
			AddRow(2, "now", "queue", true, now, now, 1, false))
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

func TestMusicRepositoryArtistClustersAndAIPlaylists(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	// Read the persisted artist -> cluster mapping.
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetArtistClustersQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"artist_key", "artist", "cluster_name"}).
			AddRow("the beatles", "The Beatles", "Classic Rock"))
	mock.ExpectRollback()
	clusters, err := repo.GetArtistClusters()
	if err != nil || len(clusters) != 1 || clusters[0].ClusterName != "Classic Rock" {
		t.Fatalf("GetArtistClusters failed: %+v err=%v", clusters, err)
	}

	// List materialized AI playlists.
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAIPlaylistsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_system", "created_at", "updated_at", "track_count", "is_ai_generated"}).
			AddRow(7, "Classic Rock", "", false, now, now, 12, true))
	mock.ExpectRollback()
	aiPlaylists, err := repo.GetAIPlaylists()
	if err != nil || len(aiPlaylists) != 1 || !aiPlaylists[0].IsAIGenerated {
		t.Fatalf("GetAIPlaylists failed: %+v err=%v", aiPlaylists, err)
	}

	// Write paths inside a single transaction: upsert, prune, create, replace.
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpsertArtistClusterQuery)).
		WithArgs("the beatles", "The Beatles", "Classic Rock").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteArtistClustersExceptQuery)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(queries.CreateAIPlaylistQuery)).
		WithArgs("Classic Rock", "").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(7, now, now))
	mock.ExpectExec(regexp.QuoteMeta(queries.ClearPlaylistTracksQuery)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 5))
	mock.ExpectExec(regexp.QuoteMeta(queries.InsertPlaylistTracksQuery)).
		WithArgs(7, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		if err := repo.UpsertArtistCluster(tx, ArtistClusterModel{ArtistKey: "the beatles", Artist: "The Beatles", ClusterName: "Classic Rock"}); err != nil {
			return err
		}
		if err := repo.DeleteArtistClustersExcept(tx, []string{"the beatles"}); err != nil {
			return err
		}
		created, err := repo.CreateAIPlaylist(tx, "Classic Rock", "")
		if err != nil {
			return err
		}
		if created.ID != 7 || !created.IsAIGenerated {
			t.Fatalf("unexpected created AI playlist: %+v", created)
		}
		return repo.ReplacePlaylistTracks(tx, 7, []int{20, 21})
	})
	if err != nil {
		t.Fatalf("cluster write paths failed: %v", err)
	}

	// ReplacePlaylistTracks with no files only clears, never inserts.
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.ClearPlaylistTracksQuery)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.ReplacePlaylistTracks(tx, 7, nil)
	})
	if err != nil {
		t.Fatalf("empty ReplacePlaylistTracks failed: %v", err)
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

func TestMusicRepositoryGetPlaylistTracksSuccessAndQueryError(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	trackValues := []driver.Value{
		1, 10, 20, 0, now,
		20, "song", "/tmp/song.mp3", "/tmp", ".mp3", int64(1024), now, now, nil, nil, 2, "sum", nil, false,
		2, 20, "/tmp/song.mp3", "audio/mpeg", 120.0, 320, 44100, 2, 1, "enc", 16, "title", "artist", "album",
		"albumArtist", "1", "rock", "composer", "2026", "2026-01-01", "lame", "pub", "2025-01-01", "orig", "lyr", "lyrics", now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlaylistTracksQuery)).
		WithArgs(10, 11, 0).
		WillReturnRows(sqlmock.NewRows(sequentialCols(len(trackValues))).AddRow(trackValues...))
	mock.ExpectRollback()
	items, err := repo.GetPlaylistTracks(10, 1, 10)
	if err != nil || len(items.Items) != 1 {
		t.Fatalf("expected GetPlaylistTracks success, len=%d err=%v", len(items.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetPlaylistTracksQuery)).
		WithArgs(10, 11, 0).
		WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()
	if _, err := repo.GetPlaylistTracks(10, 1, 10); err == nil {
		t.Fatalf("expected GetPlaylistTracks query error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func newAudioMetadataRepoWithMock(t *testing.T) (*AudioMetadataRepository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewAudioMetadataRepository(database.NewDbContext(db)), mock, db
}

func TestMusicRepositoryBrowseAggregates(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicArtistsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"artist", "track_count", "album_count"}).AddRow("a", 1, 1))
	mock.ExpectRollback()
	if out, err := repo.GetMusicArtists(1, 10); err != nil || len(out.Items) != 1 {
		t.Fatalf("GetMusicArtists failed len=%d err=%v", len(out.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicAlbumsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"album", "artist", "year", "track_count"}).AddRow("al", "ar", "2025", 3))
	mock.ExpectRollback()
	if out, err := repo.GetMusicAlbums(1, 10); err != nil || len(out.Items) != 1 {
		t.Fatalf("GetMusicAlbums failed len=%d err=%v", len(out.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicGenresQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"genre", "track_count"}).AddRow("rock", 3))
	mock.ExpectRollback()
	if out, err := repo.GetMusicGenres(1, 10); err != nil || len(out.Items) != 1 {
		t.Fatalf("GetMusicGenres failed len=%d err=%v", len(out.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicFoldersQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"folder", "track_count"}).AddRow("/music", 3))
	mock.ExpectRollback()
	if out, err := repo.GetMusicFolders(1, 10); err != nil || len(out.Items) != 1 {
		t.Fatalf("GetMusicFolders failed len=%d err=%v", len(out.Items), err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestMusicRepositoryBrowseTrackQueries(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()
	now := time.Now()
	fileType := 1

	audioValues := []driver.Value{
		3, "song", "/tmp/song.mp3", "/tmp", ".mp3", int64(20), now, now, nil, nil, fileType, "sum2", nil, true,
		4, 3, "/tmp/song.mp3", "audio/mpeg", 123.4, 320, 44100, 2, 1, "enc", 16, "title", "artist", "album",
		"albumArtist", "1", "rock", "composer", "2026", "2026-01-01", "lame", "pub", "2025-12-01", "orig", "lyr", "text", now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicQuery)).
		WillReturnRows(sqlmock.NewRows(sequentialCols(len(audioValues))).AddRow(audioValues...))
	mock.ExpectRollback()
	music, err := repo.GetMusic(1, 10)
	if err != nil || len(music.Items) != 1 {
		t.Fatalf("GetMusic failed len=%d err=%v", len(music.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicByArtistQuery)).
		WillReturnRows(sqlmock.NewRows(sequentialCols(len(audioValues))).AddRow(audioValues...))
	mock.ExpectRollback()
	byArtist, err := repo.GetMusicByArtist("artist", 1, 10)
	if err != nil || len(byArtist.Items) != 1 {
		t.Fatalf("GetMusicByArtist failed len=%d err=%v", len(byArtist.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicByAlbumQuery)).
		WillReturnRows(sqlmock.NewRows(sequentialCols(len(audioValues))).AddRow(audioValues...))
	mock.ExpectRollback()
	byAlbum, err := repo.GetMusicByAlbum("album", 1, 10)
	if err != nil || len(byAlbum.Items) != 1 {
		t.Fatalf("GetMusicByAlbum failed len=%d err=%v", len(byAlbum.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicByGenreQuery)).
		WillReturnRows(sqlmock.NewRows(sequentialCols(len(audioValues))).AddRow(audioValues...))
	mock.ExpectRollback()
	byGenre, err := repo.GetMusicByGenre("rock", 1, 10)
	if err != nil || len(byGenre.Items) != 1 {
		t.Fatalf("GetMusicByGenre failed len=%d err=%v", len(byGenre.Items), err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestMusicRepositoryBrowseScanErrors(t *testing.T) {
	repo, mock, db := newMusicRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()
	if _, err := repo.GetMusic(1, 10); err == nil {
		t.Fatalf("expected GetMusic scan error")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicByArtistQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()
	if _, err := repo.GetMusicByArtist("artist", 1, 10); err == nil {
		t.Fatalf("expected GetMusicByArtist scan error")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicByAlbumQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()
	if _, err := repo.GetMusicByAlbum("album", 1, 10); err == nil {
		t.Fatalf("expected GetMusicByAlbum scan error")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicByGenreQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()
	if _, err := repo.GetMusicByGenre("genre", 1, 10); err == nil {
		t.Fatalf("expected GetMusicByGenre scan error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestAudioMetadataRepositorySuccessPaths(t *testing.T) {
	repo, mock, db := newAudioMetadataRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	if repo == nil || repo.GetDbContext() == nil {
		t.Fatalf("expected initialized audio metadata repository")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAudioMetadataByIDQuery)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "file_id", "path", "mime", "length", "bitrate", "sample_rate", "channels", "bitrate_mode", "encoder_info",
			"bit_depth", "title", "artist", "album", "album_artist", "track_number", "genre", "composer", "year",
			"recording_date", "encoder", "publisher", "original_release_date", "original_artist", "lyricist", "lyrics", "created_at",
		}).AddRow(
			2, 20, "/a.mp3", "audio/mpeg", 120.0, 320, 44100, 2, 1, "enc", 16, "title", "artist", "album",
			"album artist", "1", "genre", "composer", "2026", "2026-01-01", "encoder", "publisher", "2025-01-01",
			"original", "lyricist", "lyrics", now,
		))
	mock.ExpectRollback()
	audioMeta, err := repo.GetAudioMetadataByID(2)
	if err != nil || audioMeta.ID != 2 {
		t.Fatalf("GetAudioMetadataByID failed meta=%+v err=%v", audioMeta, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertAudioMetadataQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(22, now))
	mock.ExpectCommit()
	err = repo.Db.ExecTx(func(tx *sql.Tx) error {
		upserted, err := repo.UpsertAudioMetadata(tx, AudioMetadataModel{FileId: 20, Path: "/a.mp3"})
		if err != nil {
			return err
		}
		if upserted.ID != 22 {
			t.Fatalf("expected audio metadata ID 22")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("UpsertAudioMetadata failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteAudioMetadataQuery)).
		WithArgs(22).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.DeleteAudioMetadata(22); err != nil {
		t.Fatalf("DeleteAudioMetadata failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestAudioMetadataRepositoryErrorPaths(t *testing.T) {
	repo, mock, db := newAudioMetadataRepoWithMock(t)
	defer db.Close()

	scanErr := errors.New("scan failed")
	execErr := errors.New("exec failed")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAudioMetadataByIDQuery)).
		WithArgs(1).
		WillReturnError(scanErr)
	mock.ExpectRollback()
	_, err := repo.GetAudioMetadataByID(1)
	if err == nil || !strings.Contains(err.Error(), "falha ao obter metadados de audio") {
		t.Fatalf("expected wrapped audio error, got: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpsertAudioMetadataQuery)).
		WillReturnError(scanErr)
	mock.ExpectRollback()
	err = repo.Db.ExecTx(func(tx *sql.Tx) error {
		_, err := repo.UpsertAudioMetadata(tx, AudioMetadataModel{})
		return err
	})
	if !errors.Is(err, scanErr) {
		t.Fatalf("expected raw upsert audio error, got: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteAudioMetadataQuery)).
		WithArgs(1).
		WillReturnError(execErr)
	mock.ExpectRollback()
	err = repo.DeleteAudioMetadata(1)
	if err == nil || !strings.Contains(err.Error(), "falha ao deletar metadados de audio") {
		t.Fatalf("expected wrapped delete audio error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
