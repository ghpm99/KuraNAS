package worker

import (
	"log"
	"nas-go/api/internal/api/v1/files"
	"path/filepath"
	"sync"
)

func convertToDto(fw FileWalk) files.FileDto {
	fileDto := files.FileDto{
		Path:       fw.Path,
		ParentPath: filepath.Dir(fw.Path),
	}
	fileDto.ParseFileInfoToFileDto(fw.Info)
	return fileDto
}

func StartDtoConverterWorker(fileWalkChannel <-chan FileWalk, fileDtoChannel chan<- files.FileDto, workerGroup *sync.WaitGroup) {
	defer workerGroup.Done()

	for fileWalkItem := range fileWalkChannel {
		log.Println("StartDtoConverterWorker, Recendo arquivo de fila", fileWalkItem.Path)
		fileDto := convertToDto(fileWalkItem)
		log.Println("StartDtoConverterWorker, Enviando arquivo para fila")
		fileDtoChannel <- fileDto
	}
}
