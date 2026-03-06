package worker

import (
	"database/sql"
	"errors"
	"testing"

	"nas-go/api/internal/api/v1/files"
	jobs "nas-go/api/internal/api/v1/jobs"
)

func TestExecutePersistStep_CreateAndUpdate(t *testing.T) {
	t.Run("create when not found", func(t *testing.T) {
		created := 0
		svc := &workerFilesServiceMock{
			getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
				return files.FileDto{}, sql.ErrNoRows
			},
			createFileFn: func(fileDto files.FileDto) (files.FileDto, error) {
				created++
				return fileDto, nil
			},
		}

		step := jobs.StepModel{Payload: []byte(`{"file":{"name":"a.txt","path":"/tmp/a.txt","parent_path":"/tmp","type":2,"format":".txt","size":1}}`)}
		err := executePersistStep(&WorkerContext{FilesService: svc}, step)
		if err != nil {
			t.Fatalf("unexpected persist create error: %v", err)
		}
		if created != 1 {
			t.Fatalf("expected one create call, got %d", created)
		}
	})

	t.Run("update when existing", func(t *testing.T) {
		updated := 0
		svc := &workerFilesServiceMock{
			getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
				return files.FileDto{ID: 9, Name: name, Path: path}, nil
			},
			updateFileFn: func(file files.FileDto) (bool, error) {
				updated++
				return true, nil
			},
		}

		step := jobs.StepModel{Payload: []byte(`{"file":{"name":"b.txt","path":"/tmp/b.txt","parent_path":"/tmp","type":2,"format":".txt","size":1}}`)}
		err := executePersistStep(&WorkerContext{FilesService: svc}, step)
		if err != nil {
			t.Fatalf("unexpected persist update error: %v", err)
		}
		if updated != 1 {
			t.Fatalf("expected one update call, got %d", updated)
		}
	})
}

func TestExecuteMetadataAndChecksumSkipCases(t *testing.T) {
	ctx := &WorkerContext{FilesService: &workerFilesServiceMock{}}

	if err := executeMetadataStep(ctx, jobs.StepModel{Payload: []byte(`{"path":"/path/does-not-exist"}`)}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected metadata step skip for missing file, got %v", err)
	}

	if err := executeChecksumStep(ctx, jobs.StepModel{Payload: []byte(`{}`)}); !errors.Is(err, ErrStepSkipped) {
		t.Fatalf("expected checksum step skip for empty payload, got %v", err)
	}
}

func TestExecuteThumbnailAndPlaylistExecutors(t *testing.T) {
	thumbCalls := 0
	playlistCalls := 0

	filesService := &workerFilesServiceMock{
		getFileByIDFn: func(id int) (files.FileDto, error) {
			return files.FileDto{ID: id, Type: files.File, Format: ".jpg"}, nil
		},
		getFileThumbFn: func(fileDto files.FileDto, width, height int) ([]byte, error) {
			thumbCalls++
			return []byte("thumb"), nil
		},
	}

	videoService := &workerVideoServiceMock{rebuildFn: func() error {
		playlistCalls++
		return nil
	}}

	if err := executeThumbnailStep(&WorkerContext{FilesService: filesService}, jobs.StepModel{Payload: []byte(`{"file_id":10}`)}); err != nil {
		t.Fatalf("unexpected thumbnail step error: %v", err)
	}
	if thumbCalls == 0 {
		t.Fatalf("expected thumbnail generation call")
	}

	if err := executePlaylistIndexStep(&WorkerContext{VideoService: videoService}, jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected playlist step error: %v", err)
	}
	if playlistCalls != 1 {
		t.Fatalf("expected playlist rebuild to be called once, got %d", playlistCalls)
	}
}
