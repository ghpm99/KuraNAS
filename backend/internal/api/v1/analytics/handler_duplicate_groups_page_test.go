package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func registerDuplicateGroups(router *gin.Engine, handler *Handler) {
	router.GET("/x", handler.GetDuplicateGroupsHandler)
}

func TestHandlerDuplicateGroupsWithoutPageKeepsFlatArray(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &serviceStub{groupsToReturn: []DuplicateGroupDto{
		{Signature: "abc", Copies: 3, SizeBytes: 100, ReclaimableSize: 200, Paths: []string{"/a", "/b", "/c"}},
	}}
	router := gin.New()
	registerDuplicateGroups(router, NewHandler(stub))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/x?limit=7", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var groups []map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &groups); err != nil {
		t.Fatalf("response without page must be a flat array: %v body=%s", err, recorder.Body.String())
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	expectedKeys := []string{"signature", "copies", "size_bytes", "reclaimable_size", "paths"}
	if len(groups[0]) != len(expectedKeys) {
		t.Fatalf("expected exactly keys %v, got %v", expectedKeys, groups[0])
	}
	for _, key := range expectedKeys {
		if _, exists := groups[0][key]; !exists {
			t.Fatalf("missing key %q in %v", key, groups[0])
		}
	}
	if stub.calledGroupsPage || stub.capturedLimit != 7 {
		t.Fatalf("expected legacy path with limit 7, got page=%v limit=%d", stub.calledGroupsPage, stub.capturedLimit)
	}
}

func TestHandlerDuplicateGroupsWithPageReturnsPaginatedEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &serviceStub{pageToReturn: utils.PaginationResponse[DuplicateGroupDetailDto]{
		Items: []DuplicateGroupDetailDto{{
			DuplicateGroupDto: DuplicateGroupDto{Signature: "abc", Copies: 2, SizeBytes: 100, ReclaimableSize: 100, Paths: []string{"/a", "/b"}},
			Files: []DuplicateCopyDto{
				{ID: 1, Path: "/a", UpdatedAt: "2026-01-02T03:04:05Z", Tier: "hot"},
				{ID: 2, Path: "/b", UpdatedAt: "2026-01-02T03:04:06Z", Tier: "cold"},
			},
		}},
		Pagination: utils.Pagination{Page: 2, PageSize: 10, HasNext: true, HasPrev: true},
	}}
	router := gin.New()
	registerDuplicateGroups(router, NewHandler(stub))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/x?page=2&page_size=10&type=video", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	var envelope utils.PaginationResponse[DuplicateGroupDetailDto]
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Pagination != (utils.Pagination{Page: 2, PageSize: 10, HasNext: true, HasPrev: true}) {
		t.Fatalf("unexpected pagination: %+v", envelope.Pagination)
	}
	if len(envelope.Items) != 1 || len(envelope.Items[0].Files) != 2 {
		t.Fatalf("unexpected items: %+v", envelope.Items)
	}
	firstCopy := envelope.Items[0].Files[0]
	if firstCopy.ID != 1 || firstCopy.Path != "/a" || firstCopy.Tier != "hot" || firstCopy.UpdatedAt != "2026-01-02T03:04:05Z" {
		t.Fatalf("unexpected copy: %+v", firstCopy)
	}
	if stub.capturedCategory != DuplicateCategoryVideo || stub.capturedPage != 2 || stub.capturedPageSize != 10 {
		t.Fatalf("unexpected forwarded args: %q %d %d", stub.capturedCategory, stub.capturedPage, stub.capturedPageSize)
	}
	if stub.calledImageGroups {
		t.Fatal("paginated request must not use the legacy image path")
	}
}

func TestHandlerDuplicateGroupsPageUsesDefaultPageSizeAndAllCategories(t *testing.T) {
	stub, recorder := doRequest(t, registerDuplicateGroups, "/x?page=1")
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if stub.capturedCategory != DuplicateCategoryAll || stub.capturedPage != 1 || stub.capturedPageSize != defaultDuplicateGroupsPageSize {
		t.Fatalf("unexpected defaults: %q %d %d", stub.capturedCategory, stub.capturedPage, stub.capturedPageSize)
	}
}

func TestHandlerDuplicateGroupsPageAcceptsEveryDocumentedType(t *testing.T) {
	for _, category := range []DuplicateCategory{
		DuplicateCategoryImage, DuplicateCategoryAudio, DuplicateCategoryVideo, DuplicateCategoryDocument, DuplicateCategoryOther,
	} {
		stub, recorder := doRequest(t, registerDuplicateGroups, "/x?page=1&type="+string(category))
		if recorder.Code != http.StatusOK || stub.capturedCategory != category {
			t.Fatalf("type %q: code=%d captured=%q", category, recorder.Code, stub.capturedCategory)
		}
	}
}

func TestHandlerDuplicateGroupsPageRejectsInvalidInput(t *testing.T) {
	for _, url := range []string{
		"/x?page=abc",
		"/x?page=1&page_size=abc",
		"/x?page=1&type=archive",
	} {
		stub, recorder := doRequest(t, registerDuplicateGroups, url)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", url, recorder.Code)
		}
		if stub.calledGroupsPage {
			t.Fatalf("%s: service must not be called", url)
		}
		var body map[string]string
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["error"] == "" {
			t.Fatalf("%s: expected error body, got %s", url, recorder.Body.String())
		}
	}
}

func TestHandlerDuplicateGroupsPageServiceFailureReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerDuplicateGroups(router, NewHandler(&serviceStub{err: http.ErrAbortHandler}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/x?page=1", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
