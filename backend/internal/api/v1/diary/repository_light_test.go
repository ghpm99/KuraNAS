package diary

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/diary"

	"github.com/DATA-DOG/go-sqlmock"
)

func newDiaryRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewRepository(database.NewDbContext(db)), mock, db
}

func TestDiaryRepositoryBasics(t *testing.T) {
	repo, mock, db := newDiaryRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	if repo.GetDbContext() == nil {
		t.Fatalf("expected db context")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetDiaryQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "start_time", "end_time"}).
			AddRow(1, "d", "desc", now, nil))
	mock.ExpectRollback()
	out, err := repo.GetDiary(DiaryFilter{}, 1, 10)
	if err != nil || len(out.Items) != 1 {
		t.Fatalf("GetDiary failed len=%d err=%v", len(out.Items), err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetDiarySummaryQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"date", "total_activities", "total_time_spent_seconds", "longest_name", "longest_seconds"}).
			AddRow(now, 2, 3600, "work", 1800))
	mock.ExpectRollback()
	summary, err := repo.GetSummary(now)
	if err != nil || summary.TotalActivities != 2 {
		t.Fatalf("GetSummary failed summary=%+v err=%v", summary, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestDiaryRepositoryCreateAndUpdate(t *testing.T) {
	repo, mock, db := newDiaryRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertDiaryQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectCommit()
	err := repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		created, err := repo.CreateDiary(tx, DiaryModel{Name: "d", Description: "x", StartTime: now})
		if err != nil {
			return err
		}
		if created.ID != 10 {
			t.Fatalf("expected ID 10")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("CreateDiary failed: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateDiaryQuery)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = repo.GetDbContext().ExecTx(func(tx *sql.Tx) error {
		_, err := repo.UpdateDiary(tx, DiaryModel{ID: 10, Name: "d2", Description: "x2", StartTime: now})
		return err
	})
	if err != nil {
		t.Fatalf("UpdateDiary failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestDiaryRepositoryUpdateDiaryErrorBranches(t *testing.T) {
	repo, mock, db := newDiaryRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateDiaryQuery)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectRollback()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	_, err = repo.UpdateDiary(tx, DiaryModel{ID: 1, Name: "x", Description: "y", StartTime: now})
	if err == nil || !strings.Contains(err.Error(), "multiple rows affected") {
		t.Fatalf("expected multiple rows affected error, got %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpdateDiaryQuery)).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected failed")))

	tx2, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx2: %v", err)
	}
	defer tx2.Rollback()
	_, err = repo.UpdateDiary(tx2, DiaryModel{ID: 1, Name: "x", Description: "y", StartTime: now})
	if err == nil || !strings.Contains(err.Error(), "rows affected failed") {
		t.Fatalf("expected rows affected error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestDiaryRepositoryGetSummaryAndGetDiaryErrorBranches(t *testing.T) {
	repo, mock, db := newDiaryRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetDiarySummaryQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"date", "total_activities", "total_time_spent_seconds", "longest_name", "longest_seconds"}).
			AddRow(now, 0, 0, nil, nil))
	mock.ExpectRollback()

	summary, err := repo.GetSummary(now)
	if err != nil {
		t.Fatalf("expected GetSummary success with null longest activity, got %v", err)
	}
	if summary.LongestActivity != nil {
		t.Fatalf("expected nil longest activity when database returns null values")
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetDiaryQuery)).
		WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()

	_, err = repo.GetDiary(DiaryFilter{}, 1, 10)
	if err == nil || !strings.Contains(err.Error(), "falha ao obter diário") {
		t.Fatalf("expected wrapped get diary error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
