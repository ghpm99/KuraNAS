package video

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type hiddenPlaylistsServiceStub struct {
	ServiceInterface
	receivedRequest HiddenPlaylistsRequest
	page            utils.PaginationResponse[VideoPlaylistDto]
}

func (s *hiddenPlaylistsServiceStub) GetHiddenPlaylists(request HiddenPlaylistsRequest) (utils.PaginationResponse[VideoPlaylistDto], error) {
	s.receivedRequest = request
	return s.page, nil
}

func performHiddenPlaylistsRequest(stub *hiddenPlaylistsServiceStub, path string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/video/playlists/hidden", NewHandler(stub, nil, nil, nil).GetHiddenPlaylistsHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

func TestGetHiddenPlaylistsHandlerPassesPaginationAndReturnsPage(t *testing.T) {
	stub := &hiddenPlaylistsServiceStub{page: utils.PaginationResponse[VideoPlaylistDto]{
		Items:      []VideoPlaylistDto{{ID: 5, Name: "Oculta", IsHidden: true}},
		Pagination: utils.Pagination{Page: 2, PageSize: 3, HasPrev: true},
	}}
	recorder := performHiddenPlaylistsRequest(stub, "/video/playlists/hidden?page=2&page_size=3")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if stub.receivedRequest.Page != 2 || stub.receivedRequest.PageSize != 3 {
		t.Fatalf("request = %+v", stub.receivedRequest)
	}
	var responseBody utils.PaginationResponse[VideoPlaylistDto]
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("unexpected body %s (%v)", recorder.Body.String(), err)
	}
	if len(responseBody.Items) != 1 || responseBody.Items[0].ID != 5 || !responseBody.Items[0].IsHidden || !responseBody.Pagination.HasPrev {
		t.Fatalf("unexpected page %+v", responseBody)
	}
}

func TestGetHiddenPlaylistsHandlerRejectsInvalidPagination(t *testing.T) {
	recorder := performHiddenPlaylistsRequest(&hiddenPlaylistsServiceStub{}, "/video/playlists/hidden?page=abc")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
}
