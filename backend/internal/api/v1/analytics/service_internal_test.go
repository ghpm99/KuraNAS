package analytics

import (
	"context"
	"database/sql"
	"errors"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/database"
	"testing"
	"time"
)

type repositoryStub struct {
	storageKpis      StorageKpisModel
	timeSeries       []StorageTimeSeriesModel
	types            []TypeDistributionModel
	extensions       []ExtensionDistributionModel
	recentFiles      []RecentFileModel
	topFolders       []FolderUsageModel
	hotFolders       []FolderHotModel
	duplicatesSum    DuplicatesSummaryModel
	duplicateGroups  []DuplicateGroupModel
	librarySummary   LibrarySummaryModel
	processing       ProcessingSummaryModel
	health           HealthModel
	err              error
	capturedPeriod   PeriodConfig
	capturedLimit    int
	capturedHotLimit int
}

func (stub *repositoryStub) GetDbContext() *database.DbContext { return nil }

func (stub *repositoryStub) GetStorageKpis(period PeriodConfig) (StorageKpisModel, error) {
	stub.capturedPeriod = period
	return stub.storageKpis, stub.err
}

func (stub *repositoryStub) GetStorageTimeSeries(period PeriodConfig) ([]StorageTimeSeriesModel, error) {
	stub.capturedPeriod = period
	return stub.timeSeries, stub.err
}

func (stub *repositoryStub) GetTypeDistribution() ([]TypeDistributionModel, error) {
	return stub.types, stub.err
}

func (stub *repositoryStub) GetExtensionDistribution(limit int) ([]ExtensionDistributionModel, error) {
	stub.capturedLimit = limit
	return stub.extensions, stub.err
}

func (stub *repositoryStub) GetRecentFiles(limit int) ([]RecentFileModel, error) {
	stub.capturedLimit = limit
	return stub.recentFiles, stub.err
}

func (stub *repositoryStub) GetTopFolders(limit int) ([]FolderUsageModel, error) {
	stub.capturedLimit = limit
	return stub.topFolders, stub.err
}

func (stub *repositoryStub) GetHotFolders(period PeriodConfig, limit int) ([]FolderHotModel, error) {
	stub.capturedPeriod = period
	stub.capturedHotLimit = limit
	return stub.hotFolders, stub.err
}

func (stub *repositoryStub) GetDuplicatesSummary() (DuplicatesSummaryModel, error) {
	return stub.duplicatesSum, stub.err
}

func (stub *repositoryStub) GetDuplicateGroups(limit int) ([]DuplicateGroupModel, error) {
	stub.capturedLimit = limit
	return stub.duplicateGroups, stub.err
}

func (stub *repositoryStub) GetLibrarySummary() (LibrarySummaryModel, error) {
	return stub.librarySummary, stub.err
}

func (stub *repositoryStub) GetProcessingSummary() (ProcessingSummaryModel, error) {
	return stub.processing, stub.err
}

func (stub *repositoryStub) GetHealth() (HealthModel, error) {
	return stub.health, stub.err
}

type analyticsAIMock struct {
	executeFn func(ctx context.Context, req ai.Request) (ai.Response, error)
}

func (m *analyticsAIMock) Execute(ctx context.Context, req ai.Request) (ai.Response, error) {
	return m.executeFn(ctx, req)
}

func TestServiceGetStorage(t *testing.T) {
	stub := &repositoryStub{storageKpis: StorageKpisModel{UsedBytes: 100, GrowthBytes: 10, FilesAdded: 2, FilesTotal: 5, FoldersTotal: 3}}
	service := NewService(stub, nil)

	result, err := service.GetStorage("30d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.capturedPeriod.Interval != "30 days" {
		t.Fatalf("expected interval 30 days, got %s", stub.capturedPeriod.Interval)
	}
	if result.Storage.UsedBytes != 100 || result.Counts.FilesTotal != 5 || result.Counts.Folders != 3 {
		t.Fatalf("unexpected storage stats: %+v", result)
	}
}

func TestServiceGetStorageInvalidPeriod(t *testing.T) {
	service := NewService(&repositoryStub{}, nil)
	if _, err := service.GetStorage("2d"); !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("expected ErrInvalidPeriod, got %v", err)
	}
}

func TestServiceGetStorageRepoError(t *testing.T) {
	service := NewService(&repositoryStub{err: sql.ErrConnDone}, nil)
	if _, err := service.GetStorage("7d"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetTimeSeries(t *testing.T) {
	now := time.Now()
	stub := &repositoryStub{timeSeries: []StorageTimeSeriesModel{{Date: now, UsedBytes: 100}}}
	service := NewService(stub, nil)

	result, err := service.GetTimeSeries("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].UsedBytes != 100 {
		t.Fatalf("unexpected time series: %+v", result)
	}

	if _, err := service.GetTimeSeries("bad"); !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("expected ErrInvalidPeriod, got %v", err)
	}

	service = NewService(&repositoryStub{err: sql.ErrConnDone}, nil)
	if _, err := service.GetTimeSeries("7d"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetTypes(t *testing.T) {
	stub := &repositoryStub{types: []TypeDistributionModel{{Category: "image", Count: 10, Bytes: 1000}}}
	service := NewService(stub, nil)

	result, err := service.GetTypes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Type != "image" {
		t.Fatalf("unexpected types: %+v", result)
	}

	if _, err := NewService(&repositoryStub{err: sql.ErrConnDone}, nil).GetTypes(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetExtensions(t *testing.T) {
	stub := &repositoryStub{extensions: []ExtensionDistributionModel{{Extension: ".jpg", Count: 5, Bytes: 500}}}
	service := NewService(stub, nil)

	result, err := service.GetExtensions(12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.capturedLimit != 12 || len(result) != 1 || result[0].Ext != ".jpg" {
		t.Fatalf("unexpected extensions: %+v limit=%d", result, stub.capturedLimit)
	}

	if _, err := NewService(&repositoryStub{err: sql.ErrConnDone}, nil).GetExtensions(5); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetRecentFiles(t *testing.T) {
	now := time.Now()
	stub := &repositoryStub{recentFiles: []RecentFileModel{{ID: 1, Name: "f.jpg", Size: 1024, CreatedAt: now, UpdatedAt: now}}}
	service := NewService(stub, nil)

	result, err := service.GetRecentFiles(50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.capturedLimit != 50 || len(result) != 1 || result[0].Name != "f.jpg" {
		t.Fatalf("unexpected recent files: %+v", result)
	}

	if _, err := NewService(&repositoryStub{err: sql.ErrConnDone}, nil).GetRecentFiles(5); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetTopFolders(t *testing.T) {
	now := time.Now()
	stub := &repositoryStub{topFolders: []FolderUsageModel{{ParentPath: "/m", TotalFiles: 10, TotalBytes: 5000, LastModified: sql.NullTime{Time: now, Valid: true}}}}
	service := NewService(stub, nil)

	result, err := service.GetTopFolders(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.capturedLimit != 20 || len(result) != 1 || result[0].Path != "/m" {
		t.Fatalf("unexpected top folders: %+v", result)
	}

	if _, err := NewService(&repositoryStub{err: sql.ErrConnDone}, nil).GetTopFolders(5); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetHotFolders(t *testing.T) {
	now := time.Now()
	stub := &repositoryStub{hotFolders: []FolderHotModel{{ParentPath: "/hot", NewFiles: 5, AddedBytes: 100, LastEvent: sql.NullTime{Time: now, Valid: true}}}}
	service := NewService(stub, nil)

	result, err := service.GetHotFolders("7d", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.capturedHotLimit != 3 || len(result) != 1 || result[0].Path != "/hot" {
		t.Fatalf("unexpected hot folders: %+v", result)
	}

	if _, err := service.GetHotFolders("bad", 3); !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("expected ErrInvalidPeriod, got %v", err)
	}

	if _, err := NewService(&repositoryStub{err: sql.ErrConnDone}, nil).GetHotFolders("7d", 3); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetDuplicatesSummary(t *testing.T) {
	stub := &repositoryStub{duplicatesSum: DuplicatesSummaryModel{GroupsTotal: 2, FilesTotal: 6, ReclaimableBytes: 4096}}
	service := NewService(stub, nil)

	result, err := service.GetDuplicatesSummary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Groups != 2 || result.Files != 6 || result.ReclaimableSize != 4096 {
		t.Fatalf("unexpected duplicates summary: %+v", result)
	}

	if _, err := NewService(&repositoryStub{err: sql.ErrConnDone}, nil).GetDuplicatesSummary(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetDuplicateGroups(t *testing.T) {
	stub := &repositoryStub{duplicateGroups: []DuplicateGroupModel{{Signature: "abc", Copies: 3, ItemSize: 1024, ReclaimableSize: 2048, Paths: []string{"/a", "/b", "/c"}}}}
	service := NewService(stub, nil)

	result, err := service.GetDuplicateGroups(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.capturedLimit != 20 || len(result) != 1 || result[0].Copies != 3 {
		t.Fatalf("unexpected duplicate groups: %+v", result)
	}

	if _, err := NewService(&repositoryStub{err: sql.ErrConnDone}, nil).GetDuplicateGroups(5); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetLibrary(t *testing.T) {
	stub := &repositoryStub{librarySummary: LibrarySummaryModel{CategorizedMedia: 4, AudioWithMetadata: 1, VideoWithMetadata: 2, ImageWithMetadata: 1, ImageClassified: 1}}
	service := NewService(stub, nil)

	result, err := service.GetLibrary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CategorizedMedia != 4 || result.VideoWithMetadata != 2 {
		t.Fatalf("unexpected library: %+v", result)
	}

	if _, err := NewService(&repositoryStub{err: sql.ErrConnDone}, nil).GetLibrary(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetProcessing(t *testing.T) {
	stub := &repositoryStub{processing: ProcessingSummaryModel{MetadataPending: 2, MetadataFailed: 1, ThumbnailPending: 3, ThumbnailFailed: 1, RecurringTimeouts: 0}}
	service := NewService(stub, nil)

	result, err := service.GetProcessing()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ThumbnailPending != 3 || result.MetadataPending != 2 {
		t.Fatalf("unexpected processing: %+v", result)
	}

	if _, err := NewService(&repositoryStub{err: sql.ErrConnDone}, nil).GetProcessing(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetHealth(t *testing.T) {
	now := time.Now()
	stub := &repositoryStub{health: HealthModel{
		Status:        sql.NullString{String: "FAILED", Valid: true},
		LastScanStart: sql.NullTime{Time: now.Add(-2 * time.Minute), Valid: true},
		LastScanEnd:   sql.NullTime{Time: now.Add(-1 * time.Minute), Valid: true},
		IndexedFiles:  42,
		ErrorsLast24h: 3,
		RecentErrors:  []LogErrorModel{{Name: "ScanFiles", Description: sql.NullString{String: "boom", Valid: true}, CreatedAt: now}},
	}}
	service := NewService(stub, nil)

	result, err := service.GetHealth()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "error" || result.IndexedFiles != 42 || result.LastScanSeconds != 60 || len(result.RecentErrors) != 1 {
		t.Fatalf("unexpected health: %+v", result)
	}

	if _, err := NewService(&repositoryStub{err: sql.ErrConnDone}, nil).GetHealth(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestServiceGetInsightsInvalidPeriod(t *testing.T) {
	service := NewService(&repositoryStub{}, nil)
	if _, err := service.GetInsights("2d"); !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("expected ErrInvalidPeriod, got %v", err)
	}
}

func TestServiceGetInsightsNilAI(t *testing.T) {
	service := NewService(&repositoryStub{}, nil)
	result, err := service.GetInsights("7d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty insights when AI is nil")
	}
}

func TestServiceGetInsightsSuccess(t *testing.T) {
	stub := &repositoryStub{
		storageKpis:   StorageKpisModel{UsedBytes: 500, FilesTotal: 100},
		duplicatesSum: DuplicatesSummaryModel{GroupsTotal: 2, ReclaimableBytes: 1000},
		hotFolders:    []FolderHotModel{{ParentPath: "/hot", NewFiles: 5}},
	}
	aiMock := &analyticsAIMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			if req.TaskType != ai.TaskSummarization {
				t.Fatalf("expected summarization task, got %s", req.TaskType)
			}
			return ai.Response{Content: `["Storage is healthy", "Consider removing duplicates"]`}, nil
		},
	}

	service := NewService(stub, aiMock)
	result, err := service.GetInsights("7d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 insights, got %d", len(result))
	}
}

func TestServiceGetInsightsAIError(t *testing.T) {
	stub := &repositoryStub{storageKpis: StorageKpisModel{UsedBytes: 500, FilesTotal: 100}}
	aiMock := &analyticsAIMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{}, errors.New("timeout")
		},
	}

	service := NewService(stub, aiMock)
	result, err := service.GetInsights("7d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty insights on AI error, got %d", len(result))
	}
}

func TestServiceGetInsightsRepoError(t *testing.T) {
	aiMock := &analyticsAIMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{}, nil
		},
	}
	service := NewService(&repositoryStub{err: sql.ErrConnDone}, aiMock)
	if _, err := service.GetInsights("7d"); err == nil {
		t.Fatalf("expected error")
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

func TestResolveHealthStatus(t *testing.T) {
	if resolveHealthStatus(sql.NullString{Valid: false}) != "ok" {
		t.Fatalf("expected ok for invalid")
	}
	if resolveHealthStatus(sql.NullString{String: "PENDING", Valid: true}) != "scanning" {
		t.Fatalf("expected scanning for PENDING")
	}
	if resolveHealthStatus(sql.NullString{String: "FAILED", Valid: true}) != "error" {
		t.Fatalf("expected error for FAILED")
	}
	if resolveHealthStatus(sql.NullString{String: "COMPLETED", Valid: true}) != "ok" {
		t.Fatalf("expected ok for unknown")
	}
}

func TestToTimeSeriesDtoEmpty(t *testing.T) {
	if len(toTimeSeriesDto(nil)) != 0 {
		t.Fatalf("expected empty result")
	}
}

func TestToHotFolderDtoValidity(t *testing.T) {
	now := time.Now()
	models := []FolderHotModel{
		{ParentPath: "/photos", NewFiles: 3, AddedBytes: 300, LastEvent: sql.NullTime{Time: now, Valid: true}},
		{ParentPath: "/docs", NewFiles: 1, AddedBytes: 100, LastEvent: sql.NullTime{Valid: false}},
	}
	result := toHotFolderDto(models)
	if result[0].LastEvent == "" || result[1].LastEvent != "" {
		t.Fatalf("unexpected last event handling: %+v", result)
	}
}

func TestToFolderUsageDtoValidity(t *testing.T) {
	now := time.Now()
	models := []FolderUsageModel{
		{ParentPath: "/media", LastModified: sql.NullTime{Time: now, Valid: true}},
		{ParentPath: "/tmp", LastModified: sql.NullTime{Valid: false}},
	}
	result := toFolderUsageDto(models)
	if result[0].LastModified == "" || result[1].LastModified != "" {
		t.Fatalf("unexpected last modified handling: %+v", result)
	}
}

func TestToHealthDtoVariants(t *testing.T) {
	t.Run("ok when no scan and no errors", func(t *testing.T) {
		result := toHealthDto(HealthModel{Status: sql.NullString{Valid: false}})
		if result.Status != "ok" || result.LastScanAt != "" || result.LastScanSeconds != 0 {
			t.Fatalf("unexpected: %+v", result)
		}
	})

	t.Run("error without description keeps name", func(t *testing.T) {
		result := toHealthDto(HealthModel{
			Status:       sql.NullString{Valid: false},
			RecentErrors: []LogErrorModel{{Name: "Error1", Description: sql.NullString{Valid: false}, CreatedAt: time.Now()}},
		})
		if len(result.RecentErrors) != 1 || result.RecentErrors[0] != "Error1" {
			t.Fatalf("unexpected errors: %v", result.RecentErrors)
		}
	})

	t.Run("error with empty description keeps name", func(t *testing.T) {
		result := toHealthDto(HealthModel{
			Status:       sql.NullString{Valid: false},
			RecentErrors: []LogErrorModel{{Name: "Error2", Description: sql.NullString{String: "", Valid: true}, CreatedAt: time.Now()}},
		})
		if result.RecentErrors[0] != "Error2" {
			t.Fatalf("expected just name, got %s", result.RecentErrors[0])
		}
	})
}

func TestParseInsightsResponse(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		if len(parseInsightsResponse(`["insight 1", "insight 2"]`)) != 2 {
			t.Fatalf("expected 2 insights")
		}
	})
	t.Run("invalid JSON returns empty", func(t *testing.T) {
		if len(parseInsightsResponse("not json")) != 0 {
			t.Fatalf("expected empty insights for invalid JSON")
		}
	})
	t.Run("markdown code fence stripped", func(t *testing.T) {
		if len(parseInsightsResponse("```json\n[\"insight\"]\n```")) != 1 {
			t.Fatalf("expected 1 insight")
		}
	})
}
