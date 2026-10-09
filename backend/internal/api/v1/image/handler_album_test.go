package image

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type fakeAlbumService struct {
	err          error
	albumPage    utils.PaginationResponse[AlbumDto]
	album        AlbumDto
	itemsChange  AlbumItemsChangeDto
	itemsPage    LibraryPageDto
	callCount    int
	listRequest  AlbumListRequest
	createdName  string
	update       AlbumUpdate
	readAlbumID  int
	deletedID    int
	changedID    int
	changedFiles []int
	changeKind   string
	itemsRequest AlbumItemsRequest
}

func (f *fakeAlbumService) ListAlbums(request AlbumListRequest) (utils.PaginationResponse[AlbumDto], error) {
	f.callCount++
	f.listRequest = request
	return f.albumPage, f.err
}

func (f *fakeAlbumService) GetAlbum(albumID int) (AlbumDto, error) {
	f.callCount++
	f.readAlbumID = albumID
	return f.album, f.err
}

func (f *fakeAlbumService) CreateAlbum(name string) (AlbumDto, error) {
	f.callCount++
	f.createdName = name
	return f.album, f.err
}

func (f *fakeAlbumService) UpdateAlbum(update AlbumUpdate) (AlbumDto, error) {
	f.callCount++
	f.update = update
	return f.album, f.err
}

func (f *fakeAlbumService) DeleteAlbum(albumID int) error {
	f.callCount++
	f.deletedID = albumID
	return f.err
}

func (f *fakeAlbumService) AddAlbumItems(albumID int, fileIDs []int) (AlbumItemsChangeDto, error) {
	f.callCount++
	f.changeKind, f.changedID, f.changedFiles = "add", albumID, fileIDs
	return f.itemsChange, f.err
}

func (f *fakeAlbumService) RemoveAlbumItems(albumID int, fileIDs []int) (AlbumItemsChangeDto, error) {
	f.callCount++
	f.changeKind, f.changedID, f.changedFiles = "remove", albumID, fileIDs
	return f.itemsChange, f.err
}

func (f *fakeAlbumService) ListAlbumItems(request AlbumItemsRequest) (LibraryPageDto, error) {
	f.callCount++
	f.itemsRequest = request
	return f.itemsPage, f.err
}

func serveAlbumRequest(service *fakeAlbumService, method string, requestURL string, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewAlbumHandler(service, &imageLoggerMock{})
	router := gin.New()
	router.GET("/image/albums", handler.ListAlbumsHandler)
	router.POST("/image/albums", handler.CreateAlbumHandler)
	router.GET("/image/albums/:id", handler.GetAlbumHandler)
	router.PUT("/image/albums/:id", handler.UpdateAlbumHandler)
	router.DELETE("/image/albums/:id", handler.DeleteAlbumHandler)
	router.GET("/image/albums/:id/items", handler.ListAlbumItemsHandler)
	router.POST("/image/albums/:id/items", handler.AddAlbumItemsHandler)
	router.DELETE("/image/albums/:id/items", handler.RemoveAlbumItemsHandler)

	request := httptest.NewRequest(method, requestURL, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestListAlbumsHandlerDecodesPagination(t *testing.T) {
	service := &fakeAlbumService{albumPage: utils.PaginationResponse[AlbumDto]{Items: []AlbumDto{{ID: 1, Name: "Viagem"}}}}
	recorder := serveAlbumRequest(service, http.MethodGet, "/image/albums?page=2&page_size=5", "")

	if recorder.Code != http.StatusOK || service.listRequest.Page != 2 || service.listRequest.PageSize != 5 {
		t.Fatalf("status %d request %+v", recorder.Code, service.listRequest)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body["items"] == nil || body["pagination"] == nil {
		t.Fatalf("unexpected body %s", recorder.Body.String())
	}

	defaults := &fakeAlbumService{}
	serveAlbumRequest(defaults, http.MethodGet, "/image/albums", "")
	if defaults.listRequest.PageSize != defaultAlbumPageSize || defaults.listRequest.Page != 1 {
		t.Fatalf("defaults = %+v", defaults.listRequest)
	}

	invalid := &fakeAlbumService{}
	if recorder := serveAlbumRequest(invalid, http.MethodGet, "/image/albums?page=abc", ""); recorder.Code != http.StatusBadRequest || invalid.callCount != 0 {
		t.Fatalf("invalid pagination status %d calls %d", recorder.Code, invalid.callCount)
	}
}

func TestCreateAlbumHandlerDecodesName(t *testing.T) {
	service := &fakeAlbumService{album: AlbumDto{ID: 3, Name: "Viagem"}}
	recorder := serveAlbumRequest(service, http.MethodPost, "/image/albums", `{"name":"Viagem"}`)

	if recorder.Code != http.StatusCreated || service.createdName != "Viagem" {
		t.Fatalf("status %d name %q", recorder.Code, service.createdName)
	}
	var created AlbumDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil || created.ID != 3 {
		t.Fatalf("created = %+v err %v", created, err)
	}
}

func TestCreateAlbumHandlerRejectsMalformedBody(t *testing.T) {
	service := &fakeAlbumService{}
	recorder := serveAlbumRequest(service, http.MethodPost, "/image/albums", `{"name":`)
	if recorder.Code != http.StatusBadRequest || service.callCount != 0 {
		t.Fatalf("status %d calls %d", recorder.Code, service.callCount)
	}
}

func TestUpdateAlbumHandlerDecodesNameAndCover(t *testing.T) {
	service := &fakeAlbumService{}
	recorder := serveAlbumRequest(service, http.MethodPut, "/image/albums/4", `{"name":"Férias","cover_file_id":12}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
	}
	update := service.update
	if update.AlbumID != 4 || update.Name == nil || *update.Name != "Férias" || update.CoverFileID == nil || *update.CoverFileID != 12 {
		t.Fatalf("update = %+v", update)
	}
}

func TestUpdateAlbumHandlerLeavesAbsentFieldsNil(t *testing.T) {
	service := &fakeAlbumService{}
	serveAlbumRequest(service, http.MethodPut, "/image/albums/4", `{"cover_file_id":12}`)
	if service.update.Name != nil || service.update.CoverFileID == nil {
		t.Fatalf("update = %+v", service.update)
	}
	serveAlbumRequest(service, http.MethodPut, "/image/albums/4", `{"name":"Só nome"}`)
	if service.update.CoverFileID != nil || service.update.Name == nil {
		t.Fatalf("update = %+v", service.update)
	}
}

func TestDeleteAlbumHandlerDecodesAlbumID(t *testing.T) {
	service := &fakeAlbumService{}
	recorder := serveAlbumRequest(service, http.MethodDelete, "/image/albums/9", "")
	if recorder.Code != http.StatusNoContent || service.deletedID != 9 {
		t.Fatalf("status %d deleted %d", recorder.Code, service.deletedID)
	}
}

func TestAlbumItemsMutationHandlersDecodeFileIDs(t *testing.T) {
	cases := map[string]struct {
		method       string
		expectedKind string
	}{
		"add":    {http.MethodPost, "add"},
		"remove": {http.MethodDelete, "remove"},
	}
	for name, testCase := range cases {
		service := &fakeAlbumService{itemsChange: AlbumItemsChangeDto{Requested: 3, Changed: 2}}
		recorder := serveAlbumRequest(service, testCase.method, "/image/albums/7/items", `{"file_ids":[3,4,5]}`)

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s: status %d body %s", name, recorder.Code, recorder.Body.String())
		}
		if service.changeKind != testCase.expectedKind || service.changedID != 7 || fmt.Sprint(service.changedFiles) != "[3 4 5]" {
			t.Fatalf("%s: kind %q album %d files %v", name, service.changeKind, service.changedID, service.changedFiles)
		}
		var change AlbumItemsChangeDto
		if err := json.Unmarshal(recorder.Body.Bytes(), &change); err != nil || change.Requested != 3 || change.Changed != 2 {
			t.Fatalf("%s: change = %+v err %v", name, change, err)
		}
	}
}

func TestAlbumItemsMutationHandlersRejectMalformedBody(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodDelete} {
		service := &fakeAlbumService{}
		recorder := serveAlbumRequest(service, method, "/image/albums/7/items", `{"file_ids":"x"}`)
		if recorder.Code != http.StatusBadRequest || service.callCount != 0 {
			t.Fatalf("%s: status %d calls %d", method, recorder.Code, service.callCount)
		}
	}
}

func TestListAlbumItemsHandlerReusesLibraryListingParameters(t *testing.T) {
	service := &fakeAlbumService{itemsPage: LibraryPageDto{Items: []LibraryItemDto{{FileID: 1}}, PageSize: 25}}
	cursor := LibraryCursor{FileID: 11}.Encode()
	recorder := serveAlbumRequest(service, http.MethodGet, "/image/albums/5/items?page_size=25&starred=true&cursor="+cursor, "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d body %s", recorder.Code, recorder.Body.String())
	}
	request := service.itemsRequest
	if request.AlbumID != 5 || request.Listing.PageSize != 25 || !request.Listing.Filter.OnlyStarred || request.Listing.Cursor == nil || request.Listing.Cursor.FileID != 11 {
		t.Fatalf("request = %+v", request)
	}
	if request.Listing.Sort != LibrarySortTakenAt || request.Listing.Order != LibrarySortOrderDesc {
		t.Fatalf("default ordering = %+v", request.Listing)
	}
}

func TestListAlbumItemsHandlerRejectsInvalidInput(t *testing.T) {
	for name, requestURL := range map[string]string{
		"cursor": "/image/albums/5/items?cursor=%21%21",
		"page":   "/image/albums/5/items?page=abc",
		"id":     "/image/albums/abc/items",
	} {
		service := &fakeAlbumService{}
		recorder := serveAlbumRequest(service, http.MethodGet, requestURL, "")
		if recorder.Code != http.StatusBadRequest || service.callCount != 0 {
			t.Fatalf("%s: status %d calls %d", name, recorder.Code, service.callCount)
		}
	}
}

func TestAlbumHandlersRejectInvalidAlbumID(t *testing.T) {
	for _, invalidID := range []string{"abc", "0", "-3"} {
		for method, body := range map[string]string{http.MethodPut: `{"name":"x"}`, http.MethodDelete: ""} {
			service := &fakeAlbumService{}
			recorder := serveAlbumRequest(service, method, "/image/albums/"+invalidID, body)
			if recorder.Code != http.StatusBadRequest || service.callCount != 0 {
				t.Fatalf("%s %s: status %d calls %d", method, invalidID, recorder.Code, service.callCount)
			}
		}
		service := &fakeAlbumService{}
		recorder := serveAlbumRequest(service, http.MethodPost, "/image/albums/"+invalidID+"/items", `{"file_ids":[1]}`)
		if recorder.Code != http.StatusBadRequest || service.callCount != 0 {
			t.Fatalf("items %s: status %d calls %d", invalidID, recorder.Code, service.callCount)
		}
	}
}

func TestAlbumHandlersMapServiceErrorsToTranslatedStatuses(t *testing.T) {
	cases := map[string]struct {
		err            error
		expectedStatus int
	}{
		"not found":        {fmt.Errorf("wrapped: %w", ErrAlbumNotFound), http.StatusNotFound},
		"name taken":       {ErrAlbumNameTaken, http.StatusConflict},
		"name required":    {ErrAlbumNameRequired, http.StatusBadRequest},
		"name too long":    {ErrAlbumNameTooLong, http.StatusBadRequest},
		"invalid file ids": {ErrAlbumInvalidFileIDs, http.StatusBadRequest},
		"cover not in":     {ErrAlbumCoverNotInAlbum, http.StatusBadRequest},
		"nothing":          {ErrAlbumNothingToUpdate, http.StatusBadRequest},
		"library cursor":   {ErrInvalidLibraryCursor, http.StatusBadRequest},
		"internal":         {errors.New("db down"), http.StatusInternalServerError},
	}
	for name, testCase := range cases {
		recorder := serveAlbumRequest(&fakeAlbumService{err: testCase.err}, http.MethodPost, "/image/albums", `{"name":"x"}`)
		if recorder.Code != testCase.expectedStatus {
			t.Fatalf("%s: status = %d", name, recorder.Code)
		}
		var body map[string]string
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: body %s", name, recorder.Body.String())
		}
		if _, hasError := body["error"]; !hasError {
			t.Fatalf("%s: body without error %s", name, recorder.Body.String())
		}
	}
}

func TestAlbumReadHandlersMapServiceErrors(t *testing.T) {
	if recorder := serveAlbumRequest(&fakeAlbumService{err: errors.New("db down")}, http.MethodGet, "/image/albums", ""); recorder.Code != http.StatusInternalServerError {
		t.Fatalf("list status = %d", recorder.Code)
	}
	if recorder := serveAlbumRequest(&fakeAlbumService{err: ErrAlbumNotFound}, http.MethodGet, "/image/albums/3/items", ""); recorder.Code != http.StatusNotFound {
		t.Fatalf("items status = %d", recorder.Code)
	}
	if recorder := serveAlbumRequest(&fakeAlbumService{err: ErrAlbumNotFound}, http.MethodDelete, "/image/albums/3", ""); recorder.Code != http.StatusNotFound {
		t.Fatalf("delete status = %d", recorder.Code)
	}
	if recorder := serveAlbumRequest(&fakeAlbumService{err: ErrAlbumNotFound}, http.MethodPut, "/image/albums/3", `{"name":"x"}`); recorder.Code != http.StatusNotFound {
		t.Fatalf("update status = %d", recorder.Code)
	}
	if recorder := serveAlbumRequest(&fakeAlbumService{err: ErrAlbumNotFound}, http.MethodDelete, "/image/albums/3/items", `{"file_ids":[1]}`); recorder.Code != http.StatusNotFound {
		t.Fatalf("remove status = %d", recorder.Code)
	}
}

func TestGetAlbumHandlerReturnsTheAlbum(t *testing.T) {
	service := &fakeAlbumService{album: AlbumDto{ID: 6, Name: "Viagem", ItemCount: 4}}
	recorder := serveAlbumRequest(service, http.MethodGet, "/image/albums/6", "")

	var album AlbumDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &album); err != nil || recorder.Code != http.StatusOK || album.ItemCount != 4 || service.readAlbumID != 6 {
		t.Fatalf("status %d album %+v err %v", recorder.Code, album, err)
	}
	if recorder := serveAlbumRequest(&fakeAlbumService{err: ErrAlbumNotFound}, http.MethodGet, "/image/albums/6", ""); recorder.Code != http.StatusNotFound {
		t.Fatalf("unknown album status = %d", recorder.Code)
	}
	if recorder := serveAlbumRequest(&fakeAlbumService{}, http.MethodGet, "/image/albums/abc", ""); recorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid id status = %d", recorder.Code)
	}
}
