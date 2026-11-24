package video

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	files "nas-go/api/internal/api/v1/files"

	"github.com/gin-gonic/gin"
)

func TestBuildRemuxStartProbeArgumentsBoundsTheReadAroundTarget(t *testing.T) {
	arguments := strings.Join(buildRemuxStartProbeArguments("/m/a.mkv", 100), " ")

	if !strings.Contains(arguments, "-read_intervals 70.000%101.000") || !strings.Contains(arguments, "-select_streams v:0") {
		t.Fatalf("unexpected arguments %s", arguments)
	}
}

func TestBuildRemuxStartProbeArgumentsClampsWindowStartAtZero(t *testing.T) {
	arguments := strings.Join(buildRemuxStartProbeArguments("/m/a.mkv", 12), " ")

	if !strings.Contains(arguments, "-read_intervals 0.000%13.000") {
		t.Fatalf("unexpected arguments %s", arguments)
	}
}

func TestParseRemuxStartSecondsPicksLastKeyframeAtOrBeforeTarget(t *testing.T) {
	probeOutput := []byte("0.000000,K__\n0.160000,___\n10.000000,K__\n10.040000,___\n20.000000,K__\n")

	if got := parseRemuxStartSeconds(probeOutput, 15); got != 10 {
		t.Fatalf("expected 10, got %v", got)
	}
	if got := parseRemuxStartSeconds(probeOutput, 20); got != 20 {
		t.Fatalf("expected 20, got %v", got)
	}
	if got := parseRemuxStartSeconds(probeOutput, 9.9); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestParseRemuxStartSecondsFallsBackToTargetWithoutKeyframes(t *testing.T) {
	if got := parseRemuxStartSeconds([]byte("garbage\n\n1.0,___\n"), 7.5); got != 7.5 {
		t.Fatalf("expected 7.5, got %v", got)
	}
}

func TestBuildRemuxArgumentsNudgesInputSeekPastTheKeyframe(t *testing.T) {
	arguments := strings.Join(BuildRemuxArguments("/m/a.mkv", 10, defaultAudioStreamIndex), " ")

	if !strings.Contains(arguments, "-ss 10.500 -i /m/a.mkv") {
		t.Fatalf("unexpected arguments %s", arguments)
	}
}

func TestBuildRemuxArgumentsOmitsSeekAtStart(t *testing.T) {
	if strings.Contains(strings.Join(BuildRemuxArguments("/m/a.mkv", 0, defaultAudioStreamIndex), " "), "-ss") {
		t.Fatal("expected no -ss at start 0")
	}
}

func performRemuxStartRequest(handler *RemuxStartHandler, target string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/video/stream/:file_id/remux/start", handler.GetRemuxStartHandler)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder
}

func remuxStartProbeReturning(probeOutput string, err error) SubtitleProbeRunner {
	return func(ctx context.Context, arguments []string) ([]byte, error) {
		return []byte(probeOutput), err
	}
}

func TestRemuxStartHandlerReturnsKeyframeStart(t *testing.T) {
	fileSource := &remuxFileSourceFake{file: files.FileDto{ID: 1, Path: "/m/a.mkv"}}
	handler := NewRemuxStartHandler(fileSource, remuxStartProbeReturning("0.0,K__\n10.0,K__\n12.0,___\n", nil))

	recorder := performRemuxStartRequest(handler, "/video/stream/1/remux/start?at=15")

	var body RemuxStartDto
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || recorder.Code != http.StatusOK || body.StartSeconds != 10 {
		t.Fatalf("unexpected response %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestRemuxStartHandlerFailures(t *testing.T) {
	existingFile := &remuxFileSourceFake{file: files.FileDto{ID: 1, Path: "/m/a.mkv"}}
	testCases := []struct {
		name         string
		fileSource   *remuxFileSourceFake
		probeRunner  SubtitleProbeRunner
		target       string
		expectedCode int
	}{
		{"invalid file id", existingFile, remuxStartProbeReturning("", nil), "/video/stream/abc/remux/start?at=1", http.StatusBadRequest},
		{"missing at", existingFile, remuxStartProbeReturning("", nil), "/video/stream/1/remux/start", http.StatusBadRequest},
		{"negative at", existingFile, remuxStartProbeReturning("", nil), "/video/stream/1/remux/start?at=-1", http.StatusBadRequest},
		{"non numeric at", existingFile, remuxStartProbeReturning("", nil), "/video/stream/1/remux/start?at=x", http.StatusBadRequest},
		{"unknown file", &remuxFileSourceFake{err: sql.ErrNoRows}, remuxStartProbeReturning("", nil), "/video/stream/1/remux/start?at=1", http.StatusNotFound},
		{"file lookup failure", &remuxFileSourceFake{err: errors.New("db down")}, remuxStartProbeReturning("", nil), "/video/stream/1/remux/start?at=1", http.StatusInternalServerError},
		{"ffprobe missing", existingFile, remuxStartProbeReturning("", exec.ErrNotFound), "/video/stream/1/remux/start?at=1", http.StatusNotImplemented},
		{"ffprobe failure", existingFile, remuxStartProbeReturning("", errors.New("boom")), "/video/stream/1/remux/start?at=1", http.StatusInternalServerError},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := performRemuxStartRequest(NewRemuxStartHandler(testCase.fileSource, testCase.probeRunner), testCase.target)

			if recorder.Code != testCase.expectedCode {
				t.Fatalf("expected %d, got %d %s", testCase.expectedCode, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestFFprobeRemuxStartHandlerIsWired(t *testing.T) {
	if NewFFprobeRemuxStartHandler(&remuxFileSourceFake{}).probeRunner == nil {
		t.Fatal("expected ffprobe runner")
	}
}

func TestRemuxStartsExactlyAtKeyframeReturnedByProbeWithRealFFmpeg(t *testing.T) {
	for _, executableName := range []string{ffmpegExecutableName, ffprobeExecutableName} {
		if _, err := exec.LookPath(executableName); err != nil {
			t.Skipf("%s not installed", executableName)
		}
	}
	const sourceDurationSeconds = 30.0
	sourcePath := filepath.Join(t.TempDir(), "keyframes.mkv")
	buildCommand := exec.Command(ffmpegExecutableName, "-v", "error", "-y",
		"-f", "lavfi", "-i", "testsrc=s=64x64:r=25:d=30",
		"-c:v", "libx264", "-g", "250", "-keyint_min", "250", "-force_key_frames", "10,20", sourcePath)
	if output, err := buildCommand.CombinedOutput(); err != nil {
		t.Skipf("ffmpeg could not build the fixture: %v %s", err, output)
	}
	fileSource := &remuxFileSourceFake{file: files.FileDto{ID: 1, Path: sourcePath}}
	handler := NewFFprobeRemuxStartHandler(fileSource)

	for _, testCase := range []struct{ atSeconds, keyframeSeconds float64 }{{10, 10}, {15, 10}, {20, 20}, {5, 0}} {
		target := "/video/stream/1/remux/start?at=" + strconv.FormatFloat(testCase.atSeconds, 'f', -1, 64)
		recorder := performRemuxStartRequest(handler, target)
		var body RemuxStartDto
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body.StartSeconds != testCase.keyframeSeconds {
			t.Fatalf("at=%v: expected keyframe %v, got %s", testCase.atSeconds, testCase.keyframeSeconds, recorder.Body.String())
		}

		remuxedDuration := remuxedDurationSeconds(t, sourcePath, body.StartSeconds)

		expectedDuration := sourceDurationSeconds - testCase.keyframeSeconds
		if remuxedDuration < expectedDuration-0.1 || remuxedDuration > expectedDuration+0.1 {
			t.Fatalf("at=%v: remux from %v lasts %v, expected %v", testCase.atSeconds, body.StartSeconds, remuxedDuration, expectedDuration)
		}
	}
}

func remuxedDurationSeconds(t *testing.T, sourcePath string, startSeconds float64) float64 {
	t.Helper()
	remuxedPath := filepath.Join(t.TempDir(), "remuxed.mp4")
	remuxedFile, err := os.Create(remuxedPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := RunFFmpegRemux(context.Background(), BuildRemuxArguments(sourcePath, startSeconds, defaultAudioStreamIndex), remuxedFile); err != nil {
		t.Fatal(err)
	}
	remuxedFile.Close()
	durationOutput, err := exec.Command(ffprobeExecutableName, "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", remuxedPath).Output()
	if err != nil {
		t.Fatal(err)
	}
	durationSeconds, err := strconv.ParseFloat(strings.TrimSpace(string(durationOutput)), 64)
	if err != nil {
		t.Fatal(err)
	}
	return durationSeconds
}
