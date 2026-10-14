package video

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type subtitleFileSourceFake struct {
	filesByID map[int]files.FileDto
	err       error
}

func (fake *subtitleFileSourceFake) GetFileById(id int) (files.FileDto, error) {
	if fake.err != nil {
		return files.FileDto{}, fake.err
	}
	file, exists := fake.filesByID[id]
	if !exists {
		return files.FileDto{}, sql.ErrNoRows
	}
	return file, nil
}

type subtitleServiceFake struct {
	subtitles []SubtitleDto
	err       error
	parent    string
	name      string
}

func (fake *subtitleServiceFake) ListVideoSubtitles(parentPath string, videoName string) ([]SubtitleDto, error) {
	fake.parent, fake.name = parentPath, videoName
	return fake.subtitles, fake.err
}

func performSubtitleRequest(handler *SubtitleHandler, target string, headers map[string]string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/video/subtitles/:id", handler.ListVideoSubtitlesHandler)
	router.GET("/video/subtitles/:id/vtt", handler.GetSubtitleWebVttHandler)
	request := httptest.NewRequest(http.MethodGet, target, nil)
	for headerName, headerValue := range headers {
		request.Header.Set(headerName, headerValue)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func writeSubtitleOnDisk(t *testing.T, fileName string, content []byte) files.FileDto {
	t.Helper()
	subtitlePath := filepath.Join(t.TempDir(), fileName)
	if err := os.WriteFile(subtitlePath, content, 0o644); err != nil {
		t.Fatal(err)
	}
	return files.FileDto{ID: 5, Name: fileName, Path: subtitlePath, Size: int64(len(content)), UpdatedAt: time.Unix(1700000000, 0)}
}

func TestListVideoSubtitlesHandlerReturnsSubtitles(t *testing.T) {
	service := &subtitleServiceFake{subtitles: []SubtitleDto{{ID: 9, Source: "external", Label: "en.forced", Language: "en", Format: "srt", IsForced: true}}}
	source := &subtitleFileSourceFake{filesByID: map[int]files.FileDto{1: {ID: 1, Name: "Filme.mkv", ParentPath: "/data"}}}

	recorder := performSubtitleRequest(NewSubtitleHandler(source, service), "/video/subtitles/1", nil)

	expectedBody := `[{"id":9,"source":"external","label":"en.forced","language":"en","format":"srt","is_forced":true}]`
	if recorder.Code != http.StatusOK || recorder.Body.String() != expectedBody {
		t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
	}
	if service.parent != "/data" || service.name != "Filme.mkv" {
		t.Fatalf("service received %q %q", service.parent, service.name)
	}
}

func TestListVideoSubtitlesHandlerReturnsEmptyListWhenNoSubtitles(t *testing.T) {
	service := &subtitleServiceFake{subtitles: []SubtitleDto{}}
	source := &subtitleFileSourceFake{filesByID: map[int]files.FileDto{1: {ID: 1, Name: "Filme.mkv"}}}

	recorder := performSubtitleRequest(NewSubtitleHandler(source, service), "/video/subtitles/1", nil)

	if recorder.Code != http.StatusOK || strings.TrimSpace(recorder.Body.String()) != "[]" {
		t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestListVideoSubtitlesHandlerErrors(t *testing.T) {
	deletedVideo := files.FileDto{ID: 2, Name: "Gone.mkv", DeletedAt: utils.Optional[time.Time]{Value: time.Now(), HasValue: true}}
	testCases := []struct {
		name       string
		source     *subtitleFileSourceFake
		service    *subtitleServiceFake
		target     string
		statusCode int
		messageKey string
	}{
		{"invalid id", &subtitleFileSourceFake{}, &subtitleServiceFake{}, "/video/subtitles/abc", http.StatusBadRequest, "ERROR_INVALID_REQUEST"},
		{"unknown video", &subtitleFileSourceFake{}, &subtitleServiceFake{}, "/video/subtitles/99", http.StatusNotFound, "ERROR_VIDEO_NOT_FOUND"},
		{"deleted video", &subtitleFileSourceFake{filesByID: map[int]files.FileDto{2: deletedVideo}}, &subtitleServiceFake{}, "/video/subtitles/2", http.StatusNotFound, "ERROR_VIDEO_NOT_FOUND"},
		{"source failure", &subtitleFileSourceFake{err: errors.New("db down")}, &subtitleServiceFake{}, "/video/subtitles/1", http.StatusInternalServerError, "ERROR_VIDEO_OPERATION_FAILED"},
		{"service failure", &subtitleFileSourceFake{filesByID: map[int]files.FileDto{1: {ID: 1, Name: "Filme.mkv"}}}, &subtitleServiceFake{err: errors.New("db down")}, "/video/subtitles/1", http.StatusInternalServerError, "ERROR_VIDEO_OPERATION_FAILED"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := performSubtitleRequest(NewSubtitleHandler(testCase.source, testCase.service), testCase.target, nil)

			if recorder.Code != testCase.statusCode || !strings.Contains(recorder.Body.String(), i18n.GetMessage(testCase.messageKey)) {
				t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestGetSubtitleWebVttHandlerConvertsSrtWithHeaders(t *testing.T) {
	subtitleFile := writeSubtitleOnDisk(t, "Filme.pt.srt", []byte("1\r\n00:00:01,000 --> 00:00:02,000\r\nA\xE7\xE3o\r\n"))
	source := &subtitleFileSourceFake{filesByID: map[int]files.FileDto{5: subtitleFile}}

	recorder := performSubtitleRequest(NewSubtitleHandler(source, &subtitleServiceFake{}), "/video/subtitles/5/vtt", nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status %d %q", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.String() != "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nAção\n" {
		t.Fatalf("unexpected body %q", recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "text/vtt; charset=utf-8" ||
		recorder.Header().Get("Cache-Control") != "public, max-age=3600" ||
		recorder.Header().Get("ETag") == "" {
		t.Fatalf("unexpected headers %v", recorder.Header())
	}
}

func TestGetSubtitleWebVttHandlerPassesVttThrough(t *testing.T) {
	content := "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nOla\n"
	source := &subtitleFileSourceFake{filesByID: map[int]files.FileDto{5: writeSubtitleOnDisk(t, "Filme.vtt", []byte(content))}}

	recorder := performSubtitleRequest(NewSubtitleHandler(source, &subtitleServiceFake{}), "/video/subtitles/5/vtt", nil)

	if recorder.Code != http.StatusOK || recorder.Body.String() != content {
		t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestGetSubtitleWebVttHandlerAnswersNotModifiedForMatchingETag(t *testing.T) {
	source := &subtitleFileSourceFake{filesByID: map[int]files.FileDto{5: writeSubtitleOnDisk(t, "Filme.srt", []byte("x"))}}
	handler := NewSubtitleHandler(source, &subtitleServiceFake{})
	firstResponse := performSubtitleRequest(handler, "/video/subtitles/5/vtt", nil)

	secondResponse := performSubtitleRequest(handler, "/video/subtitles/5/vtt", map[string]string{"If-None-Match": firstResponse.Header().Get("ETag")})

	if secondResponse.Code != http.StatusNotModified || secondResponse.Body.Len() != 0 {
		t.Fatalf("unexpected response %d %q", secondResponse.Code, secondResponse.Body.String())
	}
}

func TestGetSubtitleWebVttHandlerErrors(t *testing.T) {
	oversizedContent := make([]byte, maxSubtitleBytes+1)
	missingOnDisk := files.FileDto{ID: 6, Name: "Gone.srt", Path: filepath.Join(t.TempDir(), "Gone.srt")}
	deletedSubtitle := files.FileDto{ID: 7, Name: "Del.srt", DeletedAt: utils.Optional[time.Time]{Value: time.Now(), HasValue: true}}
	testCases := []struct {
		name       string
		source     *subtitleFileSourceFake
		target     string
		statusCode int
		messageKey string
	}{
		{"invalid id", &subtitleFileSourceFake{}, "/video/subtitles/abc/vtt", http.StatusBadRequest, "ERROR_INVALID_REQUEST"},
		{"unknown id", &subtitleFileSourceFake{}, "/video/subtitles/99/vtt", http.StatusNotFound, "ERROR_VIDEO_SUBTITLE_NOT_FOUND"},
		{"deleted subtitle", &subtitleFileSourceFake{filesByID: map[int]files.FileDto{7: deletedSubtitle}}, "/video/subtitles/7/vtt", http.StatusNotFound, "ERROR_VIDEO_SUBTITLE_NOT_FOUND"},
		{"not a subtitle", &subtitleFileSourceFake{filesByID: map[int]files.FileDto{5: writeSubtitleOnDisk(t, "Filme.mkv", []byte("x"))}}, "/video/subtitles/5/vtt", http.StatusBadRequest, "ERROR_VIDEO_SUBTITLE_UNSUPPORTED"},
		{"missing on disk", &subtitleFileSourceFake{filesByID: map[int]files.FileDto{6: missingOnDisk}}, "/video/subtitles/6/vtt", http.StatusNotFound, "ERROR_VIDEO_SUBTITLE_NOT_FOUND"},
		{"too large", &subtitleFileSourceFake{filesByID: map[int]files.FileDto{5: writeSubtitleOnDisk(t, "Big.srt", oversizedContent)}}, "/video/subtitles/5/vtt", http.StatusRequestEntityTooLarge, "ERROR_VIDEO_SUBTITLE_TOO_LARGE"},
		{"source failure", &subtitleFileSourceFake{err: errors.New("db down")}, "/video/subtitles/5/vtt", http.StatusInternalServerError, "ERROR_VIDEO_OPERATION_FAILED"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := performSubtitleRequest(NewSubtitleHandler(testCase.source, &subtitleServiceFake{}), testCase.target, nil)

			if recorder.Code != testCase.statusCode || !strings.Contains(recorder.Body.String(), i18n.GetMessage(testCase.messageKey)) {
				t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
			}
		})
	}
}
