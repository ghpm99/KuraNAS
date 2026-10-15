package files

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nas-go/api/pkg/i18n"
)

type setStarredResponse struct {
	ID      int  `json:"id"`
	Starred bool `json:"starred"`
}

func putStarred(service *filesHandlerServiceFuncMock, path string, body string) *httptest.ResponseRecorder {
	router := newFilesHandlerRouter(NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{}))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPut, path, strings.NewReader(body)))
	return response
}

func TestSetFileStarredHandlerDecodesBodyAndRespondsWithFinalState(t *testing.T) {
	for _, requestedState := range []bool{true, false} {
		var receivedID int
		var receivedState bool
		service := &filesHandlerServiceFuncMock{
			setFileStarredFn: func(id int, isStarred bool) (bool, error) {
				receivedID, receivedState = id, isStarred
				return isStarred, nil
			},
		}

		response := putStarred(service, "/files/starred/7", fmt.Sprintf(`{"starred":%t}`, requestedState))

		if response.Code != http.StatusOK {
			t.Fatalf("unexpected status %d %s", response.Code, response.Body.String())
		}
		if receivedID != 7 || receivedState != requestedState {
			t.Fatalf("service received id=%d starred=%v", receivedID, receivedState)
		}
		var decoded map[string]any
		if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if len(decoded) != 2 || decoded["id"] != float64(7) || decoded["starred"] != requestedState {
			t.Fatalf("unexpected response body %s", response.Body.String())
		}
	}
}

func TestSetFileStarredHandlerRejectsBodyWithoutStarredField(t *testing.T) {
	service := &filesHandlerServiceFuncMock{
		setFileStarredFn: func(id int, isStarred bool) (bool, error) {
			t.Fatalf("SetFileStarred must not be called")
			return false, nil
		},
	}
	for _, body := range []string{`{}`, ``, `not json`, `{"starred":"yes"}`} {
		response := putStarred(service, "/files/starred/7", body)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("body %q: unexpected status %d", body, response.Code)
		}
	}
}

func TestSetFileStarredHandlerRejectsInvalidID(t *testing.T) {
	response := putStarred(&filesHandlerServiceFuncMock{}, "/files/starred/abc", `{"starred":true}`)

	expectedBody := `{"error":"` + i18n.GetMessage("ERROR_INVALID_REQUEST") + `"}`
	if response.Code != http.StatusBadRequest || strings.TrimSpace(response.Body.String()) != expectedBody {
		t.Fatalf("unexpected response %d %s", response.Code, response.Body.String())
	}
}

func TestSetFileStarredHandlerReturnsNotFoundWhenFileIsMissingOrTrashed(t *testing.T) {
	service := &filesHandlerServiceFuncMock{
		setFileStarredFn: func(id int, isStarred bool) (bool, error) {
			return false, fmt.Errorf("set starred: %w", sql.ErrNoRows)
		},
	}

	response := putStarred(service, "/files/starred/7", `{"starred":true}`)

	expectedBody := `{"error":"` + i18n.GetMessage("ERROR_FILE_NOT_FOUND") + `"}`
	if response.Code != http.StatusNotFound || strings.TrimSpace(response.Body.String()) != expectedBody {
		t.Fatalf("unexpected response %d %s", response.Code, response.Body.String())
	}
}

func TestSetFileStarredHandlerReturnsInternalErrorForGenericFailure(t *testing.T) {
	service := &filesHandlerServiceFuncMock{
		setFileStarredFn: func(id int, isStarred bool) (bool, error) { return false, errors.New("database unavailable") },
	}

	response := putStarred(service, "/files/starred/7", `{"starred":true}`)

	expectedBody := `{"error":"` + i18n.GetMessage("ERROR_INTERNAL") + `"}`
	if response.Code != http.StatusInternalServerError || strings.TrimSpace(response.Body.String()) != expectedBody {
		t.Fatalf("unexpected response %d %s", response.Code, response.Body.String())
	}
}
