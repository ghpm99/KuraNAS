package assistant

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/assistant"

	"github.com/DATA-DOG/go-sqlmock"
)

func newRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewRepository(database.NewDbContext(db)), mock, db
}

func TestRepoCreateConversation(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertConversationQuery)).
		WithArgs("Olá").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "created_at", "updated_at"}).AddRow(1, "Olá", now, now))
	mock.ExpectCommit()

	model, err := repo.CreateConversation("Olá")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.ID != 1 || model.Title != "Olá" {
		t.Fatalf("unexpected model: %+v", model)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRepoCreateConversationError(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertConversationQuery)).
		WithArgs("x").WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	if _, err := repo.CreateConversation("x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRepoConversationExists(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ConversationExistsQuery)).
		WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.ConversationExists(5)
	if err != nil || !exists {
		t.Fatalf("expected exists true, got %v err=%v", exists, err)
	}
}

func TestRepoConversationExistsError(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ConversationExistsQuery)).
		WithArgs(5).WillReturnError(errors.New("boom"))

	if _, err := repo.ConversationExists(5); err == nil {
		t.Fatal("expected error")
	}
}

func TestRepoListConversations(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListConversationsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "created_at", "updated_at"}).
			AddRow(1, "A", now, now).
			AddRow(2, "B", now, now))

	conversations, err := repo.ListConversations()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conversations) != 2 {
		t.Fatalf("expected 2 conversations, got %d", len(conversations))
	}
}

func TestRepoListConversationsQueryError(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListConversationsQuery)).
		WillReturnError(errors.New("boom"))

	if _, err := repo.ListConversations(); err == nil {
		t.Fatal("expected error")
	}
}

func TestRepoListConversationsScanError(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListConversationsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "created_at", "updated_at"}).
			AddRow("not-an-int", "A", time.Now(), time.Now()))

	if _, err := repo.ListConversations(); err == nil {
		t.Fatal("expected scan error")
	}
}

func TestRepoTouchConversation(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.TouchConversationQuery)).
		WithArgs(3).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.TouchConversation(3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRepoTouchConversationError(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.TouchConversationQuery)).
		WithArgs(3).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	if err := repo.TouchConversation(3); err == nil {
		t.Fatal("expected error")
	}
}

func TestRepoDeleteConversation(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteConversationQuery)).
		WithArgs(9).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.DeleteConversation(9); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRepoDeleteConversationError(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.DeleteConversationQuery)).
		WithArgs(9).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	if err := repo.DeleteConversation(9); err == nil {
		t.Fatal("expected error")
	}
}

func TestRepoAddMessage(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertMessageQuery)).
		WithArgs(2, RoleUser, "oi").
		WillReturnRows(sqlmock.NewRows([]string{"id", "conversation_id", "role", "content", "created_at"}).
			AddRow(10, 2, RoleUser, "oi", now))
	mock.ExpectCommit()

	model, err := repo.AddMessage(2, RoleUser, "oi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.ID != 10 || model.Content != "oi" {
		t.Fatalf("unexpected model: %+v", model)
	}
}

func TestRepoAddMessageError(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.InsertMessageQuery)).
		WithArgs(2, RoleUser, "oi").WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	if _, err := repo.AddMessage(2, RoleUser, "oi"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRepoListMessages(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListMessagesQuery)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "conversation_id", "role", "content", "created_at"}).
			AddRow(1, 2, RoleUser, "oi", now).
			AddRow(2, 2, RoleAssistant, "olá", now))

	messages, err := repo.ListMessages(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
}

func TestRepoListMessagesError(t *testing.T) {
	repo, mock, _ := newRepoWithMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListMessagesQuery)).
		WithArgs(2).WillReturnError(errors.New("boom"))

	if _, err := repo.ListMessages(2); err == nil {
		t.Fatal("expected error")
	}
}
