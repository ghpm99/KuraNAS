package tiering

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/tiering"

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

func TestRepositoryGetSettingsDocumentFound(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetTieringSettingsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"setting_value"}).AddRow(`{"enabled":true}`))
	mock.ExpectRollback()

	document, found, err := repo.GetSettingsDocument()
	if err != nil || !found || document != `{"enabled":true}` {
		t.Fatalf("unexpected result: %q %v %v", document, found, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryGetSettingsDocumentMissing(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetTieringSettingsQuery)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, found, err := repo.GetSettingsDocument()
	if err != nil || found {
		t.Fatalf("expected clean miss, got found=%v err=%v", found, err)
	}
}

func TestRepositoryUpsertSettingsDocumentCommits(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.UpsertTieringSettingsQuery)).
		WithArgs(`{"enabled":true}`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.UpsertSettingsDocument(`{"enabled":true}`); err != nil {
		t.Fatalf("UpsertSettingsDocument: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryListDemotionCandidates(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	cutoff := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListDemotionCandidatesQuery)).
		WithArgs(int64(1024), cutoff).
		WillReturnRows(sqlmock.NewRows([]string{"id", "path", "size"}).
			AddRow(1, "/mnt/dados/a.txt", int64(2048)))
	mock.ExpectRollback()

	candidates, err := repo.ListDemotionCandidates(1024, cutoff)
	if err != nil || len(candidates) != 1 || candidates[0].FileID != 1 || candidates[0].LogicalPath != "/mnt/dados/a.txt" {
		t.Fatalf("unexpected result: %+v %v", candidates, err)
	}
}

func TestRepositoryListPromotionCandidates(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	cutoff := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ListPromotionCandidatesQuery)).
		WithArgs(cutoff).
		WillReturnRows(sqlmock.NewRows([]string{"id", "path", "physical_path", "size"}).
			AddRow(7, "/mnt/dados/c.txt", "/mnt/cold/Casa/c.txt", int64(4096)))
	mock.ExpectRollback()

	candidates, err := repo.ListPromotionCandidates(cutoff)
	if err != nil || len(candidates) != 1 || candidates[0].PhysicalPath != "/mnt/cold/Casa/c.txt" {
		t.Fatalf("unexpected result: %+v %v", candidates, err)
	}
}

func TestRepositorySetPhysicalPathDemote(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.SetPhysicalPathQuery)).
		WithArgs(3, sql.NullString{String: "/mnt/cold/Casa/a.txt", Valid: true}).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.SetPhysicalPath(3, "/mnt/cold/Casa/a.txt"); err != nil {
		t.Fatalf("SetPhysicalPath demote: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositorySetPhysicalPathPromoteClearsToNull(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(queries.SetPhysicalPathQuery)).
		WithArgs(3, sql.NullString{Valid: false}).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.SetPhysicalPath(3, ""); err != nil {
		t.Fatalf("SetPhysicalPath promote: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryGetLastRun(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetLastTieringJobQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "created_at", "started_at", "ended_at", "last_error"}).
			AddRow(7, "completed", now, now, now, ""))
	mock.ExpectRollback()

	run, found, err := repo.GetLastRun()
	if err != nil || !found || run.JobID != 7 || run.StartedAt == nil || run.EndedAt == nil {
		t.Fatalf("unexpected run: %+v %v %v", run, found, err)
	}
}

func TestRepositoryGetTierCounts(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetTierCountsQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"hot_files", "hot_bytes", "cold_files", "cold_bytes"}).
			AddRow(10, int64(1000), 4, int64(400)))
	mock.ExpectRollback()

	counts, err := repo.GetTierCounts()
	if err != nil || counts.HotFiles != 10 || counts.ColdBytes != 400 {
		t.Fatalf("unexpected counts: %+v %v", counts, err)
	}
}

func TestRepositoryGetFileByIdFoundCold(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetTieringFileByIdQuery)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "path", "physical_path", "size"}).
			AddRow(7, "/data/a.pdf", "/cold/a.pdf", 99))
	mock.ExpectRollback()

	file, found, err := repo.GetFileById(7)

	if err != nil || !found || !file.isCold() || file.PhysicalPath != "/cold/a.pdf" || file.Size != 99 {
		t.Fatalf("unexpected result: %+v %v %v", file, found, err)
	}
}

func TestRepositoryGetFileByIdHotAndMissing(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetTieringFileByIdQuery)).
		WithArgs(8).
		WillReturnRows(sqlmock.NewRows([]string{"id", "path", "physical_path", "size"}).
			AddRow(8, "/data/b.pdf", nil, 10))
	mock.ExpectRollback()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetTieringFileByIdQuery)).
		WithArgs(9).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	hotFile, hotFound, hotErr := repo.GetFileById(8)
	_, missingFound, missingErr := repo.GetFileById(9)

	if hotErr != nil || !hotFound || hotFile.isCold() {
		t.Fatalf("expected hot file, got %+v %v %v", hotFile, hotFound, hotErr)
	}
	if missingErr != nil || missingFound {
		t.Fatalf("expected clean miss, got found=%v err=%v", missingFound, missingErr)
	}
}

func TestRepositoryGetFileByIdWrapsError(t *testing.T) {
	repo, mock, db := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.GetTieringFileByIdQuery)).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	if _, _, err := repo.GetFileById(1); err == nil {
		t.Fatalf("expected error")
	}
}
