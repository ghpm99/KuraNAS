package diary

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type diaryHandlerServiceMock struct{}

func (m *diaryHandlerServiceMock) CreateDiary(diaryDto DiaryDto) (DiaryDto, error) {
	diaryDto.ID = 1
	return diaryDto, nil
}
func (m *diaryHandlerServiceMock) GetDiary(filter DiaryFilter, page int, pageSize int) (utils.PaginationResponse[DiaryDto], error) {
	return utils.PaginationResponse[DiaryDto]{Items: []DiaryDto{{ID: 1, Name: "d"}}}, nil
}
func (m *diaryHandlerServiceMock) UpdateDiary(diaryDto DiaryDto) (result bool, err error) {
	return true, nil
}
func (m *diaryHandlerServiceMock) GetSummary() (DiarySummary, error) {
	return DiarySummary{Date: time.Now(), TotalActivities: 1}, nil
}
func (m *diaryHandlerServiceMock) DuplicateDiary(id int) (DiaryDto, error) {
	return DiaryDto{ID: id, Name: "copy"}, nil
}

type diaryLoggerMock struct{ logger.LoggerServiceInterface }

func (m *diaryLoggerMock) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	return logger.LoggerModel{}, nil
}
func (m *diaryLoggerMock) CompleteWithSuccessLog(log logger.LoggerModel) error { return nil }
func (m *diaryLoggerMock) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	return nil
}

func TestDiaryHandlerEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&diaryHandlerServiceMock{}, &diaryLoggerMock{})
	router := gin.New()

	router.POST("/diary", handler.CreateDiaryHandler)
	router.POST("/diary/copy", handler.DuplicateDiaryHandler)
	router.GET("/diary", handler.GetDiaryHandler)
	router.PUT("/diary/:id", handler.UpdateDiaryHandler)
	router.GET("/diary/summary", handler.GetSummaryHandler)

	tests := []struct {
		method string
		path   string
		body   string
		ct     string
		code   int
	}{
		{http.MethodPost, "/diary", `{"name":"x"}`, "application/json", http.StatusOK},
		{http.MethodPost, "/diary/copy", `{"id":1}`, "application/json", http.StatusOK},
		{http.MethodGet, "/diary?page=1&page_size=10", "", "", http.StatusOK},
		{http.MethodPut, "/diary/1", "data=hello", "application/x-www-form-urlencoded", http.StatusOK},
		{http.MethodGet, "/diary/summary", "", "", http.StatusOK},
		{http.MethodPost, "/diary", `{}`, "application/json", http.StatusBadRequest},
		{http.MethodPost, "/diary/copy", `{}`, "application/json", http.StatusOK},
		{http.MethodPut, "/diary/1", "", "", http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.ct != "" {
				req.Header.Set("Content-Type", tc.ct)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d, body=%s", tc.code, w.Code, w.Body.String())
			}
		})
	}
}
