package files

import (
	"database/sql"
	"errors"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/files"

	"github.com/DATA-DOG/go-sqlmock"
)

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

func TestRepositoryCreateAndUpdateFile(t *testing.T) {
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryUpdateDescendantPathsAndMarkDeletedSubtree(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	separator := string(filepath.Separator)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateDescendantPathsQuery)).
		WithArgs("/tmp/old", "/tmp/new", "/tmp/old"+separator).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		affected, err := repo.UpdateDescendantPaths(tx, "/tmp/old", "/tmp/new")
		if err != nil {
			return err
		}
		if affected != 3 {
			t.Fatalf("expected 3 affected rows, got %d", affected)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("UpdateDescendantPaths failed: %v", err)
	}

	deletedAt := time.Now()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.MarkDeletedSubtreeQuery)).
		WithArgs("/tmp/gone", deletedAt, "/tmp/gone"+separator).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		affected, err := repo.MarkDeletedSubtree(tx, "/tmp/gone", deletedAt)
		if err != nil {
			return err
		}
		if affected != 2 {
			t.Fatalf("expected 2 affected rows, got %d", affected)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("MarkDeletedSubtree failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func fileRowColumns() []string {
	return []string{
		"id", "name", "path", "parent_path", "format", "size", "updated_at", "created_at",
		"last_interaction", "last_backup", "type", "check_sum", "deleted_at", "starred",
		"physical_path",
	}
}

func addFileRow(rows *sqlmock.Rows, id int, name, path string) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(id, name, path, "/tmp", ".txt", 1, now, now, nil, nil, int(File), "abc", nil, false, nil)
}

func TestRepositoryDecomposedListingQueries(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetFileByIDQuery)).
		WithArgs(7).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 7, "a", "/tmp/a"))
	mock.ExpectRollback()
	file, found, err := repo.GetFileByID(7)
	if err != nil || !found || file.ID != 7 {
		t.Fatalf("GetFileByID failed file=%+v found=%v err=%v", file, found, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetFileByIDQuery)).
		WithArgs(8).
		WillReturnRows(sqlmock.NewRows(fileRowColumns()))
	mock.ExpectRollback()
	_, found, err = repo.GetFileByID(8)
	if err != nil || found {
		t.Fatalf("GetFileByID expected not found, got found=%v err=%v", found, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetFilesByNameAndPathQuery)).
		WithArgs("a", "/tmp/a", 5).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 7, "a", "/tmp/a"))
	mock.ExpectRollback()
	rows, err := repo.GetFilesByNameAndPath("a", "/tmp/a", 5)
	if err != nil || len(rows) != 1 {
		t.Fatalf("GetFilesByNameAndPath failed len=%d err=%v", len(rows), err)
	}
	if _, err := repo.GetFilesByNameAndPath("a", "/tmp/a", 0); err == nil {
		t.Fatalf("expected limit validation error")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetChildrenByParentPathQuery)).
		WithArgs("/tmp", 11, 0).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 1, "a", "/tmp/a"))
	mock.ExpectRollback()
	pageResult, err := repo.GetActiveChildrenByParentPath("/tmp", AllCategory, 1, 10)
	if err != nil || len(pageResult.Items) != 1 {
		t.Fatalf("GetActiveChildrenByParentPath failed len=%d err=%v", len(pageResult.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetStarredChildrenByParentPathQuery)).
		WithArgs("/tmp", 11, 0).
		WillReturnRows(sqlmock.NewRows(fileRowColumns()))
	mock.ExpectRollback()
	if _, err := repo.GetActiveChildrenByParentPath("/tmp", StarredCategory, 1, 10); err != nil {
		t.Fatalf("starred children query failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetRecentChildrenByParentPathQuery)).
		WithArgs("/tmp", 11, 0).
		WillReturnRows(sqlmock.NewRows(fileRowColumns()))
	mock.ExpectRollback()
	if _, err := repo.GetActiveChildrenByParentPath("/tmp", RecentCategory, 1, 10); err != nil {
		t.Fatalf("recent children query failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetFilesByPathQuery)).
		WithArgs("/tmp/a", 11, 0).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 1, "a", "/tmp/a"))
	mock.ExpectRollback()
	if _, err := repo.GetActiveFilesByPath("/tmp/a", 1, 10); err != nil {
		t.Fatalf("GetActiveFilesByPath failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetActiveFilesQuery)).
		WithArgs(11, 0).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 1, "a", "/tmp/a"))
	mock.ExpectRollback()
	if _, err := repo.GetActiveFiles(1, 10); err != nil {
		t.Fatalf("GetActiveFiles failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetFilesByPathPrefixQuery)).
		WithArgs("/tmp", 11, 0).
		WillReturnRows(addFileRow(sqlmock.NewRows(fileRowColumns()), 1, "a", "/tmp/a"))
	mock.ExpectRollback()
	if _, err := repo.GetFilesByPathPrefix("/tmp", 1, 10); err != nil {
		t.Fatalf("GetFilesByPathPrefix failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryGetFileStatByPath(t *testing.T) {
	now := time.Now()

	t.Run("found", func(t *testing.T) {
		repo, mock, db := newRepoWithMock(t)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(queries.GetFileStatByPathQuery)).
			WithArgs("/tmp/f").
			WillReturnRows(sqlmock.NewRows([]string{"size", "updated_at"}).AddRow(int64(42), now))
		mock.ExpectRollback()

		stat, found, err := repo.GetFileStatByPath("/tmp/f")
		if err != nil {
			t.Fatalf("GetFileStatByPath failed: %v", err)
		}
		if !found {
			t.Fatalf("expected found=true")
		}
		if stat.Size != 42 || !stat.UpdatedAt.Equal(now) {
			t.Fatalf("unexpected stat: %+v", stat)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sqlmock expectations: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock, db := newRepoWithMock(t)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(queries.GetFileStatByPathQuery)).
			WithArgs("/tmp/missing").
			WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()

		stat, found, err := repo.GetFileStatByPath("/tmp/missing")
		if err != nil {
			t.Fatalf("expected no error for missing row, got %v", err)
		}
		if found {
			t.Fatalf("expected found=false")
		}
		if stat.Size != 0 || !stat.UpdatedAt.IsZero() {
			t.Fatalf("expected zero stat, got %+v", stat)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet sqlmock expectations: %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		repo, mock, db := newRepoWithMock(t)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(queries.GetFileStatByPathQuery)).
			WithArgs("/tmp/boom").
			WillReturnError(errors.New("boom"))
		mock.ExpectRollback()

		if _, _, err := repo.GetFileStatByPath("/tmp/boom"); err == nil {
			t.Fatalf("expected error from failing query")
		}
	})
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
