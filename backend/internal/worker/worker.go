package worker

import (
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker/domain"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"time"
)

type WorkerContext struct {
	Tasks           chan utils.Task
	FilesService    files.ServiceInterface
	VideoService    video.ServiceInterface
	MetadataService files.MetadataRepositoryInterface
	JobsRepository  jobs.RepositoryInterface
	StepExecutor    StepAtomicExecutor
	Orchestrator    *JobOrchestrator
	Scheduler       *JobScheduler
	Logger          logger.LoggerServiceInterface
}

func StartWorkers(context *WorkerContext, numWorkers int) {
	if !config.AppConfig.EnableWorkers {
		return
	}

	configureJobScheduler(context)

	for i := range numWorkers {
		go worker(i, context)
	}

	go startWorkersScheduler(context)
	startEntryPointWatcher(context)
}

func configureJobScheduler(context *WorkerContext) {
	if context == nil || context.JobsRepository == nil {
		return
	}

	if context.StepExecutor == nil {
		context.StepExecutor = NewDefaultStepExecutor()
	}

	context.Orchestrator = NewJobOrchestrator(context.JobsRepository, NewDefaultJobPlanner())
	context.Scheduler = NewJobScheduler(context.JobsRepository, context.StepExecutor, context)

	go context.Scheduler.Start(2 * time.Second)
}

func startWorkersScheduler(context *WorkerContext) {
	if context != nil && context.Orchestrator != nil {
		_, err := context.Orchestrator.CreateJob(
			domain.JobTypeStartupScan,
			domain.JobPriorityNormal,
			domain.NewRootScopePayload(config.AppConfig.EntryPoint),
		)
		if err == nil {
			log.Println("startup_scan job criado para scheduler")
			return
		}

		log.Printf("erro ao criar startup_scan job, fallback para fila legada: %v\n", err)
	}

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
			go StartFileProcessingPipeline(context.FilesService, context.Tasks, context.Logger)
		case utils.ScanDir:
			go ScanDirWorker(context.FilesService, task.Data)
		case utils.UpdateCheckSum:
			go UpdateCheckSumWorker(context.FilesService, task.Data, context.Logger)
		case utils.CreateThumbnail:
			go CreateThumbnailWorker(context.FilesService, task.Data, context.Logger)
		case utils.GenerateVideoPlaylists:
			go GenerateVideoPlaylistsWorker(context.VideoService, context.Logger)
		default:
			log.Println("Tipo de tarefa desconhecido")
		}
		log.Printf("Worker %d: Tarefa %s completa\n", id, task.Data)
	}
}
