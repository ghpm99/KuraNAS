package video

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type embeddedSubtitleServiceFake struct {
	subtitles   []SubtitleDto
	webVtt      []byte
	err         error
	readCalls   int
	lastStream  int
	lastVideoID int
}

func (fake *embeddedSubtitleServiceFake) ListEmbeddedSubtitles(videoFile files.FileDto) []SubtitleDto {
	return fake.subtitles
}

func (fake *embeddedSubtitleServiceFake) ReadEmbeddedSubtitleWebVtt(videoFile files.FileDto, streamIndex int) ([]byte, error) {
	fake.readCalls++
	fake.lastStream, fake.lastVideoID = streamIndex, videoFile.ID
	return fake.webVtt, fake.err
}

func performEmbeddedSubtitleRequest(handler *SubtitleHandler, target string, headers map[string]string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/video/subtitles/:id", handler.ListVideoSubtitlesHandler)
	router.GET("/video/subtitles/:id/vtt", handler.GetSubtitleWebVttHandler)
	router.GET("/video/subtitles/:id/embedded/:stream_index/vtt", handler.GetEmbeddedSubtitleWebVttHandler)
	request := httptest.NewRequest(http.MethodGet, target, nil)
	for headerName, headerValue := range headers {
		request.Header.Set(headerName, headerValue)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func videoSourceForEmbeddedTest() *subtitleFileSourceFake {
	return &subtitleFileSourceFake{filesByID: map[int]files.FileDto{
		1: {ID: 1, Name: "Filme.mkv", ParentPath: "/data", Size: 10, UpdatedAt: time.Unix(1700000000, 0)},
	}}
}

func TestListVideoSubtitlesHandlerMergesExternalAndEmbeddedTracks(t *testing.T) {
	streamIndex := 3
	external := &subtitleServiceFake{subtitles: []SubtitleDto{{ID: 9, Source: "external", Label: "en", Language: "en", Format: "srt"}}}
	embedded := &embeddedSubtitleServiceFake{subtitles: []SubtitleDto{
		{Source: "embedded", Key: "1:3", StreamIndex: &streamIndex, Label: "Full", Language: "pt", Format: "vtt", IsForced: true},
	}}
	handler := NewSubtitleHandler(videoSourceForEmbeddedTest(), external)
	handler.UseEmbeddedSubtitles(embedded)

	recorder := performEmbeddedSubtitleRequest(handler, "/video/subtitles/1", nil)

	expectedBody := `[{"id":9,"source":"external","label":"en","language":"en","format":"srt","is_forced":false},` +
		`{"id":0,"source":"embedded","key":"1:3","stream_index":3,"label":"Full","language":"pt","format":"vtt","is_forced":true}]`
	if recorder.Code != http.StatusOK || recorder.Body.String() != expectedBody {
		t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestGetEmbeddedSubtitleWebVttHandlerServesVttWithHeaders(t *testing.T) {
	embedded := &embeddedSubtitleServiceFake{webVtt: []byte("WEBVTT\n\n")}
	handler := NewSubtitleHandler(videoSourceForEmbeddedTest(), &subtitleServiceFake{})
	handler.UseEmbeddedSubtitles(embedded)

	recorder := performEmbeddedSubtitleRequest(handler, "/video/subtitles/1/embedded/3/vtt", nil)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "WEBVTT\n\n" {
		t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "text/vtt; charset=utf-8" ||
		recorder.Header().Get("Cache-Control") != "public, max-age=3600" ||
		recorder.Header().Get("ETag") == "" {
		t.Fatalf("unexpected headers %v", recorder.Header())
	}
	if embedded.lastStream != 3 || embedded.lastVideoID != 1 {
		t.Fatalf("service received video %d stream %d", embedded.lastVideoID, embedded.lastStream)
	}
}

func TestGetEmbeddedSubtitleWebVttHandlerAnswersNotModifiedWithoutExtracting(t *testing.T) {
	embedded := &embeddedSubtitleServiceFake{webVtt: []byte("WEBVTT\n\n")}
	handler := NewSubtitleHandler(videoSourceForEmbeddedTest(), &subtitleServiceFake{})
	handler.UseEmbeddedSubtitles(embedded)
	firstResponse := performEmbeddedSubtitleRequest(handler, "/video/subtitles/1/embedded/3/vtt", nil)

	secondResponse := performEmbeddedSubtitleRequest(handler, "/video/subtitles/1/embedded/3/vtt", map[string]string{"If-None-Match": firstResponse.Header().Get("ETag")})

	if secondResponse.Code != http.StatusNotModified || secondResponse.Body.Len() != 0 || embedded.readCalls != 1 {
		t.Fatalf("unexpected response %d %q calls=%d", secondResponse.Code, secondResponse.Body.String(), embedded.readCalls)
	}
}

func TestGetEmbeddedSubtitleWebVttHandlerErrors(t *testing.T) {
	deletedVideo := files.FileDto{ID: 2, Name: "Gone.mkv", DeletedAt: utils.Optional[time.Time]{Value: time.Now(), HasValue: true}}
	testCases := []struct {
		name       string
		source     *subtitleFileSourceFake
		embedded   *embeddedSubtitleServiceFake
		target     string
		statusCode int
		messageKey string
	}{
		{"invalid video id", videoSourceForEmbeddedTest(), &embeddedSubtitleServiceFake{}, "/video/subtitles/abc/embedded/3/vtt", http.StatusBadRequest, "ERROR_INVALID_REQUEST"},
		{"invalid stream index", videoSourceForEmbeddedTest(), &embeddedSubtitleServiceFake{}, "/video/subtitles/1/embedded/x/vtt", http.StatusBadRequest, "ERROR_INVALID_REQUEST"},
		{"unknown video", &subtitleFileSourceFake{}, &embeddedSubtitleServiceFake{}, "/video/subtitles/99/embedded/3/vtt", http.StatusNotFound, "ERROR_VIDEO_NOT_FOUND"},
		{"deleted video", &subtitleFileSourceFake{filesByID: map[int]files.FileDto{2: deletedVideo}}, &embeddedSubtitleServiceFake{}, "/video/subtitles/2/embedded/3/vtt", http.StatusNotFound, "ERROR_VIDEO_NOT_FOUND"},
		{"missing or image track", videoSourceForEmbeddedTest(), &embeddedSubtitleServiceFake{err: ErrEmbeddedSubtitleNotFound}, "/video/subtitles/1/embedded/4/vtt", http.StatusNotFound, "ERROR_VIDEO_SUBTITLE_NOT_FOUND"},
		{"extraction failure", videoSourceForEmbeddedTest(), &embeddedSubtitleServiceFake{err: errors.New("ffmpeg failed")}, "/video/subtitles/1/embedded/3/vtt", http.StatusInternalServerError, "ERROR_VIDEO_OPERATION_FAILED"},
		{"source failure", &subtitleFileSourceFake{err: errors.New("db down")}, &embeddedSubtitleServiceFake{}, "/video/subtitles/1/embedded/3/vtt", http.StatusInternalServerError, "ERROR_VIDEO_OPERATION_FAILED"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := NewSubtitleHandler(testCase.source, &subtitleServiceFake{})
			handler.UseEmbeddedSubtitles(testCase.embedded)

			recorder := performEmbeddedSubtitleRequest(handler, testCase.target, nil)

			if recorder.Code != testCase.statusCode || !strings.Contains(recorder.Body.String(), i18n.GetMessage(testCase.messageKey)) {
				t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
			}
			if recorder.Header().Get("ETag") != "" {
				t.Fatalf("error response must not carry an ETag: %v", recorder.Header())
			}
		})
	}
}

func TestGetEmbeddedSubtitleWebVttHandlerAnswersNotFoundWithoutEmbeddedService(t *testing.T) {
	handler := NewSubtitleHandler(videoSourceForEmbeddedTest(), &subtitleServiceFake{})

	recorder := performEmbeddedSubtitleRequest(handler, "/video/subtitles/1/embedded/3/vtt", nil)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("unexpected status %d", recorder.Code)
	}
}
