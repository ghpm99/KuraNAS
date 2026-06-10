package scan

import (
	"log"
	"nas-go/api/internal/api/v1/files"
	videodom "nas-go/api/internal/api/v1/video"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

func CreateThumbnailWorker(service files.ServiceInterface, videoService videodom.ServiceInterface, data any, logService logger.LoggerServiceInterface) {
	fileID, ok := data.(int)
	if !ok || fileID <= 0 {
		log.Println("CreateThumbnailWorker: data inválido, esperado fileID int")
		return
	}

	fileDto, err := service.GetFileById(fileID)
	if err != nil {
		log.Printf("CreateThumbnailWorker: erro ao carregar arquivo %d: %v\n", fileID, err)
		return
	}

	if fileDto.Type != files.File {
		return
	}

	formatType := utils.GetFormatTypeByExtension(fileDto.Format)
	if formatType.Type == utils.FormatTypeVideo {
		if videoService == nil {
			log.Printf("CreateThumbnailWorker: video service indisponível, pulando thumbnail de video fileID=%d\n", fileID)
			return
		}
		if _, err := videoService.GetVideoThumbnail(fileDto, 320, 180); err != nil {
			log.Printf("CreateThumbnailWorker: erro ao gerar thumbnail de video fileID=%d: %v\n", fileID, err)
		}
		if _, err := videoService.GetVideoPreviewGif(fileDto, 320, 180); err != nil {
			log.Printf("CreateThumbnailWorker: erro ao gerar preview gif fileID=%d: %v\n", fileID, err)
		}
		return
	}

	if _, err := service.GetFileThumbnail(fileDto, 320, 320); err != nil {
		log.Printf("CreateThumbnailWorker: erro ao gerar thumbnail padrao fileID=%d: %v\n", fileID, err)
	}
}
