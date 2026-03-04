package database

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewDbContextAndGetDatabase(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	ctx := NewDbContext(db)
	if ctx.GetDatabase() != db {
		t.Fatalf("expected same database pointer")
	}
}

func TestExecTx(t *testing.T) {
	t.Run("begin error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer db.Close()

		mock.ExpectBegin().WillReturnError(errors.New("begin failed"))
		ctx := NewDbContext(db)
		err = ctx.ExecTx(func(tx *sql.Tx) error { return nil })
		if err == nil {
			t.Fatalf("expected begin error")
		}
	})

	t.Run("callback error rolls back", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectRollback()

		ctx := NewDbContext(db)
		err = ctx.ExecTx(func(tx *sql.Tx) error { return errors.New("callback failed") })
		if err == nil {
			t.Fatalf("expected callback error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})

	t.Run("success commits", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectCommit()

		ctx := NewDbContext(db)
		err = ctx.ExecTx(func(tx *sql.Tx) error { return nil })
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})

	t.Run("commit error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

		ctx := NewDbContext(db)
		err = ctx.ExecTx(func(tx *sql.Tx) error { return nil })
		if err == nil {
			t.Fatalf("expected commit error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})
}

func TestQueryTx(t *testing.T) {
	t.Run("begin error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer db.Close()

		mock.ExpectBegin().WillReturnError(errors.New("begin failed"))
		ctx := NewDbContext(db)
		err = ctx.QueryTx(func(tx *sql.Tx) error { return nil })
		if err == nil {
			t.Fatalf("expected begin error")
		}
	})

	t.Run("callback error rolls back", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectRollback()

		ctx := NewDbContext(db)
		err = ctx.QueryTx(func(tx *sql.Tx) error { return errors.New("query callback failed") })
		if err == nil {
			t.Fatalf("expected callback error")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})

	t.Run("success returns nil and rolls back deferred tx", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectRollback()

		ctx := NewDbContext(db)
		err = ctx.QueryTx(func(tx *sql.Tx) error { return nil })
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})
}
