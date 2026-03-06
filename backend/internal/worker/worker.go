package worker

import (
	stdcontext "context"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker/domain"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"strings"
	"sync"
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

	jobExecutionMutex    sync.Mutex
	jobExecutionContexts map[string]stdcontext.Context
	jobExecutionCancel   map[string]stdcontext.CancelFunc
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

	pollInterval := time.Duration(config.AppConfig.WorkerSchedulerPollIntervalSecond) * time.Second
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	go context.Scheduler.Start(pollInterval)
}

func startWorkersScheduler(context *WorkerContext) {
	if context != nil && context.Orchestrator != nil {
		job, err := context.Orchestrator.CreateJob(
			domain.JobTypeStartupScan,
			domain.JobPriorityLow,
			domain.NewRootScopePayload(config.AppConfig.EntryPoint),
		)
		if err == nil {
			log.Printf("startup_scan job criado para scheduler job_id=%s", job.ID)
			return
		}

		log.Printf("erro ao criar startup_scan job: %v\n", err)
		return
	}
	log.Println("job orchestrator indisponivel; startup_scan nao foi enfileirado")
}

func worker(id int, context *WorkerContext) {
	for task := range context.Tasks {
		log.Printf("Worker %d: Processando tarefa %s\n", id, task.Data)

		switch task.Type {
		case utils.ScanFiles:
			enqueueStartupScanJob(context)
		case utils.ScanDir:
			enqueueReindexFolderJob(context, task.Data)
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

func enqueueStartupScanJob(context *WorkerContext) {
	if context == nil || context.Orchestrator == nil {
		log.Println("scan_files task ignorada: job orchestrator indisponivel")
		return
	}

	job, err := context.Orchestrator.CreateJob(
		domain.JobTypeStartupScan,
		domain.JobPriorityLow,
		domain.NewRootScopePayload(config.AppConfig.EntryPoint),
	)
	if err != nil {
		log.Printf("erro ao enfileirar startup_scan via task scan_files: %v", err)
		return
	}

	log.Printf("scan_files convertida para startup_scan job_id=%s", job.ID)
}

func enqueueReindexFolderJob(context *WorkerContext, data any) {
	if context == nil || context.Orchestrator == nil {
		log.Println("scan_dir task ignorada: job orchestrator indisponivel")
		return
	}

	scopePath, ok := data.(string)
	scopePath = strings.TrimSpace(scopePath)
	if !ok || scopePath == "" {
		log.Printf("scan_dir task ignorada: path invalido (%T)", data)
		return
	}

	job, err := context.Orchestrator.CreateJob(
		domain.JobTypeReindexFolder,
		domain.JobPriorityNormal,
		domain.NewPathScopePayload(scopePath),
	)
	if err != nil {
		log.Printf("erro ao enfileirar reindex_folder via task scan_dir path=%s: %v", scopePath, err)
		return
	}

	log.Printf("scan_dir convertida para reindex_folder job_id=%s path=%s", job.ID, scopePath)
}

func (workerContext *WorkerContext) EnsureJobExecutionContext(jobID string) stdcontext.Context {
	if workerContext == nil || jobID == "" {
		return contextBackground()
	}

	workerContext.jobExecutionMutex.Lock()
	defer workerContext.jobExecutionMutex.Unlock()

	if workerContext.jobExecutionContexts == nil {
		workerContext.jobExecutionContexts = map[string]stdcontext.Context{}
	}
	if workerContext.jobExecutionCancel == nil {
		workerContext.jobExecutionCancel = map[string]stdcontext.CancelFunc{}
	}

	existing, exists := workerContext.jobExecutionContexts[jobID]
	if exists {
		return existing
	}

	jobContext, cancel := stdcontext.WithCancel(contextBackground())
	workerContext.jobExecutionContexts[jobID] = jobContext
	workerContext.jobExecutionCancel[jobID] = cancel
	return jobContext
}

func (workerContext *WorkerContext) CancelJobExecution(jobID string) {
	if workerContext == nil || jobID == "" {
		return
	}

	workerContext.jobExecutionMutex.Lock()
	defer workerContext.jobExecutionMutex.Unlock()

	if cancel, exists := workerContext.jobExecutionCancel[jobID]; exists {
		cancel()
	}
}

func (workerContext *WorkerContext) ReleaseJobExecution(jobID string) {
	if workerContext == nil || jobID == "" {
		return
	}

	workerContext.jobExecutionMutex.Lock()
	defer workerContext.jobExecutionMutex.Unlock()

	if cancel, exists := workerContext.jobExecutionCancel[jobID]; exists {
		cancel()
		delete(workerContext.jobExecutionCancel, jobID)
	}
	delete(workerContext.jobExecutionContexts, jobID)
}

func (workerContext *WorkerContext) CheckJobCancellation(jobID string) error {
	if workerContext == nil || jobID == "" {
		return nil
	}

	jobContext := workerContext.EnsureJobExecutionContext(jobID)
	select {
	case <-jobContext.Done():
		return jobContext.Err()
	default:
		return nil
	}
}

func contextBackground() stdcontext.Context {
	return stdcontext.Background()
}
