package video

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
)

const thumbnailFrameStrategyVersion = "f2"

type FrameCommandRunner func(ctx context.Context, arguments []string) error

func RunFFmpegFrameCommand(ctx context.Context, arguments []string) error {
	return exec.CommandContext(ctx, ffmpegExecutableName, arguments...).Run()
}

func (s *Service) runFrameCommand(ctx context.Context, arguments []string) error {
	if s.FrameCommandRunner != nil {
		return s.FrameCommandRunner(ctx, arguments)
	}
	return RunFFmpegFrameCommand(ctx, arguments)
}

func (s *Service) loadVideoDuration(fileID int) sql.NullFloat64 {
	duration, err := s.Repository.GetVideoDurationSecondsByFileID(fileID)
	if err != nil {
		return sql.NullFloat64{}
	}
	return duration
}

func (s *Service) renderFrameWithRetryAtStart(buildArguments func(offsetSeconds float64) []string, offsetSeconds float64, outputPath string) ([]byte, bool) {
	attemptOffsets := []float64{offsetSeconds}
	if offsetSeconds > 0 {
		attemptOffsets = append(attemptOffsets, 0)
	}
	for _, attemptOffset := range attemptOffsets {
		if renderedFrame, isRendered := s.renderFrameAt(buildArguments(attemptOffset), outputPath); isRendered {
			return renderedFrame, true
		}
	}
	return nil, false
}

func (s *Service) renderFrameAt(arguments []string, outputPath string) ([]byte, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), ffmpegTimeout())
	defer cancel()
	if err := s.runFrameCommand(ctx, arguments); err != nil {
		return nil, false
	}
	renderedFrame, err := os.ReadFile(outputPath)
	return renderedFrame, err == nil
}
