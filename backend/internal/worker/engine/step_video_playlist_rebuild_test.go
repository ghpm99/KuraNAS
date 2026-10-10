package engine

import (
	"database/sql"
	"testing"

	"nas-go/api/internal/api/v1/files"
	jobs "nas-go/api/internal/api/v1/jobs"
	jobdomain "nas-go/api/internal/worker/job"
)

func newPlaylistRebuildContext(jobsRepository *fakeJobsRepository, rebuildCalls *int, onRebuild func()) *WorkerContext {
	return &WorkerContext{
		JobOrchestrator:      NewJobOrchestrator(jobsRepository, nil),
		VideoPlaylistRebuild: NewVideoPlaylistRebuildCoordinator(),
		VideoService: &workerVideoServiceMock{rebuildFn: func() error {
			*rebuildCalls++
			if onRebuild != nil {
				onRebuild()
			}
			return nil
		}},
	}
}

func countRebuildJobs(jobsRepository *fakeJobsRepository) int {
	rebuildJobs := 0
	for _, createdJob := range jobsRepository.jobs {
		if createdJob.Type == string(jobdomain.JobTypeVideoPlaylistRebuild) {
			rebuildJobs++
		}
	}
	return rebuildJobs
}

func TestEnqueueVideoPlaylistRebuildIfVideo_ManyVideosEnqueueOneJob(t *testing.T) {
	jobsRepository := newFakeJobsRepository()
	rebuildCalls := 0
	workerContext := newPlaylistRebuildContext(jobsRepository, &rebuildCalls, nil)

	for videoIndex := 0; videoIndex < 50; videoIndex++ {
		enqueueVideoPlaylistRebuildIfVideo(workerContext, files.FileDto{Path: "/v.mkv", Format: ".mkv"})
	}

	if got := countRebuildJobs(jobsRepository); got != 1 {
		t.Fatalf("expected one coalesced rebuild job, got %d", got)
	}
}

func TestEnqueueVideoPlaylistRebuildIfVideo_IgnoresNonVideo(t *testing.T) {
	jobsRepository := newFakeJobsRepository()
	rebuildCalls := 0
	workerContext := newPlaylistRebuildContext(jobsRepository, &rebuildCalls, nil)

	enqueueVideoPlaylistRebuildIfVideo(workerContext, files.FileDto{Path: "/a.txt", Format: ".txt"})
	enqueueVideoPlaylistRebuildIfVideo(nil, files.FileDto{Path: "/v.mkv", Format: ".mkv"})

	if got := countRebuildJobs(jobsRepository); got != 0 {
		t.Fatalf("expected no rebuild job, got %d", got)
	}
}

func TestExecuteVideoPlaylistRebuildStep_RerunsOnceWhenVideosArriveWhileRunning(t *testing.T) {
	jobsRepository := newFakeJobsRepository()
	rebuildCalls := 0
	var workerContext *WorkerContext
	workerContext = newPlaylistRebuildContext(jobsRepository, &rebuildCalls, func() {
		if rebuildCalls == 1 {
			enqueueVideoPlaylistRebuildIfVideo(workerContext, files.FileDto{Path: "/late.mkv", Format: ".mkv"})
			enqueueVideoPlaylistRebuildIfVideo(workerContext, files.FileDto{Path: "/late2.mkv", Format: ".mkv"})
		}
	})
	enqueueVideoPlaylistRebuildIfVideo(workerContext, files.FileDto{Path: "/first.mkv", Format: ".mkv"})

	if err := executeVideoPlaylistRebuildStep(workerContext, jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rebuildCalls != 2 {
		t.Fatalf("expected one rebuild plus one rerun, got %d", rebuildCalls)
	}
	if got := countRebuildJobs(jobsRepository); got != 1 {
		t.Fatalf("expected a single rebuild job, got %d", got)
	}
}

func TestExecuteVideoPlaylistRebuildStep_NoRerunWithoutNewVideos(t *testing.T) {
	jobsRepository := newFakeJobsRepository()
	rebuildCalls := 0
	workerContext := newPlaylistRebuildContext(jobsRepository, &rebuildCalls, nil)
	enqueueVideoPlaylistRebuildIfVideo(workerContext, files.FileDto{Path: "/first.mkv", Format: ".mkv"})
	enqueueVideoPlaylistRebuildIfVideo(workerContext, files.FileDto{Path: "/second.mkv", Format: ".mkv"})

	if err := executeVideoPlaylistRebuildStep(workerContext, jobs.StepModel{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rebuildCalls != 1 {
		t.Fatalf("expected a single rebuild, got %d", rebuildCalls)
	}
}

func TestExecuteVideoPlaylistRebuildStep_RequiresVideoService(t *testing.T) {
	if err := executeVideoPlaylistRebuildStep(&WorkerContext{}, jobs.StepModel{}); err == nil {
		t.Fatal("expected error without video service")
	}
}

func TestExecutePersistStep_VideoEnqueuesCoalescedRebuild(t *testing.T) {
	jobsRepository := newFakeJobsRepository()
	rebuildCalls := 0
	workerContext := newPlaylistRebuildContext(jobsRepository, &rebuildCalls, nil)
	workerContext.FilesService = &workerFilesServiceMock{
		getFileByNamePathFn: func(name, path string) (files.FileDto, error) {
			return files.FileDto{}, sql.ErrNoRows
		},
		createFileFn: func(fileDto files.FileDto) (files.FileDto, error) { return fileDto, nil },
	}

	for _, payload := range []string{
		`{"file":{"name":"a.mkv","path":"/tmp/a.mkv","parent_path":"/tmp","type":2,"format":".mkv","size":1}}`,
		`{"file":{"name":"b.mkv","path":"/tmp/b.mkv","parent_path":"/tmp","type":2,"format":".mkv","size":1}}`,
	} {
		if err := executePersistStep(workerContext, jobs.StepModel{Payload: []byte(payload)}); err != nil {
			t.Fatalf("unexpected persist error: %v", err)
		}
	}

	if got := countRebuildJobs(jobsRepository); got != 1 {
		t.Fatalf("expected one rebuild job after persisting two videos, got %d", got)
	}
}
