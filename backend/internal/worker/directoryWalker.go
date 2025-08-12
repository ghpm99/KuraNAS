package worker

import (
	"errors"
	"log"
	"nas-go/api/pkg/i18n"
	"os"
	"path/filepath"
	"sync"
)

func StartDirectoryWalker(targetDirectory string, fileWalkChannel chan<- FileWalk, workerGroup *sync.WaitGroup) {
	defer workerGroup.Done()

	walkCallback := func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				i18n.LogTranslate("ERROR_PERMISSION_DENIED", filePath)

				return nil
			}
			msg := i18n.GetMessage("ERROR_GET_FILE")
			log.Printf(msg, filePath, err)
		}
		log.Println("Enviando arquivo para canal", filePath)
		fileWalkChannel <- FileWalk{
			Path: filePath,
			Info: fileInfo,
		}
		return nil
	}

	if err := filepath.Walk(targetDirectory, walkCallback); err != nil {
		log.Printf("Erro na exploração do diretório: %v\n", err)
	}
}
