package captures

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func fullyPopulatedCaptureDto() CaptureDto {
	fileID, season, episode, releaseYear := 9, 2, 3, 2021
	return CaptureDto{
		ID:            1,
		Name:          "name",
		FileName:      "file.mp4",
		FilePath:      "/captures/file.mp4",
		MediaType:     "video",
		MimeType:      "video/mp4",
		Size:          10,
		EpisodeKey:    "show:s02e03",
		CreatedAt:     time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		FileID:        &fileID,
		Status:        "failed",
		FailureReason: "reason",
		Title:         "title",
		EpisodeTitle:  "episode title",
		Season:        &season,
		Episode:       &episode,
		Description:   "description",
		ReleaseYear:   &releaseYear,
		Genres:        []string{"drama"},
		Cast:          []string{"actor"},
		Directors:     []string{"director"},
		Studio:        "studio",
		ContentRating: "PG",
		Platform:      "platform",
		SourceURL:     "https://example.test/watch",
		ThumbnailURL:  "https://example.test/thumb.jpg",
		ContentType:   "episode",
	}
}

type capturesResponseContractServiceMock struct {
	capturesHandlerServiceMock
}

func (m *capturesResponseContractServiceMock) InitCaptureUpload(dto InitCaptureUploadDto) (InitCaptureUploadResultDto, error) {
	return InitCaptureUploadResultDto{
		UploadID:        "upload-1",
		ChunkSize:       1024,
		ReceivedSize:    512,
		Resumed:         true,
		AlreadyComplete: true,
	}, nil
}

func (m *capturesResponseContractServiceMock) CompleteCaptureUpload(dto CompleteCaptureUploadDto) (CaptureDto, error) {
	return fullyPopulatedCaptureDto(), nil
}

func (m *capturesResponseContractServiceMock) GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureDto], error) {
	return utils.PaginationResponse[CaptureDto]{
		Items:      []CaptureDto{fullyPopulatedCaptureDto()},
		Pagination: utils.Pagination{Page: 2, PageSize: 5, HasNext: true, HasPrev: true},
	}, nil
}

func (m *capturesResponseContractServiceMock) GetCaptureByID(id int) (CaptureDto, error) {
	return fullyPopulatedCaptureDto(), nil
}

func serveCapturesContractRequest(t *testing.T, request *http.Request) map[string]any {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&capturesResponseContractServiceMock{}, &capturesLoggerMock{})
	router := gin.New()
	router.POST("/captures/upload/init", handler.InitCaptureUploadHandler)
	router.POST("/captures/upload/complete", handler.CompleteCaptureUploadHandler)
	router.GET("/captures", handler.GetCapturesHandler)
	router.GET("/captures/:id", handler.GetCaptureByIDHandler)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	var decodedBody map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &decodedBody); err != nil {
		t.Fatalf("decode body %q: %v", recorder.Body.String(), err)
	}
	return decodedBody
}

func assertExactKeys(t *testing.T, decodedObject map[string]any, expectedKeys []string) {
	t.Helper()
	actualKeys := make([]string, 0, len(decodedObject))
	for key := range decodedObject {
		actualKeys = append(actualKeys, key)
	}
	sort.Strings(actualKeys)
	sort.Strings(expectedKeys)
	if len(actualKeys) != len(expectedKeys) {
		t.Fatalf("expected keys %v, got %v", expectedKeys, actualKeys)
	}
	for index := range expectedKeys {
		if actualKeys[index] != expectedKeys[index] {
			t.Fatalf("expected keys %v, got %v", expectedKeys, actualKeys)
		}
	}
}

var captureResponseKeys = []string{
	"id", "name", "file_name", "file_path", "media_type", "mime_type", "size",
	"episode_key", "created_at", "file_id", "status", "failure_reason", "title",
	"episode_title", "season", "episode", "description", "release_year", "genres",
	"cast", "directors", "studio", "content_rating", "platform", "source_url",
	"thumbnail_url", "content_type",
}

func TestCapturesListResponsePinsEnvelopeAndItemFieldNames(t *testing.T) {
	body := serveCapturesContractRequest(t, httptest.NewRequest(http.MethodGet, "/captures?page=2&page_size=5", nil))

	assertExactKeys(t, body, []string{"items", "pagination"})

	paginationBody, isObject := body["pagination"].(map[string]any)
	if !isObject {
		t.Fatalf("pagination is not an object: %v", body["pagination"])
	}
	assertExactKeys(t, paginationBody, []string{"page", "page_size", "has_next", "has_prev"})
	if paginationBody["page"] != float64(2) || paginationBody["page_size"] != float64(5) ||
		paginationBody["has_next"] != true || paginationBody["has_prev"] != true {
		t.Fatalf("unexpected pagination values: %v", paginationBody)
	}

	items, isList := body["items"].([]any)
	if !isList || len(items) != 1 {
		t.Fatalf("items is not a single-element list: %v", body["items"])
	}
	firstItem, isObject := items[0].(map[string]any)
	if !isObject {
		t.Fatalf("item is not an object: %v", items[0])
	}
	assertExactKeys(t, firstItem, captureResponseKeys)
}

func TestCapturesDetailResponsePinsEveryFieldNameAndValue(t *testing.T) {
	body := serveCapturesContractRequest(t, httptest.NewRequest(http.MethodGet, "/captures/1", nil))

	assertExactKeys(t, body, captureResponseKeys)
	expectedValues := map[string]any{
		"id":             float64(1),
		"name":           "name",
		"file_name":      "file.mp4",
		"file_path":      "/captures/file.mp4",
		"media_type":     "video",
		"mime_type":      "video/mp4",
		"size":           float64(10),
		"episode_key":    "show:s02e03",
		"created_at":     "2026-01-02T03:04:05Z",
		"file_id":        float64(9),
		"status":         "failed",
		"failure_reason": "reason",
		"title":          "title",
		"episode_title":  "episode title",
		"season":         float64(2),
		"episode":        float64(3),
		"description":    "description",
		"release_year":   float64(2021),
		"studio":         "studio",
		"content_rating": "PG",
		"platform":       "platform",
		"source_url":     "https://example.test/watch",
		"thumbnail_url":  "https://example.test/thumb.jpg",
		"content_type":   "episode",
	}
	for fieldName, expectedValue := range expectedValues {
		if body[fieldName] != expectedValue {
			t.Fatalf("field %q: expected %v, got %v", fieldName, expectedValue, body[fieldName])
		}
	}
	for _, listFieldName := range []string{"genres", "cast", "directors"} {
		values, isList := body[listFieldName].([]any)
		if !isList || len(values) != 1 {
			t.Fatalf("field %q is not a single-element list: %v", listFieldName, body[listFieldName])
		}
	}
}

func TestCapturesUploadInitResponsePinsFieldNames(t *testing.T) {
	request := buildJSONRequest(t, http.MethodPost, "/captures/upload/init", map[string]any{"name": "capture"})
	body := serveCapturesContractRequest(t, request)

	assertExactKeys(t, body, []string{"upload_id", "chunk_size", "received_size", "resumed", "already_complete"})
	if body["upload_id"] != "upload-1" || body["chunk_size"] != float64(1024) ||
		body["received_size"] != float64(512) || body["resumed"] != true || body["already_complete"] != true {
		t.Fatalf("unexpected init values: %v", body)
	}
}

func TestCapturesUploadCompleteResponsePinsCaptureFieldNames(t *testing.T) {
	request := buildJSONRequest(t, http.MethodPost, "/captures/upload/complete", map[string]any{"upload_id": "upload-1"})
	body := serveCapturesContractRequest(t, request)

	assertExactKeys(t, body, captureResponseKeys)
	if body["id"] != float64(1) || body["status"] != "failed" || body["file_name"] != "file.mp4" {
		t.Fatalf("unexpected complete values: %v", body)
	}
}
