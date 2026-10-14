package video

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type recordingLogService struct {
	logger.LoggerServiceInterface
	mutex         sync.Mutex
	createdLogs   int
	successLogs   int
	errorMessages []string
}

func (service *recordingLogService) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.createdLogs++
	return log, nil
}

func (service *recordingLogService) CompleteWithSuccessLog(log logger.LoggerModel) error {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.successLogs++
	return nil
}

func (service *recordingLogService) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.errorMessages = append(service.errorMessages, err.Error())
	return nil
}

type recordingRecentService struct {
	videoRecentServiceMock
	mutex          sync.Mutex
	registeredKeys []playbackAccessKey
}

func (service *recordingRecentService) RegisterAccess(ip string, fileID int) error {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.registeredKeys = append(service.registeredKeys, playbackAccessKey{clientIP: ip, fileID: fileID})
	return nil
}

type streamAccessFixture struct {
	handler       *Handler
	router        *gin.Engine
	recentService *recordingRecentService
	logService    *recordingLogService
	clock         *fakeClock
}

func newStreamAccessFixture(t *testing.T) *streamAccessFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	videoPath := filepath.Join(t.TempDir(), "v.mp4")
	if err := os.WriteFile(videoPath, []byte("0123456789abcdefghijklmnopqrstuvwxyz"), 0644); err != nil {
		t.Fatal(err)
	}
	fixture := &streamAccessFixture{
		recentService: &recordingRecentService{},
		logService:    &recordingLogService{},
		clock:         newFakeClock(),
	}
	fixture.handler = NewHandler(&videoHandlerServiceMock{}, &videoFilesServiceMock{filePath: videoPath, format: ".mp4"}, fixture.recentService, fixture.logService)
	fixture.handler.playbackAccess = newPlaybackAccessTracker(fixture.clock.now, 5*time.Minute, 100)
	fixture.router = gin.New()
	fixture.router.GET("/files/video-stream/:id", fixture.handler.StreamVideoHandler)
	return fixture
}

func (fixture *streamAccessFixture) stream(rangeHeader string, clientIP string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/files/video-stream/1", nil)
	request.RemoteAddr = clientIP + ":1234"
	if rangeHeader != "" {
		request.Header.Set("Range", rangeHeader)
	}
	recorder := httptest.NewRecorder()
	fixture.router.ServeHTTP(recorder, request)
	return recorder
}

func (fixture *streamAccessFixture) assertCounts(t *testing.T, accesses, createdLogs, successLogs int) {
	t.Helper()
	if len(fixture.recentService.registeredKeys) != accesses || fixture.logService.createdLogs != createdLogs || fixture.logService.successLogs != successLogs {
		t.Fatalf("expected accesses=%d logs=%d successes=%d, got accesses=%d logs=%d successes=%d",
			accesses, createdLogs, successLogs, len(fixture.recentService.registeredKeys), fixture.logService.createdLogs, fixture.logService.successLogs)
	}
}

func TestStreamVideoRegistersAccessWhenRequestHasNoRange(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	fixture.stream("", "10.0.0.1")
	fixture.assertCounts(t, 1, 1, 1)
}

func TestStreamVideoRegistersAccessWhenRangeStartsAtZero(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	fixture.stream("bytes=0-", "10.0.0.1")
	fixture.assertCounts(t, 1, 1, 1)
}

func TestStreamVideoSkipsAccessAndLogWhenRangeIsContinuation(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	recorder := fixture.stream("bytes=10-", "10.0.0.1")
	if recorder.Code != http.StatusPartialContent {
		t.Fatalf("unexpected status %d", recorder.Code)
	}
	fixture.assertCounts(t, 0, 0, 0)
}

func TestStreamVideoRegistersBrowserProbeFollowedByFullRangeOnce(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	fixture.stream("bytes=0-1", "10.0.0.1")
	fixture.stream("bytes=0-", "10.0.0.1")
	fixture.assertCounts(t, 1, 1, 1)
}

func TestStreamVideoRegistersAgainAfterWindowExpires(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	fixture.stream("bytes=0-", "10.0.0.1")
	fixture.clock.advance(5 * time.Minute)
	fixture.stream("bytes=0-", "10.0.0.1")
	fixture.assertCounts(t, 2, 2, 2)
}

func TestStreamVideoRegistersDifferentClientsSeparately(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	fixture.stream("bytes=0-", "10.0.0.1")
	fixture.stream("bytes=0-", "10.0.0.2")
	fixture.assertCounts(t, 2, 2, 2)
}

func TestStreamVideoContinuationErrorStillReachesErrorLog(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	missingFileHandler := NewHandler(&videoHandlerServiceMock{}, &videoFilesServiceMock{filePath: "/nonexistent.mp4"}, fixture.recentService, fixture.logService)
	router := gin.New()
	router.GET("/files/video-stream/:id", missingFileHandler.StreamVideoHandler)
	request := httptest.NewRequest(http.MethodGet, "/files/video-stream/1", nil)
	request.Header.Set("Range", "bytes=1000-")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError && recorder.Code != http.StatusNotFound {
		t.Fatalf("unexpected status %d", recorder.Code)
	}
	if len(fixture.logService.errorMessages) != 1 || fixture.logService.createdLogs != 1 || fixture.logService.successLogs != 0 {
		t.Fatalf("expected one error log, got errors=%v created=%d successes=%d", fixture.logService.errorMessages, fixture.logService.createdLogs, fixture.logService.successLogs)
	}
}

func TestRemuxAndTranscodeShareDeduplicationWithDirectStream(t *testing.T) {
	fixture := newStreamAccessFixture(t)
	shared := &SharedPlaybackAccess{tracker: fixture.handler.playbackAccess}
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		_, err := output.Write([]byte("bytes"))
		return err
	}
	remuxHandler := NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), runner, isPresent, 2, fixture.logService)
	transcodeHandler := NewTranscodeHandler(newRemuxSourceFile(t), h264Codec(), runner, isPresent, 2, fixture.logService)
	remuxHandler.UseSharedPlaybackAccess(shared)
	transcodeHandler.UseSharedPlaybackAccess(shared)

	fixture.stream("bytes=0-", "192.0.2.1")
	performRemuxRequest(remuxHandler, "/video/stream/1/remux")
	performTranscodeRequestFromIP(transcodeHandler, "/video/stream/1/transcode", "192.0.2.1")
	fixture.assertCounts(t, 1, 1, 1)
}

func TestRemuxLogsOnlyPlaybackStartAndStillLogsContinuationErrors(t *testing.T) {
	logService := &recordingLogService{}
	failingRunner := func(ctx context.Context, arguments []string, output io.Writer) error {
		return errors.New("ffmpeg exploded")
	}
	handler := NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), failingRunner, isPresent, 2, logService)

	performRemuxRequest(handler, "/video/stream/1/remux?start=30")

	if logService.createdLogs != 1 || len(logService.errorMessages) != 1 || logService.successLogs != 0 {
		t.Fatalf("expected one forensic error log, got created=%d errors=%v successes=%d", logService.createdLogs, logService.errorMessages, logService.successLogs)
	}
}

func TestRemuxSeekRequestDoesNotCreateSuccessLog(t *testing.T) {
	logService := &recordingLogService{}
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		_, err := output.Write([]byte("bytes"))
		return err
	}
	handler := NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), runner, isPresent, 2, logService)

	performRemuxRequest(handler, "/video/stream/1/remux?start=30")

	if logService.createdLogs != 0 || logService.successLogs != 0 {
		t.Fatalf("seek must not log, got created=%d successes=%d", logService.createdLogs, logService.successLogs)
	}
}

func performTranscodeRequestFromIP(handler *TranscodeHandler, target string, clientIP string) *httptest.ResponseRecorder {
	router := gin.New()
	router.GET("/video/stream/:file_id/transcode", handler.StreamTranscodedVideoHandler)
	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.RemoteAddr = clientIP + ":1234"
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
