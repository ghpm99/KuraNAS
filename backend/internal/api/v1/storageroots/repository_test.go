package storageroots

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/storageroots"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
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

func storageRootRows(rows ...StorageRootModel) *sqlmock.Rows {
	out := sqlmock.NewRows([]string{"id", "path", "label", "enabled", "created_at"})
	for _, row := range rows {
		out.AddRow(row.ID, row.Path, row.Label, row.Enabled, row.CreatedAt)
	}
	return out
}

func TestRepositoryGetAll(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetStorageRootsQuery)).
		WillReturnRows(storageRootRows(
			StorageRootModel{ID: 1, Path: "/data", Label: "data", Enabled: true, CreatedAt: now},
			StorageRootModel{ID: 2, Path: "/midia", Label: "midia", Enabled: false, CreatedAt: now},
		))
	mock.ExpectRollback()

	models, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(models) != 2 || models[0].Path != "/data" || models[1].Enabled {
		t.Fatalf("unexpected models: %+v", models)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRepositoryGetAllQueryError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetStorageRootsQuery)).
		WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	if _, err := repo.GetAll(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRepositoryGetAllWithoutDatabaseFailsSoft(t *testing.T) {
	repo := NewRepository(database.NewDbContext(nil))
	if _, err := repo.GetAll(); err == nil {
		t.Fatalf("expected error with no database")
	}

	repoNilContext := NewRepository(nil)
	if _, err := repoNilContext.GetAll(); err == nil {
		t.Fatalf("expected error with nil context")
	}
}

func TestRepositoryGetByID(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetStorageRootByIDQuery)).
		WithArgs(1).
		WillReturnRows(storageRootRows(StorageRootModel{ID: 1, Path: "/data", Label: "data", Enabled: true, CreatedAt: now}))
	mock.ExpectRollback()

	model, found, err := repo.GetByID(1)
	if err != nil || !found || model.Path != "/data" {
		t.Fatalf("GetByID: model=%+v found=%v err=%v", model, found, err)
	}
}

func TestRepositoryGetByIDNotFound(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetStorageRootByIDQuery)).
		WithArgs(9).
		WillReturnRows(storageRootRows())
	mock.ExpectRollback()

	_, found, err := repo.GetByID(9)
	if err != nil || found {
		t.Fatalf("expected not found without error, got found=%v err=%v", found, err)
	}
}

func TestRepositoryCreate(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertStorageRootQuery)).
		WithArgs("/midia", "midia", true).
		WillReturnRows(storageRootRows(StorageRootModel{ID: 2, Path: "/midia", Label: "midia", Enabled: true, CreatedAt: now}))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		created, createErr := repo.Create(tx, StorageRootModel{Path: "/midia", Label: "midia", Enabled: true})
		if createErr != nil {
			return createErr
		}
		if created.ID != 2 {
			t.Fatalf("unexpected created: %+v", created)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestRepositoryUpdate(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpdateStorageRootQuery)).
		WithArgs(2, "renamed", false).
		WillReturnRows(storageRootRows(StorageRootModel{ID: 2, Path: "/midia", Label: "renamed", Enabled: false, CreatedAt: now}))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		updated, updateErr := repo.Update(tx, StorageRootModel{ID: 2, Label: "renamed", Enabled: false})
		if updateErr != nil {
			return updateErr
		}
		if updated.Label != "renamed" || updated.Enabled {
			t.Fatalf("unexpected updated: %+v", updated)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestRepositoryUpdateNotFound(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpdateStorageRootQuery)).
		WithArgs(9, "x", true).
		WillReturnRows(storageRootRows())
	mock.ExpectRollback()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, updateErr := repo.Update(tx, StorageRootModel{ID: 9, Label: "x", Enabled: true})
		return updateErr
	})
	if !errors.Is(err, ErrRootNotFound) {
		t.Fatalf("expected ErrRootNotFound, got %v", err)
	}
}

func TestRepositoryDelete(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteStorageRootQuery)).
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.Delete(tx, 2)
	})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestRepositoryDeleteNotFound(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteStorageRootQuery)).
		WithArgs(9).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.Delete(tx, 9)
	})
	if !errors.Is(err, ErrRootNotFound) {
		t.Fatalf("expected ErrRootNotFound, got %v", err)
	}
}
