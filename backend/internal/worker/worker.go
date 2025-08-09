package worker

import (
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

type WorkerContext struct {
	Tasks           chan utils.Task
	FilesService    files.ServiceInterface
	MetadataService files.MetadataRepositoryInterface
	Logger          logger.LoggerServiceInterface
}

func StartWorkers(context *WorkerContext, numWorkers int) {
	if !config.AppConfig.EnableWorkers {
		return
	}
	for i := range numWorkers {
		go worker(i, context)
	}

	go startWorkersScheduler(context)
}

func startWorkersScheduler(context *WorkerContext) {

	log.Println("Escaneamento de arquivos")
	context.Tasks <- utils.Task{
		Type: utils.ScanFiles,
		Data: "Escaneamento de arquivos",
	}
	log.Println("📁 Tarefa de escaneamento de arquivos enviada para a fila")

}

func worker(id int, context *WorkerContext) {
	for task := range context.Tasks {
		log.Printf("Worker %d: Processando tarefa %s\n", id, task.Data)

		switch task.Type {
		case utils.ScanFiles:
			go StartFileProcessingPipeline(context.FilesService, context.Logger)
		case utils.ScanDir:
			go ScanDirWorker(context.FilesService, task.Data)
		case utils.UpdateCheckSum:
			go UpdateCheckSumWorker(context.FilesService, task.Data, context.Logger)
		case utils.CreateThumbnail:
			go CreateThumbnailWorker(context.FilesService, task.Data, context.Logger)
		default:
			log.Println("Tipo de tarefa desconhecido")
		}
		log.Printf("Worker %d: Tarefa %s completa\n", id, task.Data)
	}
}
