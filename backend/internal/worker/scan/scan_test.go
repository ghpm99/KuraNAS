package scan

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

// --- mocks ---

type scanFilesServiceMock struct {
	files.ServiceInterface
	getFilesFn          func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error)
	getFileByNamePathFn func(name, path string) (files.FileDto, error)
	createFileFn        func(fileDto files.FileDto) (files.FileDto, error)
	updateFileFn        func(file files.FileDto) (bool, error)
	updateCheckSumFn    func(fileID int) error
	getFileByIDFn       func(id int) (files.FileDto, error)
	getFileThumbFn      func(fileDto files.FileDto, width, height int) ([]byte, error)
	getVideoThumbFn     func(fileDto files.FileDto, width, height int) ([]byte, error)
	getVideoGifFn       func(fileDto files.FileDto, width, height int) ([]byte, error)
}

func (m *scanFilesServiceMock) GetFiles(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	if m.getFilesFn != nil {
		return m.getFilesFn(filter, page, pageSize)
	}
	return utils.PaginationResponse[files.FileDto]{}, nil
}
func (m *scanFilesServiceMock) GetFileByNameAndPath(name, path string) (files.FileDto, error) {
	if m.getFileByNamePathFn != nil {
		return m.getFileByNamePathFn(name, path)
	}
	return files.FileDto{}, sql.ErrNoRows
}
func (m *scanFilesServiceMock) CreateFile(fileDto files.FileDto) (files.FileDto, error) {
	if m.createFileFn != nil {
		return m.createFileFn(fileDto)
	}
	return fileDto, nil
}
func (m *scanFilesServiceMock) UpdateFile(file files.FileDto) (bool, error) {
	if m.updateFileFn != nil {
		return m.updateFileFn(file)
	}
	return true, nil
}
func (m *scanFilesServiceMock) UpdateCheckSum(fileID int) error {
	if m.updateCheckSumFn != nil {
		return m.updateCheckSumFn(fileID)
	}
	return nil
}
func (m *scanFilesServiceMock) GetFileById(id int) (files.FileDto, error) {
	if m.getFileByIDFn != nil {
		return m.getFileByIDFn(id)
	}
	return files.FileDto{}, nil
}
func (m *scanFilesServiceMock) GetFileThumbnail(fileDto files.FileDto, width, height int) ([]byte, error) {
	if m.getFileThumbFn != nil {
		return m.getFileThumbFn(fileDto, width, height)
	}
	return []byte("thumb"), nil
}
func (m *scanFilesServiceMock) GetVideoThumbnail(fileDto files.FileDto, width, height int) ([]byte, error) {
	if m.getVideoThumbFn != nil {
		return m.getVideoThumbFn(fileDto, width, height)
	}
	return []byte("thumb"), nil
}
func (m *scanFilesServiceMock) GetVideoPreviewGif(fileDto files.FileDto, width, height int) ([]byte, error) {
	if m.getVideoGifFn != nil {
		return m.getVideoGifFn(fileDto, width, height)
	}
	return []byte("gif"), nil
}

type scanLoggerMock struct {
	logger.LoggerServiceInterface
	createLogFn            func(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error)
	completeWithSuccessFn  func(log logger.LoggerModel) error
	completeWithErrorLogFn func(log logger.LoggerModel, err error) error
}

func (m *scanLoggerMock) CreateLog(logModel logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	if m.createLogFn != nil {
		return m.createLogFn(logModel, object)
	}
	return logModel, nil
}
func (m *scanLoggerMock) CompleteWithSuccessLog(logModel logger.LoggerModel) error {
	if m.completeWithSuccessFn != nil {
		return m.completeWithSuccessFn(logModel)
	}
	return nil
}
func (m *scanLoggerMock) CompleteWithErrorLog(logModel logger.LoggerModel, err error) error {
	if m.completeWithErrorLogFn != nil {
		return m.completeWithErrorLogFn(logModel, err)
	}
	return nil
}

// --- checksum tests ---

func TestGetCheckSum(t *testing.T) {
	file := files.FileDto{Type: files.File, Path: "/tmp/a"}
	dir := files.FileDto{Type: files.Directory, Path: "/tmp/d"}

	fileHash, err := GetCheckSum(
		file,
		func(path string) (string, error) { return "file-hash", nil },
		func(path string) (string, error) { return "dir-hash", nil },
	)
	if err != nil || fileHash != "file-hash" {
		t.Fatalf("expected file hash, got %q err=%v", fileHash, err)
	}

	dirHash, err := GetCheckSum(
		dir,
		func(path string) (string, error) { return "file-hash", nil },
		func(path string) (string, error) { return "dir-hash", nil },
	)
	if err != nil || dirHash != "dir-hash" {
		t.Fatalf("expected dir hash, got %q err=%v", dirHash, err)
	}

	_, err = GetCheckSum(
		files.FileDto{Type: files.FileType(99), Path: "/tmp/x"},
		func(path string) (string, error) { return "", nil },
		func(path string) (string, error) { return "", nil },
	)
	if err == nil {
		t.Fatalf("expected unknown type error")
	}
}

func TestStartChecksumWorker(t *testing.T) {
	in := make(chan files.FileDto, 3)
	out := make(chan files.FileDto, 3)
	monitor := make(chan ResultWorkerData, 3)
	var wg sync.WaitGroup

	in <- files.FileDto{ID: 1, Type: files.File, Path: "/tmp/file"}
	in <- files.FileDto{ID: 2, Type: files.Directory, Path: "/tmp/dir"}
	in <- files.FileDto{ID: 3, Type: files.FileType(99), Path: "/tmp/unknown"}
	close(in)

	wg.Add(1)
	go StartChecksumWorker(
		in,
		out,
		func(path string) (string, error) { return "fh", nil },
		func(path string) (string, error) { return "dh", nil },
		monitor,
		&wg,
	)
	wg.Wait()
	close(out)
	close(monitor)

	var items []files.FileDto
	for item := range out {
		items = append(items, item)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 processed items, got %d", len(items))
	}
	if items[0].CheckSum == "" || items[1].CheckSum == "" {
		t.Fatalf("expected checksums on first two items")
	}

	var errorsCount int
	for r := range monitor {
		if !r.Success {
			errorsCount++
		}
	}
	if errorsCount != 1 {
		t.Fatalf("expected 1 checksum error, got %d", errorsCount)
	}
}

// --- dto converter and monitor ---

func TestDtoConverterAndMonitor(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "a.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat temp file: %v", err)
	}

	in := make(chan FileWalk, 1)
	out := make(chan files.FileDto, 1)
	var wg sync.WaitGroup
	in <- FileWalk{Path: filePath, Info: info}
	close(in)
	wg.Add(1)
	go StartDtoConverterWorker(in, out, &wg)
	wg.Wait()
	close(out)

	count := 0
	for range out {
		count++
	}
	if count != 1 {
		t.Fatalf("expected one dto from converter")
	}

	monitor := make(chan ResultWorkerData, 2)
	monitor <- ResultWorkerData{Path: "ok", Success: true}
	monitor <- ResultWorkerData{Path: "err", Success: false, Error: "boom"}
	close(monitor)
	wg.Add(1)
	go StartResultMonitorWorker(monitor, &wg)
	wg.Wait()
}

// --- metadata tests ---

func TestGetMetadataDispatchByFormat(t *testing.T) {
	runner := func(scriptType utils.ScriptType, filePath string) (string, error) {
		switch scriptType {
		case utils.ImageMetadata:
			return `{"id":1,"file_id":1,"path":"` + filePath + `","format":"jpeg","mode":"RGB","width":1,"height":1,"created_at":"2026-01-01T00:00:00Z"}`, nil
		case utils.AudioMetadata:
			return `{"id":2,"file_id":1,"path":"` + filePath + `","mime":"audio/mpeg","length":1,"bitrate":320,"sample_rate":44100,"channels":2,"created_at":"2026-01-01T00:00:00Z"}`, nil
		case utils.VideoMetadata:
			return `{"id":3,"file_id":1,"path":"` + filePath + `","format_name":"mp4","size":"1","duration":"1","width":1,"height":1,"created_at":"2026-01-01T00:00:00Z"}`, nil
		default:
			return `{}`, nil
		}
	}

	imgMeta, err := GetMetadata(files.FileDto{ID: 1, Path: "/x.jpg", Format: ".jpg"}, runner, nil)
	if err != nil || imgMeta == nil {
		t.Fatalf("expected image metadata dispatch success, err=%v", err)
	}

	audioMeta, err := GetMetadata(files.FileDto{ID: 1, Path: "/x.mp3", Format: ".mp3"}, runner, nil)
	if err != nil || audioMeta == nil {
		t.Fatalf("expected audio metadata dispatch success, err=%v", err)
	}

	videoMeta, err := GetMetadata(files.FileDto{ID: 1, Path: "/x.mp4", Format: ".mp4"}, runner, nil)
	if err != nil || videoMeta == nil {
		t.Fatalf("expected video metadata dispatch success, err=%v", err)
	}

	nilMeta, err := GetMetadata(files.FileDto{Format: ".txt"}, runner, nil)
	if err != nil || nilMeta != nil {
		t.Fatalf("expected nil metadata for unsupported format, got meta=%v err=%v", nilMeta, err)
	}
}

func TestMetadataWorkerAndHelpers(t *testing.T) {
	runner := func(scriptType utils.ScriptType, filePath string) (string, error) {
		switch scriptType {
		case utils.ImageMetadata:
			b, _ := json.Marshal(files.ImageMetadataModel{Format: "PNG", Path: filePath})
			return string(b), nil
		case utils.AudioMetadata:
			b, _ := json.Marshal(files.AudioMetadataModel{Mime: "mp3", Path: filePath})
			return string(b), nil
		case utils.VideoMetadata:
			b, _ := json.Marshal(files.VideoMetadataModel{FormatName: "mp4", Path: filePath})
			return string(b), nil
		default:
			return "", errors.New("unknown")
		}
	}

	imgMeta, err := getImageMetadata(files.FileDto{ID: 1, Path: "/img.png"}, runner, nil)
	if err != nil || imgMeta.Format != "PNG" {
		t.Fatalf("expected image metadata, err=%v", err)
	}
	if imgMeta.Classification.Category != files.ImageClassificationCategoryOther {
		t.Fatalf("expected default image classification, got %s", imgMeta.Classification.Category)
	}
	audioMeta, err := getAudioMetadata(files.FileDto{ID: 1, Path: "/a.mp3"}, runner)
	if err != nil || audioMeta.Mime != "mp3" {
		t.Fatalf("expected audio metadata, err=%v", err)
	}
	videoMeta, err := getVideoMetadata(files.FileDto{ID: 1, Path: "/v.mp4"}, runner)
	if err != nil || videoMeta.FormatName != "mp4" {
		t.Fatalf("expected video metadata, err=%v", err)
	}

	in := make(chan files.FileDto, 2)
	out := make(chan files.FileDto, 2)
	monitor := make(chan ResultWorkerData, 2)
	var wg sync.WaitGroup

	in <- files.FileDto{ID: 1, Path: "/x.png", Format: ".png", Type: files.File}
	in <- files.FileDto{ID: 2, Path: "/x.txt", Format: ".txt", Type: files.File}
	close(in)

	wg.Add(1)
	go StartMetadataWorker(in, out, runner, monitor, &wg, nil)
	wg.Wait()
	close(out)
	close(monitor)

	processed := 0
	for item := range out {
		processed++
		if item.Format == ".png" {
			if item.Metadata == nil {
				t.Fatalf("expected metadata for png")
			}
			metadata, ok := item.Metadata.(files.ImageMetadataModel)
			if !ok {
				t.Fatalf("expected image metadata model, got %T", item.Metadata)
			}
			if metadata.Classification.Category != files.ImageClassificationCategoryOther {
				t.Fatalf("expected classified image metadata, got %s", metadata.Classification.Category)
			}
		}
	}
	if processed != 2 {
		t.Fatalf("expected 2 processed files, got %d", processed)
	}

	errRunner := func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "", errors.New("runner failed")
	}
	if _, err := getImageMetadata(files.FileDto{ID: 2, Path: "/err.png"}, errRunner, nil); err == nil {
		t.Fatalf("expected image metadata runner error")
	}
	if _, err := getAudioMetadata(files.FileDto{ID: 2, Path: "/err.mp3"}, errRunner); err == nil {
		t.Fatalf("expected audio metadata runner error")
	}
	if _, err := getVideoMetadata(files.FileDto{ID: 2, Path: "/err.mp4"}, errRunner); err == nil {
		t.Fatalf("expected video metadata runner error")
	}
	if _, err := getAudioMetadata(files.FileDto{ID: 3, Path: "/bad.mp3"}, func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "{invalid-json", nil
	}); err == nil {
		t.Fatalf("expected audio metadata json parse error")
	}
	if _, err := getVideoMetadata(files.FileDto{ID: 3, Path: "/bad.mp4"}, func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "{invalid-json", nil
	}); err == nil {
		t.Fatalf("expected video metadata json parse error")
	}
}

func TestMetadataWorkerErrorBranch(t *testing.T) {
	runner := func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "{invalid-json", nil
	}
	in := make(chan files.FileDto, 1)
	out := make(chan files.FileDto, 1)
	monitor := make(chan ResultWorkerData, 1)
	var wg sync.WaitGroup

	in <- files.FileDto{ID: 10, Path: "/x.png", Format: ".png", Type: files.File}
	close(in)
	wg.Add(1)
	go StartMetadataWorker(in, out, runner, monitor, &wg, nil)
	wg.Wait()
	close(out)
	close(monitor)

	if len(out) != 1 {
		t.Fatalf("expected processed file output")
	}
	if len(monitor) != 1 {
		t.Fatalf("expected one monitor error item")
	}
}

// --- directory walker ---

func TestStartDirectoryWalker(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "f.txt")
	if err := os.WriteFile(filePath, []byte("abc"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	fileWalkChannel := make(chan FileWalk, 10)
	monitor := make(chan ResultWorkerData, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go StartDirectoryWalker(tmpDir, fileWalkChannel, monitor, &wg)
	wg.Wait()

	close(fileWalkChannel)
	walked := 0
	for range fileWalkChannel {
		walked++
	}
	if walked < 1 {
		t.Fatalf("expected walked entries, got %d", walked)
	}

	errCh := make(chan FileWalk, 1)
	monErr := make(chan ResultWorkerData, 2)
	wg.Add(1)
	go StartDirectoryWalker(filepath.Join(tmpDir, "missing"), errCh, monErr, &wg)
	wg.Wait()
	close(errCh)
	close(monErr)
	receivedAny := false
	for range errCh {
		receivedAny = true
	}
	if receivedAny {
		t.Fatalf("did not expect file walk items for missing path")
	}
	monitorErrors := 0
	for item := range monErr {
		if !item.Success {
			monitorErrors++
		}
	}
	if monitorErrors == 0 {
		t.Fatalf("expected at least one monitor error for missing path")
	}
}

func TestStartDirectoryWalkerPermissionDenied(t *testing.T) {
	tmpDir := t.TempDir()
	restrictedDir := filepath.Join(tmpDir, "restricted")
	if err := os.MkdirAll(restrictedDir, 0700); err != nil {
		t.Fatalf("failed to create restricted dir: %v", err)
	}
	innerFile := filepath.Join(restrictedDir, "hidden.txt")
	if err := os.WriteFile(innerFile, []byte("x"), 0600); err != nil {
		t.Fatalf("failed to create file in restricted dir: %v", err)
	}
	if err := os.Chmod(restrictedDir, 0000); err != nil {
		t.Fatalf("failed to remove permissions: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(restrictedDir, 0700)
	})

	fileWalkChannel := make(chan FileWalk, 10)
	monitor := make(chan ResultWorkerData, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go StartDirectoryWalker(tmpDir, fileWalkChannel, monitor, &wg)
	wg.Wait()
	close(fileWalkChannel)
	close(monitor)

	permissionErrors := 0
	for item := range monitor {
		if !item.Success && strings.Contains(strings.ToLower(item.Error), "permission") {
			permissionErrors++
		}
	}

	if permissionErrors == 0 {
		t.Skip("environment did not surface permission-denied during walk")
	}
}

// --- pipeline (file processing) ---

type scanPipelineFilesServiceMock struct {
	scanFilesServiceMock
	updated []files.FileDto
	created []files.FileDto
}

func (m *scanPipelineFilesServiceMock) UpdateFile(file files.FileDto) (bool, error) {
	m.updated = append(m.updated, file)
	if m.updateFileFn != nil {
		return m.updateFileFn(file)
	}
	return true, nil
}

func (m *scanPipelineFilesServiceMock) CreateFile(fileDto files.FileDto) (files.FileDto, error) {
	fileDto.ID = len(m.created) + 1
	m.created = append(m.created, fileDto)
	return fileDto, nil
}

func (m *scanPipelineFilesServiceMock) GetFileByNameAndPath(name, path string) (files.FileDto, error) {
	return files.FileDto{}, sql.ErrNoRows
}

type pipelineLoggerMock struct{ logger.LoggerServiceInterface }

func (m *pipelineLoggerMock) CreateLog(log logger.LoggerModel, object interface{}) (logger.LoggerModel, error) {
	return logger.LoggerModel{}, nil
}
func (m *pipelineLoggerMock) CompleteWithSuccessLog(log logger.LoggerModel) error { return nil }
func (m *pipelineLoggerMock) CompleteWithErrorLog(log logger.LoggerModel, err error) error {
	return nil
}

func TestStartFileProcessingPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	prev := config.AppConfig
	t.Cleanup(func() { config.AppConfig = prev })
	config.AppConfig.EntryPoint = tmpDir

	mock := &scanPipelineFilesServiceMock{}
	tasks := make(chan utils.Task, 4)

	SetPythonScriptRunnerForTesting(func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "{}", nil
	})
	defer SetPythonScriptRunnerForTesting(nil)

	StartFileProcessingPipeline(mock, tasks, &pipelineLoggerMock{}, nil)

	if len(mock.created) == 0 && len(mock.updated) == 0 {
		t.Fatalf("expected pipeline to persist at least one file")
	}

	foundVideoTask := false
	close(tasks)
	for task := range tasks {
		if task.Type == utils.GenerateVideoPlaylists {
			foundVideoTask = true
			break
		}
	}
	if !foundVideoTask {
		t.Fatalf("expected GenerateVideoPlaylists task to be enqueued")
	}
}

// --- database persistence ---

func TestDatabasePersistenceWorker(t *testing.T) {
	now := time.Now()
	created := 0
	updated := 0
	svc := &scanFilesServiceMock{
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
	svc := &scanFilesServiceMock{
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
	tasks <- utils.Task{Type: utils.ScanFiles, Data: "fill"}

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
	EnqueueVideoThumbnailTask(localTasks, files.FileDto{Type: files.Directory, Format: ".mp4"}, 1)
	EnqueueVideoThumbnailTask(localTasks, files.FileDto{Type: files.File, Format: ".txt"}, 1)
	EnqueueVideoThumbnailTask(localTasks, files.FileDto{Type: files.File, Format: ".mp4"}, 0)
	if len(localTasks) != 0 {
		t.Fatalf("expected no tasks enqueued for invalid guards")
	}
}

// --- ScanFilesWorker & helpers ---

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
	svc := &scanFilesServiceMock{
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
	logSvc := &scanLoggerMock{
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

func TestScanDirWorkerAndHelpers(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "new.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	var created, updated int
	svc := &scanFilesServiceMock{
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

	if !FileExists(filepath.Join(tmpDir, "new.txt")) {
		t.Fatalf("expected FileExists true for existing file")
	}
	if FileExists(filepath.Join(tmpDir, "missing.txt")) {
		t.Fatalf("expected FileExists false for missing file")
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
	svc := &scanFilesServiceMock{
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

	deleted := FindFilesDeleted(svc)
	if deleted != 1 {
		t.Fatalf("expected 1 deleted file, got %d", deleted)
	}
	if updates == 0 {
		t.Fatalf("expected at least one update call")
	}
}

func TestCreateAndUpdateFileDtoHelpers(t *testing.T) {
	svc := &scanFilesServiceMock{
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

	created, err := CreateFileDto(svc, "/tmp/a.txt", files.FileDto{Name: "a.txt"}, fail)
	if err != nil || created.ID != 10 {
		t.Fatalf("CreateFileDto failed, created=%+v err=%v", created, err)
	}

	ok, err := UpdateFileRecord(svc, files.FileDto{Name: "a", Format: ".txt"}, files.FileDto{ID: 1})
	if err != nil || !ok {
		t.Fatalf("UpdateFileRecord failed, ok=%v err=%v", ok, err)
	}

	if err := UpdateFileDto(svc, files.FileDto{ID: 1}, fail); err != nil {
		t.Fatalf("UpdateFileDto returned error: %v", err)
	}
}

func TestCreateAndUpdateFileDtoHelpers_ErrorPaths(t *testing.T) {
	expectedCreateErr := errors.New("create failed")
	expectedUpdateErr := errors.New("update failed")

	createSvc := &scanFilesServiceMock{
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
	if _, err := CreateFileDto(createSvc, "/tmp/a.txt", files.FileDto{Name: "a.txt"}, fail); !errors.Is(err, expectedCreateErr) {
		t.Fatalf("expected create error to propagate, got %v", err)
	}
	if createFailCalled != 1 {
		t.Fatalf("expected create fail callback once, got %d", createFailCalled)
	}

	updateSvcErr := &scanFilesServiceMock{
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
	if err := UpdateFileDto(updateSvcErr, files.FileDto{ID: 1}, updateFail); !errors.Is(err, expectedUpdateErr) {
		t.Fatalf("expected update error to propagate, got %v", err)
	}
	if updateFailCalled != 1 {
		t.Fatalf("expected update fail callback once, got %d", updateFailCalled)
	}

	updateSvcFalse := &scanFilesServiceMock{
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
	err := UpdateFileDto(updateSvcFalse, files.FileDto{ID: 1}, notUpdatedFail)
	if err == nil || err.Error() != "file was not updated" {
		t.Fatalf("expected not-updated error propagation, got %v", err)
	}
	if notUpdatedCallbackCalls != 1 {
		t.Fatalf("expected not-updated callback once, got %d", notUpdatedCallbackCalls)
	}
}
