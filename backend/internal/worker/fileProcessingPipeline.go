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

type ResultWorkerData struct {
	Path    string
	Success bool
	Error   string
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
	monitorChannel := make(chan ResultWorkerData, 100)
	fileWalkChannel := make(chan FileWalk, 100)
	fileDtoChannel := make(chan files.FileDto, 100)
	metadataProcessedChannel := make(chan files.FileDto, 100)
	checksumCompletedChannel := make(chan files.FileDto, 100)

	log.Println("criando worker group")

	var monitorWG sync.WaitGroup
	monitorWG.Add(1)
	go StartResultMonitorWorker(monitorChannel, &monitorWG)

	var walkerWG sync.WaitGroup
	walkerWG.Add(1)
	go func() {
		StartDirectoryWalker(config.AppConfig.EntryPoint, fileWalkChannel, monitorChannel, &walkerWG)
		walkerWG.Wait()
		close(fileWalkChannel)
	}()

	var dtoWG sync.WaitGroup
	for range 5 {
		dtoWG.Add(1)
		go StartDtoConverterWorker(fileWalkChannel, fileDtoChannel, &dtoWG)
	}
	go func() {
		dtoWG.Wait()
		close(fileDtoChannel)
	}()

	var metaWG sync.WaitGroup
	for range 3 {
		metaWG.Add(1)
		go StartMetadataWorker(fileDtoChannel, metadataProcessedChannel, pythonScriptRunner, monitorChannel, &metaWG)
	}
	go func() {
		metaWG.Wait()
		close(metadataProcessedChannel)
	}()

	var checksumWG sync.WaitGroup
	for range 5 {
		checksumWG.Add(1)
		go StartChecksumWorker(
			metadataProcessedChannel,
			checksumCompletedChannel,
			utils.GetFileChecksum,
			utils.GetDirectoryChecksum,
			monitorChannel,
			&checksumWG,
		)
	}
	go func() {
		checksumWG.Wait()
		close(checksumCompletedChannel)
	}()

	var dbWG sync.WaitGroup
	dbWG.Add(1)
	go StartDatabasePersistenceWorker(service, checksumCompletedChannel, monitorChannel, &dbWG)

	log.Println("Esperando processamento concluir")
	dbWG.Wait()

	log.Println("Pipeline de processamento concluído.")
	Logger.CompleteWithSuccessLog(logger)

}
