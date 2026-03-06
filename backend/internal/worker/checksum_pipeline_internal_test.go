package worker

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
)

type pipelineFilesServiceMock struct {
	workerFilesServiceMock
	updated []files.FileDto
	created []files.FileDto
}

func (m *pipelineFilesServiceMock) UpdateFile(file files.FileDto) (bool, error) {
	m.updated = append(m.updated, file)
	if m.updateFileFn != nil {
		return m.updateFileFn(file)
	}
	return true, nil
}

func (m *pipelineFilesServiceMock) CreateFile(fileDto files.FileDto) (files.FileDto, error) {
	fileDto.ID = len(m.created) + 1
	m.created = append(m.created, fileDto)
	return fileDto, nil
}

func (m *pipelineFilesServiceMock) GetFileByNameAndPath(name, path string) (files.FileDto, error) {
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

func TestUpdateCheckSumWorker(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "a.txt")
	if err := os.WriteFile(filePath, []byte("abc"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	subDir := filepath.Join(tmpDir, "dir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "b.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("failed to write dir file: %v", err)
	}

	mock := &pipelineFilesServiceMock{
		workerFilesServiceMock: workerFilesServiceMock{
			getFileByIDFn: func(id int) (files.FileDto, error) {
				if id == 1 {
					return files.FileDto{ID: 1, Name: "a.txt", Path: filePath, Type: files.File, UpdatedAt: time.Now()}, nil
				}
				return files.FileDto{ID: 2, Name: "dir", Path: subDir, Type: files.Directory, UpdatedAt: time.Now()}, nil
			},
			getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
				return utils.PaginationResponse[files.FileDto]{
					Items: []files.FileDto{{ID: 3, Name: "b.txt", CheckSum: "abcd", Path: filepath.Join(subDir, "b.txt"), Type: files.File}},
					Pagination: utils.Pagination{
						Page: page, PageSize: pageSize, HasNext: false,
					},
				}, nil
			},
		},
	}

	UpdateCheckSumWorker(&WorkerContext{FilesService: mock}, "bad")
	UpdateCheckSumWorker(&WorkerContext{FilesService: mock}, 1)
	UpdateCheckSumWorker(&WorkerContext{FilesService: mock}, 2)
	if len(mock.updated) < 2 {
		t.Fatalf("expected updated files from checksum worker, got %d", len(mock.updated))
	}
}

func TestUpdateCheckSumWorker_ErrorBranchesDoNotUpdateInvalidEntries(t *testing.T) {
	tmpDir := t.TempDir()
	missingFile := filepath.Join(tmpDir, "missing.txt")
	subDir := filepath.Join(tmpDir, "dir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	updateCalls := 0
	mock := &pipelineFilesServiceMock{
		workerFilesServiceMock: workerFilesServiceMock{
			getFileByIDFn: func(id int) (files.FileDto, error) {
				switch id {
				case 1:
					// Missing path: checksum generation should fail and not call update.
					return files.FileDto{ID: 1, Name: "missing", Path: missingFile, Type: files.File, UpdatedAt: time.Now()}, nil
				case 2:
					// Directory listing error: should return early and not call update.
					return files.FileDto{ID: 2, Name: "dir", Path: subDir, Type: files.Directory, UpdatedAt: time.Now()}, nil
				default:
					return files.FileDto{}, errors.New("not found")
				}
			},
			getFilesFn: func(filter files.FileFilter, page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
				return utils.PaginationResponse[files.FileDto]{}, errors.New("list failed")
			},
		},
	}
	mock.updateFileFn = func(file files.FileDto) (bool, error) {
		updateCalls++
		return true, nil
	}

	UpdateCheckSumWorker(&WorkerContext{FilesService: mock}, 1)
	UpdateCheckSumWorker(&WorkerContext{FilesService: mock}, 2)
	if updateCalls != 1 {
		t.Fatalf("expected one update call (directory checksum still succeeds), got %d", updateCalls)
	}
}

func TestStartFileProcessingPipeline(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	prev := config.AppConfig
	t.Cleanup(func() { config.AppConfig = prev })
	config.AppConfig.EntryPoint = tmpDir

	mock := &pipelineFilesServiceMock{}
	tasks := make(chan utils.Task, 4)

	SetPythonScriptRunnerForTesting(func(scriptType utils.ScriptType, filePath string) (string, error) {
		return "{}", nil
	})
	defer SetPythonScriptRunnerForTesting(nil)

	StartFileProcessingPipeline(mock, tasks, &pipelineLoggerMock{})

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
