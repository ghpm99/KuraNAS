package watchfolders

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/watch_folders"

	"github.com/DATA-DOG/go-sqlmock"
)

func newWatchFoldersRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewRepository(database.NewDbContext(db))
	return repo, mock, db
}

func TestNewRepositoryAndGetDbContext(t *testing.T) {
	repo, _, _ := newWatchFoldersRepoWithMock(t)
	if repo.GetDbContext() == nil {
		t.Fatalf("expected db context to be set")
	}
}

func TestRepositoryGetAllAndGetByID(t *testing.T) {
	repo, mock, _ := newWatchFoldersRepoWithMock(t)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetWatchFoldersQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "path", "label", "enabled", "last_scan_at", "created_at", "updated_at"}).
			AddRow(1, "/watch/a", "Folder A", true, now, now, now))
	mock.ExpectRollback()

	all, err := repo.GetAll()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(all) != 1 || all[0].ID != 1 {
		t.Fatalf("unexpected result from GetAll: %+v", all)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetWatchFolderByIDQuery)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "path", "label", "enabled", "last_scan_at", "created_at", "updated_at"}).
			AddRow(7, "/watch/b", "Folder B", false, now, now, now))
	mock.ExpectRollback()

	item, err := repo.GetByID(7)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item.ID != 7 || item.Path != "/watch/b" {
		t.Fatalf("unexpected result from GetByID: %+v", item)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryGetAllAndGetByIDErrors(t *testing.T) {
	repo, mock, _ := newWatchFoldersRepoWithMock(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetWatchFoldersQuery)).
		WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()

	if _, err := repo.GetAll(); err == nil || !regexp.MustCompile(`GetAll:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped GetAll error, got %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetWatchFolderByIDQuery)).
		WithArgs(9).
		WillReturnError(errors.New("row failed"))
	mock.ExpectRollback()

	if _, err := repo.GetByID(9); err == nil || !regexp.MustCompile(`GetByID:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped GetByID error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryCreateAndUpdate(t *testing.T) {
	repo, mock, db := newWatchFoldersRepoWithMock(t)
	now := time.Now()

	model := WatchFolderModel{
		ID:         11,
		Path:       "/watch/c",
		Label:      "Folder C",
		Enabled:    true,
		LastScanAt: &now,
	}

	mock.ExpectBegin()
	txCreate, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(queries.CreateWatchFolderQuery)).
		WithArgs(model.Path, model.Label, model.Enabled, model.LastScanAt).
		WillReturnRows(sqlmock.NewRows([]string{"id", "path", "label", "enabled", "last_scan_at", "created_at", "updated_at"}).
			AddRow(11, model.Path, model.Label, model.Enabled, now, now, now))
	created, err := repo.Create(txCreate, model)
	if err != nil {
		t.Fatalf("expected no create error, got %v", err)
	}
	if created.ID != 11 {
		t.Fatalf("unexpected created model: %+v", created)
	}
	mock.ExpectRollback()
	_ = txCreate.Rollback()

	mock.ExpectBegin()
	txUpdate, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(queries.UpdateWatchFolderQuery)).
		WithArgs(model.ID, model.Path, model.Label, model.Enabled).
		WillReturnRows(sqlmock.NewRows([]string{"id", "path", "label", "enabled", "last_scan_at", "created_at", "updated_at"}).
			AddRow(model.ID, model.Path, model.Label, model.Enabled, now, now, now))
	updated, err := repo.Update(txUpdate, model)
	if err != nil {
		t.Fatalf("expected no update error, got %v", err)
	}
	if updated.ID != model.ID {
		t.Fatalf("unexpected updated model: %+v", updated)
	}
	mock.ExpectRollback()
	_ = txUpdate.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryDeleteAndUpdateLastScan(t *testing.T) {
	repo, mock, db := newWatchFoldersRepoWithMock(t)
	now := time.Now()

	mock.ExpectBegin()
	txDelete, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteWatchFolderQuery)).
		WithArgs(4).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Delete(txDelete, 4); err != nil {
		t.Fatalf("expected no delete error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txDelete.Rollback()

	mock.ExpectBegin()
	txLastScan, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateWatchFolderLastScanQuery)).
		WithArgs(4, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.UpdateLastScan(txLastScan, 4, now); err != nil {
		t.Fatalf("expected no UpdateLastScan error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txLastScan.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryMutationErrors(t *testing.T) {
	repo, mock, db := newWatchFoldersRepoWithMock(t)
	now := time.Now()

	mock.ExpectBegin()
	txCreate, _ := db.Begin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CreateWatchFolderQuery)).
		WillReturnError(errors.New("create failed"))
	if _, err := repo.Create(txCreate, WatchFolderModel{}); err == nil || !regexp.MustCompile(`Create:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped create error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txCreate.Rollback()

	mock.ExpectBegin()
	txUpdate, _ := db.Begin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpdateWatchFolderQuery)).
		WillReturnError(errors.New("update failed"))
	if _, err := repo.Update(txUpdate, WatchFolderModel{}); err == nil || !regexp.MustCompile(`Update:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped update error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txUpdate.Rollback()

	mock.ExpectBegin()
	txDeleteExecErr, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteWatchFolderQuery)).
		WithArgs(1).
		WillReturnError(errors.New("delete exec failed"))
	if err := repo.Delete(txDeleteExecErr, 1); err == nil || !regexp.MustCompile(`Delete:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped delete exec error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txDeleteExecErr.Rollback()

	mock.ExpectBegin()
	txDeleteNoRows, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteWatchFolderQuery)).
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 0))
	if err := repo.Delete(txDeleteNoRows, 2); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	mock.ExpectRollback()
	_ = txDeleteNoRows.Rollback()

	mock.ExpectBegin()
	txLastScanExecErr, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateWatchFolderLastScanQuery)).
		WithArgs(3, now).
		WillReturnError(errors.New("last scan failed"))
	if err := repo.UpdateLastScan(txLastScanExecErr, 3, now); err == nil || !regexp.MustCompile(`UpdateLastScan:`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped UpdateLastScan error, got %v", err)
	}
	mock.ExpectRollback()
	_ = txLastScanExecErr.Rollback()

	mock.ExpectBegin()
	txLastScanNoRows, _ := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateWatchFolderLastScanQuery)).
		WithArgs(5, now).
		WillReturnResult(sqlmock.NewResult(0, 0))
	if err := repo.UpdateLastScan(txLastScanNoRows, 5, now); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	mock.ExpectRollback()
	_ = txLastScanNoRows.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
