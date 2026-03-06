package worker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"
	"reflect"
	"sync"
	"time"
)

type PersistStepInput struct {
	File files.FileDto
}

type PersistStepOutput struct {
	File    files.FileDto
	Skipped bool
	Created bool
	Updated bool
}

type PersistStepExecutor struct {
	service files.ServiceInterface
}

func NewPersistStepExecutor(service files.ServiceInterface) *PersistStepExecutor {
	return &PersistStepExecutor{service: service}
}

func (e *PersistStepExecutor) Execute(input PersistStepInput) (PersistStepOutput, error) {
	if e == nil || e.service == nil {
		return PersistStepOutput{}, fmt.Errorf("persist step: file service is required")
	}

	finalizedFile := input.File
	existingRecord, err := e.service.GetFileByNameAndPath(finalizedFile.Name, finalizedFile.Path)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return PersistStepOutput{File: finalizedFile}, err
	}

	found := err == nil
	if found {
		if isPersistUpToDate(existingRecord, finalizedFile) {
			finalizedFile.ID = existingRecord.ID
			return PersistStepOutput{File: finalizedFile, Skipped: true}, newStepSkipped("persist up-to-date")
		}

		_, updateErr := UpdateFileRecord(e.service, finalizedFile, existingRecord)
		if updateErr != nil {
			return PersistStepOutput{File: finalizedFile}, updateErr
		}
		finalizedFile.ID = existingRecord.ID
		return PersistStepOutput{File: finalizedFile, Updated: true}, nil
	}

	createdFile, createErr := createFileRecord(e.service, finalizedFile)
	if createErr != nil {
		return PersistStepOutput{File: finalizedFile}, createErr
	}

	finalizedFile.ID = createdFile.ID
	return PersistStepOutput{File: finalizedFile, Created: true}, nil
}

func isPersistUpToDate(existingRecord files.FileDto, finalizedFile files.FileDto) bool {
	return existingRecord.Format == finalizedFile.Format &&
		existingRecord.Size == finalizedFile.Size &&
		existingRecord.UpdatedAt.Equal(finalizedFile.UpdatedAt) &&
		existingRecord.CreatedAt.Equal(finalizedFile.CreatedAt) &&
		existingRecord.CheckSum == finalizedFile.CheckSum &&
		existingRecord.DirectoryContentCount == finalizedFile.DirectoryContentCount &&
		existingRecord.Starred == finalizedFile.Starred &&
		optionalTimeEqual(existingRecord.DeletedAt, finalizedFile.DeletedAt) &&
		optionalTimeEqual(existingRecord.LastInteraction, finalizedFile.LastInteraction) &&
		optionalTimeEqual(existingRecord.LastBackup, finalizedFile.LastBackup) &&
		reflect.DeepEqual(existingRecord.Metadata, finalizedFile.Metadata)
}

func optionalTimeEqual(a utils.Optional[time.Time], b utils.Optional[time.Time]) bool {
	if a.HasValue != b.HasValue {
		return false
	}
	if !a.HasValue {
		return true
	}
	return a.Value.Equal(b.Value)
}

func StartDatabasePersistenceWorker(
	service files.ServiceInterface,
	tasks chan utils.Task,
	checksumCompletedChannel <-chan files.FileDto,
	monitorChannel chan<- ResultWorkerData,
	workerGroup *sync.WaitGroup,
) {
	defer workerGroup.Done()
	executor := NewPersistStepExecutor(service)

	for finalizedFile := range checksumCompletedChannel {
		output, err := executor.Execute(PersistStepInput{File: finalizedFile})
		if err != nil {
			if !isStepSkipped(err) {
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

		persistedFileID := output.File.ID

		enqueueVideoThumbnailTask(tasks, output.File, persistedFileID)

		monitorChannel <- ResultWorkerData{
			Path:    finalizedFile.Path,
			Success: true,
			Error:   "",
		}
	}
}

func enqueueVideoThumbnailTask(tasks chan utils.Task, finalizedFile files.FileDto, fileID int) {
	if tasks == nil || fileID <= 0 || finalizedFile.Type != files.File {
		return
	}

	formatType := utils.GetFormatTypeByExtension(finalizedFile.Format)
	if formatType.Type != utils.FormatTypeVideo {
		return
	}

	task := utils.Task{
		Type: utils.CreateThumbnail,
		Data: fileID,
	}

	select {
	case tasks <- task:
	default:
		log.Printf("fila de tasks cheia, thumbnail de video ignorada para fileID=%d", fileID)
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
