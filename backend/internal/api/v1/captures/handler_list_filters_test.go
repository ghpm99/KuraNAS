package captures

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type capturesFilterRecordingServiceMock struct {
	capturesHandlerServiceMock
	receivedFilter CaptureFilter
}

func (m *capturesFilterRecordingServiceMock) GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureDto], error) {
	m.receivedFilter = filter
	return utils.PaginationResponse[CaptureDto]{
		Items:      []CaptureDto{{ID: 3, Name: "failed-one", Status: "failed"}},
		Pagination: utils.Pagination{Page: page, PageSize: pageSize, HasNext: true},
	}, nil
}

func TestGetCapturesHandlerAppliesStatusAndPlatformFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &capturesFilterRecordingServiceMock{}
	router := gin.New()
	router.GET("/captures", NewHandler(service, &capturesLoggerMock{}).GetCapturesHandler)

	request := httptest.NewRequest(http.MethodGet, "/captures?status=failed&platform=netflix", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, recorder.Code)
	}
	var body struct {
		Items      []CaptureDto `json:"items"`
		Pagination struct {
			HasNext bool `json:"has_next"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].Status != "failed" || !body.Pagination.HasNext {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
	if service.receivedFilter.Status.Value != "failed" || service.receivedFilter.Platform.Value != "netflix" {
		t.Fatalf("filter not forwarded: %+v", service.receivedFilter)
	}
}

func TestGetCapturesHandlerRejectsInvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/captures", NewHandler(&capturesFilterRecordingServiceMock{}, &capturesLoggerMock{}).GetCapturesHandler)

	request := httptest.NewRequest(http.MethodGet, "/captures?status=bogus", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}
