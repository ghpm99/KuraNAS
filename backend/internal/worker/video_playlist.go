package worker

import (
	"log"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/pkg/logger"
)

func GenerateVideoPlaylistsWorker(service video.ServiceInterface, logService logger.LoggerServiceInterface) {
	if service == nil {
		log.Println("GenerateVideoPlaylistsWorker: video service nulo")
		return
	}

	if err := service.RebuildSmartPlaylists(); err != nil {
		log.Printf("GenerateVideoPlaylistsWorker: erro ao gerar playlists inteligentes: %v\n", err)
		return
	}

	log.Println("GenerateVideoPlaylistsWorker: playlists inteligentes atualizadas")
}
