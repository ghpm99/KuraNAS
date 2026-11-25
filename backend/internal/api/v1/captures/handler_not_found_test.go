package captures

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type captureLookupFailureServiceMock struct {
	capturesHandlerErrServiceMock
	lookupErr error
}

func (m *captureLookupFailureServiceMock) GetCaptureByID(id int) (CaptureDto, error) {
	return CaptureDto{}, m.lookupErr
}

func (m *captureLookupFailureServiceMock) DeleteCapture(id int) (bool, error) {
	return false, m.lookupErr
}

func TestCaptureByIDHandlersStatusByLookupError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testCases := []struct {
		name           string
		method         string
		lookupErr      error
		expectedStatus int
	}{
		{"get missing capture", http.MethodGet, ErrCaptureNotFound, http.StatusNotFound},
		{"get database failure", http.MethodGet, errors.New("db down"), http.StatusInternalServerError},
		{"delete missing capture", http.MethodDelete, ErrCaptureNotFound, http.StatusNotFound},
		{"delete database failure", http.MethodDelete, errors.New("db down"), http.StatusInternalServerError},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := NewHandler(&captureLookupFailureServiceMock{lookupErr: testCase.lookupErr}, &capturesLoggerMock{})
			router := gin.New()
			router.GET("/captures/:id", handler.GetCaptureByIDHandler)
			router.DELETE("/captures/:id", handler.DeleteCaptureHandler)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(testCase.method, "/captures/7", nil))

			if recorder.Code != testCase.expectedStatus {
				t.Fatalf("expected %d, got %d", testCase.expectedStatus, recorder.Code)
			}
		})
	}
}
