package files

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"encoding/json"

	"nas-go/api/pkg/i18n"
)

func postStarredBatch(service *filesHandlerServiceFuncMock, body string) *httptest.ResponseRecorder {
	router := newFilesHandlerRouter(NewHandler(service, &filesRecentServiceMock{}, &filesLoggerMock{}))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/files/starred", strings.NewReader(body)))
	return response
}

func decodeStarredBatchResponse(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	return decoded
}

func failingBatchService(t *testing.T) *filesHandlerServiceFuncMock {
	return &filesHandlerServiceFuncMock{
		setFilesStarredFn: func(ids []int, isStarred bool) ([]int, error) {
			t.Fatalf("SetFilesStarred must not be called")
			return nil, nil
		},
	}
}

func TestSetFilesStarredHandlerReportsUpdatedAndNotFoundIDs(t *testing.T) {
	var receivedIDs []int
	var receivedState bool
	service := &filesHandlerServiceFuncMock{
		setFilesStarredFn: func(ids []int, isStarred bool) ([]int, error) {
			receivedIDs, receivedState = ids, isStarred
			return []int{4, 2}, nil
		},
	}

	response := postStarredBatch(service, `{"ids":[4,9,2,7],"starred":true}`)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d %s", response.Code, response.Body.String())
	}
	if !reflect.DeepEqual(receivedIDs, []int{4, 9, 2, 7}) || !receivedState {
		t.Fatalf("service received ids=%v starred=%v", receivedIDs, receivedState)
	}
	decoded := decodeStarredBatchResponse(t, response)
	expected := map[string]any{
		"starred":       true,
		"updated_ids":   []any{float64(4), float64(2)},
		"not_found_ids": []any{float64(9), float64(7)},
	}
	if !reflect.DeepEqual(decoded, expected) {
		t.Fatalf("unexpected response body %s", response.Body.String())
	}
}

func TestSetFilesStarredHandlerAcceptsStarredFalse(t *testing.T) {
	var receivedState = true
	service := &filesHandlerServiceFuncMock{
		setFilesStarredFn: func(ids []int, isStarred bool) ([]int, error) {
			receivedState = isStarred
			return ids, nil
		},
	}

	response := postStarredBatch(service, `{"ids":[1],"starred":false}`)

	decoded := decodeStarredBatchResponse(t, response)
	if response.Code != http.StatusOK || receivedState || decoded["starred"] != false {
		t.Fatalf("unexpected response %d %s (service state=%v)", response.Code, response.Body.String(), receivedState)
	}
	if !reflect.DeepEqual(decoded["not_found_ids"], []any{}) {
		t.Fatalf("not_found_ids must be an empty array: %s", response.Body.String())
	}
}

func TestSetFilesStarredHandlerDeduplicatesIDs(t *testing.T) {
	var receivedIDs []int
	service := &filesHandlerServiceFuncMock{
		setFilesStarredFn: func(ids []int, isStarred bool) ([]int, error) {
			receivedIDs = ids
			return ids, nil
		},
	}

	response := postStarredBatch(service, `{"ids":[3,3,5,3,5],"starred":true}`)

	if response.Code != http.StatusOK || !reflect.DeepEqual(receivedIDs, []int{3, 5}) {
		t.Fatalf("unexpected response %d, service received %v", response.Code, receivedIDs)
	}
}

func TestSetFilesStarredHandlerRejectsInvalidBodies(t *testing.T) {
	tooManyIDs := make([]string, MaxBatchStarredIDs+1)
	for index := range tooManyIDs {
		tooManyIDs[index] = "1"
	}
	bodies := map[string]string{
		"empty list":      `{"ids":[],"starred":true}`,
		"missing ids":     `{"starred":true}`,
		"above limit":     `{"ids":[` + strings.Join(tooManyIDs, ",") + `],"starred":true}`,
		"missing starred": `{"ids":[1]}`,
		"not json":        `not json`,
		"empty body":      ``,
	}
	for name, body := range bodies {
		response := postStarredBatch(failingBatchService(t), body)

		expectedBody := `{"error":"` + i18n.GetMessage("ERROR_INVALID_REQUEST") + `"}`
		if response.Code != http.StatusBadRequest || strings.TrimSpace(response.Body.String()) != expectedBody {
			t.Fatalf("%s: unexpected response %d %s", name, response.Code, response.Body.String())
		}
	}
}

func TestSetFilesStarredHandlerAcceptsExactlyTheLimit(t *testing.T) {
	ids := make([]string, MaxBatchStarredIDs)
	for index := range ids {
		ids[index] = strconv.Itoa(index + 1)
	}
	service := &filesHandlerServiceFuncMock{
		setFilesStarredFn: func(ids []int, isStarred bool) ([]int, error) { return ids, nil },
	}

	response := postStarredBatch(service, `{"ids":[`+strings.Join(ids, ",")+`],"starred":true}`)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", response.Code)
	}
}

func TestSetFilesStarredHandlerReturnsInternalErrorOnServiceFailure(t *testing.T) {
	service := &filesHandlerServiceFuncMock{
		setFilesStarredFn: func(ids []int, isStarred bool) ([]int, error) {
			return nil, errors.New("database unavailable")
		},
	}

	response := postStarredBatch(service, `{"ids":[1],"starred":true}`)

	expectedBody := `{"error":"` + i18n.GetMessage("ERROR_INTERNAL") + `"}`
	if response.Code != http.StatusInternalServerError || strings.TrimSpace(response.Body.String()) != expectedBody {
		t.Fatalf("unexpected response %d %s", response.Code, response.Body.String())
	}
}
