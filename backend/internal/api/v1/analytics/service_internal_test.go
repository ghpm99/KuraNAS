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

func TestResolvePeriodAllBranches(t *testing.T) {
	tests := []struct {
		input    string
		label    string
		interval string
		wantErr  bool
	}{
		{"", "7d", "7 days", false},
		{"7d", "7d", "7 days", false},
		{"24h", "24h", "24 hours", false},
		{"30d", "30d", "30 days", false},
		{"90d", "90d", "90 days", false},
		{"invalid", "", "", true},
	}

	for _, tc := range tests {
		t.Run("period="+tc.input, func(t *testing.T) {
			cfg, err := resolvePeriod(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Label != tc.label || cfg.Interval != tc.interval {
				t.Fatalf("got label=%s interval=%s, want label=%s interval=%s", cfg.Label, cfg.Interval, tc.label, tc.interval)
			}
		})
	}
}

func TestToTimeSeriesDto(t *testing.T) {
	now := time.Now()
	models := []StorageTimeSeriesModel{
		{Date: now, UsedBytes: 100},
		{Date: now.Add(24 * time.Hour), UsedBytes: 200},
	}
	result := toTimeSeriesDto(models)
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
	if result[0].UsedBytes != 100 || result[1].UsedBytes != 200 {
		t.Fatalf("unexpected used bytes")
	}
}

func TestToTimeSeriesDtoEmpty(t *testing.T) {
	result := toTimeSeriesDto(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty result")
	}
}

func TestToTypeBreakdownDto(t *testing.T) {
	models := []TypeDistributionModel{
		{Category: "image", Count: 10, Bytes: 1000},
	}
	result := toTypeBreakdownDto(models)
	if len(result) != 1 || result[0].Type != "image" || result[0].Count != 10 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestToExtensionDto(t *testing.T) {
	models := []ExtensionDistributionModel{
		{Extension: ".jpg", Count: 5, Bytes: 500},
	}
	result := toExtensionDto(models)
	if len(result) != 1 || result[0].Ext != ".jpg" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestToHotFolderDto(t *testing.T) {
	now := time.Now()
	models := []FolderHotModel{
		{ParentPath: "/photos", NewFiles: 3, AddedBytes: 300, LastEvent: sql.NullTime{Time: now, Valid: true}},
		{ParentPath: "/docs", NewFiles: 1, AddedBytes: 100, LastEvent: sql.NullTime{Valid: false}},
	}
	result := toHotFolderDto(models)
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
	if result[0].LastEvent == "" {
		t.Fatalf("expected non-empty last event for valid time")
	}
	if result[1].LastEvent != "" {
		t.Fatalf("expected empty last event for invalid time")
	}
}

func TestToFolderUsageDto(t *testing.T) {
	now := time.Now()
	models := []FolderUsageModel{
		{ParentPath: "/media", TotalFiles: 10, TotalBytes: 5000, LastModified: sql.NullTime{Time: now, Valid: true}},
		{ParentPath: "/tmp", TotalFiles: 2, TotalBytes: 100, LastModified: sql.NullTime{Valid: false}},
	}
	result := toFolderUsageDto(models)
	if len(result) != 2 {
		t.Fatalf("expected 2 items")
	}
	if result[0].LastModified == "" {
		t.Fatalf("expected non-empty last modified for valid time")
	}
	if result[1].LastModified != "" {
		t.Fatalf("expected empty last modified for invalid time")
	}
}

func TestToRecentFilesDto(t *testing.T) {
	now := time.Now()
	models := []RecentFileModel{
		{ID: 1, Name: "file.jpg", Path: "/photos/file.jpg", ParentPath: "/photos", Size: 1024, Format: "image/jpeg", CreatedAt: now, UpdatedAt: now},
	}
	result := toRecentFilesDto(models)
	if len(result) != 1 || result[0].Name != "file.jpg" || result[0].SizeBytes != 1024 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestToDuplicateGroupDto(t *testing.T) {
	models := []DuplicateGroupModel{
		{Signature: "abc123", Copies: 3, ItemSize: 1024, ReclaimableSize: 2048, Paths: []string{"/a", "/b", "/c"}},
	}
	result := toDuplicateGroupDto(models)
	if len(result) != 1 || result[0].Copies != 3 || len(result[0].Paths) != 3 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestToHealthDto(t *testing.T) {
	t.Run("ok status when invalid health", func(t *testing.T) {
		data := OverviewDataModel{
			HealthStatus: sql.NullString{Valid: false},
		}
		result := toHealthDto(data)
		if result.Status != "ok" {
			t.Fatalf("expected ok, got %s", result.Status)
		}
	})

	t.Run("scanning status for PENDING", func(t *testing.T) {
		data := OverviewDataModel{
			HealthStatus: sql.NullString{String: "PENDING", Valid: true},
		}
		result := toHealthDto(data)
		if result.Status != "scanning" {
			t.Fatalf("expected scanning, got %s", result.Status)
		}
	})

	t.Run("ok status for unknown value", func(t *testing.T) {
		data := OverviewDataModel{
			HealthStatus: sql.NullString{String: "COMPLETED", Valid: true},
		}
		result := toHealthDto(data)
		if result.Status != "ok" {
			t.Fatalf("expected ok, got %s", result.Status)
		}
	})

	t.Run("last scan times", func(t *testing.T) {
		now := time.Now()
		data := OverviewDataModel{
			HealthStatus:  sql.NullString{Valid: false},
			LastScanStart: sql.NullTime{Time: now.Add(-60 * time.Second), Valid: true},
			LastScanEnd:   sql.NullTime{Time: now, Valid: true},
		}
		result := toHealthDto(data)
		if result.LastScanAt == "" {
			t.Fatalf("expected non-empty last scan at")
		}
		if result.LastScanSeconds != 60 {
			t.Fatalf("expected 60 seconds, got %d", result.LastScanSeconds)
		}
	})

	t.Run("no scan times", func(t *testing.T) {
		data := OverviewDataModel{
			HealthStatus:  sql.NullString{Valid: false},
			LastScanStart: sql.NullTime{Valid: false},
			LastScanEnd:   sql.NullTime{Valid: false},
		}
		result := toHealthDto(data)
		if result.LastScanAt != "" {
			t.Fatalf("expected empty last scan at")
		}
		if result.LastScanSeconds != 0 {
			t.Fatalf("expected 0 seconds")
		}
	})

	t.Run("errors without description", func(t *testing.T) {
		data := OverviewDataModel{
			HealthStatus: sql.NullString{Valid: false},
			RecentErrors: []LogErrorModel{
				{Name: "Error1", Description: sql.NullString{Valid: false}, CreatedAt: time.Now()},
			},
		}
		result := toHealthDto(data)
		if len(result.RecentErrors) != 1 || result.RecentErrors[0] != "Error1" {
			t.Fatalf("unexpected errors: %v", result.RecentErrors)
		}
	})

	t.Run("errors with empty description", func(t *testing.T) {
		data := OverviewDataModel{
			HealthStatus: sql.NullString{Valid: false},
			RecentErrors: []LogErrorModel{
				{Name: "Error2", Description: sql.NullString{String: "", Valid: true}, CreatedAt: time.Now()},
			},
		}
		result := toHealthDto(data)
		if result.RecentErrors[0] != "Error2" {
			t.Fatalf("expected just name, got %s", result.RecentErrors[0])
		}
	})
}

func TestServiceGetOverviewRepoError(t *testing.T) {
	stub := &repositoryStub{err: sql.ErrNoRows}
	service := NewService(stub)
	_, err := service.GetOverview("7d")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetOverviewWithAllData(t *testing.T) {
	now := time.Now().UTC()
	stub := &repositoryStub{response: OverviewDataModel{
		StorageKpis: StorageKpisModel{UsedBytes: 500, GrowthBytes: 50, FilesAdded: 10, FilesTotal: 100, FoldersTotal: 20},
		TimeSeries:  []StorageTimeSeriesModel{{Date: now, UsedBytes: 500}},
		Types:       []TypeDistributionModel{{Category: "image", Count: 50, Bytes: 250}},
		Extensions:  []ExtensionDistributionModel{{Extension: ".png", Count: 30, Bytes: 150}},
		HotFolders:  []FolderHotModel{{ParentPath: "/hot", NewFiles: 5, AddedBytes: 100, LastEvent: sql.NullTime{Time: now, Valid: true}}},
		TopFolders:  []FolderUsageModel{{ParentPath: "/top", TotalFiles: 40, TotalBytes: 2000, LastModified: sql.NullTime{Time: now, Valid: true}}},
		RecentFiles: []RecentFileModel{{ID: 1, Name: "new.jpg", Path: "/new.jpg", ParentPath: "/", Size: 1024, Format: "image/jpeg", CreatedAt: now, UpdatedAt: now}},
		Duplicates:  DuplicatesSummaryModel{GroupsTotal: 2, FilesTotal: 6, ReclaimableBytes: 4096},
		TopDuplicateSets: []DuplicateGroupModel{
			{Signature: "abc", Copies: 3, ItemSize: 1024, ReclaimableSize: 2048, Paths: []string{"/a", "/b", "/c"}},
		},
		HealthStatus: sql.NullString{String: "PENDING", Valid: true},
	}}

	service := NewService(stub)
	result, err := service.GetOverview("")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if result.Period != "7d" {
		t.Fatalf("expected default period 7d, got %s", result.Period)
	}
	if len(result.TimeSeries) != 1 {
		t.Fatalf("expected 1 time series point")
	}
	if len(result.Types) != 1 {
		t.Fatalf("expected 1 type breakdown")
	}
	if len(result.Extensions) != 1 {
		t.Fatalf("expected 1 extension")
	}
	if len(result.HotFolders) != 1 {
		t.Fatalf("expected 1 hot folder")
	}
	if len(result.TopFolders) != 1 {
		t.Fatalf("expected 1 top folder")
	}
	if len(result.RecentFiles) != 1 {
		t.Fatalf("expected 1 recent file")
	}
	if result.Duplicates.Groups != 2 {
		t.Fatalf("expected 2 duplicate groups")
	}
	if result.Health.Status != "scanning" {
		t.Fatalf("expected scanning status, got %s", result.Health.Status)
	}
}

func TestNewServiceReturnsCorrectType(t *testing.T) {
	repo := &repositoryStub{}
	svc := NewService(repo)
	typed, ok := svc.(*Service)
	if !ok || typed.Repository != repo {
		t.Fatalf("expected concrete service with repository")
	}
}
