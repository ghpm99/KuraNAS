package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
)

type populatedService struct {
	ServiceInterface
}

func (populatedService) GetStorage(string) (StorageStatsDto, error) {
	return StorageStatsDto{
		Storage: StorageDto{TotalBytes: 1000, UsedBytes: 400, FreeBytes: 600, GrowthBytes: 50},
		Counts:  CountsDto{FilesTotal: 10, FilesAdded: 2, Folders: 3},
	}, nil
}

func (populatedService) GetTimeSeries(string) ([]TimeSeriesPointDto, error) {
	return []TimeSeriesPointDto{{Date: "2026-11-25", UsedBytes: 400}}, nil
}

func (populatedService) GetTypes() ([]TypeBreakdownDto, error) {
	return []TypeBreakdownDto{{Type: "image", Count: 5, Bytes: 300}}, nil
}

func (populatedService) GetExtensions(int) ([]ExtensionDto, error) {
	return []ExtensionDto{{Ext: ".jpg", Count: 5, Bytes: 300}}, nil
}

func (populatedService) GetRecentFiles(int) ([]RecentFileDto, error) {
	return []RecentFileDto{{
		ID: 7, Name: "a.jpg", Path: "/p/a.jpg", ParentPath: "/p", Format: ".jpg",
		SizeBytes: 12, CreatedAt: "2026-11-25T10:00:00Z", UpdatedAt: "2026-11-25T11:00:00Z",
	}}, nil
}

func (populatedService) GetTopFolders(int) ([]FolderUsageDto, error) {
	return []FolderUsageDto{{Path: "/p", Files: 4, Bytes: 90, LastModified: "2026-11-25T10:00:00Z"}}, nil
}

func (populatedService) GetHotFolders(string, int) ([]HotFolderDto, error) {
	return []HotFolderDto{{Path: "/p", NewFiles: 2, AddedBytes: 80, LastEvent: "2026-11-25T10:00:00Z"}}, nil
}

func (populatedService) GetDuplicatesSummary() (DuplicatesSummaryDto, error) {
	return DuplicatesSummaryDto{Groups: 1, Files: 2, ReclaimableSize: 30}, nil
}

func (populatedService) GetDuplicateGroups(int) ([]DuplicateGroupDto, error) {
	return []DuplicateGroupDto{{Signature: "abc", Copies: 2, SizeBytes: 30, ReclaimableSize: 30, Paths: []string{"/a", "/b"}}}, nil
}

func (populatedService) GetLibrary() (LibraryDto, error) {
	return LibraryDto{CategorizedMedia: 1, AudioWithMetadata: 2, VideoWithMetadata: 3, ImageWithMetadata: 4, ImageClassified: 5}, nil
}

func (populatedService) GetProcessing() (ProcessingDto, error) {
	return ProcessingDto{MetadataPending: 1, MetadataFailed: 2, ThumbnailPending: 3, ThumbnailFailed: 4, RecurringTimeouts: 5}, nil
}

func (populatedService) GetHealth() (HealthDto, error) {
	return HealthDto{
		Status: "ok", LastScanAt: "2026-11-25T10:00:00Z", LastScanSeconds: 9,
		IndexedFiles: 100, ErrorsLast24h: 1, RecentErrors: []string{"boom"},
	}, nil
}

func (populatedService) GetAIUsage() (AIUsageDto, error) {
	return AIUsageDto{Total: 3, Success: 2, Failure: 1, TotalTokens: 400, AvgLatencyMs: 120}, nil
}

func requestPopulatedJSON(t *testing.T, route string, url string, registerRoute func(*gin.Engine, *Handler)) any {
	t.Helper()
	router := gin.New()
	registerRoute(router, NewHandler(populatedService{}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, url, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("%s: status %d", route, recorder.Code)
	}

	var decodedBody any
	if err := json.Unmarshal(recorder.Body.Bytes(), &decodedBody); err != nil {
		t.Fatalf("%s: invalid json: %v", route, err)
	}
	return decodedBody
}

func sortedKeys(object any) []string {
	objectAsMap, isMap := object.(map[string]any)
	if !isMap {
		return nil
	}
	keys := make([]string, 0, len(objectAsMap))
	for key := range objectAsMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func assertKeys(t *testing.T, label string, expectedKeys []string, object any) {
	t.Helper()
	if actualKeys := sortedKeys(object); !reflect.DeepEqual(expectedKeys, actualKeys) {
		t.Errorf("%s: chaves esperadas %v, recebidas %v", label, expectedKeys, actualKeys)
	}
}

func firstElement(t *testing.T, list any) any {
	t.Helper()
	elements, isList := list.([]any)
	if !isList || len(elements) == 0 {
		t.Fatalf("resposta deveria ser uma lista não vazia, veio %v", list)
	}
	return elements[0]
}

func TestAnalyticsJSONContractObjects(t *testing.T) {
	testCases := []struct {
		name         string
		url          string
		register     func(*gin.Engine, *Handler)
		expectedKeys []string
		nestedKeys   map[string][]string
	}{
		{
			name: "storage",
			url:  "/analytics/storage",
			register: func(router *gin.Engine, handler *Handler) {
				router.GET("/analytics/storage", handler.GetStorageHandler)
			},
			expectedKeys: []string{"counts", "storage"},
			nestedKeys: map[string][]string{
				"storage": {"free_bytes", "growth_bytes", "total_bytes", "used_bytes"},
				"counts":  {"files_added", "files_total", "folders"},
			},
		},
		{
			name: "duplicates",
			url:  "/analytics/duplicates",
			register: func(router *gin.Engine, handler *Handler) {
				router.GET("/analytics/duplicates", handler.GetDuplicatesHandler)
			},
			expectedKeys: []string{"files", "groups", "reclaimable_size"},
		},
		{
			name: "library",
			url:  "/analytics/library",
			register: func(router *gin.Engine, handler *Handler) {
				router.GET("/analytics/library", handler.GetLibraryHandler)
			},
			expectedKeys: []string{"audio_with_metadata", "categorized_media", "image_classified", "image_with_metadata", "video_with_metadata"},
		},
		{
			name: "processing",
			url:  "/analytics/processing",
			register: func(router *gin.Engine, handler *Handler) {
				router.GET("/analytics/processing", handler.GetProcessingHandler)
			},
			expectedKeys: []string{"metadata_failed", "metadata_pending", "recurring_timeouts", "thumbnail_failed", "thumbnail_pending"},
		},
		{
			name:         "health",
			url:          "/analytics/health",
			register:     func(router *gin.Engine, handler *Handler) { router.GET("/analytics/health", handler.GetHealthHandler) },
			expectedKeys: []string{"errors_last_24h", "indexed_files", "last_scan_at", "last_scan_seconds", "recent_errors", "status"},
		},
		{
			name: "ai-usage",
			url:  "/analytics/ai-usage",
			register: func(router *gin.Engine, handler *Handler) {
				router.GET("/analytics/ai-usage", handler.GetAIUsageHandler)
			},
			expectedKeys: []string{"avg_latency_ms", "failure", "success", "total", "total_tokens"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body := requestPopulatedJSON(t, testCase.name, testCase.url, testCase.register)
			assertKeys(t, testCase.name, testCase.expectedKeys, body)
			for nestedName, expectedNested := range testCase.nestedKeys {
				assertKeys(t, nestedName, expectedNested, body.(map[string]any)[nestedName])
			}
		})
	}
}

func TestAnalyticsJSONContractLists(t *testing.T) {
	testCases := []struct {
		name         string
		url          string
		register     func(*gin.Engine, *Handler)
		expectedKeys []string
	}{
		{
			name: "timeseries",
			url:  "/analytics/timeseries",
			register: func(router *gin.Engine, handler *Handler) {
				router.GET("/analytics/timeseries", handler.GetTimeSeriesHandler)
			},
			expectedKeys: []string{"date", "used_bytes"},
		},
		{
			name:         "types",
			url:          "/analytics/types",
			register:     func(router *gin.Engine, handler *Handler) { router.GET("/analytics/types", handler.GetTypesHandler) },
			expectedKeys: []string{"bytes", "count", "type"},
		},
		{
			name: "extensions",
			url:  "/analytics/extensions",
			register: func(router *gin.Engine, handler *Handler) {
				router.GET("/analytics/extensions", handler.GetExtensionsHandler)
			},
			expectedKeys: []string{"bytes", "count", "ext"},
		},
		{
			name: "recent-files",
			url:  "/analytics/recent-files",
			register: func(router *gin.Engine, handler *Handler) {
				router.GET("/analytics/recent-files", handler.GetRecentFilesHandler)
			},
			expectedKeys: []string{"created_at", "format", "id", "name", "parent_path", "path", "size_bytes", "updated_at"},
		},
		{
			name: "top-folders",
			url:  "/analytics/top-folders",
			register: func(router *gin.Engine, handler *Handler) {
				router.GET("/analytics/top-folders", handler.GetTopFoldersHandler)
			},
			expectedKeys: []string{"bytes", "files", "last_modified", "path"},
		},
		{
			name: "hot-folders",
			url:  "/analytics/hot-folders",
			register: func(router *gin.Engine, handler *Handler) {
				router.GET("/analytics/hot-folders", handler.GetHotFoldersHandler)
			},
			expectedKeys: []string{"added_bytes", "last_event", "new_files", "path"},
		},
		{
			name: "duplicates-groups",
			url:  "/analytics/duplicates/groups",
			register: func(router *gin.Engine, handler *Handler) {
				router.GET("/analytics/duplicates/groups", handler.GetDuplicateGroupsHandler)
			},
			expectedKeys: []string{"copies", "paths", "reclaimable_size", "signature", "size_bytes"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body := requestPopulatedJSON(t, testCase.name, testCase.url, testCase.register)
			assertKeys(t, testCase.name, testCase.expectedKeys, firstElement(t, body))
		})
	}
}
