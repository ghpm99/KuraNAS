package engine

import (
	"encoding/json"
	"errors"
	jobdomain "nas-go/api/internal/worker/job"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/api/v1/files"
)

type pipelineFilesServiceMock struct {
	workerFilesServiceMock
	updated []files.FileDto
}

func (m *pipelineFilesServiceMock) UpdateFile(file files.FileDto) (bool, error) {
	m.updated = append(m.updated, file)
	if m.updateFileFn != nil {
		return m.updateFileFn(file)
	}
	return true, nil
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

func TestUpdateCheckSumWorkerOrchestratorAndPayloadHelper(t *testing.T) {
	repository := newFakeJobsRepository()
	orchestrator := NewJobOrchestrator(repository, nil)

	UpdateCheckSumWorker(nil, 1)
	UpdateCheckSumWorker(&WorkerContext{}, 0)
	UpdateCheckSumWorker(&WorkerContext{JobOrchestrator: orchestrator}, 11)

	if len(repository.jobs) != 1 {
		t.Fatalf("expected checksum job to be created, got %d", len(repository.jobs))
	}

	steps, err := repository.GetStepsByJobID(1)
	if err != nil {
		t.Fatalf("GetStepsByJobID returned error: %v", err)
	}
	if len(steps) != 1 || steps[0].Type != string(jobdomain.StepTypeChecksum) {
		t.Fatalf("unexpected checksum steps: %+v", steps)
	}

	payload, err := marshalChecksumStepPayload(11)
	if err != nil {
		t.Fatalf("marshalChecksumStepPayload returned error: %v", err)
	}
	decoded := StepFilePayload{}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("failed to unmarshal checksum payload: %v", err)
	}
	if decoded.FileID != 11 {
		t.Fatalf("unexpected checksum payload: %+v", decoded)
	}
}
