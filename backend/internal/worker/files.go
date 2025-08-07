package worker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"time"
)

type failCallback func(err error) error

func ScanFilesWorker(service files.ServiceInterface, Logger logger.LoggerServiceInterface) {
	i18n.LogTranslate("SCAN_FILES_START")
	logger, _ := Logger.CreateLog(logger.LoggerModel{
		Name:        "ScanFilesWorker",
		Description: i18n.GetMessage("SCAN_FILES_START"),
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
	}, nil)

	var successFilesCount = 0
	var failedFilesCount = 0

	fail := func(path string, err error) error {
		Logger.CompleteWithErrorLog(logger, err)
		msg := i18n.GetMessage("ERROR_GET_FILE")
		log.Printf(msg, path, err)
		return nil
	}

	err := filepath.Walk(config.AppConfig.EntryPoint, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				i18n.LogTranslate("ERROR_PERMISSION_DENIED", path)
				failedFilesCount++
				return nil
			}
			fail(path, err)
		}
		name := info.Name()
		fileDto, fileDtoError := service.GetFileByNameAndPath(name, path)

		if fileDtoError != nil {
			if !errors.Is(fileDtoError, sql.ErrNoRows) {
				failedFilesCount++
				return fail(path, err)
			}
			i18n.LogTranslate("FILE_NOT_FOUND_IN_DATABASE", path)
		}

		if err := fileDto.ParseFileInfoToFileDto(info); err != nil {
			failedFilesCount++
			return fail(path, err)
		}

		if fileDtoError == nil {
			err = updateFileDto(service, fileDto, func(err error) error {
				return fail(path, err)
			})
		} else {
			fileDto, err = createFileDto(service, path, fileDto, func(err error) error {
				return fail(path, err)
			})
		}

		if err != nil {
			failedFilesCount++
			return err
		}

		go func() {
			err = service.UpdateCheckSum(fileDto.ID)
			if err != nil {
				fail(path, err)
			}
			fmt.Println("checksum atualizado com sucesso", fileDto.ID)
		}()
		successFilesCount++
		return nil
	})

	if err != nil {
		i18n.LogTranslate("ERROR_SCAN_FILES", err)
	} else {
		i18n.LogTranslate("SCAN_FILES_SUCCESS")
	}

	deletedFilesCount := findFilesDeleted(service)
	Logger.CompleteWithSuccessLog(logger)
	log.Printf(
		"%d Arquivos processados com sucesso. %d Arquivos falharam no processamento. %d Arquivos deletados",
		successFilesCount,
		failedFilesCount,
		deletedFilesCount,
	)
}

func updateFileDto(service files.ServiceInterface, fileDto files.FileDto, failCallback failCallback) error {
	fileDto.DeletedAt = utils.Optional[time.Time]{
		HasValue: false,
	}
	updated, err := service.UpdateFile(fileDto)
	if err != nil || !updated {
		return failCallback(err)
	}
	i18n.LogTranslate("FILE_UPDATE_SUCCESS", fileDto.ID)

	return nil
}

func createFileDto(service files.ServiceInterface, path string, fileDto files.FileDto, failCallback failCallback) (files.FileDto, error) {
	fileDto.Path = path
	fileDto.ParentPath = filepath.Dir(path)

	fileCreated, err := service.CreateFile(fileDto)

	if err != nil {
		return fileCreated, failCallback(err)
	}
	i18n.LogTranslate("FILE_CREATE_SUCCESS", fileCreated.ID)
	return fileCreated, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func findFilesDeleted(service files.ServiceInterface) int {
	var deletedFilesCount = 0
	var currentPage = 1
	var pagination, error = service.GetFiles(files.FileFilter{
		DeletedAt: utils.Optional[time.Time]{
			HasValue: true,
		},
	}, currentPage, 20)
	if error != nil {
		i18n.LogTranslate("ERROR_GET_FILES", error)
		return deletedFilesCount
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
				deletedFilesCount++
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
	return deletedFilesCount

}
