package analytics

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"nas-go/api/pkg/database"
	queries "nas-go/api/pkg/database/queries/analytics"

	"github.com/DATA-DOG/go-sqlmock"
)

func newAnalyticsRepoWithMock(t *testing.T) (*Repository, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewRepository(database.NewDbContext(db)), mock, db
}

var period7d = PeriodConfig{Label: "7d", Interval: "7 days"}

func TestRepositoryGetStorageKpis(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.StorageKPIsQuery)).
		WithArgs("7 days").
		WillReturnRows(sqlmock.NewRows([]string{"used", "growth", "added", "total", "folders"}).
			AddRow(100, 10, 2, 5, 3))
	mock.ExpectRollback()

	result, err := repo.GetStorageKpis(period7d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UsedBytes != 100 || result.FoldersTotal != 3 {
		t.Fatalf("unexpected kpis: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryGetStorageKpisError(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.StorageKPIsQuery)).
		WithArgs("7 days").
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	if _, err := repo.GetStorageKpis(period7d); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRepositoryGetStorageTimeSeries(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.StorageTimeSeriesQuery)).
		WithArgs("7 days").
		WillReturnRows(sqlmock.NewRows([]string{"day", "used"}).AddRow(now, 500))
	mock.ExpectRollback()

	result, err := repo.GetStorageTimeSeries(period7d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].UsedBytes != 500 {
		t.Fatalf("unexpected timeseries: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryGetTypeDistribution(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.TypeDistributionQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"category", "count", "bytes"}).AddRow("image", 10, 1000))
	mock.ExpectRollback()

	result, err := repo.GetTypeDistribution()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Category != "image" {
		t.Fatalf("unexpected types: %+v", result)
	}
}

func TestRepositoryGetExtensionDistributionNormalizesNone(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ExtensionDistributionQuery)).
		WithArgs(12).
		WillReturnRows(sqlmock.NewRows([]string{"ext", "count", "bytes"}).
			AddRow("<none>", 3, 30).
			AddRow(".jpg", 5, 500))
	mock.ExpectRollback()

	result, err := repo.GetExtensionDistribution(12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].Extension != "unknown" || result[1].Extension != ".jpg" {
		t.Fatalf("unexpected normalization: %+v", result)
	}
}

func TestRepositoryGetRecentFiles(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.RecentFilesQuery)).
		WithArgs(50).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "parent", "size", "format", "created", "updated"}).
			AddRow(1, "f.jpg", "/f.jpg", "/", 1024, ".jpg", now, now))
	mock.ExpectRollback()

	result, err := repo.GetRecentFiles(50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Name != "f.jpg" {
		t.Fatalf("unexpected recent files: %+v", result)
	}
}

func TestRepositoryGetTopFolders(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.FolderSizeRankQuery)).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows([]string{"parent", "files", "bytes", "modified"}).
			AddRow("/media", 10, 5000, now))
	mock.ExpectRollback()

	result, err := repo.GetTopFolders(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].ParentPath != "/media" {
		t.Fatalf("unexpected top folders: %+v", result)
	}
}

func TestRepositoryGetHotFolders(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.FolderHotRankQuery)).
		WithArgs("7 days", 3).
		WillReturnRows(sqlmock.NewRows([]string{"parent", "new", "bytes", "event"}).
			AddRow("/hot", 5, 100, now))
	mock.ExpectRollback()

	result, err := repo.GetHotFolders(period7d, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].ParentPath != "/hot" {
		t.Fatalf("unexpected hot folders: %+v", result)
	}
}

func TestRepositoryGetDuplicatesSummary(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.DuplicatesSummaryQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"groups", "files", "reclaimable"}).AddRow(2, 6, 4096))
	mock.ExpectRollback()

	result, err := repo.GetDuplicatesSummary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GroupsTotal != 2 || result.ReclaimableBytes != 4096 {
		t.Fatalf("unexpected duplicates summary: %+v", result)
	}
}

func TestRepositoryGetDuplicateGroups(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.DuplicatesTopGroupsQuery)).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows([]string{"sig", "copies", "size", "reclaimable", "paths"}).
			AddRow("abc", 3, 1024, 2048, "{/a,/b,/c}"))
	mock.ExpectRollback()

	result, err := repo.GetDuplicateGroups(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Copies != 3 || len(result[0].Paths) != 3 {
		t.Fatalf("unexpected duplicate groups: %+v", result)
	}
}

func TestRepositoryGetLibrarySummary(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.LibraryMetadataSummaryQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"cat", "audio", "video", "image", "classified"}).
			AddRow(4, 1, 2, 1, 1))
	mock.ExpectRollback()

	result, err := repo.GetLibrarySummary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CategorizedMedia != 4 || result.VideoWithMetadata != 2 {
		t.Fatalf("unexpected library summary: %+v", result)
	}
}

func TestRepositoryGetProcessingSummary(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.ProcessingQueueSummaryQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"mp", "mf", "tp", "tf", "rt"}).AddRow(2, 1, 3, 1, 0))
	mock.ExpectRollback()

	result, err := repo.GetProcessingSummary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MetadataPending != 2 || result.ThumbnailPending != 3 {
		t.Fatalf("unexpected processing summary: %+v", result)
	}
}

func TestRepositoryGetHealth(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.IndexHealthStatusQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "start", "end"}).
			AddRow("failed", now.Add(-time.Minute), now))
	mock.ExpectQuery(regexp.QuoteMeta(queries.IndexFilesTotalQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))
	mock.ExpectQuery(regexp.QuoteMeta(queries.IndexErrorsRecentQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(regexp.QuoteMeta(queries.IndexErrorsLatestQuery)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"name", "desc", "created"}).
			AddRow("ScanFiles", "boom", now))
	mock.ExpectRollback()

	result, err := repo.GetHealth()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IndexedFiles != 42 || result.ErrorsLast24h != 3 || len(result.RecentErrors) != 1 {
		t.Fatalf("unexpected health: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepositoryGetHealthNoScanRow(t *testing.T) {
	repo, mock, db := newAnalyticsRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(queries.IndexHealthStatusQuery)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(queries.IndexFilesTotalQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(queries.IndexErrorsRecentQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(queries.IndexErrorsLatestQuery)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"name", "desc", "created"}))
	mock.ExpectRollback()

	result, err := repo.GetHealth()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status.Valid {
		t.Fatalf("expected invalid status when no scan row")
	}
}
