package worker

import (
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"os"
	"sync"
)

type FileWalk struct {
	Path string
	Info os.FileInfo
}

var pythonScriptRunner = func(scriptType utils.ScriptType, filePath string) (string, error) {
	return utils.RunPythonScript(scriptType, filePath)
}

func StartFileProcessingPipeline(service files.ServiceInterface, Logger logger.LoggerServiceInterface) {

	log.Println("Iniciando o pipeline de processamento de arquivos...")
	logger, _ := Logger.CreateLog(logger.LoggerModel{
		Name:        "StartFileProcessingPipeline",
		Description: i18n.GetMessage("SCAN_FILES_START"),
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
	}, nil)

	log.Println("criando canais")
	fileWalkChannel := make(chan FileWalk, 100)
	fileDtoChannel := make(chan files.FileDto, 100)
	metadataProcessedChannel := make(chan files.FileDto, 100)
	checksumCompletedChannel := make(chan files.FileDto, 100)

	log.Println("criando worker group")
	var workerGroup sync.WaitGroup

	workerGroup.Add(1)
	go StartDirectoryWalker(config.AppConfig.EntryPoint, fileWalkChannel, &workerGroup)

	for range 5 {
		workerGroup.Add(1)
		go StartDtoConverterWorker(fileWalkChannel, fileDtoChannel, &workerGroup)
	}

	for range 5 {
		workerGroup.Add(1)
		go StartChecksumWorker(
			metadataProcessedChannel,
			checksumCompletedChannel,
			utils.GetFileChecksum,
			utils.GetDirectoryChecksum,
			&workerGroup,
		)
	}

	for range 3 {
		workerGroup.Add(1)
		go StartMetadataWorker(fileDtoChannel, metadataProcessedChannel, pythonScriptRunner, &workerGroup)
	}

	workerGroup.Add(1)
	go StartDatabasePersistenceWorker(service, checksumCompletedChannel, &workerGroup)

	log.Println("Esperando processamento concluir")
	workerGroup.Wait()

	log.Println("Pipeline de processamento concluído.")
	Logger.CompleteWithSuccessLog(logger)

}
