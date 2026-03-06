package worker

import (
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/pkg/logger"
)

type PlaylistIndexStepInput struct{}

type PlaylistIndexStepOutput struct {
	Skipped bool
}

type PlaylistIndexStepExecutor struct {
	service video.ServiceInterface
}

func NewPlaylistIndexStepExecutor(service video.ServiceInterface) *PlaylistIndexStepExecutor {
	return &PlaylistIndexStepExecutor{service: service}
}

func (e *PlaylistIndexStepExecutor) Execute(input PlaylistIndexStepInput) (PlaylistIndexStepOutput, error) {
	_ = input

	if e == nil || e.service == nil {
		return PlaylistIndexStepOutput{}, fmt.Errorf("playlist_index step: video service is required")
	}

	if err := e.service.RebuildSmartPlaylists(); err != nil {
		return PlaylistIndexStepOutput{}, err
	}

	return PlaylistIndexStepOutput{}, nil
}

func GenerateVideoPlaylistsWorker(service video.ServiceInterface, logService logger.LoggerServiceInterface) {
	_ = logService

	executor := NewPlaylistIndexStepExecutor(service)
	_, err := executor.Execute(PlaylistIndexStepInput{})
	if err != nil {
		log.Printf("GenerateVideoPlaylistsWorker: erro ao gerar playlists inteligentes: %v\n", err)
		return
	}

	log.Println("GenerateVideoPlaylistsWorker: playlists inteligentes atualizadas")
}
