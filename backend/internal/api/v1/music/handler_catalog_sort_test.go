package music

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type sortRecordingServiceMock struct {
	musicHandlerServiceMock
	receivedSort CatalogSort
}

func (m *sortRecordingServiceMock) GetLibraryAlbums(page int, pageSize int, sort CatalogSort) (utils.PaginationResponse[MusicAlbumGroupDto], error) {
	m.receivedSort = sort
	return utils.PaginationResponse[MusicAlbumGroupDto]{}, nil
}

func (m *sortRecordingServiceMock) GetLibraryArtists(page int, pageSize int, sort CatalogSort) (utils.PaginationResponse[MusicArtistGroupDto], error) {
	m.receivedSort = sort
	return utils.PaginationResponse[MusicArtistGroupDto]{}, nil
}

func (m *sortRecordingServiceMock) GetHomeCatalog(clientID string, limit int, sort CatalogSort) (MusicHomeCatalogDto, error) {
	m.receivedSort = sort
	return MusicHomeCatalogDto{}, nil
}

func TestCatalogListHandlersParseSortAndOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		path     string
		wantCode int
		wantSort CatalogSort
	}{
		{"/music/library/artists", http.StatusOK, CatalogSort{Field: CatalogSortByTracks, IsDescending: true}},
		{"/music/library/artists?sort=name", http.StatusOK, CatalogSort{Field: CatalogSortByName}},
		{"/music/library/artists?sort=recent&order=asc", http.StatusOK, CatalogSort{Field: CatalogSortByRecent}},
		{"/music/library/albums?sort=year", http.StatusOK, CatalogSort{Field: CatalogSortByYear, IsDescending: true}},
		{"/music/library/home?sort=recent", http.StatusOK, CatalogSort{Field: CatalogSortByRecent, IsDescending: true}},
		{"/music/library/artists?sort=year", http.StatusBadRequest, CatalogSort{}},
		{"/music/library/artists?sort=bogus", http.StatusBadRequest, CatalogSort{}},
		{"/music/library/albums?order=sideways", http.StatusBadRequest, CatalogSort{}},
		{"/music/library/home?sort=year", http.StatusBadRequest, CatalogSort{}},
	}

	for _, testCase := range cases {
		t.Run(testCase.path, func(t *testing.T) {
			serviceMock := &sortRecordingServiceMock{}
			handler := NewHandler(serviceMock, nil, &musicRecentServiceMock{}, &musicLoggerMock{})
			router := gin.New()
			router.GET("/music/library/artists", handler.GetLibraryArtistsHandler)
			router.GET("/music/library/albums", handler.GetLibraryAlbumsHandler)
			router.GET("/music/library/home", handler.GetHomeCatalogHandler)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, testCase.path, nil))

			if recorder.Code != testCase.wantCode {
				t.Fatalf("expected %d, got %d body=%s", testCase.wantCode, recorder.Code, recorder.Body.String())
			}
			if testCase.wantCode == http.StatusOK && serviceMock.receivedSort != testCase.wantSort {
				t.Fatalf("service received %+v, want %+v", serviceMock.receivedSort, testCase.wantSort)
			}
		})
	}
}
