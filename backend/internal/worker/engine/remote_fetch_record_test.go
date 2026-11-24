package engine

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	ingestapi "nas-go/api/internal/api/v1/ingest"
	jobs "nas-go/api/internal/api/v1/jobs"
	jobdomain "nas-go/api/internal/worker/job"
	"nas-go/api/pkg/i18n"
)

type recordedDownload struct {
	state       string
	progress    int
	speed       *int64
	eta         *int
	title       string
	finalPath   string
	failure     string
	transitions []string
}

type fakeDownloadRecorder struct {
	mutex   sync.Mutex
	records map[int]*recordedDownload
}

func newFakeDownloadRecorder() *fakeDownloadRecorder {
	return &fakeDownloadRecorder{records: map[int]*recordedDownload{}}
}

func (recorder *fakeDownloadRecorder) record(downloadID int) *recordedDownload {
	if recorder.records[downloadID] == nil {
		recorder.records[downloadID] = &recordedDownload{state: "queued"}
	}
	return recorder.records[downloadID]
}

func (recorder *fakeDownloadRecorder) MarkRunning(downloadID int) error {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()
	stored := recorder.record(downloadID)
	stored.state = "running"
	stored.transitions = append(stored.transitions, "running")
	return nil
}

func (recorder *fakeDownloadRecorder) UpdateProgress(downloadID int, progress ingestapi.DownloadProgress) error {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()
	stored := recorder.record(downloadID)
	stored.progress, stored.speed, stored.eta = progress.Percent, progress.SpeedBytesPerSecond, progress.ETASeconds
	return nil
}

func (recorder *fakeDownloadRecorder) UpdateTitle(downloadID int, title string) error {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()
	recorder.record(downloadID).title = title
	return nil
}

func (recorder *fakeDownloadRecorder) MarkCompleted(downloadID int, title string, finalPath string) error {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()
	stored := recorder.record(downloadID)
	stored.state, stored.progress, stored.title, stored.finalPath = "completed", 100, title, finalPath
	stored.transitions = append(stored.transitions, "completed")
	return nil
}

func (recorder *fakeDownloadRecorder) MarkFailed(downloadID int, translatedError string) error {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()
	stored := recorder.record(downloadID)
	stored.state, stored.failure = "failed", translatedError
	stored.transitions = append(stored.transitions, "failed")
	return nil
}

func (recorder *fakeDownloadRecorder) MarkCanceled(downloadID int) error {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()
	stored := recorder.record(downloadID)
	stored.state = "canceled"
	stored.transitions = append(stored.transitions, "canceled")
	return nil
}

func (recorder *fakeDownloadRecorder) snapshot(downloadID int) recordedDownload {
	recorder.mutex.Lock()
	defer recorder.mutex.Unlock()
	return *recorder.record(downloadID)
}

func trackedFetchStep(t *testing.T, downloadID int, fakeYtDlpScript string) (jobs.StepModel, string) {
	t.Helper()
	outputDir := t.TempDir()
	rawPayload, _ := json.Marshal(ingestapi.RemoteFetchStepPayload{
		DownloadID: downloadID,
		URL:        "https://x.test/clip",
		Preset:     "audio_mp3",
		OutputDir:  outputDir,
		Binary:     writeFakeYtDlp(t, fakeYtDlpScript),
	})
	return jobs.StepModel{Payload: rawPayload}, outputDir
}

func TestExecuteRemoteFetchStepRecordsACompletedDownload(t *testing.T) {
	script := fakeYtDlpReadsOutputDir + `echo "[download] Destination: $output_dir/Clip_Title.mp3"
echo "[download]  42.3% of ~12.34MiB at 1.00MiB/s ETA 00:12"
echo "[download] 100% of 12.34MiB at 2.00MiB/s ETA 00:00"
echo audio > "$output_dir/Clip_Title.mp3"
`
	fetchStep, outputDir := trackedFetchStep(t, 7, script)
	recorder := newFakeDownloadRecorder()

	if err := executeRemoteFetchStep(&WorkerContext{DownloadRecorder: recorder}, fetchStep); err != nil {
		t.Fatalf("download failed: %v", err)
	}

	stored := recorder.snapshot(7)
	if stored.state != "completed" || stored.progress != 100 || stored.title != "Clip_Title" {
		t.Fatalf("unexpected record: %+v", stored)
	}
	if !strings.HasPrefix(stored.finalPath, outputDir) || !strings.HasSuffix(stored.finalPath, "Clip_Title.mp3") {
		t.Fatalf("final path = %q", stored.finalPath)
	}
	if len(stored.transitions) != 2 || stored.transitions[0] != "running" {
		t.Fatalf("transitions = %v, want running then completed", stored.transitions)
	}
}

func TestExecuteRemoteFetchStepRecordsSpeedAndEtaOfTheProgressLine(t *testing.T) {
	script := fakeYtDlpReadsOutputDir + `echo "[download]  42.3% of ~12.34MiB at 1.00MiB/s ETA 00:12"
echo audio > "$output_dir/clip.mp3"
`
	fetchStep, _ := trackedFetchStep(t, 8, script)
	recorder := newFakeDownloadRecorder()
	progressSeen := &recordedDownload{}
	observing := &progressObservingRecorder{fakeDownloadRecorder: recorder, seen: progressSeen}

	if err := executeRemoteFetchStep(&WorkerContext{DownloadRecorder: observing}, fetchStep); err != nil {
		t.Fatalf("download failed: %v", err)
	}

	if progressSeen.progress != 42 || progressSeen.speed == nil || *progressSeen.speed != 1<<20 || progressSeen.eta == nil || *progressSeen.eta != 12 {
		t.Fatalf("progress line not recorded: %+v", progressSeen)
	}
}

type progressObservingRecorder struct {
	*fakeDownloadRecorder
	seen *recordedDownload
}

func (recorder *progressObservingRecorder) UpdateProgress(downloadID int, progress ingestapi.DownloadProgress) error {
	recorder.seen.progress, recorder.seen.speed, recorder.seen.eta = progress.Percent, progress.SpeedBytesPerSecond, progress.ETASeconds
	return recorder.fakeDownloadRecorder.UpdateProgress(downloadID, progress)
}

func TestExecuteRemoteFetchStepRecordsAFailedDownloadWithATranslatedErrorWithoutPaths(t *testing.T) {
	fetchStep, outputDir := trackedFetchStep(t, 9, "echo 'ERROR: /srv/secret/path unavailable' >&2\nexit 1\n")
	recorder := newFakeDownloadRecorder()

	if err := executeRemoteFetchStep(&WorkerContext{DownloadRecorder: recorder}, fetchStep); err == nil {
		t.Fatal("expected the download to fail")
	}

	stored := recorder.snapshot(9)
	if stored.state != "failed" || stored.failure != i18n.GetMessage("INGEST_DOWNLOAD_ERROR_TOOL_FAILED") {
		t.Fatalf("unexpected record: %+v", stored)
	}
	if strings.Contains(stored.failure, "/srv/secret/path") || strings.Contains(stored.failure, outputDir) {
		t.Fatalf("failure leaks a server path: %q", stored.failure)
	}
}

func TestExecuteRemoteFetchStepRecordsADownloadWithoutOutputFileAsFailed(t *testing.T) {
	fetchStep, _ := trackedFetchStep(t, 10, "exit 0\n")
	recorder := newFakeDownloadRecorder()

	if err := executeRemoteFetchStep(&WorkerContext{DownloadRecorder: recorder}, fetchStep); err == nil {
		t.Fatal("expected the download to fail")
	}

	if stored := recorder.snapshot(10); stored.state != "failed" || stored.failure != i18n.GetMessage("INGEST_DOWNLOAD_ERROR_NO_OUTPUT_FILE") {
		t.Fatalf("unexpected record: %+v", stored)
	}
}

func TestExecuteRemoteFetchStepRecordsAMissingBinaryAsFailed(t *testing.T) {
	rawPayload, _ := json.Marshal(ingestapi.RemoteFetchStepPayload{
		DownloadID: 11, URL: "https://x.test/clip", Preset: "audio_mp3", OutputDir: t.TempDir(), Binary: "/nonexistent/yt-dlp",
	})
	recorder := newFakeDownloadRecorder()

	if err := executeRemoteFetchStep(&WorkerContext{DownloadRecorder: recorder}, jobs.StepModel{Payload: rawPayload}); err == nil {
		t.Fatal("expected the download to fail")
	}

	if stored := recorder.snapshot(11); stored.state != "failed" || stored.failure != i18n.GetMessage("INGEST_DOWNLOAD_ERROR_TOOL_UNAVAILABLE") {
		t.Fatalf("unexpected record: %+v", stored)
	}
}

func TestExecuteRemoteFetchStepRecordsACanceledDownload(t *testing.T) {
	originalPollInterval := jobCancellationPollInterval
	jobCancellationPollInterval = 10 * time.Millisecond
	t.Cleanup(func() { jobCancellationPollInterval = originalPollInterval })

	repository := newFakeJobsRepository()
	recorder := newFakeDownloadRecorder()
	fetchStep := runningRemoteFetchJob(t, repository, ingestapi.RemoteFetchStepPayload{
		DownloadID: 12,
		URL:        "https://x.test/clip",
		Preset:     "audio_mp3",
		OutputDir:  t.TempDir(),
		Binary:     writeFakeYtDlp(t, "exec sleep 30\n"),
	})
	stepOutcome := make(chan error, 1)
	go func() {
		stepOutcome <- executeRemoteFetchStep(&WorkerContext{JobsRepository: repository, DownloadRecorder: recorder}, fetchStep)
	}()
	waitForDownloadState(t, recorder, 12, "running")

	cancelRequested := true
	if _, err := repository.UpdateJobExecution(nil, fetchStep.JobID, string(jobdomain.JobStatusCanceled), nil, nil, &cancelRequested, nil); err != nil {
		t.Fatalf("cancel job: %v", err)
	}

	select {
	case err := <-stepOutcome:
		if !errors.Is(err, ErrStepCanceled) {
			t.Fatalf("expected ErrStepCanceled, got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the canceled download kept running")
	}
	if stored := recorder.snapshot(12); stored.state != "canceled" {
		t.Fatalf("unexpected record: %+v", stored)
	}
}

func waitForDownloadState(t *testing.T, recorder *fakeDownloadRecorder, downloadID int, wantState string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if recorder.snapshot(downloadID).state == wantState {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("download %d never reached state %q", downloadID, wantState)
}

func TestExecuteRemoteFetchStepWithoutDownloadIdTouchesNoRecord(t *testing.T) {
	fetchStep, _ := trackedFetchStep(t, 0, fakeYtDlpReadsOutputDir+"echo audio > \"$output_dir/clip.mp3\"\n")
	recorder := newFakeDownloadRecorder()

	if err := executeRemoteFetchStep(&WorkerContext{DownloadRecorder: recorder}, fetchStep); err != nil {
		t.Fatalf("download failed: %v", err)
	}

	if len(recorder.records) != 0 {
		t.Fatalf("a payload without download id must not touch records: %+v", recorder.records)
	}
}
