package analytics

import (
	"database/sql"
	"testing"
	"time"

	"nas-go/api/pkg/database"
)

type repositoryStub struct {
	response OverviewDataModel
	err      error
	period   PeriodConfig
}

func (stub *repositoryStub) GetDbContext() *database.DbContext { return nil }

func (stub *repositoryStub) GetOverviewData(period PeriodConfig, limits OverviewLimits) (OverviewDataModel, error) {
	stub.period = period
	return stub.response, stub.err
}

func TestServiceGetOverviewRejectsInvalidPeriod(t *testing.T) {
	service := NewService(&repositoryStub{})
	_, err := service.GetOverview("2d")
	if err == nil {
		t.Fatalf("expected invalid period error")
	}
	if err != nil && err.Error() == "" {
		t.Fatalf("expected non-empty error")
	}
}

func TestServiceGetOverviewMapsHealthAndPeriod(t *testing.T) {
	now := time.Now().UTC()
	stub := &repositoryStub{response: OverviewDataModel{
		StorageKpis:    StorageKpisModel{UsedBytes: 100, GrowthBytes: 10, FilesAdded: 2, FilesTotal: 5, FoldersTotal: 3},
		LibrarySummary: LibrarySummaryModel{CategorizedMedia: 4, AudioWithMetadata: 1, VideoWithMetadata: 2, ImageWithMetadata: 1, ImageClassified: 1},
		Processing:     ProcessingSummaryModel{MetadataPending: 2, MetadataFailed: 1, ThumbnailPending: 3, ThumbnailFailed: 1},
		HealthStatus:   sql.NullString{String: "FAILED", Valid: true},
		LastScanStart:  sql.NullTime{Time: now.Add(-2 * time.Minute), Valid: true},
		LastScanEnd:    sql.NullTime{Time: now.Add(-1 * time.Minute), Valid: true},
		ErrorsLast24h:  3,
		RecentErrors:   []LogErrorModel{{Name: "ScanFiles", Description: sql.NullString{String: "boom", Valid: true}, CreatedAt: now}},
	}}

	service := NewService(stub)
	result, err := service.GetOverview("30d")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if stub.period.Interval != "30 days" {
		t.Fatalf("expected interval 30 days, got %s", stub.period.Interval)
	}
	if result.Health.Status != "error" {
		t.Fatalf("expected health status error, got %s", result.Health.Status)
	}
	if result.Counts.FilesTotal != 5 {
		t.Fatalf("expected files total 5, got %d", result.Counts.FilesTotal)
	}
	if result.Library.CategorizedMedia != 4 {
		t.Fatalf("expected categorized media 4, got %d", result.Library.CategorizedMedia)
	}
	if result.Processing.ThumbnailPending != 3 {
		t.Fatalf("expected thumbnail pending 3, got %d", result.Processing.ThumbnailPending)
	}
	if len(result.Health.RecentErrors) != 1 {
		t.Fatalf("expected 1 recent error")
	}
}
