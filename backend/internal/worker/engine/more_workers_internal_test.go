package engine

import (
	"database/sql"
	"errors"
	"testing"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/config"
	"nas-go/api/internal/worker/scan"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

type workerFilesServiceMock struct {
	files.ServiceInterface
	getFilesFn          func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	getFileStatByPathFn func(path string) (files.FileStat, bool, error)
	createFileFn        func(fileDto files.FileDto) (files.FileDto, error)
	updateFileFn        func(file files.FileDto) (bool, error)
	updateCheckSumFn    func(fileID int) error
	getFileByIDFn       func(id int) (files.FileDto, error)
	getFileByNamePathFn func(name, path string) (files.FileDto, error)
	getFileThumbFn      func(fileDto files.FileDto, width, height int) ([]byte, error)
	getVideoThumbFn     func(fileDto files.FileDto, width, height int) ([]byte, error)
	getVideoGifFn       func(fileDto files.FileDto, width, height int) ([]byte, error)
}

func (m *workerFilesServiceMock) GetFiles(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	if m.getFilesFn != nil {
		return m.getFilesFn(filter, page, pageSize)
	}
	return utils.PaginationResponse[files.FileDto]{}, nil
}
func (m *workerFilesServiceMock) GetFileStatByPath(path string) (files.FileStat, bool, error) {
	if m.getFileStatByPathFn != nil {
		return m.getFileStatByPathFn(path)
	}
	return files.FileStat{}, false, nil
}
func (m *workerFilesServiceMock) CreateFile(fileDto files.FileDto) (files.FileDto, error) {
	if m.createFileFn != nil {
		return m.createFileFn(fileDto)
	}
	return fileDto, nil
}
func (m *workerFilesServiceMock) UpdateFile(file files.FileDto) (bool, error) {
	if m.updateFileFn != nil {
		return m.updateFileFn(file)
	}
	return true, nil
}
func (m *workerFilesServiceMock) UpdateCheckSum(fileID int) error {
	if m.updateCheckSumFn != nil {
		return m.updateCheckSumFn(fileID)
	}
	return nil
}
func (m *workerFilesServiceMock) GetFileById(id int) (files.FileDto, error) {
	if m.getFileByIDFn != nil {
		return m.getFileByIDFn(id)
	}
	return files.FileDto{}, nil
}
func (m *workerFilesServiceMock) GetFileByNameAndPath(name, path string) (files.FileDto, error) {
	if m.getFileByNamePathFn != nil {
		return m.getFileByNamePathFn(name, path)
	}
	return files.FileDto{}, sql.ErrNoRows
}
func (m *workerFilesServiceMock) GetFileThumbnail(fileDto files.FileDto, width, height int) ([]byte, error) {
	if m.getFileThumbFn != nil {
		return m.getFileThumbFn(fileDto, width, height)
	}
	return []byte("thumb"), nil
}

type workerVideoServiceMock struct {
	video.ServiceInterface
	rebuildFn       func() error
	getVideoThumbFn func(fileDto files.FileDto, width, height int) ([]byte, error)
	getVideoGifFn   func(fileDto files.FileDto, width, height int) ([]byte, error)
}

func (m *workerVideoServiceMock) GetVideoThumbnail(fileDto files.FileDto, width, height int) ([]byte, error) {
	if m.getVideoThumbFn != nil {
		return m.getVideoThumbFn(fileDto, width, height)
	}
	return []byte("thumb"), nil
}
func (m *workerVideoServiceMock) GetVideoPreviewGif(fileDto files.FileDto, width, height int) ([]byte, error) {
	if m.getVideoGifFn != nil {
		return m.getVideoGifFn(fileDto, width, height)
	}
	return []byte("gif"), nil
}

func (m *workerVideoServiceMock) RebuildSmartPlaylists() error {
	if m.rebuildFn != nil {
		return m.rebuildFn()
	}
	return nil
}

type workerLoggerMock struct {
	logger.LoggerServiceInterface
	createLogFn            func(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error)
	completeWithSuccessFn  func(log logger.LoggerModel) error
	completeWithErrorLogFn func(log logger.LoggerModel, err error) error
}

func (m *workerLoggerMock) CreateLog(logModel logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	if m.createLogFn != nil {
		return m.createLogFn(logModel, object)
	}
	return logModel, nil
}

func (m *workerLoggerMock) CompleteWithSuccessLog(logModel logger.LoggerModel) error {
	if m.completeWithSuccessFn != nil {
		return m.completeWithSuccessFn(logModel)
	}
	return nil
}

func (m *workerLoggerMock) CompleteWithErrorLog(logModel logger.LoggerModel, err error) error {
	if m.completeWithErrorLogFn != nil {
		return m.completeWithErrorLogFn(logModel, err)
	}
	return nil
}

func TestWorkerSchedulerAndWorkerLoop(t *testing.T) {
	// Without an orchestrator the scheduler must refuse loudly — no legacy
	// scan task may be enqueued as a silent fallback.
	tasks := make(chan utils.Task, 2)
	ctx := &WorkerContext{Tasks: tasks}

	startWorkersScheduler(ctx)
	select {
	case task := <-tasks:
		t.Fatalf("expected no fallback task without orchestrator, got %v", task.Type)
	default:
	}

	loopTasks := make(chan utils.Task, 1)
	loopTasks <- utils.Task{Type: utils.TaskType(99), Data: "x"}
	close(loopTasks)
	worker(1, &WorkerContext{Tasks: loopTasks})
}

func TestStartWorkersRespectsConfigFlag(t *testing.T) {
	prev := config.AppConfig
	t.Cleanup(func() { config.AppConfig = prev })

	config.AppConfig.EnableWorkers = false
	ctx := &WorkerContext{Tasks: make(chan utils.Task, 1)}
	StartWorkers(ctx, 1)
}

func TestStartWorkersWithoutJobsRepositoryRefusesToStart(t *testing.T) {
	prev := config.AppConfig
	t.Cleanup(func() { config.AppConfig = prev })

	config.AppConfig.EnableWorkers = true
	ctx := &WorkerContext{Tasks: make(chan utils.Task, 2)}

	// JobsRepository is mandatory: the subsystem must refuse to start instead
	// of degrading to the removed legacy pipeline.
	StartWorkers(ctx, 0)

	if ctx.JobOrchestrator != nil || ctx.JobScheduler != nil {
		t.Fatalf("expected no orchestrator/scheduler without JobsRepository")
	}
	select {
	case task := <-ctx.Tasks:
		t.Fatalf("expected no task scheduled, got %v", task.Type)
	default:
	}
}

func TestWorkerEnqueuesJobsWhenOrchestratorIsAvailable(t *testing.T) {
	previousConfig := config.AppConfig
	t.Cleanup(func() { config.AppConfig = previousConfig })

	root := t.TempDir()
	config.AppConfig.EntryPoint = root

	repository := newFakeJobsRepository()
	context := &WorkerContext{
		Tasks:           make(chan utils.Task, 2),
		JobOrchestrator: NewJobOrchestrator(repository, nil),
	}

	context.Tasks <- utils.Task{Type: utils.ScanFiles, Data: "scan"}
	context.Tasks <- utils.Task{Type: utils.ScanDir, Data: root}
	close(context.Tasks)

	worker(1, context)

	if len(repository.jobs) != 2 {
		t.Fatalf("expected two enqueued fs_event jobs, got %d", len(repository.jobs))
	}
}

func TestWorkerKnownTaskBranches(t *testing.T) {
	tasks := make(chan utils.Task, 4)
	tasks <- utils.Task{Type: utils.ScanDir, Data: 123}
	tasks <- utils.Task{Type: utils.UpdateCheckSum, Data: "bad"}
	tasks <- utils.Task{Type: utils.CreateThumbnail, Data: "bad"}
	tasks <- utils.Task{Type: utils.GenerateVideoPlaylists, Data: nil}
	close(tasks)

	worker(2, &WorkerContext{
		Tasks:        tasks,
		FilesService: &workerFilesServiceMock{},
		Logger:       &workerLoggerMock{},
	})
}

func TestCreateThumbnailWorkerAndVideoPlaylistWorker(t *testing.T) {
	videoCalls := 0
	imageCalls := 0
	svc := &workerFilesServiceMock{
		getFileByIDFn: func(id int) (files.FileDto, error) {
			if id == 1 {
				return files.FileDto{ID: 1, Type: files.File, Format: ".mp4"}, nil
			}
			return files.FileDto{ID: 2, Type: files.File, Format: ".jpg"}, nil
		},
		getFileThumbFn: func(fileDto files.FileDto, width, height int) ([]byte, error) {
			imageCalls++
			return []byte("i"), nil
		},
	}
	thumbVideoSvc := &workerVideoServiceMock{
		getVideoThumbFn: func(fileDto files.FileDto, width, height int) ([]byte, error) {
			videoCalls++
			return []byte("v"), nil
		},
		getVideoGifFn: func(fileDto files.FileDto, width, height int) ([]byte, error) {
			videoCalls++
			return []byte("g"), nil
		},
	}

	scan.CreateThumbnailWorker(svc, thumbVideoSvc, "bad", &workerLoggerMock{})
	scan.CreateThumbnailWorker(svc, thumbVideoSvc, 1, &workerLoggerMock{})
	scan.CreateThumbnailWorker(svc, thumbVideoSvc, 2, &workerLoggerMock{})
	if videoCalls != 2 || imageCalls != 1 {
		t.Fatalf("unexpected thumbnail calls, video=%d image=%d", videoCalls, imageCalls)
	}

	scan.GenerateVideoPlaylistsWorker(nil, &workerLoggerMock{})
	playlistCalls := 0
	videoSvc := &workerVideoServiceMock{
		rebuildFn: func() error {
			playlistCalls++
			return nil
		},
	}
	scan.GenerateVideoPlaylistsWorker(videoSvc, &workerLoggerMock{})
	if playlistCalls != 1 {
		t.Fatalf("expected one rebuild call, got %d", playlistCalls)
	}

	errVideoSvc := &workerVideoServiceMock{
		rebuildFn: func() error {
			return errors.New("rebuild failed")
		},
	}
	scan.GenerateVideoPlaylistsWorker(errVideoSvc, &workerLoggerMock{})
	scan.CreateThumbnailWorker(&workerFilesServiceMock{
		getFileByIDFn: func(id int) (files.FileDto, error) {
			return files.FileDto{}, errors.New("missing")
		},
	}, nil, 100, &workerLoggerMock{})
	scan.CreateThumbnailWorker(&workerFilesServiceMock{
		getFileByIDFn: func(id int) (files.FileDto, error) {
			return files.FileDto{ID: 10, Type: files.Directory, Format: ".mp4"}, nil
		},
	}, nil, 10, &workerLoggerMock{})
	// Video file without a video service: skips gracefully.
	scan.CreateThumbnailWorker(&workerFilesServiceMock{
		getFileByIDFn: func(id int) (files.FileDto, error) {
			return files.FileDto{ID: 11, Type: files.File, Format: ".mp4"}, nil
		},
	}, nil, 11, &workerLoggerMock{})
}
