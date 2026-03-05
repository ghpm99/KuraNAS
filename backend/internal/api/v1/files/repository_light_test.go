package files

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/file"

	"github.com/DATA-DOG/go-sqlmock"
)

func numberedCols(total int) []string {
	cols := make([]string, 0, total)
	for i := 1; i <= total; i++ {
		cols = append(cols, fmt.Sprintf("c%d", i))
	}
	return cols
}

func newRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	ctx := database.NewDbContext(db)
	return NewRepository(ctx), mock, db
}

func TestRepositoryConstructorsAndSimpleQueries(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	if repo.GetDbContext() == nil {
		t.Fatalf("expected db context")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetChildrenCountQuery)).
		WithArgs("/tmp", 1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectRollback()
	if v, err := repo.GetDirectoryContentCount(1, "/tmp"); err != nil || v != 3 {
		t.Fatalf("GetDirectoryContentCount failed v=%d err=%v", v, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CountByTypeQuery)).
		WithArgs(File).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))
	mock.ExpectRollback()
	if v, err := repo.GetCountByType(File); err != nil || v != 7 {
		t.Fatalf("GetCountByType failed v=%d err=%v", v, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.TotalSpaceUsedQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(1024))
	mock.ExpectRollback()
	if v, err := repo.GetTotalSpaceUsed(); err != nil || v != 1024 {
		t.Fatalf("GetTotalSpaceUsed failed v=%d err=%v", v, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CountByFormatQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"format", "total", "size"}).AddRow(".mp3", 2, 200))
	mock.ExpectRollback()
	report, err := repo.GetReportSizeByFormat()
	if err != nil || len(report) != 1 {
		t.Fatalf("GetReportSizeByFormat failed len=%d err=%v", len(report), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.TopFilesBySizeQuery)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "size", "path"}).AddRow(1, "a", 99, "/tmp/a"))
	mock.ExpectRollback()
	top, err := repo.GetTopFilesBySize(5)
	if err != nil || len(top) != 1 {
		t.Fatalf("GetTopFilesBySize failed len=%d err=%v", len(top), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetDuplicateFilesQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"name", "size", "copies", "paths"}).AddRow("dup", 10, 2, "/a,/b"))
	mock.ExpectRollback()
	dups, err := repo.GetDuplicateFiles(1, 10)
	if err != nil || len(dups.Items) != 1 {
		t.Fatalf("GetDuplicateFiles failed len=%d err=%v", len(dups.Items), err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryCreateUpdateAndGetFiles(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	model := FileModel{
		Name:       "f",
		Path:       "/tmp/f",
		ParentPath: "/tmp",
		Format:     ".txt",
		Size:       1,
		UpdatedAt:  now,
		CreatedAt:  now,
		Type:       File,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertFileQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		created, err := repo.CreateFile(tx, model)
		if err != nil {
			return err
		}
		if created.ID != 11 {
			t.Fatalf("expected created id 11")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateFileQuery)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		ok, err := repo.UpdateFile(tx, FileModel{ID: 11, Name: "f", Path: "/tmp/f", ParentPath: "/tmp", Type: File, UpdatedAt: now, CreatedAt: now})
		if err != nil {
			return err
		}
		if !ok {
			t.Fatalf("expected updated true")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("UpdateFile failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetFilesQuery)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "path", "parent_path", "format", "size", "updated_at", "created_at",
			"last_interaction", "last_backup", "type", "check_sum", "deleted_at", "starred",
		}).AddRow(1, "n", "/tmp/n", "/tmp", ".txt", 1, now, now, nil, nil, int(File), "abc", nil, false))
	mock.ExpectRollback()
	out, err := repo.GetFiles(FileFilter{}, 1, 10)
	if err != nil || len(out.Items) != 1 {
		t.Fatalf("GetFiles failed len=%d err=%v", len(out.Items), err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryMusicAggregatesLight(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
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

func TestRepositoryMediaQueriesScanErrorPaths(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(getImagesQueryByGroup(ImageGroupByDate))).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()
	if _, err := repo.GetImages(1, 10, ImageGroupByDate); err == nil {
		t.Fatalf("expected GetImages scan error")
	}

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

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideosQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()
	if _, err := repo.GetVideos(1, 10); err == nil {
		t.Fatalf("expected GetVideos scan error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryUpdateFileBranches(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()
	base := FileModel{
		ID:         99,
		Name:       "f",
		Path:       "/tmp/f",
		ParentPath: "/tmp",
		Type:       File,
		UpdatedAt:  now,
		CreatedAt:  now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateFileQuery)).
		WillReturnError(errors.New("exec failed"))
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if ok, err := repo.UpdateFile(tx, base); err == nil || ok {
		t.Fatalf("expected UpdateFile exec error")
	}
	_ = tx.Rollback()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateFileQuery)).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected failed")))
	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if ok, err := repo.UpdateFile(tx, base); err == nil || ok {
		t.Fatalf("expected UpdateFile rows affected error")
	}
	_ = tx.Rollback()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateFileQuery)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectRollback()
	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if ok, err := repo.UpdateFile(tx, base); err == nil || ok {
		t.Fatalf("expected UpdateFile multiple rows error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryMediaQueriesSuccessPaths(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()
	fileType := int(File)

	imageValues := []driver.Value{
		1, "img", "/tmp/img.jpg", "/tmp", ".jpg", int64(10), now, now, nil, nil, fileType, "sum", nil, false,
		2, 1, "/tmp/img.jpg", "jpeg", "RGB", 10, 10, 72.0, 72.0, 72.0, 72.0, 2.0, 1.0, 6.0, 2.0, 1.0, "cfg", "icc",
		"mk", "mdl", "sw", "lens", "ser", "dt", "dto", "dtd", "sub", 1.0, 2.8, 100.0, 1.0, 2.0, 0.0, 3.0, 0.0, 50.0,
		0.0, 1.0, 2.8, 1.8, -23.0, -46.0, 800.0, "2026:01:01", "12:00:00", "desc", "comment", "copy", "artist", now,
	}

	audioValues := []driver.Value{
		3, "song", "/tmp/song.mp3", "/tmp", ".mp3", int64(20), now, now, nil, nil, fileType, "sum2", nil, true,
		4, 3, "/tmp/song.mp3", "audio/mpeg", 123.4, 320, 44100, 2, 1, "enc", 16, "title", "artist", "album",
		"albumArtist", "1", "rock", "composer", "2026", "2026-01-01", "lame", "pub", "2025-12-01", "orig", "lyr", "text", now,
	}

	videoValues := []driver.Value{
		5, "video", "/tmp/video.mp4", "/tmp", ".mp4", int64(30), now, now, nil, nil, fileType, "sum3", nil, false,
		6, 5, "/tmp/video.mp4", "mp4", "30", "60.0", 1920, 1080, 30.0, 1800, "1000000", "h264", "H.264", "yuv420p",
		40, "High", "16:9", "aac", 2, "48000", "192000", now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(getImagesQueryByGroup(ImageGroupByDate))).
		WillReturnRows(sqlmock.NewRows(numberedCols(len(imageValues))).AddRow(imageValues...))
	mock.ExpectRollback()
	images, err := repo.GetImages(1, 10, ImageGroupByDate)
	if err != nil || len(images.Items) != 1 {
		t.Fatalf("GetImages failed len=%d err=%v", len(images.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicQuery)).
		WillReturnRows(sqlmock.NewRows(numberedCols(len(audioValues))).AddRow(audioValues...))
	mock.ExpectRollback()
	music, err := repo.GetMusic(1, 10)
	if err != nil || len(music.Items) != 1 {
		t.Fatalf("GetMusic failed len=%d err=%v", len(music.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicByArtistQuery)).
		WillReturnRows(sqlmock.NewRows(numberedCols(len(audioValues))).AddRow(audioValues...))
	mock.ExpectRollback()
	byArtist, err := repo.GetMusicByArtist("artist", 1, 10)
	if err != nil || len(byArtist.Items) != 1 {
		t.Fatalf("GetMusicByArtist failed len=%d err=%v", len(byArtist.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicByAlbumQuery)).
		WillReturnRows(sqlmock.NewRows(numberedCols(len(audioValues))).AddRow(audioValues...))
	mock.ExpectRollback()
	byAlbum, err := repo.GetMusicByAlbum("album", 1, 10)
	if err != nil || len(byAlbum.Items) != 1 {
		t.Fatalf("GetMusicByAlbum failed len=%d err=%v", len(byAlbum.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetMusicByGenreQuery)).
		WillReturnRows(sqlmock.NewRows(numberedCols(len(audioValues))).AddRow(audioValues...))
	mock.ExpectRollback()
	byGenre, err := repo.GetMusicByGenre("rock", 1, 10)
	if err != nil || len(byGenre.Items) != 1 {
		t.Fatalf("GetMusicByGenre failed len=%d err=%v", len(byGenre.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetVideosQuery)).
		WillReturnRows(sqlmock.NewRows(numberedCols(len(videoValues))).AddRow(videoValues...))
	mock.ExpectRollback()
	videos, err := repo.GetVideos(1, 10)
	if err != nil || len(videos.Items) != 1 {
		t.Fatalf("GetVideos failed len=%d err=%v", len(videos.Items), err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
