package captures

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type capturePlatformsServiceMock struct {
	capturesHandlerServiceMock
	platforms []string
	err       error
}

func (m *capturePlatformsServiceMock) GetCapturePlatforms() ([]string, error) {
	return m.platforms, m.err
}

func serveCapturePlatforms(t *testing.T, service *capturePlatformsServiceMock) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, &capturesLoggerMock{})
	router := gin.New()
	router.GET("/captures/platforms", handler.GetCapturePlatformsHandler)
	router.GET("/captures/:id", handler.GetCaptureByIDHandler)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/captures/platforms", nil))
	return recorder
}

func decodeRawBody(t *testing.T, recorder *httptest.ResponseRecorder) map[string]json.RawMessage {
	t.Helper()
	var body map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body %q: %v", recorder.Body.String(), err)
	}
	return body
}

func TestGetCapturePlatformsHandlerReturnsOnlyItemsKey(t *testing.T) {
	recorder := serveCapturePlatforms(t, &capturePlatformsServiceMock{platforms: []string{"crunchyroll", "youtube"}})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	body := decodeRawBody(t, recorder)
	if len(body) != 1 {
		t.Fatalf("expected only the items key, got %v", body)
	}
	var items []string
	if err := json.Unmarshal(body["items"], &items); err != nil {
		t.Fatalf("decode items: %v", err)
	}
	if !reflect.DeepEqual(items, []string{"crunchyroll", "youtube"}) {
		t.Fatalf("unexpected items %v", items)
	}
}

func TestGetCapturePlatformsHandlerEmptyListIsArrayNotNull(t *testing.T) {
	for name, platforms := range map[string][]string{"nil": nil, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			recorder := serveCapturePlatforms(t, &capturePlatformsServiceMock{platforms: platforms})
			if got := string(decodeRawBody(t, recorder)["items"]); got != "[]" {
				t.Fatalf("expected [], got %s", got)
			}
		})
	}
}

func TestGetCapturePlatformsHandlerServiceErrorReturnsTranslatedMessage(t *testing.T) {
	recorder := serveCapturePlatforms(t, &capturePlatformsServiceMock{err: errors.New("pq: connection refused")})

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
	var errorMessage string
	if err := json.Unmarshal(decodeRawBody(t, recorder)["error"], &errorMessage); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if errorMessage != i18n.GetMessage("CAPTURES_PLATFORMS_LOAD_ERROR") || errorMessage == "pq: connection refused" {
		t.Fatalf("unexpected error message %q", errorMessage)
	}
}
