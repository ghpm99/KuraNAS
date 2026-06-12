package trash

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/trash"

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

func trashItemRows(rows ...TrashItemModel) *sqlmock.Rows {
	out := sqlmock.NewRows([]string{"id", "original_path", "trash_path", "size", "deleted_at"})
	for _, row := range rows {
		out.AddRow(row.ID, row.OriginalPath, row.TrashPath, row.Size, row.DeletedAt)
	}
	return out
}

func TestTrashRepository_CreateItem(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	if repo.GetDbContext() == nil {
		t.Fatalf("expected db context")
	}

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertTrashItemQuery)).
		WithArgs("/data/a.txt", "/data/.kuranas-trash/a.txt.1", int64(8), now).
		WillReturnRows(trashItemRows(TrashItemModel{
			ID: 1, OriginalPath: "/data/a.txt", TrashPath: "/data/.kuranas-trash/a.txt.1", Size: 8, DeletedAt: now,
		}))
	mock.ExpectCommit()

	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		created, createErr := repo.CreateItem(tx, TrashItemModel{
			OriginalPath: "/data/a.txt",
			TrashPath:    "/data/.kuranas-trash/a.txt.1",
			Size:         8,
			DeletedAt:    now,
		})
		if createErr != nil {
			return createErr
		}
		if created.ID != 1 {
			t.Fatalf("expected id 1, got %d", created.ID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestTrashRepository_GetItemsPagination(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	now := time.Now()
	// pageSize+1 rows returned → HasNext
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetTrashItemsQuery)).
		WithArgs(3, 0).
		WillReturnRows(trashItemRows(
			TrashItemModel{ID: 3, OriginalPath: "/c", TrashPath: "/t/c", DeletedAt: now},
			TrashItemModel{ID: 2, OriginalPath: "/b", TrashPath: "/t/b", DeletedAt: now},
			TrashItemModel{ID: 1, OriginalPath: "/a", TrashPath: "/t/a", DeletedAt: now},
		))
	mock.ExpectRollback()

	page, err := repo.GetItems(1, 2)
	if err != nil {
		t.Fatalf("GetItems: %v", err)
	}
	if len(page.Items) != 2 || !page.Pagination.HasNext || page.Pagination.HasPrev {
		t.Fatalf("unexpected pagination: items=%d %+v", len(page.Items), page.Pagination)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestTrashRepository_GetItemByID(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetTrashItemByIDQuery)).
		WithArgs(7).
		WillReturnRows(trashItemRows(TrashItemModel{ID: 7, OriginalPath: "/a", TrashPath: "/t/a", DeletedAt: now}))
	mock.ExpectRollback()

	item, found, err := repo.GetItemByID(7)
	if err != nil || !found || item.ID != 7 {
		t.Fatalf("GetItemByID: item=%+v found=%v err=%v", item, found, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetTrashItemByIDQuery)).
		WithArgs(8).
		WillReturnRows(trashItemRows())
	mock.ExpectRollback()

	_, found, err = repo.GetItemByID(8)
	if err != nil || found {
		t.Fatalf("missing row must report found=false, got found=%v err=%v", found, err)
	}
}

func TestTrashRepository_ExpiredAllAndDelete(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	cutoff := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetExpiredTrashItemsQuery)).
		WithArgs(cutoff).
		WillReturnRows(trashItemRows(TrashItemModel{ID: 1, OriginalPath: "/a", TrashPath: "/t/a", DeletedAt: cutoff.AddDate(0, 0, -40)}))
	mock.ExpectRollback()

	expired, err := repo.GetExpiredItems(cutoff)
	if err != nil || len(expired) != 1 {
		t.Fatalf("GetExpiredItems: len=%d err=%v", len(expired), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAllTrashItemsQuery)).
		WillReturnRows(trashItemRows(
			TrashItemModel{ID: 1, OriginalPath: "/a", TrashPath: "/t/a", DeletedAt: cutoff},
			TrashItemModel{ID: 2, OriginalPath: "/b", TrashPath: "/t/b", DeletedAt: cutoff},
		))
	mock.ExpectRollback()

	all, err := repo.GetAllItems()
	if err != nil || len(all) != 2 {
		t.Fatalf("GetAllItems: len=%d err=%v", len(all), err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteTrashItemQuery)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.DeleteItem(tx, 1)
	})
	if err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteTrashItemQuery)).
		WithArgs(9).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		return repo.DeleteItem(tx, 9)
	})
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for missing row, got %v", err)
	}
}

func TestTrashRepository_RetentionDays(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetRetentionDaysQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"setting_value"}).AddRow("45"))
	mock.ExpectRollback()

	days, found, err := repo.GetRetentionDays()
	if err != nil || !found || days != 45 {
		t.Fatalf("GetRetentionDays: days=%d found=%v err=%v", days, found, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetRetentionDaysQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"setting_value"}))
	mock.ExpectRollback()

	_, found, err = repo.GetRetentionDays()
	if err != nil || found {
		t.Fatalf("unset retention must report found=false, got found=%v err=%v", found, err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpsertRetentionDaysQuery)).
		WithArgs("30").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.SetRetentionDays(30); err != nil {
		t.Fatalf("SetRetentionDays: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
