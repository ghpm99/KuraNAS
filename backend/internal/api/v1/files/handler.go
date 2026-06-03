package files

import "nas-go/api/pkg/logger"

type Handler struct {
	service           ServiceInterface
	recentFileService RecentFileServiceInterface
	Logger            logger.LoggerServiceInterface
}

func NewHandler(
	filesService ServiceInterface,
	recentFileService RecentFileServiceInterface,
	loggerService logger.LoggerServiceInterface,
) *Handler {
	return &Handler{
		service:           filesService,
		Logger:            loggerService,
		recentFileService: recentFileService,
	}
}
