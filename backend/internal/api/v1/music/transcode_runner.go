package music

import (
	"context"
	"io"
	"os/exec"
	"strconv"
)

const ffmpegExecutableName = "ffmpeg"

type TranscodeRunner func(ctx context.Context, arguments []string, output io.Writer) error

func RunFFmpegTranscode(ctx context.Context, arguments []string, output io.Writer) error {
	command := exec.CommandContext(ctx, ffmpegExecutableName, arguments...)
	command.Stdout = output
	return command.Run()
}

func IsFFmpegInstalled() bool {
	_, err := exec.LookPath(ffmpegExecutableName)
	return err == nil
}

func formatSeconds(seconds float64) string {
	return strconv.FormatFloat(seconds, 'f', 3, 64)
}
