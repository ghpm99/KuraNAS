package worker

import (
	"database/sql"
	"errors"
	"fmt"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"time"
)

func ScanFilesWorker(service files.ServiceInterface, Logger logger.LoggerServiceInterface) {
	i18n.LogTranslate("SCAN_FILES_START")
	logger, _ := Logger.CreateLog(logger.LoggerModel{
		Name:        "ScanFilesWorker",
		Description: i18n.GetMessage("SCAN_FILES_START"),
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
	}, nil)

	fail := func(path string, err error) error {
		Logger.CompleteWithErrorLog(logger, err)
		msg := i18n.GetMessage("ERROR_GET_FILE")
		return fmt.Errorf(msg, path, err)
	}

	err := filepath.Walk(config.AppConfig.EntryPoint, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				i18n.LogTranslate("ERROR_PERMISSION_DENIED", path)
				return nil
			}
			return fail(path, err)
		}
		name := info.Name()
		fileDto, fileDtoError := service.GetFileByNameAndPath(name, path)

		if fileDtoError != nil {
			if !errors.Is(fileDtoError, sql.ErrNoRows) {
				return fail(path, err)
			}
			i18n.LogTranslate("FILE_NOT_FOUND_IN_DATABASE", path)
		}

		if err := fileDto.ParseFileInfoToFileDto(info); err != nil {
			return fail(path, err)
		}

		if fileDtoError == nil {
			fileDto.DeletedAt = utils.Optional[time.Time]{
				HasValue: false,
			}
			updated, err := service.UpdateFile(fileDto)
			if err != nil || !updated {
				return fail(path, err)
			}
			i18n.PrintTranslate("FILE_UPDATE_SUCCESS", fileDto.ID)
			service.UpdateCheckSumTask(fileDto.ID)
			return nil
		} else {
			fileDto.Path = path
			fileDto.ParentPath = filepath.Dir(path)
		}

		fileCreated, err := service.CreateFile(fileDto)

		if err != nil {
			return fail(path, err)
		}
		i18n.PrintTranslate("FILE_CREATE_SUCCESS", fileCreated.ID)
		service.UpdateCheckSumTask(fileCreated.ID)
		return nil
	})

	if err != nil {
		i18n.LogTranslate("ERROR_SCAN_FILES", err)
	} else {
		i18n.PrintTranslate("SCAN_FILES_SUCCESS")
	}

	findFilesDeleted(service)
	Logger.CompleteWithSuccessLog(logger)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func findFilesDeleted(service files.ServiceInterface) {
	var currentPage = 1
	var pagination, error = service.GetFiles(files.FileFilter{
		DeletedAt: utils.Optional[time.Time]{
			HasValue: true,
		},
	}, currentPage, 20)
	if error != nil {
		i18n.PrintTranslate("ERROR_GET_FILES", error)
		return
	}
	for {
		for _, file := range pagination.Items {
			if !fileExists(file.Path) {
				i18n.LogTranslate("FILE_DONT_EXIST", file.ID, file.Name)
				file.DeletedAt = utils.Optional[time.Time]{
					HasValue: true,
					Value:    time.Now(),
				}
				_, error := service.UpdateFile(file)
				if error != nil {
					i18n.LogTranslate("ERROR_DELETING_FILE", file.ID, file.Name)
					continue
				}
			} else {
				continue
			}
		}
		if !pagination.Pagination.HasNext {
			break
		}
		currentPage++
		pagination, error = service.GetFiles(files.FileFilter{}, currentPage, 20)
		if error != nil {
			i18n.LogTranslate("ERROR_GET_FILES", error)
			break
		}
	}

}
