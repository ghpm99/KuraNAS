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

type workerLoggerMock struct{ logger.LoggerServiceInterface }

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
