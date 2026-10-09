package video

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

	files "nas-go/api/internal/api/v1/files"

	"github.com/gin-gonic/gin"
)

type remuxFileSourceFake struct {
	file files.FileDto
	err  error
}

func (fake *remuxFileSourceFake) GetFileById(id int) (files.FileDto, error) {
	return fake.file, fake.err
}

type remuxCodecSourceFake struct {
	summary VideoSummaryDto
	err     error
}

func (fake *remuxCodecSourceFake) GetVideoSummary(fileID int) (VideoSummaryDto, error) {
	return fake.summary, fake.err
}

func newRemuxSourceFile(t *testing.T) *remuxFileSourceFake {
	t.Helper()
	sourcePath := filepath.Join(t.TempDir(), "movie.mkv")
	if err := os.WriteFile(sourcePath, []byte("raw"), 0o644); err != nil {
		t.Fatal(err)
	}
	return &remuxFileSourceFake{file: files.FileDto{ID: 1, Path: sourcePath}}
}

func h264Codec() *remuxCodecSourceFake {
	return &remuxCodecSourceFake{summary: VideoSummaryDto{CodecName: "h264"}}
}

func isPresent() bool { return true }

func performRemuxRequest(handler *RemuxHandler, target string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/video/stream/:file_id/remux", handler.StreamRemuxedVideoHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder
}

func TestRemuxHandlerPipesRunnerOutputWithFragmentedMp4Arguments(t *testing.T) {
	var receivedArguments []string
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		receivedArguments = arguments
		_, err := output.Write([]byte("fmp4-bytes"))
		return err
	}
	handler := NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), runner, isPresent, 2, &videoLoggerMock{})

	recorder := performRemuxRequest(handler, "/video/stream/1/remux?start=12.5")

	if recorder.Code != http.StatusOK || recorder.Body.String() != "fmp4-bytes" {
		t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "video/mp4" || recorder.Header().Get("Accept-Ranges") != "none" {
		t.Fatalf("unexpected headers %v", recorder.Header())
	}
	joinedArguments := strings.Join(receivedArguments, " ")
	expectedFragments := []string{"-ss 12.500 -i ", "-map 0:v:0", "-map 0:a:0?", "-c copy", "-movflags frag_keyframe+empty_moov+default_base_moof", "-f mp4", "pipe:1"}
	for _, fragment := range expectedFragments {
		if !strings.Contains(joinedArguments, fragment) {
			t.Fatalf("missing %q in ffmpeg arguments %s", fragment, joinedArguments)
		}
	}
}

func TestRemuxHandlerOmitsSeekWhenStartIsZero(t *testing.T) {
	var receivedArguments []string
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		receivedArguments = arguments
		_, err := output.Write([]byte("x"))
		return err
	}
	handler := NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), runner, isPresent, 0, &videoLoggerMock{})

	performRemuxRequest(handler, "/video/stream/1/remux")

	if strings.Contains(strings.Join(receivedArguments, " "), "-ss") {
		t.Fatalf("unexpected -ss in %v", receivedArguments)
	}
}

func TestRemuxHandlerAnswers422WhenCodecNeedsTranscode(t *testing.T) {
	runnerWasCalled := false
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		runnerWasCalled = true
		return nil
	}
	for _, codecSource := range []*remuxCodecSourceFake{
		{summary: VideoSummaryDto{CodecName: "hevc"}},
		{summary: VideoSummaryDto{CodecName: "mpeg4"}},
		{err: errors.New("no metadata")},
	} {
		handler := NewRemuxHandler(newRemuxSourceFile(t), codecSource, runner, isPresent, 2, &videoLoggerMock{})
		if code := performRemuxRequest(handler, "/video/stream/1/remux").Code; code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", code)
		}
	}
	if runnerWasCalled {
		t.Fatal("runner must not run for non remuxable codecs")
	}
}

func TestRemuxHandlerAcceptsVp9AndAv1(t *testing.T) {
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		_, err := output.Write([]byte("x"))
		return err
	}
	for _, codecName := range []string{"vp9", "AV1"} {
		handler := NewRemuxHandler(newRemuxSourceFile(t), &remuxCodecSourceFake{summary: VideoSummaryDto{CodecName: codecName}}, runner, isPresent, 2, &videoLoggerMock{})
		if code := performRemuxRequest(handler, "/video/stream/1/remux").Code; code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", codecName, code)
		}
	}
}

func TestRemuxHandlerAnswers501WhenFFmpegIsMissing(t *testing.T) {
	runnerWasCalled := false
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		runnerWasCalled = true
		return nil
	}
	handler := NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), runner, func() bool { return false }, 2, &videoLoggerMock{})

	recorder := performRemuxRequest(handler, "/video/stream/1/remux")

	if recorder.Code != http.StatusNotImplemented || runnerWasCalled {
		t.Fatalf("expected 501 without running, got %d", recorder.Code)
	}
}

func TestRemuxHandlerAnswers503WhenConcurrencyLimitIsReached(t *testing.T) {
	runnerStarted := make(chan struct{})
	releaseRunner := make(chan struct{})
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		close(runnerStarted)
		<-releaseRunner
		return nil
	}
	handler := NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), runner, isPresent, 1, &videoLoggerMock{})

	firstFinished := make(chan struct{})
	go func() {
		performRemuxRequest(handler, "/video/stream/1/remux")
		close(firstFinished)
	}()
	<-runnerStarted

	rejected := performRemuxRequest(handler, "/video/stream/1/remux")
	close(releaseRunner)
	<-firstFinished

	if rejected.Code != http.StatusServiceUnavailable || rejected.Header().Get("Retry-After") == "" {
		t.Fatalf("expected 503 with Retry-After, got %d %v", rejected.Code, rejected.Header())
	}
	if !handler.tryAcquireSlot() {
		t.Fatal("slot should be released after the remux ends")
	}
}

func TestRemuxHandlerCancelsRunnerWhenClientDisconnects(t *testing.T) {
	runnerStarted := make(chan struct{})
	runnerCancelled := make(chan struct{})
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		close(runnerStarted)
		<-ctx.Done()
		close(runnerCancelled)
		return ctx.Err()
	}
	handler := NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), runner, isPresent, 1, &videoLoggerMock{})
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/video/stream/:file_id/remux", handler.StreamRemuxedVideoHandler)
	requestContext, disconnect := context.WithCancel(context.Background())
	request := httptest.NewRequest(http.MethodGet, "/video/stream/1/remux", nil).WithContext(requestContext)
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

func TestRemuxHandlerErrors(t *testing.T) {
	failingRunner := func(ctx context.Context, arguments []string, output io.Writer) error {
		return errors.New("ffmpeg crashed")
	}
	noopRunner := func(ctx context.Context, arguments []string, output io.Writer) error { return nil }
	missingOnDisk := &remuxFileSourceFake{file: files.FileDto{ID: 1, Path: filepath.Join(t.TempDir(), "gone.mkv")}}

	testCases := []struct {
		name         string
		handler      *RemuxHandler
		target       string
		expectedCode int
	}{
		{"runner failure before output", NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), failingRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/1/remux", http.StatusInternalServerError},
		{"negative start", NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), noopRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/1/remux?start=-1", http.StatusBadRequest},
		{"invalid start", NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), noopRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/1/remux?start=abc", http.StatusBadRequest},
		{"invalid id", NewRemuxHandler(newRemuxSourceFile(t), h264Codec(), noopRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/abc/remux", http.StatusBadRequest},
		{"unknown file", NewRemuxHandler(&remuxFileSourceFake{err: errors.New("no rows")}, h264Codec(), noopRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/1/remux", http.StatusNotFound},
		{"file gone from disk", NewRemuxHandler(missingOnDisk, h264Codec(), noopRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/1/remux", http.StatusNotFound},
	}
	for _, testCase := range testCases {
		if code := performRemuxRequest(testCase.handler, testCase.target).Code; code != testCase.expectedCode {
			t.Fatalf("%s: expected %d, got %d", testCase.name, testCase.expectedCode, code)
		}
	}
}

func TestRunFFmpegRemuxFailsWithBadInput(t *testing.T) {
	var sink strings.Builder
	if IsFFmpegInstalled() {
		if err := RunFFmpegRemux(context.Background(), []string{"-i", filepath.Join(t.TempDir(), "missing.mkv"), "-f", "mp4", "pipe:1"}, &sink); err == nil {
			t.Fatal("expected ffmpeg to fail on a missing input")
		}
		return
	}
	if err := RunFFmpegRemux(context.Background(), nil, &sink); err == nil {
		t.Fatal("expected failure without ffmpeg")
	}
}
