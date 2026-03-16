package worker

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/video"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

type workerFilesServiceMock struct {
	files.ServiceInterface
	getFilesFn          func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
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
func (m *workerFilesServiceMock) GetVideoThumbnail(fileDto files.FileDto, width, height int) ([]byte, error) {
	if m.getVideoThumbFn != nil {
		return m.getVideoThumbFn(fileDto, width, height)
	}
	return []byte("thumb"), nil
}
func (m *workerFilesServiceMock) GetVideoPreviewGif(fileDto files.FileDto, width, height int) ([]byte, error) {
	if m.getVideoGifFn != nil {
		return m.getVideoGifFn(fileDto, width, height)
	}
	return []byte("gif"), nil
}

type workerVideoServiceMock struct {
	video.ServiceInterface
	rebuildFn func() error
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
	tasks := make(chan utils.Task, 2)
	ctx := &WorkerContext{Tasks: tasks}

	startWorkersScheduler(ctx)
	select {
	case task := <-tasks:
		if task.Type != utils.ScanFiles {
			t.Fatalf("expected ScanFiles task, got %v", task.Type)
		}
	default:
		t.Fatalf("expected task in queue")
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

func TestStartWorkersEnabledSchedulesScanTask(t *testing.T) {
	prev := config.AppConfig
	t.Cleanup(func() { config.AppConfig = prev })

	config.AppConfig.EnableWorkers = true
	ctx := &WorkerContext{Tasks: make(chan utils.Task, 2)}

	StartWorkers(ctx, 0)

	select {
	case task := <-ctx.Tasks:
		if task.Type != utils.ScanFiles {
			t.Fatalf("expected scheduled ScanFiles task, got %v", task.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected scheduled task from StartWorkers")
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

func TestScanDirWorkerAndHelpers(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "new.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	var created, updated int
	svc := &workerFilesServiceMock{
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{
				Items: []files.FileDto{
					{Name: "old.txt", Path: tmpDir + "/old.txt"},
				},
			}, nil
		},
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
			created++
			return fileDto, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) {
			updated++
			return true, nil
		},
	}

	ScanDirWorker(svc, 123) // invalid input branch
	ScanDirWorker(svc, tmpDir)
	if created == 0 {
		t.Fatalf("expected at least one create operation, created=%d updated=%d", created, updated)
	}

	if !fileExists(filepath.Join(tmpDir, "new.txt")) {
		t.Fatalf("expected fileExists true for existing file")
	}
	if fileExists(filepath.Join(tmpDir, "missing.txt")) {
		t.Fatalf("expected fileExists false for missing file")
	}
}

func TestScanDirWorkerErrorBranches(t *testing.T) {
	svc := &workerFilesServiceMock{
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{}, errors.New("get files failed")
		},
	}

	ScanDirWorker(svc, filepath.Join(t.TempDir(), "missing")) // read dir error
	ScanDirWorker(svc, t.TempDir())                           // get files error
}

func TestScanFilesWorker(t *testing.T) {
	prev := config.AppConfig
	t.Cleanup(func() { config.AppConfig = prev })

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "new.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	config.AppConfig.EntryPoint = tmpDir

	createCalls := 0
	checksumCalls := make(chan int, 4)
	svc := &workerFilesServiceMock{
		getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
			return files.FileDto{}, sql.ErrNoRows
		},
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
			createCalls++
			fileDto.ID = createCalls
			return fileDto, nil
		},
		updateCheckSumFn: func(fileID int) error {
			checksumCalls <- fileID
			return nil
		},
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{}, nil
		},
	}
	logSvc := &workerLoggerMock{
		createLogFn: func(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
			log.ID = 1
			return log, nil
		},
	}

	ScanFilesWorker(svc, logSvc)

	if createCalls == 0 {
		t.Fatalf("expected create calls > 0")
	}
	select {
	case <-checksumCalls:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected checksum call")
	}
}

func TestScanFilesWorker_UsesRepositoryErrorInFailureCallback(t *testing.T) {
	prev := config.AppConfig
	t.Cleanup(func() { config.AppConfig = prev })

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "item.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	config.AppConfig.EntryPoint = tmpDir

	expectedErr := errors.New("db lookup failed")
	svc := &workerFilesServiceMock{
		getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
			return files.FileDto{}, expectedErr
		},
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{}, nil
		},
	}

	var receivedErr error
	logSvc := &workerLoggerMock{
		createLogFn: func(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
			log.ID = 1
			return log, nil
		},
		completeWithErrorLogFn: func(log logger.LoggerModel, err error) error {
			receivedErr = err
			return nil
		},
	}

	ScanFilesWorker(svc, logSvc)

	if !errors.Is(receivedErr, expectedErr) {
		t.Fatalf("expected failure callback to receive repository error, got %v", receivedErr)
	}
}

func TestScanFilesWorker_UpdatePathAndWalkError(t *testing.T) {
	t.Run("existing file uses update path and triggers checksum", func(t *testing.T) {
		prev := config.AppConfig
		t.Cleanup(func() { config.AppConfig = prev })

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "existing.txt")
		if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		config.AppConfig.EntryPoint = tmpDir

		updated := 0
		checksums := make(chan int, 4)
		svc := &workerFilesServiceMock{
			getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
				return files.FileDto{ID: 99, Name: name, Path: path}, nil
			},
			updateFileFn: func(file files.FileDto) (bool, error) {
				updated++
				return true, nil
			},
			updateCheckSumFn: func(fileID int) error {
				checksums <- fileID
				return nil
			},
			getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
				return utils.PaginationResponse[files.FileDto]{}, nil
			},
		}
		logSvc := &workerLoggerMock{}

		ScanFilesWorker(svc, logSvc)
		if updated == 0 {
			t.Fatalf("expected update path to be used")
		}
		select {
		case id := <-checksums:
			if id != 99 {
				t.Fatalf("expected checksum for existing id 99, got %d", id)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("expected checksum update call")
		}
	})

	t.Run("missing root path exercises walk error branch", func(t *testing.T) {
		prev := config.AppConfig
		t.Cleanup(func() { config.AppConfig = prev })
		config.AppConfig.EntryPoint = filepath.Join(t.TempDir(), "missing-root")

		svc := &workerFilesServiceMock{
			getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
				return utils.PaginationResponse[files.FileDto]{}, nil
			},
		}

		ScanFilesWorker(svc, &workerLoggerMock{})
	})
}

func TestScanFilesWorker_CreateAndUpdateFailuresDoNotAdvanceChecksum(t *testing.T) {
	t.Run("create failure", func(t *testing.T) {
		prev := config.AppConfig
		t.Cleanup(func() { config.AppConfig = prev })

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "new.txt")
		if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		config.AppConfig.EntryPoint = tmpDir

		checksumCalls := 0
		svc := &workerFilesServiceMock{
			getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
				return files.FileDto{}, sql.ErrNoRows
			},
			createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
				return files.FileDto{}, errors.New("create failed")
			},
			updateCheckSumFn: func(fileID int) error {
				checksumCalls++
				return nil
			},
			getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
				return utils.PaginationResponse[files.FileDto]{}, nil
			},
		}
		logSvc := &workerLoggerMock{}
		ScanFilesWorker(svc, logSvc)
		if checksumCalls != 0 {
			t.Fatalf("expected no checksum calls on create failure, got %d", checksumCalls)
		}
	})

	t.Run("update returns false", func(t *testing.T) {
		prev := config.AppConfig
		t.Cleanup(func() { config.AppConfig = prev })

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "existing.txt")
		if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		config.AppConfig.EntryPoint = tmpDir

		checksumCalls := 0
		svc := &workerFilesServiceMock{
			getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
				return files.FileDto{ID: 1, Name: name, Path: path}, nil
			},
			updateFileFn: func(file files.FileDto) (bool, error) {
				return false, nil
			},
			updateCheckSumFn: func(fileID int) error {
				checksumCalls++
				return nil
			},
			getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
				return utils.PaginationResponse[files.FileDto]{}, nil
			},
		}
		logSvc := &workerLoggerMock{}
		ScanFilesWorker(svc, logSvc)
		if checksumCalls != 0 {
			t.Fatalf("expected no checksum calls when update fails, got %d", checksumCalls)
		}
	})
}

func TestCreateAndUpdateFileDtoHelpers(t *testing.T) {
	svc := &workerFilesServiceMock{
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
			fileDto.ID = 10
			return fileDto, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) { return true, nil },
	}
	fail := func(err error) error {
		if err == nil {
			return errors.New("unexpected nil error")
		}
		return err
	}

	created, err := createFileDto(svc, "/tmp/a.txt", files.FileDto{Name: "a.txt"}, fail)
	if err != nil || created.ID != 10 {
		t.Fatalf("createFileDto failed, created=%+v err=%v", created, err)
	}

	ok, err := UpdateFileRecord(svc, files.FileDto{Name: "a", Format: ".txt"}, files.FileDto{ID: 1})
	if err != nil || !ok {
		t.Fatalf("UpdateFileRecord failed, ok=%v err=%v", ok, err)
	}

	if err := updateFileDto(svc, files.FileDto{ID: 1}, fail); err != nil {
		t.Fatalf("updateFileDto returned error: %v", err)
	}
}

func TestCreateAndUpdateFileDtoHelpers_ErrorPaths(t *testing.T) {
	expectedCreateErr := errors.New("create failed")
	expectedUpdateErr := errors.New("update failed")

	createSvc := &workerFilesServiceMock{
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
			return files.FileDto{}, expectedCreateErr
		},
	}
	createFailCalled := 0
	fail := func(err error) error {
		createFailCalled++
		if !errors.Is(err, expectedCreateErr) {
			t.Fatalf("expected create error in callback, got %v", err)
		}
		return nil
	}
	if _, err := createFileDto(createSvc, "/tmp/a.txt", files.FileDto{Name: "a.txt"}, fail); !errors.Is(err, expectedCreateErr) {
		t.Fatalf("expected create error to propagate, got %v", err)
	}
	if createFailCalled != 1 {
		t.Fatalf("expected create fail callback once, got %d", createFailCalled)
	}

	updateSvcErr := &workerFilesServiceMock{
		updateFileFn: func(file files.FileDto) (bool, error) {
			return false, expectedUpdateErr
		},
	}
	updateFailCalled := 0
	updateFail := func(err error) error {
		updateFailCalled++
		if !errors.Is(err, expectedUpdateErr) {
			t.Fatalf("expected update error in callback, got %v", err)
		}
		return nil
	}
	if err := updateFileDto(updateSvcErr, files.FileDto{ID: 1}, updateFail); !errors.Is(err, expectedUpdateErr) {
		t.Fatalf("expected update error to propagate, got %v", err)
	}
	if updateFailCalled != 1 {
		t.Fatalf("expected update fail callback once, got %d", updateFailCalled)
	}

	updateSvcFalse := &workerFilesServiceMock{
		updateFileFn: func(file files.FileDto) (bool, error) {
			return false, nil
		},
	}
	notUpdatedCallbackCalls := 0
	notUpdatedFail := func(err error) error {
		notUpdatedCallbackCalls++
		if err == nil || err.Error() != "file was not updated" {
			t.Fatalf("expected not-updated error in callback, got %v", err)
		}
		return nil
	}
	err := updateFileDto(updateSvcFalse, files.FileDto{ID: 1}, notUpdatedFail)
	if err == nil || err.Error() != "file was not updated" {
		t.Fatalf("expected not-updated error propagation, got %v", err)
	}
	if notUpdatedCallbackCalls != 1 {
		t.Fatalf("expected not-updated callback once, got %d", notUpdatedCallbackCalls)
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

func TestFindFilesDeleted(t *testing.T) {
	tmpDir := t.TempDir()
	missing := filepath.Join(tmpDir, "missing.txt")
	existing := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(existing, []byte("ok"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	updates := 0
	svc := &workerFilesServiceMock{
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			return utils.PaginationResponse[files.FileDto]{
				Items: []files.FileDto{
					{ID: 1, Name: "missing", Path: missing},
					{ID: 2, Name: "existing", Path: existing},
				},
			}, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) {
			updates++
			return true, nil
		},
	}

	deleted := findFilesDeleted(svc)
	if deleted != 1 {
		t.Fatalf("expected 1 deleted file, got %d", deleted)
	}
	if updates == 0 {
		t.Fatalf("expected at least one update call")
	}
}

func TestFindFilesDeleted_CountsOnlySuccessfulUpdatesAndKeepsDeletedFilterOnPagination(t *testing.T) {
	tmpDir := t.TempDir()
	missing1 := filepath.Join(tmpDir, "missing-1.txt")
	missing2 := filepath.Join(tmpDir, "missing-2.txt")
	existing := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(existing, []byte("ok"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	pageCalls := 0
	updates := 0
	svc := &workerFilesServiceMock{
		getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
			pageCalls++
			if !filter.DeletedAt.HasValue {
				t.Fatalf("expected DeletedAt filter to be preserved on page %d", page)
			}
			if page == 1 {
				return utils.PaginationResponse[files.FileDto]{
					Items: []files.FileDto{{ID: 1, Name: "missing-1", Path: missing1}},
					Pagination: utils.Pagination{
						Page: 1, PageSize: 20, HasNext: true,
					},
				}, nil
			}
			return utils.PaginationResponse[files.FileDto]{
				Items: []files.FileDto{
					{ID: 2, Name: "missing-2", Path: missing2},
					{ID: 3, Name: "existing", Path: existing},
				},
				Pagination: utils.Pagination{
					Page: 2, PageSize: 20, HasNext: false,
				},
			}, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) {
			updates++
			if file.ID == 2 {
				return false, errors.New("update failed")
			}
			return true, nil
		},
	}

	deleted := findFilesDeleted(svc)
	if pageCalls < 2 {
		t.Fatalf("expected at least two pagination calls, got %d", pageCalls)
	}
	if updates != 2 {
		t.Fatalf("expected two update attempts for missing files, got %d", updates)
	}
	if deleted != 1 {
		t.Fatalf("expected deleted count to include only successful updates, got %d", deleted)
	}
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
		getVideoThumbFn: func(fileDto files.FileDto, width, height int) ([]byte, error) {
			videoCalls++
			return []byte("v"), nil
		},
		getVideoGifFn: func(fileDto files.FileDto, width, height int) ([]byte, error) {
			videoCalls++
			return []byte("g"), nil
		},
		getFileThumbFn: func(fileDto files.FileDto, width, height int) ([]byte, error) {
			imageCalls++
			return []byte("i"), nil
		},
	}

	CreateThumbnailWorker(svc, "bad", &workerLoggerMock{})
	CreateThumbnailWorker(svc, 1, &workerLoggerMock{})
	CreateThumbnailWorker(svc, 2, &workerLoggerMock{})
	if videoCalls != 2 || imageCalls != 1 {
		t.Fatalf("unexpected thumbnail calls, video=%d image=%d", videoCalls, imageCalls)
	}

	GenerateVideoPlaylistsWorker(nil, &workerLoggerMock{})
	playlistCalls := 0
	videoSvc := &workerVideoServiceMock{
		rebuildFn: func() error {
			playlistCalls++
			return nil
		},
	}
	GenerateVideoPlaylistsWorker(videoSvc, &workerLoggerMock{})
	if playlistCalls != 1 {
		t.Fatalf("expected one rebuild call, got %d", playlistCalls)
	}

	errVideoSvc := &workerVideoServiceMock{
		rebuildFn: func() error {
			return errors.New("rebuild failed")
		},
	}
	GenerateVideoPlaylistsWorker(errVideoSvc, &workerLoggerMock{})
	CreateThumbnailWorker(&workerFilesServiceMock{
		getFileByIDFn: func(id int) (files.FileDto, error) {
			return files.FileDto{}, errors.New("missing")
		},
	}, 100, &workerLoggerMock{})
	CreateThumbnailWorker(&workerFilesServiceMock{
		getFileByIDFn: func(id int) (files.FileDto, error) {
			return files.FileDto{ID: 10, Type: files.Directory, Format: ".mp4"}, nil
		},
	}, 10, &workerLoggerMock{})
}

func TestDatabasePersistenceWorker(t *testing.T) {
	now := time.Now()
	created := 0
	updated := 0
	svc := &workerFilesServiceMock{
		getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
			if name == "existing.mp4" {
				return files.FileDto{ID: 9, Name: name, Path: path}, nil
			}
			return files.FileDto{}, sql.ErrNoRows
		},
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
			created++
			fileDto.ID = 10
			return fileDto, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) {
			updated++
			return true, nil
		},
	}

	in := make(chan files.FileDto, 2)
	monitor := make(chan ResultWorkerData, 2)
	tasks := make(chan utils.Task, 2)
	in <- files.FileDto{Name: "new.mp4", Path: "/x/new.mp4", Type: files.File, Format: ".mp4", UpdatedAt: now}
	in <- files.FileDto{Name: "existing.mp4", Path: "/x/existing.mp4", Type: files.File, Format: ".mp4", UpdatedAt: now}
	close(in)

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		StartDatabasePersistenceWorker(svc, tasks, in, monitor, &wg)
		close(done)
	}()
	<-done
	close(monitor)

	successCount := 0
	for result := range monitor {
		if result.Success {
			successCount++
		}
	}
	if successCount != 2 || created != 1 || updated != 1 {
		t.Fatalf("unexpected persistence result success=%d created=%d updated=%d", successCount, created, updated)
	}
}

func TestDatabasePersistenceWorkerErrorPathsAndTaskGuards(t *testing.T) {
	now := time.Now()
	svc := &workerFilesServiceMock{
		getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
			if name == "lookup-error.mp4" {
				return files.FileDto{}, errors.New("lookup failed")
			}
			if name == "create-fail.mp4" {
				return files.FileDto{}, sql.ErrNoRows
			}
			if name == "update-fail.mp4" {
				return files.FileDto{ID: 99, Name: name, Path: path}, nil
			}
			return files.FileDto{}, sql.ErrNoRows
		},
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
			if fileDto.Name == "create-fail.mp4" {
				return files.FileDto{}, errors.New("create failed")
			}
			fileDto.ID = 10
			return fileDto, nil
		},
		updateFileFn: func(file files.FileDto) (bool, error) {
			if file.Name == "update-fail.mp4" {
				return false, errors.New("update failed")
			}
			return true, nil
		},
	}

	in := make(chan files.FileDto, 3)
	monitor := make(chan ResultWorkerData, 3)
	tasks := make(chan utils.Task, 1)
	tasks <- utils.Task{Type: utils.ScanFiles, Data: "fill"} // fill queue to hit default branch in enqueue

	in <- files.FileDto{Name: "lookup-error.mp4", Path: "/x/lookup-error.mp4", Type: files.File, Format: ".mp4", UpdatedAt: now}
	in <- files.FileDto{Name: "create-fail.mp4", Path: "/x/create-fail.mp4", Type: files.File, Format: ".mp4", UpdatedAt: now}
	in <- files.FileDto{Name: "update-fail.mp4", Path: "/x/update-fail.mp4", Type: files.File, Format: ".mp4", UpdatedAt: now}
	close(in)

	var wg sync.WaitGroup
	wg.Add(1)
	StartDatabasePersistenceWorker(svc, tasks, in, monitor, &wg)
	close(monitor)

	errorCount := 0
	for result := range monitor {
		if !result.Success {
			errorCount++
		}
	}
	if errorCount != 3 {
		t.Fatalf("expected three persistence errors, got %d", errorCount)
	}

	localTasks := make(chan utils.Task, 1)
	enqueueVideoThumbnailTask(localTasks, files.FileDto{Type: files.Directory, Format: ".mp4"}, 1)
	enqueueVideoThumbnailTask(localTasks, files.FileDto{Type: files.File, Format: ".txt"}, 1)
	enqueueVideoThumbnailTask(localTasks, files.FileDto{Type: files.File, Format: ".mp4"}, 0)
	if len(localTasks) != 0 {
		t.Fatalf("expected no tasks enqueued for invalid guards")
	}
}
