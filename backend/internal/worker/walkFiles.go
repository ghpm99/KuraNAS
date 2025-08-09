package worker

import (
	"errors"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"os"
	"path/filepath"
)

func WalkFilesWorker(service files.ServiceInterface, Logger logger.LoggerServiceInterface, fileWalkChannel chan FileWalk) {
	i18n.LogTranslate("SCAN_FILES_START")
	logger, _ := Logger.CreateLog(logger.LoggerModel{
		Name:        "WalkFilesWorker",
		Description: i18n.GetMessage("SCAN_FILES_START"),
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
	}, nil)

	err := filepath.Walk(config.AppConfig.EntryPoint, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				i18n.LogTranslate("ERROR_PERMISSION_DENIED", path)

				return nil
			}
			msg := i18n.GetMessage("ERROR_GET_FILE")
			log.Printf(msg, path, err)
		}

		fileWalkChannel <- FileWalk{
			path: path,
			info: info,
		}

		return nil
	})
	if err != nil {
		i18n.LogTranslate("ERROR_SCAN_FILES", err)
	} else {
		i18n.LogTranslate("SCAN_FILES_SUCCESS")
	}

	Logger.CompleteWithSuccessLog(logger)
}
