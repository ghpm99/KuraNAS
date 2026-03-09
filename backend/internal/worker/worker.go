package worker

import (
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

type WorkerContext struct {
	Tasks           chan utils.Task
	FilesService    files.ServiceInterface
	VideoService    video.ServiceInterface
	JobsRepository  jobs.RepositoryInterface
	MetadataService files.MetadataRepositoryInterface
	Logger          logger.LoggerServiceInterface
	JobScheduler    *JobScheduler
	JobOrchestrator *JobOrchestrator
}

func StartWorkers(context *WorkerContext, numWorkers int) {
	if !config.AppConfig.EnableWorkers {
		return
	}

	if context != nil && context.JobsRepository != nil {
		context.JobScheduler = NewJobScheduler(context.JobsRepository, buildStepExecutors(context))
		context.JobOrchestrator = NewJobOrchestrator(context.JobsRepository, context.JobScheduler)
		context.JobScheduler.Start()
	}

	for i := range numWorkers {
		go worker(i, context)
	}

	go startWorkersScheduler(context)
	startEntryPointWatcher(context)
}

func startWorkersScheduler(context *WorkerContext) {
	if context != nil && context.JobOrchestrator != nil {
		if err := enqueueStartupScanJob(context); err != nil {
			log.Printf("failed to enqueue startup_scan job: %v\n", err)
		}
		return
	}

	log.Println("enqueuing file scan task")
	context.Tasks <- utils.Task{
		Type: utils.ScanFiles,
		Data: "file scan",
	}
	log.Println("file scan task enqueued")
}

func enqueueStartupScanJob(context *WorkerContext) error {
	rootPath := config.AppConfig.EntryPoint
	if rootPath == "" {
		return nil
	}

	jobID, err := context.JobOrchestrator.CreateJob(buildScanPlan(rootPath, JobTypeStartupScan, JobPriorityLow))
	if err != nil {
		return err
	}

	log.Printf("startup_scan job enqueued id=%d\n", jobID)
	return nil
}

func enqueueFilesystemEventJob(context *WorkerContext, rootPath string, priority JobPriority) error {
	if context == nil || context.JobOrchestrator == nil {
		return nil
	}
	if rootPath == "" {
		return nil
	}

	_, err := context.JobOrchestrator.CreateJob(buildScanPlan(rootPath, JobTypeFSEvent, priority))
	return err
}

func buildScanPlan(rootPath string, jobType JobType, priority JobPriority) PlannedJob {
	payload := mustMarshalPayload(StepFilePayload{Path: rootPath})
	return PlannedJob{
		Type:     jobType,
		Priority: priority,
		Scope:    JobScope{Root: rootPath},
		Steps: []PlannedStep{
			{
				Key:         "scan_filesystem",
				Type:        StepTypeScanFilesystem,
				MaxAttempts: 1,
				Payload:     payload,
			},
			{
				Key:         "diff_against_db",
				Type:        StepTypeDiffAgainstDB,
				DependsOn:   []string{"scan_filesystem"},
				MaxAttempts: 1,
				Payload:     payload,
			},
			{
				Key:         "mark_deleted",
				Type:        StepTypeMarkDeleted,
				DependsOn:   []string{"diff_against_db"},
				MaxAttempts: 1,
				Payload:     payload,
			},
		},
	}
}

func worker(id int, context *WorkerContext) {
	for task := range context.Tasks {
		log.Printf("worker %d: processing task %s\n", id, task.Data)

		switch task.Type {
		case utils.ScanFiles:
			if context != nil && context.JobOrchestrator != nil {
				go func() {
					if err := enqueueFilesystemEventJob(context, config.AppConfig.EntryPoint, JobPriorityLow); err != nil {
						log.Printf("worker %d: failed to enqueue fs_event job: %v\n", id, err)
					}
				}()
			} else {
				go StartFileProcessingPipeline(context.FilesService, context.Tasks, context.Logger)
			}
		case utils.ScanDir:
			if context != nil && context.JobOrchestrator != nil {
				targetPath, ok := task.Data.(string)
				if ok {
					go func() {
						if err := enqueueFilesystemEventJob(context, targetPath, JobPriorityNormal); err != nil {
							log.Printf("worker %d: failed to enqueue fs_event job for %s: %v\n", id, targetPath, err)
						}
					}()
				}
			} else {
				go ScanDirWorker(context.FilesService, task.Data)
			}
		case utils.UpdateCheckSum:
			go UpdateCheckSumWorker(context, task.Data)
		case utils.CreateThumbnail:
			go CreateThumbnailWorker(context.FilesService, task.Data, context.Logger)
		case utils.GenerateVideoPlaylists:
			go GenerateVideoPlaylistsWorker(context.VideoService, context.Logger)
		default:
			log.Printf("worker %d: unknown task type %v\n", id, task.Type)
		}
	}
}
