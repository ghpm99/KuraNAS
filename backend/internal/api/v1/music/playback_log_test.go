package music

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"nas-go/api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type playbackLoggerSpy struct {
	logger.LoggerServiceInterface
	createdLogs int
}

func (spy *playbackLoggerSpy) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	spy.createdLogs++
	return logger.LoggerModel{ID: spy.createdLogs}, nil
}
func (spy *playbackLoggerSpy) CompleteWithSuccessLog(log logger.LoggerModel) error { return nil }
func (spy *playbackLoggerSpy) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	return nil
}

type recentAccessSpy struct {
	musicRecentServiceMock
	registeredAccesses int
}

func (spy *recentAccessSpy) RegisterAccess(ip string, fileID int) error {
	spy.registeredAccesses++
	return nil
}

func TestStreamAudioRecordsOnlyPlaybackStart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	audioPath := filepath.Join(t.TempDir(), "a.mp3")
	if err := os.WriteFile(audioPath, []byte("abcdefghijklmnopqrstuvwxyz"), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name        string
		rangeHeader string
		isRecorded  bool
	}{
		{"no range", "", true},
		{"open range from zero", "bytes=0-", true},
		{"closed range from zero", "bytes=0-5", true},
		{"range from middle", "bytes=10-", false},
		{"suffix range", "bytes=-5", false},
		{"unknown unit", "items=0-5", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			loggerSpy := &playbackLoggerSpy{}
			recentSpy := &recentAccessSpy{}
			filesService := &musicStreamFilesServiceMock{filePath: audioPath, format: ".mp3"}
			handler := NewHandler(&musicHandlerServiceMock{}, filesService, recentSpy, loggerSpy)
			router := gin.New()
			router.GET("/files/stream/:id", handler.StreamAudioHandler)

			request := httptest.NewRequest(http.MethodGet, "/files/stream/1", nil)
			if testCase.rangeHeader != "" {
				request.Header.Set("Range", testCase.rangeHeader)
			}
			router.ServeHTTP(httptest.NewRecorder(), request)

			expectedCount := 0
			if testCase.isRecorded {
				expectedCount = 1
			}
			if loggerSpy.createdLogs != expectedCount || recentSpy.registeredAccesses != expectedCount {
				t.Fatalf("expected %d logs and accesses, got %d and %d", expectedCount, loggerSpy.createdLogs, recentSpy.registeredAccesses)
			}
		})
	}
}

func TestTranscodeLogsOnlyWhenStartIsZero(t *testing.T) {
	runner := func(ctx context.Context, arguments []string, output io.Writer) error {
		_, err := output.Write([]byte("x"))
		return err
	}
	cases := []struct {
		target       string
		expectedLogs int
	}{
		{"/music/tracks/1/stream", 1},
		{"/music/tracks/1/stream?start=0", 1},
		{"/music/tracks/1/stream?start=30", 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.target, func(t *testing.T) {
			loggerSpy := &playbackLoggerSpy{}
			handler := NewTranscodeHandler(newTranscodeSourceFile(t), runner, alwaysPresent, 2, loggerSpy)

			performTranscodeRequest(handler, testCase.target)

			if loggerSpy.createdLogs != testCase.expectedLogs {
				t.Fatalf("expected %d logs, got %d", testCase.expectedLogs, loggerSpy.createdLogs)
			}
		})
	}
}
