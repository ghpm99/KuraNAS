package music

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"

	"github.com/gin-gonic/gin"
)

type transcodeTrackSourceFake struct {
	track files.FileDto
	err   error
}

func (fake *transcodeTrackSourceFake) GetFileById(id int) (files.FileDto, error) {
	return fake.track, fake.err
}

func newTranscodeSourceFile(t *testing.T) *transcodeTrackSourceFake {
	t.Helper()
	sourcePath := filepath.Join(t.TempDir(), "track.wma")
	if err := os.WriteFile(sourcePath, []byte("raw"), 0o644); err != nil {
		t.Fatal(err)
	}
	return &transcodeTrackSourceFake{track: files.FileDto{ID: 1, Path: sourcePath}}
}

func performTranscodeRequest(handler *TranscodeHandler, target string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/music/tracks/:file_id/stream", handler.StreamTranscodedTrackHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder
}

func alwaysPresent() bool { return true }

func TestTranscodeHandlerPipesRunnerOutputWithFormatHeaders(t *testing.T) {
	var receivedArguments []string
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		receivedArguments = arguments
		_, err := output.Write([]byte("encoded-bytes"))
		return err
	}
	handler := NewTranscodeHandler(newTranscodeSourceFile(t), runner, alwaysPresent, 2, &musicLoggerMock{})

	recorder := performTranscodeRequest(handler, "/music/tracks/1/stream?format=aac&start=12.5")

	if recorder.Code != http.StatusOK || recorder.Body.String() != "encoded-bytes" {
		t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "audio/aac" || recorder.Header().Get("Accept-Ranges") != "none" {
		t.Fatalf("unexpected headers %v", recorder.Header())
	}
	joinedArguments := strings.Join(receivedArguments, " ")
	if !strings.Contains(joinedArguments, "-ss 12.500") || !strings.Contains(joinedArguments, "-f adts") || !strings.Contains(joinedArguments, "-vn") {
		t.Fatalf("unexpected ffmpeg arguments %s", joinedArguments)
	}
}

func TestTranscodeHandlerDefaultsToMp3WithoutSeek(t *testing.T) {
	var receivedArguments []string
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		receivedArguments = arguments
		_, err := output.Write([]byte("x"))
		return err
	}
	handler := NewTranscodeHandler(newTranscodeSourceFile(t), runner, alwaysPresent, 0, &musicLoggerMock{})

	recorder := performTranscodeRequest(handler, "/music/tracks/1/stream")

	joinedArguments := strings.Join(receivedArguments, " ")
	if recorder.Header().Get("Content-Type") != "audio/mpeg" || strings.Contains(joinedArguments, "-ss") || !strings.Contains(joinedArguments, "-f mp3") {
		t.Fatalf("unexpected defaults %v %s", recorder.Header(), joinedArguments)
	}
}

func TestTranscodeHandlerAnswers501WhenFFmpegIsMissing(t *testing.T) {
	runnerWasCalled := false
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		runnerWasCalled = true
		return nil
	}
	handler := NewTranscodeHandler(newTranscodeSourceFile(t), runner, func() bool { return false }, 2, &musicLoggerMock{})

	recorder := performTranscodeRequest(handler, "/music/tracks/1/stream?format=mp3")

	if recorder.Code != http.StatusNotImplemented || runnerWasCalled {
		t.Fatalf("expected 501 without running, got %d", recorder.Code)
	}
}

func TestTranscodeHandlerAnswers503WhenConcurrencyLimitIsReached(t *testing.T) {
	runnerStarted := make(chan struct{})
	releaseRunner := make(chan struct{})
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		close(runnerStarted)
		<-releaseRunner
		return nil
	}
	handler := NewTranscodeHandler(newTranscodeSourceFile(t), runner, alwaysPresent, 1, &musicLoggerMock{})

	firstFinished := make(chan struct{})
	go func() {
		performTranscodeRequest(handler, "/music/tracks/1/stream")
		close(firstFinished)
	}()
	<-runnerStarted

	rejected := performTranscodeRequest(handler, "/music/tracks/1/stream")
	close(releaseRunner)
	<-firstFinished

	if rejected.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rejected.Code)
	}
	if handler.tryAcquireSlot() == false {
		t.Fatal("slot should be released after the transcode ends")
	}
}

func TestTranscodeHandlerCancelsRunnerWhenClientDisconnects(t *testing.T) {
	runnerStarted := make(chan struct{})
	runnerCancelled := make(chan struct{})
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		close(runnerStarted)
		<-ctx.Done()
		close(runnerCancelled)
		return ctx.Err()
	}
	handler := NewTranscodeHandler(newTranscodeSourceFile(t), runner, alwaysPresent, 1, &musicLoggerMock{})
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/music/tracks/:file_id/stream", handler.StreamTranscodedTrackHandler)
	requestContext, disconnect := context.WithCancel(context.Background())
	request := httptest.NewRequest(http.MethodGet, "/music/tracks/1/stream", nil).WithContext(requestContext)
	handlerFinished := make(chan struct{})
	go func() {
		router.ServeHTTP(httptest.NewRecorder(), request)
		close(handlerFinished)
	}()

	<-runnerStarted
	disconnect()

	select {
	case <-runnerCancelled:
	case <-time.After(2 * time.Second):
		t.Fatal("runner was not cancelled")
	}
	<-handlerFinished
}

func TestTranscodeHandlerErrors(t *testing.T) {
	failingRunner := func(ctx context.Context, arguments []string, output io.Writer) error {
		return errors.New("ffmpeg crashed")
	}
	noopRunner := func(ctx context.Context, arguments []string, output io.Writer) error { return nil }
	missingOnDisk := &transcodeTrackSourceFake{track: files.FileDto{ID: 1, Path: filepath.Join(t.TempDir(), "gone.wma")}}

	testCases := []struct {
		name         string
		handler      *TranscodeHandler
		target       string
		expectedCode int
	}{
		{"runner failure before output", NewTranscodeHandler(newTranscodeSourceFile(t), failingRunner, alwaysPresent, 1, &musicLoggerMock{}), "/music/tracks/1/stream", http.StatusInternalServerError},
		{"unknown format", NewTranscodeHandler(newTranscodeSourceFile(t), noopRunner, alwaysPresent, 1, &musicLoggerMock{}), "/music/tracks/1/stream?format=wav", http.StatusBadRequest},
		{"negative start", NewTranscodeHandler(newTranscodeSourceFile(t), noopRunner, alwaysPresent, 1, &musicLoggerMock{}), "/music/tracks/1/stream?start=-1", http.StatusBadRequest},
		{"invalid start", NewTranscodeHandler(newTranscodeSourceFile(t), noopRunner, alwaysPresent, 1, &musicLoggerMock{}), "/music/tracks/1/stream?start=abc", http.StatusBadRequest},
		{"invalid id", NewTranscodeHandler(newTranscodeSourceFile(t), noopRunner, alwaysPresent, 1, &musicLoggerMock{}), "/music/tracks/abc/stream", http.StatusBadRequest},
		{"unknown track", NewTranscodeHandler(&transcodeTrackSourceFake{err: errors.New("no rows")}, noopRunner, alwaysPresent, 1, &musicLoggerMock{}), "/music/tracks/1/stream", http.StatusNotFound},
		{"file gone from disk", NewTranscodeHandler(missingOnDisk, noopRunner, alwaysPresent, 1, &musicLoggerMock{}), "/music/tracks/1/stream", http.StatusNotFound},
	}
	for _, testCase := range testCases {
		if code := performTranscodeRequest(testCase.handler, testCase.target).Code; code != testCase.expectedCode {
			t.Fatalf("%s: expected %d, got %d", testCase.name, testCase.expectedCode, code)
		}
	}
}

func TestResolveTranscodeFormatRejectsUnknownName(t *testing.T) {
	if _, err := ResolveTranscodeFormat("wav"); !errors.Is(err, ErrUnsupportedTranscodeFormat) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestRunFFmpegTranscodeFailsWithoutBinaryOrBadInput(t *testing.T) {
	var sink strings.Builder
	if IsFFmpegInstalled() {
		if err := RunFFmpegTranscode(context.Background(), []string{"-i", "/nonexistent/input.wma", "-f", "mp3", "pipe:1"}, &sink); err == nil {
			t.Fatal("expected ffmpeg to fail on a missing input")
		}
		return
	}
	if err := RunFFmpegTranscode(context.Background(), []string{"-version"}, &sink); err == nil {
		t.Fatal("expected failure when ffmpeg is not installed")
	}
}
