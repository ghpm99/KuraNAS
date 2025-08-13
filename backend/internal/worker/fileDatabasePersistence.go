package worker

import (
	"database/sql"
	"errors"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"sync"
)

func StartDatabasePersistenceWorker(
	service files.ServiceInterface,
	checksumCompletedChannel <-chan files.FileDto,
	monitorChannel chan<- ResultWorkerData,
	workerGroup *sync.WaitGroup,
) {
	defer workerGroup.Done()

	for finalizedFile := range checksumCompletedChannel {
		existingRecord, err := service.GetFileByNameAndPath(finalizedFile.Name, finalizedFile.Path)

		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				msg := i18n.GetMessage("ERROR_GET_FILE")
				log.Printf(msg, finalizedFile.Path, err)
				monitorChannel <- ResultWorkerData{
					Path:    finalizedFile.Path,
					Success: false,
					Error:   err.Error(),
				}
				continue
			}
		}

		found := err == nil
		if found {
			_, err = UpdateFileRecord(service, finalizedFile, existingRecord)
		} else {
			_, err = createFileRecord(service, finalizedFile)
		}
		if err != nil {
			log.Println("StartDatabasePersistenceWorker, falhou em processar", err)
			monitorChannel <- ResultWorkerData{
				Path:    finalizedFile.Path,
				Success: false,
				Error:   err.Error(),
			}
			continue
		}
		monitorChannel <- ResultWorkerData{
			Path:    finalizedFile.Path,
			Success: true,
			Error:   "",
		}
	}
}

func createFileRecord(service files.ServiceInterface, finalizedFile files.FileDto) (files.FileDto, error) {
	return service.CreateFile(finalizedFile)
}

func UpdateFileRecord(service files.ServiceInterface, finalizedFile files.FileDto, existingRecord files.FileDto) (bool, error) {
	existingRecord.Format = finalizedFile.Format
	existingRecord.Size = finalizedFile.Size
	existingRecord.UpdatedAt = finalizedFile.UpdatedAt
	existingRecord.CreatedAt = finalizedFile.CreatedAt
	existingRecord.DeletedAt = finalizedFile.DeletedAt
	existingRecord.LastInteraction = finalizedFile.LastInteraction
	existingRecord.LastBackup = finalizedFile.LastBackup
	existingRecord.CheckSum = finalizedFile.CheckSum
	existingRecord.DirectoryContentCount = finalizedFile.DirectoryContentCount
	existingRecord.Starred = finalizedFile.Starred
	existingRecord.Metadata = finalizedFile.Metadata

	return service.UpdateFile(existingRecord)
}
