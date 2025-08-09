package worker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"sync"
)

func StartDatabasePersistenceWorker(service files.ServiceInterface, checksumCompletedChannel <-chan files.FileDto, workerGroup *sync.WaitGroup) {
	defer workerGroup.Done()

	for finalizedFile := range checksumCompletedChannel {
		log.Println("StartDatabasePersistenceWorker, Recendo arquivo de fila", finalizedFile.Path)
		existingRecord, err := service.GetFileByNameAndPath(finalizedFile.Name, finalizedFile.Path)

		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				msg := i18n.GetMessage("ERROR_GET_FILE")
				log.Printf(msg, finalizedFile.Path, err)
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
			continue
		}
		log.Println("StartDatabasePersistenceWorker, processamento concluido, existia:", found, finalizedFile.Path)
	}
	fmt.Println("Fechando o canal de checksum.")
}

func createFileRecord(service files.ServiceInterface, finalizedFile files.FileDto) (files.FileDto, error) {
	return service.CreateFile(finalizedFile)
}

func UpdateFileRecord(service files.ServiceInterface, finalizedFile files.FileDto, existingRecord files.FileDto) (bool, error) {
	existingRecord.Name = finalizedFile.Name

	return service.UpdateFile(existingRecord)
}
