package analytics

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nas-go/api/internal/api/v1/trash"

	"github.com/gin-gonic/gin"
)

type trashSummaryReaderFake struct {
	summary trash.TrashSummaryDto
	err     error
}

func (fake trashSummaryReaderFake) GetSummary() (trash.TrashSummaryDto, error) {
	return fake.summary, fake.err
}

func TestServiceGetTrashMapsSummary(t *testing.T) {
	reader := trashSummaryReaderFake{summary: trash.TrashSummaryDto{Items: 4, SizeBytes: 9000}}
	service := NewService(&repositoryStub{}, nil, reader)

	trashDto, err := service.GetTrash()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trashDto.ItemCount != 4 || trashDto.SizeBytes != 9000 {
		t.Fatalf("unexpected trash dto: %+v", trashDto)
	}
}

func TestServiceGetTrashPropagatesReaderError(t *testing.T) {
	service := NewService(&repositoryStub{}, nil, trashSummaryReaderFake{err: sql.ErrConnDone})

	if _, err := service.GetTrash(); err == nil {
		t.Fatalf("expected reader error")
	}
}

func TestHandlerGetTrashReturnsExactKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analytics/trash", NewHandler(&serviceStub{}).GetTrashHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/analytics/trash", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var body map[string]int64
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid body: %v", err)
	}
	if len(body) != 2 || body["item_count"] != 3 || body["size_bytes"] != 2048 {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestHandlerGetTrashErrorDoesNotLeakRawError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analytics/trash", NewHandler(&serviceStub{err: sql.ErrConnDone}).GetTrashHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/analytics/trash", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), sql.ErrConnDone.Error()) {
		t.Fatalf("raw error leaked: %s", recorder.Body.String())
	}
}
