package accesscontrol

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/accesscontrol"

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

func allowedIPRows(rows ...AllowedIPModel) *sqlmock.Rows {
	out := sqlmock.NewRows([]string{"id", "cidr", "label", "enabled", "created_at"})
	for _, row := range rows {
		out.AddRow(row.ID, row.CIDR, row.Label, row.Enabled, row.CreatedAt)
	}
	return out
}

func TestRepositoryGetAllAndGetByID(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	if repo.GetDbContext() == nil {
		t.Fatalf("expected db context")
	}

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAllowedIPsQuery)).
		WillReturnRows(allowedIPRows(
			AllowedIPModel{ID: 1, CIDR: "192.168.1.10/32", Label: "a", Enabled: true, CreatedAt: now},
			AllowedIPModel{ID: 2, CIDR: "10.0.0.0/8", Label: "b", Enabled: false, CreatedAt: now},
		))
	mock.ExpectRollback()

	all, err := repo.GetAll()
	if err != nil || len(all) != 2 {
		t.Fatalf("GetAll failed len=%d err=%v", len(all), err)
	}
	if all[0].CIDR != "192.168.1.10/32" || all[1].Enabled {
		t.Fatalf("unexpected rows: %+v", all)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAllowedIPByIDQuery)).
		WithArgs(1).
		WillReturnRows(allowedIPRows(AllowedIPModel{ID: 1, CIDR: "192.168.1.10/32", Enabled: true, CreatedAt: now}))
	mock.ExpectRollback()

	byID, err := repo.GetByID(1)
	if err != nil || byID.ID != 1 {
		t.Fatalf("GetByID failed %+v err=%v", byID, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAllowedIPByIDQuery)).
		WithArgs(9).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	if _, err := repo.GetByID(9); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestRepositoryCreateUpdateDelete(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.CreateAllowedIPQuery)).
		WithArgs("192.168.1.10/32", "notebook", true).
		WillReturnRows(allowedIPRows(AllowedIPModel{ID: 7, CIDR: "192.168.1.10/32", Label: "notebook", Enabled: true, CreatedAt: now}))
	mock.ExpectCommit()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		created, createErr := repo.Create(tx, AllowedIPModel{CIDR: "192.168.1.10/32", Label: "notebook", Enabled: true})
		if createErr != nil {
			return createErr
		}
		if created.ID != 7 {
			t.Fatalf("expected created id 7, got %d", created.ID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpdateAllowedIPQuery)).
		WithArgs(7, "192.168.1.0/24", "lan", false).
		WillReturnRows(allowedIPRows(AllowedIPModel{ID: 7, CIDR: "192.168.1.0/24", Label: "lan", Enabled: false, CreatedAt: now}))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		updated, updateErr := repo.Update(tx, AllowedIPModel{ID: 7, CIDR: "192.168.1.0/24", Label: "lan", Enabled: false})
		if updateErr != nil {
			return updateErr
		}
		if updated.Enabled {
			t.Fatalf("expected disabled entry, got %+v", updated)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteAllowedIPQuery)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.Delete(tx, 7)
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteAllowedIPQuery)).
		WithArgs(8).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.Delete(tx, 8)
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for missing row, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
