package worker

import (
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

func CreateThumbnailWorker(service files.ServiceInterface, data any, logService logger.LoggerServiceInterface) {
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
		if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
			log.Printf("CreateThumbnailWorker: erro ao gerar thumbnail de video fileID=%d: %v\n", fileID, err)
		}
		if _, err := service.GetVideoPreviewGif(fileDto, 320, 180); err != nil {
			log.Printf("CreateThumbnailWorker: erro ao gerar preview gif fileID=%d: %v\n", fileID, err)
		}
		return
	}

	if _, err := service.GetFileThumbnail(fileDto, 320, 320); err != nil {
		log.Printf("CreateThumbnailWorker: erro ao gerar thumbnail padrao fileID=%d: %v\n", fileID, err)
	}
}
