package img

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

const stillConversionTimeout = 60 * time.Second

var ErrStillConversionUnavailable = errors.New("still image conversion unavailable")

type StillImageConverter interface {
	ConvertToJPEG(sourcePath string) ([]byte, error)
}

type commandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

type FFmpegStillConverter struct {
	runCommand commandRunner
	timeout    time.Duration
}

func NewFFmpegStillConverter() *FFmpegStillConverter {
	return &FFmpegStillConverter{runCommand: runCommandCapturingStdout, timeout: stillConversionTimeout}
}

func (converter *FFmpegStillConverter) ConvertToJPEG(sourcePath string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), converter.timeout)
	defer cancel()

	jpegBytes, err := converter.runCommand(ctx, "ffmpeg", ffmpegStillArguments(sourcePath)...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStillConversionUnavailable, err)
	}
	if len(jpegBytes) == 0 {
		return nil, fmt.Errorf("%w: ffmpeg produced no output", ErrStillConversionUnavailable)
	}
	return jpegBytes, nil
}

func ffmpegStillArguments(sourcePath string) []string {
	return []string{
		"-hide_banner",
		"-loglevel", "error",
		"-i", sourcePath,
		"-frames:v", "1",
		"-pix_fmt", "yuvj420p",
		"-q:v", "2",
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"pipe:1",
	}
}

func runCommandCapturingStdout(ctx context.Context, name string, args ...string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	command := exec.CommandContext(ctx, name, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, bytes.TrimSpace(stderr.Bytes()))
	}
	return stdout.Bytes(), nil
}
