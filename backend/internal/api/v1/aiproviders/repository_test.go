package aiproviders

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/aiproviders"

	"github.com/DATA-DOG/go-sqlmock"
)

func newRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewRepository(database.NewDbContext(db)), mock, db
}

var providerColumns = []string{
	"id", "name", "enabled", "model", "base_url", "priority", "params", "created_at", "updated_at",
}

func TestRepositoryGetAll(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAIProvidersQuery)).
		WillReturnRows(sqlmock.NewRows(providerColumns).
			AddRow(1, "ollama", true, "llama3.1", "http://localhost:11434", 0, []byte(`{"timeout_seconds":120,"keep_alive":"5m"}`), now, now).
			AddRow(2, "openai", false, "gpt-4o-mini", "https://api.openai.com/v1", 1, []byte(`{}`), now, now))
	mock.ExpectRollback()

	models, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll error: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(models))
	}
	if models[0].Name != ProviderOllama || models[0].Params.TimeoutSeconds != 120 || models[0].Params.KeepAlive != "5m" {
		t.Fatalf("unexpected first provider: %+v", models[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryGetByName(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetAIProviderByNameQuery)).
		WithArgs("ollama").
		WillReturnRows(sqlmock.NewRows(providerColumns).
			AddRow(1, "ollama", true, "llama3.1", "http://localhost:11434", 0, []byte(`{"max_retries":2}`), now, now))
	mock.ExpectRollback()

	model, err := repo.GetByName(ProviderOllama)
	if err != nil {
		t.Fatalf("GetByName error: %v", err)
	}
	if model.Name != ProviderOllama || model.Params.MaxRetries != 2 {
		t.Fatalf("unexpected provider: %+v", model)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestRepositoryInsertIfAbsentCommits guards against the write being rolled back
// (the seed must persist), so it asserts a Commit is issued.
func TestRepositoryInsertIfAbsentCommits(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.InsertAIProviderIfAbsentQuery)).
		WithArgs("ollama", true, "llama3.1", "http://localhost:11434", 0, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.InsertIfAbsent(ProviderModel{
		Name:    ProviderOllama,
		Enabled: true,
		Model:   "llama3.1",
		BaseURL: "http://localhost:11434",
		Params:  ProviderParams{TimeoutSeconds: 120},
	})
	if err != nil {
		t.Fatalf("InsertIfAbsent error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryUpdateCommits(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.UpdateAIProviderQuery)).
		WithArgs("ollama", true, "qwen2.5", "http://nas:11434", 0, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(providerColumns).
			AddRow(1, "ollama", true, "qwen2.5", "http://nas:11434", 0, []byte(`{"timeout_seconds":300}`), now, now))
	mock.ExpectCommit()

	updated, err := repo.Update(ProviderModel{
		Name:    ProviderOllama,
		Enabled: true,
		Model:   "qwen2.5",
		BaseURL: "http://nas:11434",
		Params:  ProviderParams{TimeoutSeconds: 300},
	})
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if updated.Model != "qwen2.5" || updated.Params.TimeoutSeconds != 300 {
		t.Fatalf("unexpected updated provider: %+v", updated)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
