package video

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func performTranscodeRequest(handler *TranscodeHandler, target string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/video/stream/:file_id/transcode", handler.StreamTranscodedVideoHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder
}

func capturingRunner(receivedArguments *[]string) RemuxRunner {
	return func(ctx context.Context, arguments []string, output io.Writer) error {
		*receivedArguments = arguments
		_, err := output.Write([]byte("fmp4-bytes"))
		return err
	}
}

func TestBuildTranscodeArgumentsUsesH264AacFragmentedMp4(t *testing.T) {
	joinedArguments := strings.Join(BuildTranscodeArguments("/media/movie.mkv", 12.5, 0, 2160), " ")

	expectedFragments := []string{
		"-ss 12.500 -i /media/movie.mkv", "-map 0:v:0", "-map 0:a:0?",
		"-c:v libx264", "-preset veryfast", "-crf 23", "-vf scale=-2:1080",
		"-c:a aac", "-b:a 160k", "-ac 2",
		"-movflags frag_keyframe+empty_moov+default_base_moof", "-f mp4", "pipe:1",
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(joinedArguments, fragment) {
			t.Fatalf("missing %q in %s", fragment, joinedArguments)
		}
	}
}

func TestBuildTranscodeArgumentsOmitsSeekWhenStartIsZero(t *testing.T) {
	if strings.Contains(strings.Join(BuildTranscodeArguments("/a.mkv", 0, 0, 720), " "), "-ss") {
		t.Fatal("unexpected -ss")
	}
}

func TestClampTranscodeHeight(t *testing.T) {
	testCases := []struct {
		name            string
		requestedHeight int
		sourceHeight    int
		expectedHeight  int
	}{
		{"source above ceiling", 0, 2160, 1080},
		{"request above ceiling", 4000, 0, 1080},
		{"request below source", 480, 1080, 480},
		{"source below request", 720, 360, 360},
		{"source only", 0, 576, 576},
		{"odd height rounds down", 481, 1080, 480},
		{"nothing known", 0, 0, 1080},
	}
	for _, testCase := range testCases {
		if got := ClampTranscodeHeight(testCase.requestedHeight, testCase.sourceHeight); got != testCase.expectedHeight {
			t.Fatalf("%s: expected %d, got %d", testCase.name, testCase.expectedHeight, got)
		}
	}
}

func TestBuildTranscodeScaleFilterNeverUpscalesWhenSourceHeightIsUnknown(t *testing.T) {
	if got := buildTranscodeScaleFilter(720, 0); got != `scale=-2:min(720\,ih)` {
		t.Fatalf("unexpected filter %s", got)
	}
}

func TestTranscodeHandlerPipesOutputAndUsesSourceHeight(t *testing.T) {
	var receivedArguments []string
	codecSource := &remuxCodecSourceFake{summary: VideoSummaryDto{CodecName: "hevc", Height: 540}}
	handler := NewTranscodeHandler(newRemuxSourceFile(t), codecSource, capturingRunner(&receivedArguments), isPresent, 2, &videoLoggerMock{})

	recorder := performTranscodeRequest(handler, "/video/stream/1/transcode?start=30&height=720")

	if recorder.Code != http.StatusOK || recorder.Body.String() != "fmp4-bytes" {
		t.Fatalf("unexpected response %d %q", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "video/mp4" {
		t.Fatalf("unexpected headers %v", recorder.Header())
	}
	joinedArguments := strings.Join(receivedArguments, " ")
	if !strings.Contains(joinedArguments, "-ss 30.000") || !strings.Contains(joinedArguments, "scale=-2:540") {
		t.Fatalf("unexpected arguments %s", joinedArguments)
	}
}

func TestTranscodeHandlerTranscodesWhenSummaryIsUnavailable(t *testing.T) {
	var receivedArguments []string
	codecSource := &remuxCodecSourceFake{err: errors.New("no metadata")}
	handler := NewTranscodeHandler(newRemuxSourceFile(t), codecSource, capturingRunner(&receivedArguments), isPresent, 2, &videoLoggerMock{})

	recorder := performTranscodeRequest(handler, "/video/stream/1/transcode?height=480")

	if recorder.Code != http.StatusOK || !strings.Contains(strings.Join(receivedArguments, " "), `scale=-2:min(480\,ih)`) {
		t.Fatalf("unexpected response %d %v", recorder.Code, receivedArguments)
	}
}

func TestTranscodeHandlerAnswers501WhenFFmpegIsMissing(t *testing.T) {
	runnerWasCalled := false
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		runnerWasCalled = true
		return nil
	}
	handler := NewTranscodeHandler(newRemuxSourceFile(t), h264Codec(), runner, func() bool { return false }, 2, &videoLoggerMock{})

	recorder := performTranscodeRequest(handler, "/video/stream/1/transcode")

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
	handler := NewTranscodeHandler(newRemuxSourceFile(t), h264Codec(), runner, isPresent, 1, &videoLoggerMock{})

	firstFinished := make(chan struct{})
	go func() {
		performTranscodeRequest(handler, "/video/stream/1/transcode")
		close(firstFinished)
	}()
	<-runnerStarted

	rejected := performTranscodeRequest(handler, "/video/stream/1/transcode")
	close(releaseRunner)
	<-firstFinished

	if rejected.Code != http.StatusServiceUnavailable || rejected.Header().Get("Retry-After") == "" {
		t.Fatalf("expected 503 with Retry-After, got %d %v", rejected.Code, rejected.Header())
	}
	if !handler.transcodeSlots.tryAcquire() {
		t.Fatal("slot should be released after the transcode ends")
	}
}

func TestNewTranscodeHandlerDefaultsToTwoConcurrentTranscodes(t *testing.T) {
	handler := NewTranscodeHandler(nil, nil, nil, isPresent, 0, &videoLoggerMock{})

	if cap(handler.transcodeSlots) != 2 {
		t.Fatalf("expected 2 slots, got %d", cap(handler.transcodeSlots))
	}
}

func TestNewFFmpegTranscodeHandlerUsesDefaultLimit(t *testing.T) {
	handler := NewFFmpegTranscodeHandler(nil, nil, &videoLoggerMock{})

	if cap(handler.transcodeSlots) != defaultMaxConcurrentTranscodes {
		t.Fatalf("unexpected slots %d", cap(handler.transcodeSlots))
	}
}

func TestTranscodeHandlerErrors(t *testing.T) {
	failingRunner := func(ctx context.Context, arguments []string, output io.Writer) error {
		return errors.New("ffmpeg crashed")
	}
	noopRunner := func(ctx context.Context, arguments []string, output io.Writer) error { return nil }
	missingOnDisk := &remuxFileSourceFake{err: errors.New("no rows")}

	testCases := []struct {
		name         string
		handler      *TranscodeHandler
		target       string
		expectedCode int
	}{
		{"runner failure before output", NewTranscodeHandler(newRemuxSourceFile(t), h264Codec(), failingRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/1/transcode", http.StatusInternalServerError},
		{"negative start", NewTranscodeHandler(newRemuxSourceFile(t), h264Codec(), noopRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/1/transcode?start=-1", http.StatusBadRequest},
		{"invalid height", NewTranscodeHandler(newRemuxSourceFile(t), h264Codec(), noopRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/1/transcode?height=abc", http.StatusBadRequest},
		{"negative height", NewTranscodeHandler(newRemuxSourceFile(t), h264Codec(), noopRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/1/transcode?height=-5", http.StatusBadRequest},
		{"invalid id", NewTranscodeHandler(newRemuxSourceFile(t), h264Codec(), noopRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/abc/transcode", http.StatusBadRequest},
		{"unknown file", NewTranscodeHandler(missingOnDisk, h264Codec(), noopRunner, isPresent, 1, &videoLoggerMock{}), "/video/stream/1/transcode", http.StatusNotFound},
	}
	for _, testCase := range testCases {
		if code := performTranscodeRequest(testCase.handler, testCase.target).Code; code != testCase.expectedCode {
			t.Fatalf("%s: expected %d, got %d", testCase.name, testCase.expectedCode, code)
		}
	}
}

func TestTranscodeHandlerAnswers404WhenFileIsGoneFromDisk(t *testing.T) {
	source := newRemuxSourceFile(t)
	source.file.Path = source.file.Path + ".gone"
	handler := NewTranscodeHandler(source, h264Codec(), nil, isPresent, 1, &videoLoggerMock{})

	if code := performTranscodeRequest(handler, "/video/stream/1/transcode").Code; code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", code)
	}
}
