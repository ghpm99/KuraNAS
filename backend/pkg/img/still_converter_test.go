package img

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

func converterWithRunner(runner commandRunner) *FFmpegStillConverter {
	return &FFmpegStillConverter{runCommand: runner, timeout: time.Second}
}

func TestFFmpegStillConverterBuildsCommandAndReturnsOutput(t *testing.T) {
	var executedName string
	var executedArguments []string
	converter := converterWithRunner(func(_ context.Context, name string, args ...string) ([]byte, error) {
		executedName, executedArguments = name, args
		return []byte{0xFF, 0xD8, 0xFF}, nil
	})

	output, err := converter.ConvertToJPEG("/photos/IMG_0001.HEIC")
	if err != nil || len(output) != 3 {
		t.Fatalf("unexpected output=%v err=%v", output, err)
	}
	if executedName != "ffmpeg" || !slices.Contains(executedArguments, "/photos/IMG_0001.HEIC") || !slices.Contains(executedArguments, "pipe:1") {
		t.Fatalf("unexpected command %s %v", executedName, executedArguments)
	}
}

func TestFFmpegStillConverterReportsUnavailable(t *testing.T) {
	failing := converterWithRunner(func(context.Context, string, ...string) ([]byte, error) {
		return nil, errors.New("executable file not found")
	})
	if _, err := failing.ConvertToJPEG("/x.heic"); !errors.Is(err, ErrStillConversionUnavailable) {
		t.Fatalf("expected ErrStillConversionUnavailable, got %v", err)
	}

	empty := converterWithRunner(func(context.Context, string, ...string) ([]byte, error) { return nil, nil })
	if _, err := empty.ConvertToJPEG("/x.heic"); !errors.Is(err, ErrStillConversionUnavailable) {
		t.Fatalf("expected ErrStillConversionUnavailable for empty output, got %v", err)
	}
}

func TestRunCommandCapturingStdout(t *testing.T) {
	output, err := runCommandCapturingStdout(context.Background(), "sh", "-c", "printf abc")
	if err != nil || string(output) != "abc" {
		t.Fatalf("unexpected output=%q err=%v", output, err)
	}
	if _, err := runCommandCapturingStdout(context.Background(), "sh", "-c", "echo boom >&2; exit 3"); err == nil {
		t.Fatal("expected failure to be reported")
	}
}

func TestNewFFmpegStillConverterUsesBoundedTimeout(t *testing.T) {
	converter := NewFFmpegStillConverter()
	if converter.timeout != stillConversionTimeout || converter.runCommand == nil {
		t.Fatalf("unexpected defaults %+v", converter)
	}
}
