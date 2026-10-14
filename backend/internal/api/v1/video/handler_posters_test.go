package video

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type posterServiceStub struct {
	ServiceInterface
	posterFileID int
	lookupErr    error
}

func (s *posterServiceStub) FindPlaylistPosterFileID(playlistID int) (int, error) {
	return s.posterFileID, s.lookupErr
}

func (s *posterServiceStub) FindMoviePosterFileID(videoID int) (int, error) {
	return s.posterFileID, s.lookupErr
}

type posterFilesStub struct {
	files.ServiceInterface
	thumbnail    []byte
	thumbnailErr error
	fileErr      error
	receivedSize [2]int
}

func (s *posterFilesStub) GetFileById(id int) (files.FileDto, error) {
	return files.FileDto{ID: id, Name: "poster.jpg", Format: ".jpg"}, s.fileErr
}

func (s *posterFilesStub) GetFileThumbnail(fileDto files.FileDto, width, height int) ([]byte, error) {
	s.receivedSize = [2]int{width, height}
	return s.thumbnail, s.thumbnailErr
}

func performPosterRequest(service *posterServiceStub, filesService *posterFilesStub, path string, header http.Header) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service, filesService, nil, nil)
	router := gin.New()
	router.GET("/video/playlists/:id/poster", handler.GetPlaylistPosterHandler)
	router.GET("/video/movies/:file_id/poster", handler.GetMoviePosterHandler)
	request := httptest.NewRequest(http.MethodGet, path, nil)
	for name, values := range header {
		request.Header[name] = values
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

var jpegSignature = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 0x10, 'J', 'F', 'I', 'F', 0}

func TestPosterHandlersServeThumbnailWithCacheHeaders(t *testing.T) {
	for _, path := range []string{"/video/playlists/3/poster", "/video/movies/9/poster"} {
		filesService := &posterFilesStub{thumbnail: jpegSignature}
		recorder := performPosterRequest(&posterServiceStub{posterFileID: 12}, filesService, path, nil)

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, want 200", path, recorder.Code)
		}
		if recorder.Header().Get("Content-Type") != "image/jpeg" {
			t.Errorf("%s: content type = %q", path, recorder.Header().Get("Content-Type"))
		}
		if recorder.Header().Get("Cache-Control") == "" || recorder.Header().Get("ETag") == "" {
			t.Errorf("%s: missing cache headers: %v", path, recorder.Header())
		}
		if filesService.receivedSize != [2]int{400, 600} {
			t.Errorf("%s: default size = %v, want 400x600", path, filesService.receivedSize)
		}
	}
}

func TestPosterHandlerHonorsRequestedSize(t *testing.T) {
	filesService := &posterFilesStub{thumbnail: jpegSignature}
	performPosterRequest(&posterServiceStub{posterFileID: 12}, filesService, "/video/movies/9/poster?width=200&height=300", nil)
	if filesService.receivedSize != [2]int{200, 300} {
		t.Fatalf("size = %v, want 200x300", filesService.receivedSize)
	}
}

func TestPosterHandlerAnswersNotModifiedForMatchingETag(t *testing.T) {
	first := performPosterRequest(&posterServiceStub{posterFileID: 12}, &posterFilesStub{thumbnail: jpegSignature}, "/video/movies/9/poster", nil)
	etag := first.Header().Get("ETag")

	filesService := &posterFilesStub{thumbnail: jpegSignature}
	second := performPosterRequest(&posterServiceStub{posterFileID: 12}, filesService, "/video/movies/9/poster", http.Header{"If-None-Match": {etag}})
	if second.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", second.Code)
	}
	if filesService.receivedSize != [2]int{} {
		t.Fatal("thumbnail must not be rendered on 304")
	}
}

func TestPosterHandlerReturnsTranslatedNotFoundWithoutPoster(t *testing.T) {
	recorder := performPosterRequest(&posterServiceStub{lookupErr: ErrPosterNotFound}, &posterFilesStub{}, "/video/playlists/3/poster", nil)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("body: %v", err)
	}
	if body["error"] != i18n.GetMessage("ERROR_VIDEO_POSTER_NOT_FOUND") {
		t.Fatalf("error = %q", body["error"])
	}
}

func TestPosterHandlerReturnsNotFoundForUnknownPlaylistAndMissingPosterFile(t *testing.T) {
	unknown := performPosterRequest(&posterServiceStub{lookupErr: sql.ErrNoRows}, &posterFilesStub{}, "/video/playlists/3/poster", nil)
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("unknown playlist: status = %d, want 404", unknown.Code)
	}

	missingOnDisk := performPosterRequest(&posterServiceStub{posterFileID: 12}, &posterFilesStub{thumbnailErr: files.ErrFileMissingDisk}, "/video/movies/9/poster", nil)
	if missingOnDisk.Code != http.StatusNotFound || missingOnDisk.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("missing on disk: status = %d cache = %q", missingOnDisk.Code, missingOnDisk.Header().Get("Cache-Control"))
	}

	staleIndex := performPosterRequest(&posterServiceStub{posterFileID: 12}, &posterFilesStub{fileErr: errors.New("gone")}, "/video/movies/9/poster", nil)
	if staleIndex.Code != http.StatusNotFound {
		t.Fatalf("stale index: status = %d, want 404", staleIndex.Code)
	}
}

func TestPosterHandlerRejectsInvalidIDAndInternalFailures(t *testing.T) {
	if recorder := performPosterRequest(&posterServiceStub{}, &posterFilesStub{}, "/video/playlists/abc/poster", nil); recorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid id: status = %d, want 400", recorder.Code)
	}
	if recorder := performPosterRequest(&posterServiceStub{posterFileID: 1}, &posterFilesStub{}, "/video/movies/1/poster?width=x", nil); recorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid size: status = %d, want 400", recorder.Code)
	}
	if recorder := performPosterRequest(&posterServiceStub{lookupErr: errors.New("db down")}, &posterFilesStub{}, "/video/movies/1/poster", nil); recorder.Code != http.StatusInternalServerError {
		t.Fatalf("lookup failure: status = %d, want 500", recorder.Code)
	}
	if recorder := performPosterRequest(&posterServiceStub{posterFileID: 1}, &posterFilesStub{thumbnailErr: errors.New("decode")}, "/video/movies/1/poster", nil); recorder.Code != http.StatusInternalServerError {
		t.Fatalf("thumbnail failure: status = %d, want 500", recorder.Code)
	}
}

func TestPosterContentTypeFallsBackToPngForNonJpeg(t *testing.T) {
	if posterContentType([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}) != "image/png" {
		t.Fatal("png expected")
	}
}
