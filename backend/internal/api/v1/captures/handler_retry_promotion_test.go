package captures

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type retryPromotionServiceMock struct {
	capturesHandlerServiceMock
	retryErr error
}

func (m *retryPromotionServiceMock) RetryCapturePromotion(captureID int) (int, error) {
	if m.retryErr != nil {
		return 0, m.retryErr
	}
	return 7, nil
}

func performRetryPromotionRequest(retryErr error, idSegment string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&retryPromotionServiceMock{retryErr: retryErr}, &capturesLoggerMock{})
	router := gin.New()
	router.POST("/captures/:id/promote", handler.RetryCapturePromotionHandler)
	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/captures/%s/promote", idSegment), nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestRetryCapturePromotionHandlerAccepted(t *testing.T) {
	recorder := performRetryPromotionRequest(nil, "3")
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected %d, got %d, body=%s", http.StatusAccepted, recorder.Code, recorder.Body.String())
	}
	var body map[string]int
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid body: %v", err)
	}
	if body["job_id"] != 7 {
		t.Fatalf("expected job_id 7, got body=%s", recorder.Body.String())
	}
}

func TestRetryCapturePromotionHandlerErrorStatuses(t *testing.T) {
	testCases := []struct {
		name           string
		retryErr       error
		idSegment      string
		expectedStatus int
	}{
		{"not retryable", fmt.Errorf("wrapped: %w", ErrCaptureNotRetryable), "3", http.StatusConflict},
		{"not found", ErrCaptureNotFound, "3", http.StatusNotFound},
		{"invalid id", nil, "abc", http.StatusBadRequest},
		{"unexpected error", errors.New("boom secret"), "3", http.StatusInternalServerError},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := performRetryPromotionRequest(testCase.retryErr, testCase.idSegment)
			if recorder.Code != testCase.expectedStatus {
				t.Fatalf("expected %d, got %d", testCase.expectedStatus, recorder.Code)
			}
			var body map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid body: %v", err)
			}
			if body["error"] == "" || body["error"] == "boom secret" {
				t.Fatalf("unexpected error message %q", body["error"])
			}
		})
	}
}
