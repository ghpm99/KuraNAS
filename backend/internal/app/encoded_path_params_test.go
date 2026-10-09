package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/music"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type catalogKeyRecorderService struct {
	music.ServiceInterface
	receivedKeys map[string]string
}

func (service *catalogKeyRecorderService) record(catalog string, key string) (utils.PaginationResponse[files.FileDto], error) {
	service.receivedKeys[catalog] = key
	return utils.PaginationResponse[files.FileDto]{}, nil
}

func (service *catalogKeyRecorderService) GetLibraryTracksByFolder(folderKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return service.record("folders", folderKey)
}

func (service *catalogKeyRecorderService) GetLibraryTracksByArtist(artistKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return service.record("artists", artistKey)
}

func (service *catalogKeyRecorderService) GetLibraryTracksByAlbum(albumKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return service.record("albums", albumKey)
}

func (service *catalogKeyRecorderService) GetLibraryTracksByGenre(genreKey string, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	return service.record("genres", genreKey)
}

type silentLogService struct {
	logger.LoggerServiceInterface
}

func (silentLogService) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	return log, nil
}

func (silentLogService) CompleteWithSuccessLog(log logger.LoggerModel) error {
	return nil
}

func (silentLogService) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	return nil
}

func TestMusicCatalogKeysSurviveUrlEncoding(t *testing.T) {
	testCases := []struct {
		name        string
		catalog     string
		encodedKey  string
		expectedKey string
	}{
		{"folder path with slashes", "folders", "%2Fdata%2FRock", "/data/Rock"},
		{"artist with slash", "artists", "AC%2FDC", "AC/DC"},
		{"album with literal percent", "albums", "100%25%20Hits", "100% Hits"},
		{"genre with spaces and accents", "genres", "M%C3%BAsica%20Popular", "Música Popular"},
		{"folder with accents and spaces", "folders", "%2Fdata%2FM%C3%BAsica%20Brasileira", "/data/Música Brasileira"},
		{"plain key", "artists", "Beatles", "Beatles"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := &catalogKeyRecorderService{receivedKeys: map[string]string{}}
			context := buildRouteContext()
			context.Music.Handler = music.NewHandler(recorder, nil, nil, silentLogService{})
			router := SetUpRouter()
			RegisterMusicRoutes(router.Group("/api/v1"), context)

			request := httptest.NewRequest(http.MethodGet, "/api/v1/music/library/"+testCase.catalog+"/"+testCase.encodedKey+"/tracks", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
			}
			if recorder.receivedKeys[testCase.catalog] != testCase.expectedKey {
				t.Fatalf("expected key %q, got %q", testCase.expectedKey, recorder.receivedKeys[testCase.catalog])
			}
		})
	}
}

func TestCatchAllPathsKeepSpacesHashAndEncodedSlashes(t *testing.T) {
	router := SetUpRouter()
	router.GET("/files/*filePath", func(c *gin.Context) {
		c.String(http.StatusOK, c.Param("filePath"))
	})

	testCases := []struct {
		name         string
		requestPath  string
		expectedPath string
	}{
		{"spaces", "/files/My%20Music/song.mp3", "/My Music/song.mp3"},
		{"hash", "/files/Top%20%2310/song.mp3", "/Top #10/song.mp3"},
		{"encoded slash", "/files/a%2Fb/c.txt", "/a/b/c.txt"},
		{"literal percent", "/files/100%25/c.txt", "/100%/c.txt"},
		{"accents", "/files/M%C3%BAsica/c.txt", "/Música/c.txt"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, testCase.requestPath, nil))

			if response.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", response.Code)
			}
			if response.Body.String() != testCase.expectedPath {
				t.Fatalf("expected %q, got %q", testCase.expectedPath, response.Body.String())
			}
		})
	}
}

func (service *catalogKeyRecorderService) recordQueue(catalog string, key string) (music.MusicQueueDto, error) {
	service.receivedKeys[catalog] = key
	return music.MusicQueueDto{Items: []music.MusicQueueEntryDto{}}, nil
}

func (service *catalogKeyRecorderService) GetLibraryQueueByFolder(folderKey string) (music.MusicQueueDto, error) {
	return service.recordQueue("folders", folderKey)
}

func (service *catalogKeyRecorderService) GetLibraryQueueByArtist(artistKey string) (music.MusicQueueDto, error) {
	return service.recordQueue("artists", artistKey)
}

func (service *catalogKeyRecorderService) GetLibraryQueueByAlbum(albumKey string) (music.MusicQueueDto, error) {
	return service.recordQueue("albums", albumKey)
}

func (service *catalogKeyRecorderService) GetLibraryQueueByGenre(genreKey string) (music.MusicQueueDto, error) {
	return service.recordQueue("genres", genreKey)
}

func TestMusicQueueRoutesDecodeCatalogKeys(t *testing.T) {
	testCases := []struct {
		catalog     string
		encodedKey  string
		expectedKey string
	}{
		{"folders", "%2Fdata%2FRock", "/data/Rock"},
		{"artists", "AC%2FDC", "AC/DC"},
		{"albums", "100%25%20Hits", "100% Hits"},
		{"genres", "M%C3%BAsica%20Popular", "Música Popular"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.catalog, func(t *testing.T) {
			recorder := &catalogKeyRecorderService{receivedKeys: map[string]string{}}
			context := buildRouteContext()
			context.Music.Handler = music.NewHandler(recorder, nil, nil, silentLogService{})
			router := SetUpRouter()
			RegisterMusicRoutes(router.Group("/api/v1"), context)

			request := httptest.NewRequest(http.MethodGet, "/api/v1/music/library/"+testCase.catalog+"/"+testCase.encodedKey+"/queue", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
			}
			if recorder.receivedKeys[testCase.catalog] != testCase.expectedKey {
				t.Fatalf("expected key %q, got %q", testCase.expectedKey, recorder.receivedKeys[testCase.catalog])
			}
		})
	}
}
