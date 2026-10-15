package files

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/i18n"
)

func getStarredCounts(service *filesHandlerServiceFuncMock) *httptest.ResponseRecorder {
	router := newFilesHandlerRouter(NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{}))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/files/starred/counts", nil))
	return response
}

func TestGetStarredCountsHandlerReturnsCounts(t *testing.T) {
	service := &filesHandlerServiceFuncMock{
		countStarredFn: func() (StarredCountsDto, error) {
			return StarredCountsDto{Total: 6, Folders: 1, Files: 2, Media: 3}, nil
		},
	}

	response := getStarredCounts(service)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d %s", response.Code, response.Body.String())
	}
	var decoded map[string]int
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	expected := map[string]int{"total": 6, "folders": 1, "files": 2, "media": 3}
	if len(decoded) != len(expected) {
		t.Fatalf("unexpected body %v", decoded)
	}
	for key, expectedCount := range expected {
		if decoded[key] != expectedCount {
			t.Fatalf("key %s: expected %d, got %v", key, expectedCount, decoded)
		}
	}
}

func TestGetStarredCountsHandlerReturnsInternalErrorWhenServiceFails(t *testing.T) {
	service := &filesHandlerServiceFuncMock{
		countStarredFn: func() (StarredCountsDto, error) {
			return StarredCountsDto{}, errors.New("db down")
		},
	}

	response := getStarredCounts(service)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d", response.Code)
	}
	var decoded map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded["error"] != i18n.GetMessage("ERROR_INTERNAL") {
		t.Fatalf("unexpected body %v", decoded)
	}
}
