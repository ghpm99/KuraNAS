package worker

import (
	"fmt"
	"log"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

type WorkerContext struct {
	Tasks               chan utils.Task
	FilesService        files.ServiceInterface
	LibrariesService    libraries.ServiceInterface
	VideoService        video.ServiceInterface
	JobsRepository      jobs.RepositoryInterface
	MetadataService     files.MetadataRepositoryInterface
	Logger              logger.LoggerServiceInterface
	NotificationService notifications.ServiceInterface
	AIService           ai.ServiceInterface
	JobScheduler        *JobScheduler
	JobOrchestrator     *JobOrchestrator
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
	go startNotificationCleanup(context)
	startEntryPointWatcher(context)
}

func startNotificationCleanup(context *WorkerContext) {
	if context == nil || context.NotificationService == nil {
		return
	}
	for {
		time.Sleep(1 * time.Hour)
		if err := context.NotificationService.CleanupOldNotifications(); err != nil {
			log.Printf("notification cleanup error: %v\n", err)
		}
	}
}

func emitNotification(context *WorkerContext, notifType string, title string, message string, groupKey string) {
	if context == nil || context.NotificationService == nil {
		return
	}
	_, err := context.NotificationService.GroupOrCreate(notifications.CreateNotificationDto{
		Type:     notifType,
		Title:    title,
		Message:  message,
		GroupKey: groupKey,
	})
	if err != nil {
		log.Printf("failed to emit notification: %v\n", err)
	}
}

func startWorkersScheduler(context *WorkerContext) {
	if context != nil && context.JobOrchestrator != nil {
		if err := enqueueStartupScanJob(context); err != nil {
			log.Printf("failed to enqueue startup_scan job: %v\n", err)
			emitNotification(
				context,
				"error",
				i18n.GetMessage("NOTIFICATION_STARTUP_SCAN_FAILED_TITLE"),
				err.Error(),
				"",
			)
		} else {
			emitNotification(
				context,
				"info",
				i18n.GetMessage("NOTIFICATION_FILE_SCAN_STARTED_TITLE"),
				i18n.GetMessage("NOTIFICATION_STARTUP_FILE_SCAN_ENQUEUED_MESSAGE"),
				"file_scan",
			)
		}
		return
	}

	log.Println("enqueuing file scan task")
	context.Tasks <- utils.Task{
		Type: utils.ScanFiles,
		Data: "file scan",
	}
	emitNotification(
		context,
		"info",
		i18n.GetMessage("NOTIFICATION_FILE_SCAN_STARTED_TITLE"),
		i18n.GetMessage("NOTIFICATION_FILE_SCAN_TASK_ENQUEUED_MESSAGE"),
		"file_scan",
	)
	log.Println("file scan task enqueued")
}

func enqueueStartupScanJob(context *WorkerContext) error {
	rootPath := config.AppConfig.EntryPoint
	if rootPath == "" {
		return nil
	}

	plan, planErr := buildScanPlan(rootPath, JobTypeStartupScan, JobPriorityLow)
	if planErr != nil {
		return planErr
	}

	jobID, err := context.JobOrchestrator.CreateJob(plan)
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

	plan, planErr := buildScanPlan(rootPath, JobTypeFSEvent, priority)
	if planErr != nil {
		return planErr
	}

	_, err := context.JobOrchestrator.CreateJob(plan)
	return err
}

func buildScanPlan(rootPath string, jobType JobType, priority JobPriority) (PlannedJob, error) {
	payload, err := marshalPayload(StepFilePayload{Path: rootPath})
	if err != nil {
		return PlannedJob{}, fmt.Errorf("marshal scan payload: %w", err)
	}
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
	}, nil
}

func worker(id int, context *WorkerContext) {
	for task := range context.Tasks {
		log.Printf("worker %d: processing task %s\n", id, task.Data)

		switch task.Type {
		case utils.ScanFiles:
			if context != nil && context.JobOrchestrator != nil {
				if err := enqueueFilesystemEventJob(context, config.AppConfig.EntryPoint, JobPriorityLow); err != nil {
					log.Printf("worker %d: failed to enqueue fs_event job: %v\n", id, err)
					emitNotification(
						context,
						"error",
						i18n.GetMessage("NOTIFICATION_FILE_SCAN_FAILED_TITLE"),
						err.Error(),
						"",
					)
				}
			} else {
				go StartFileProcessingPipeline(context.FilesService, context.Tasks, context.Logger, context.AIService)
			}
		case utils.ScanDir:
			if context != nil && context.JobOrchestrator != nil {
				targetPath, ok := task.Data.(string)
				if ok {
					if err := enqueueFilesystemEventJob(context, targetPath, JobPriorityNormal); err != nil {
						log.Printf("worker %d: failed to enqueue fs_event job for %s: %v\n", id, targetPath, err)
						emitNotification(
							context,
							"error",
							i18n.GetMessage("NOTIFICATION_DIRECTORY_SCAN_FAILED_TITLE"),
							i18n.Translate("NOTIFICATION_DIRECTORY_SCAN_FAILED_MESSAGE", targetPath, err),
							"",
						)
					}
				}
			} else {
				go ScanDirWorker(context.FilesService, task.Data)
			}
		case utils.UpdateCheckSum:
			UpdateCheckSumWorker(context, task.Data)
		case utils.CreateThumbnail:
			CreateThumbnailWorker(context.FilesService, task.Data, context.Logger)
		case utils.GenerateVideoPlaylists:
			GenerateVideoPlaylistsWorker(context.VideoService, context.Logger)
		default:
			log.Printf("worker %d: unknown task type %v\n", id, task.Type)
		}
	}
}
