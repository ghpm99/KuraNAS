package video

import (
	"bytes"
	"context"
	"os/exec"
)

const ffprobeExecutableName = "ffprobe"

type SubtitleProbeRunner func(ctx context.Context, arguments []string) ([]byte, error)

type SubtitleExtractRunner func(ctx context.Context, arguments []string) error

func RunFFprobeSubtitleProbe(ctx context.Context, arguments []string) ([]byte, error) {
	var probeOutput bytes.Buffer
	command := exec.CommandContext(ctx, ffprobeExecutableName, arguments...)
	command.Stdout = &probeOutput
	if err := command.Run(); err != nil {
		return nil, err
	}
	return probeOutput.Bytes(), nil
}

func RunFFmpegSubtitleExtract(ctx context.Context, arguments []string) error {
	return exec.CommandContext(ctx, ffmpegExecutableName, arguments...).Run()
}
