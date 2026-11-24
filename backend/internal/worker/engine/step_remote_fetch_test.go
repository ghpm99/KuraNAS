package engine

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	ingestapi "nas-go/api/internal/api/v1/ingest"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/notifications"
	jobdomain "nas-go/api/internal/worker/job"
)

func TestExecuteRemoteFetchStepEarlyErrors(t *testing.T) {
	step := func(p ingestapi.RemoteFetchStepPayload) jobs.StepModel {
		raw, _ := json.Marshal(p)
		return jobs.StepModel{Payload: raw}
	}

	if err := executeRemoteFetchStep(nil, jobs.StepModel{Payload: []byte("{bad")}); err == nil {
		t.Fatal("expected a decode error for malformed payload")
	}
	if err := executeRemoteFetchStep(nil, step(ingestapi.RemoteFetchStepPayload{URL: "https://x.test/v", Preset: "nope", OutputDir: t.TempDir()})); err == nil {
		t.Fatal("expected an error for an unknown preset")
	}
	if err := executeRemoteFetchStep(nil, step(ingestapi.RemoteFetchStepPayload{URL: "https://x.test/v", Preset: "audio_mp3", OutputDir: ""})); err == nil {
		t.Fatal("expected an error for an empty output dir")
	}
	// A missing binary makes cmd.Start fail — exercises the exec path without
	// needing a real yt-dlp installed.
	err := executeRemoteFetchStep(nil, step(ingestapi.RemoteFetchStepPayload{
		URL:       "https://x.test/v",
		Preset:    "audio_mp3",
		OutputDir: t.TempDir(),
		Binary:    "kuranas-no-such-binary-xyz",
	}))
	if err == nil {
		t.Fatal("expected an error when the yt-dlp binary is missing")
	}
}

func TestBuildYtDlpArgs(t *testing.T) {
	args := buildYtDlpArgs([]string{"-x", "--audio-format", "mp3"}, "/tmp/out", "https://x.test/v")
	joined := strings.Join(args, " ")

	for _, want := range []string{"-x", "--audio-format", "mp3", "--no-playlist", "--restrict-filenames", "--no-overwrites", "--newline", "https://x.test/v"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("args missing %q: %v", want, args)
		}
	}
	// The URL must be last so flags never get parsed as the positional arg.
	if args[len(args)-1] != "https://x.test/v" {
		t.Fatalf("expected url last, got %q", args[len(args)-1])
	}
	if !strings.Contains(joined, filepath.Join("/tmp/out", "%(title)s.%(ext)s")) {
		t.Fatalf("output template not rooted at output dir: %v", args)
	}
}

func TestParseDownloadProgress(t *testing.T) {
	cases := []struct {
		line    string
		wantPct int
		wantOK  bool
	}{
		{"[download]   0.0% of 10MiB", 0, true},
		{"[download]  45.6% of 10MiB at 1MiB/s", 45, true},
		{"[download] 100% of 10MiB", 100, true},
		{"[info] something else", 0, false},
		{"no marker 50%", 0, false},
		{"[download] no percent here", 0, false},
	}
	for _, tc := range cases {
		pct, ok := parseDownloadProgress(tc.line)
		if ok != tc.wantOK || (ok && pct != tc.wantPct) {
			t.Errorf("parseDownloadProgress(%q) = (%d,%v), want (%d,%v)", tc.line, pct, ok, tc.wantPct, tc.wantOK)
		}
	}
}

func TestIsTempArtifact(t *testing.T) {
	for _, name := range []string{"video.mp4.part", "audio.m4a.ytdl", "x.TEMP", "y.tmp"} {
		if !isTempArtifact(name) {
			t.Errorf("expected %q to be a temp artifact", name)
		}
	}
	for _, name := range []string{"song.mp3", "movie.mkv"} {
		if isTempArtifact(name) {
			t.Errorf("expected %q to be a finished file", name)
		}
	}
}

func TestMoveFetchedFilesSkipsTempAndHidden(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	write := func(dir, name string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(src, "song.mp3")
	write(src, "song.mp3.part")
	write(src, ".hidden")

	moved, err := moveFetchedFiles(src, dst)
	if err != nil {
		t.Fatalf("moveFetchedFiles: %v", err)
	}
	if moved != 1 {
		t.Fatalf("expected 1 moved, got %d", moved)
	}
	if _, err := os.Stat(filepath.Join(dst, "song.mp3")); err != nil {
		t.Fatalf("finished file not moved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "song.mp3.part")); !os.IsNotExist(err) {
		t.Fatal("temp artifact should not have moved")
	}
}

func TestMoveFileCopyFallback(t *testing.T) {
	src := filepath.Join(t.TempDir(), "a.txt")
	dst := filepath.Join(t.TempDir(), "b.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := moveFile(src, dst); err != nil {
		t.Fatalf("moveFile: %v", err)
	}
	data, err := os.ReadFile(dst)
	if err != nil || string(data) != "hello" {
		t.Fatalf("dst content wrong: %q, %v", data, err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("src should be gone after move")
	}
}

func TestLastLine(t *testing.T) {
	if got := lastLine("a\nb\n  c  "); got != "c" {
		t.Fatalf("expected 'c', got %q", got)
	}
	if got := lastLine("solo"); got != "solo" {
		t.Fatalf("expected 'solo', got %q", got)
	}
}

func TestDecodeRemoteFetchPayload(t *testing.T) {
	good, _ := json.Marshal(ingestapi.RemoteFetchStepPayload{URL: "https://x.test/v", Preset: "audio_mp3", OutputDir: "/o"})
	payload, err := decodeRemoteFetchPayload(good)
	if err != nil || payload.URL != "https://x.test/v" {
		t.Fatalf("good payload: %+v, %v", payload, err)
	}

	if _, err := decodeRemoteFetchPayload([]byte(`{"preset":"audio_mp3"}`)); err == nil {
		t.Fatal("expected error for payload without url")
	}
	if _, err := decodeRemoteFetchPayload([]byte(`{bad`)); err == nil {
		t.Fatal("expected error for malformed json")
	}
}

func writeFakeYtDlp(t *testing.T, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake yt-dlp is a POSIX shell script")
	}
	binaryPath := filepath.Join(t.TempDir(), "fake-yt-dlp")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("write fake yt-dlp: %v", err)
	}
	return binaryPath
}

func runningRemoteFetchJob(t *testing.T, repository *fakeJobsRepository, payload ingestapi.RemoteFetchStepPayload) jobs.StepModel {
	t.Helper()
	fetchJob, err := repository.CreateJob(nil, jobs.JobModel{Type: string(jobdomain.JobTypeRemoteFetch), Status: string(jobdomain.JobStatusRunning)})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	rawPayload, _ := json.Marshal(payload)
	fetchStep, err := repository.CreateStep(nil, jobs.StepModel{
		JobID:       fetchJob.ID,
		Type:        string(jobdomain.StepTypeRemoteFetch),
		Status:      string(jobdomain.StepStatusRunning),
		MaxAttempts: 1,
		Payload:     rawPayload,
	})
	if err != nil {
		t.Fatalf("create step: %v", err)
	}
	return fetchStep
}

func TestExecuteRemoteFetchStepStopsTheDownloadWhenTheJobIsCanceled(t *testing.T) {
	originalPollInterval := jobCancellationPollInterval
	jobCancellationPollInterval = 20 * time.Millisecond
	t.Cleanup(func() { jobCancellationPollInterval = originalPollInterval })

	outputDir := t.TempDir()
	repository := newFakeJobsRepository()
	fetchStep := runningRemoteFetchJob(t, repository, ingestapi.RemoteFetchStepPayload{
		URL:       "https://x.test/v",
		Preset:    "audio_mp3",
		OutputDir: outputDir,
		Binary:    writeFakeYtDlp(t, "echo '[download]  10.0% of 1MiB'\nexec sleep 60\n"),
	})

	stepOutcome := make(chan error, 1)
	go func() {
		stepOutcome <- executeRemoteFetchStep(&WorkerContext{JobsRepository: repository}, fetchStep)
	}()

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
		t.Fatal("the download kept running after the job was canceled")
	}

	if leftovers, _ := os.ReadDir(outputDir); len(leftovers) != 0 {
		t.Fatalf("a canceled download must leave nothing in the library, found %d entries", len(leftovers))
	}
}

func TestSchedulerRecordsACanceledStepAsCanceledNotFailed(t *testing.T) {
	repository := newFakeJobsRepository()
	fetchStep := runningRemoteFetchJob(t, repository, ingestapi.RemoteFetchStepPayload{URL: "https://x.test/v"})
	fetchStep.Status = string(jobdomain.StepStatusQueued)
	scheduler := &JobScheduler{
		repository: repository,
		executors: map[jobdomain.StepType]StepExecutor{
			jobdomain.StepTypeRemoteFetch: func(jobs.StepModel) error { return ErrStepCanceled },
		},
	}

	if err := scheduler.executeStep(fetchStep); err != nil {
		t.Fatalf("a canceled step is not a scheduler error, got %v", err)
	}

	steps, _ := repository.GetStepsByJobID(fetchStep.JobID)
	if steps[0].Status != string(jobdomain.StepStatusCanceled) || steps[0].LastError != "" {
		t.Fatalf("expected a canceled step without error, got status=%s error=%q", steps[0].Status, steps[0].LastError)
	}
}

func TestExecuteRemoteFetchStepFailsADownloadThatOutlivesItsTimeout(t *testing.T) {
	originalTimeout := remoteFetchTimeout
	remoteFetchTimeout = func() time.Duration { return 50 * time.Millisecond }
	t.Cleanup(func() { remoteFetchTimeout = originalTimeout })

	outputDir := t.TempDir()
	rawPayload, _ := json.Marshal(ingestapi.RemoteFetchStepPayload{
		URL:       "https://x.test/v",
		Preset:    "audio_mp3",
		OutputDir: outputDir,
		Binary:    writeFakeYtDlp(t, "exec sleep 60\n"),
	})

	err := executeRemoteFetchStep(nil, jobs.StepModel{Payload: rawPayload})

	if err == nil || !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("expected a timeout failure, got %v", err)
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrStepCanceled) {
		t.Fatalf("a download timeout must fail the step, not requeue or cancel it: %v", err)
	}
	if leftovers, _ := os.ReadDir(outputDir); len(leftovers) != 0 {
		t.Fatalf("a timed out download must leave nothing in the library, found %d entries", len(leftovers))
	}
}

func TestMoveFetchedFilesKeepsWhatIsAlreadyInTheLibrary(t *testing.T) {
	downloadDir := t.TempDir()
	libraryDir := t.TempDir()
	writeTextFile := func(path string, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	readTextFile := func(path string) string {
		t.Helper()
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(content)
	}
	writeTextFile(filepath.Join(libraryDir, "clip.mp4"), "first download")
	writeTextFile(filepath.Join(libraryDir, "clip (2).mp4"), "second download")
	writeTextFile(filepath.Join(downloadDir, "clip.mp4"), "third download")

	movedCount, err := moveFetchedFiles(downloadDir, libraryDir)
	if err != nil || movedCount != 1 {
		t.Fatalf("expected one file moved, got count=%d err=%v", movedCount, err)
	}

	if readTextFile(filepath.Join(libraryDir, "clip.mp4")) != "first download" ||
		readTextFile(filepath.Join(libraryDir, "clip (2).mp4")) != "second download" {
		t.Fatal("files already in the library must keep their content")
	}
	if readTextFile(filepath.Join(libraryDir, "clip (3).mp4")) != "third download" {
		t.Fatal("the new download must land under the next free name")
	}
}

const fakeYtDlpReadsOutputDir = `while [ $# -gt 0 ]; do
  if [ "$1" = "-o" ]; then output_dir=$(dirname "$2"); fi
  shift
done
`

func TestExecuteRemoteFetchStepSurvivesAnOutputLineBeyondTheProgressBuffer(t *testing.T) {
	originalTimeout := remoteFetchTimeout
	remoteFetchTimeout = func() time.Duration { return 10 * time.Second }
	t.Cleanup(func() { remoteFetchTimeout = originalTimeout })

	outputDir := t.TempDir()
	rawPayload, _ := json.Marshal(ingestapi.RemoteFetchStepPayload{
		URL:       "https://x.test/v",
		Preset:    "audio_mp3",
		OutputDir: outputDir,
		Binary: writeFakeYtDlp(t, fakeYtDlpReadsOutputDir+
			"head -c 3000000 /dev/zero | tr '\\0' 'a'\necho\necho audio > \"$output_dir/clip.mp3\"\n"),
	})

	if err := executeRemoteFetchStep(nil, jobs.StepModel{Payload: rawPayload}); err != nil {
		t.Fatalf("a very long output line must not break the download, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "clip.mp3")); err != nil {
		t.Fatalf("the downloaded file must reach the library: %v", err)
	}
}

func TestCopyFileContentWritesTheWholeFile(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "clip.mp4")
	stagingPath := filepath.Join(t.TempDir(), "clip.mp4"+stagedCopySuffix)
	if err := os.WriteFile(sourcePath, []byte("whole download"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyFileContent(sourcePath, stagingPath); err != nil {
		t.Fatalf("copyFileContent: %v", err)
	}

	copied, err := os.ReadFile(stagingPath)
	if err != nil || string(copied) != "whole download" {
		t.Fatalf("staged copy content wrong: %q, %v", copied, err)
	}
	if !isTempArtifact(filepath.Base(stagingPath)) {
		t.Fatal("a staged copy must be recognised as an unfinished artifact")
	}
}

func TestMoveFileLeavesNoStagedCopyBehindWhenItFails(t *testing.T) {
	libraryDir := t.TempDir()
	destinationPath := filepath.Join(libraryDir, "clip.mp4")

	if err := moveFile(filepath.Join(t.TempDir(), "never-downloaded.mp4"), destinationPath); err == nil {
		t.Fatal("moving a missing file must fail")
	}

	if leftovers, _ := os.ReadDir(libraryDir); len(leftovers) != 0 {
		t.Fatalf("a failed move must leave the library untouched, found %d entries", len(leftovers))
	}
}

type recordingNotificationService struct {
	notifications.ServiceInterface
	raised []notifications.CreateNotificationDto
}

func (service *recordingNotificationService) GroupOrCreate(dto notifications.CreateNotificationDto) (notifications.NotificationDto, error) {
	service.raised = append(service.raised, dto)
	return notifications.NotificationDto{}, nil
}

func remoteFetchStepRunning(t *testing.T, fakeYtDlpScript string) (jobs.StepModel, *recordingNotificationService) {
	t.Helper()
	rawPayload, _ := json.Marshal(ingestapi.RemoteFetchStepPayload{
		URL:       "https://x.test/clip",
		Preset:    "audio_mp3",
		OutputDir: t.TempDir(),
		Binary:    writeFakeYtDlp(t, fakeYtDlpScript),
	})
	return jobs.StepModel{Payload: rawPayload}, &recordingNotificationService{}
}

func TestExecuteRemoteFetchStepNotifiesAFinishedDownload(t *testing.T) {
	fetchStep, notifier := remoteFetchStepRunning(t, fakeYtDlpReadsOutputDir+"echo audio > \"$output_dir/clip.mp3\"\n")

	if err := executeRemoteFetchStep(&WorkerContext{NotificationService: notifier}, fetchStep); err != nil {
		t.Fatalf("download failed: %v", err)
	}

	if len(notifier.raised) != 1 || notifier.raised[0].Type != string(notifications.NotificationTypeSuccess) {
		t.Fatalf("expected one success notification, got %+v", notifier.raised)
	}
}

func TestExecuteRemoteFetchStepNotifiesAFailedDownloadWithoutTheRawError(t *testing.T) {
	fetchStep, notifier := remoteFetchStepRunning(t, "echo 'ERROR: /srv/secret/path unavailable' >&2\nexit 1\n")

	if err := executeRemoteFetchStep(&WorkerContext{NotificationService: notifier}, fetchStep); err == nil {
		t.Fatal("expected the download to fail")
	}

	if len(notifier.raised) != 1 || notifier.raised[0].Type != string(notifications.NotificationTypeError) {
		t.Fatalf("expected one error notification, got %+v", notifier.raised)
	}
	if strings.Contains(notifier.raised[0].Message, "/srv/secret/path") {
		t.Fatalf("the notification must not carry the raw yt-dlp error: %q", notifier.raised[0].Message)
	}
}

func TestNotifyRemoteFetchOutcomeStaysSilentForACanceledDownload(t *testing.T) {
	fetchStep, notifier := remoteFetchStepRunning(t, "exit 0\n")

	notifyRemoteFetchOutcome(&WorkerContext{NotificationService: notifier}, fetchStep, ErrStepCanceled)

	if len(notifier.raised) != 0 {
		t.Fatalf("a canceled download must not notify, got %+v", notifier.raised)
	}
}

func TestExecuteRemoteFetchStepKeepsServerPathsOutOfTheStoredError(t *testing.T) {
	fetchStep, _ := remoteFetchStepRunning(t, fakeYtDlpReadsOutputDir+"echo \"ERROR: unable to write $output_dir/clip.mp3\" >&2\nexit 1\n")

	err := executeRemoteFetchStep(nil, fetchStep)

	if err == nil || !strings.Contains(err.Error(), "unable to write clip.mp3") {
		t.Fatalf("expected the yt-dlp reason with a bare file name, got %v", err)
	}
	if strings.Contains(err.Error(), os.TempDir()) {
		t.Fatalf("the stored error must not expose server directories: %v", err)
	}
}

func TestWithoutServerPathsIgnoresEmptyDirs(t *testing.T) {
	if cleaned := withoutServerPaths("failed at /data/media/clip.mp3", "", "/data/media"); cleaned != "failed at clip.mp3" {
		t.Fatalf("unexpected message: %q", cleaned)
	}
}

func remoteFetchStepInto(t *testing.T, libraryDir string, fakeYtDlpScript string) jobs.StepModel {
	t.Helper()
	rawPayload, _ := json.Marshal(ingestapi.RemoteFetchStepPayload{
		URL:       "https://x.test/clip",
		Preset:    "audio_mp3",
		OutputDir: libraryDir,
		Binary:    writeFakeYtDlp(t, fakeYtDlpScript),
	})
	return jobs.StepModel{Payload: rawPayload}
}

func namesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func TestExecuteRemoteFetchStepPlacesTheDownloadedContentInTheLibrary(t *testing.T) {
	libraryDir := t.TempDir()
	fetchStep := remoteFetchStepInto(t, libraryDir, fakeYtDlpReadsOutputDir+"printf 'audio-bytes' > \"$output_dir/clip.mp3\"\n")

	if err := executeRemoteFetchStep(nil, fetchStep); err != nil {
		t.Fatalf("download failed: %v", err)
	}

	placedContent, err := os.ReadFile(filepath.Join(libraryDir, "clip.mp3"))
	if err != nil || string(placedContent) != "audio-bytes" {
		t.Fatalf("expected the finished file in the library, got %q err=%v", placedContent, err)
	}
}

func TestExecuteRemoteFetchStepFailsWithItsOwnReasonWhenNoFileIsProduced(t *testing.T) {
	libraryDir := t.TempDir()
	fetchStep := remoteFetchStepInto(t, libraryDir, "exit 0\n")

	err := executeRemoteFetchStep(nil, fetchStep)

	if err == nil || !strings.Contains(err.Error(), "produced no output file") {
		t.Fatalf("expected the no-output failure, got %v", err)
	}
	if strings.Contains(err.Error(), libraryDir) {
		t.Fatalf("the failure must not expose the library path: %v", err)
	}
	if leftovers := namesIn(t, libraryDir); len(leftovers) != 0 {
		t.Fatalf("nothing may reach the library, got %v", leftovers)
	}
}

func TestExecuteRemoteFetchStepNeverPlacesAPartialDownloadInTheLibrary(t *testing.T) {
	libraryDir := t.TempDir()
	fetchStep := remoteFetchStepInto(t, libraryDir, fakeYtDlpReadsOutputDir+
		"echo half > \"$output_dir/clip.mp3.part\"\necho half > \"$output_dir/clip.ytdl\"\nexit 0\n")

	err := executeRemoteFetchStep(nil, fetchStep)

	if err == nil || !strings.Contains(err.Error(), "produced no output file") {
		t.Fatalf("a run that leaves only partial files must fail, got %v", err)
	}
	if leftovers := namesIn(t, libraryDir); len(leftovers) != 0 {
		t.Fatalf("partial files must never reach the library, got %v", leftovers)
	}
}

func TestExecuteRemoteFetchStepLeavesTheLibraryUntouchedWhenTheDownloadFails(t *testing.T) {
	libraryDir := t.TempDir()
	fetchStep := remoteFetchStepInto(t, libraryDir, fakeYtDlpReadsOutputDir+
		"echo half > \"$output_dir/clip.mp3.part\"\necho \"ERROR: unable to download $output_dir/clip.mp3\" >&2\nexit 1\n")

	err := executeRemoteFetchStep(nil, fetchStep)

	if err == nil || strings.Contains(err.Error(), libraryDir) || strings.Contains(err.Error(), os.TempDir()) {
		t.Fatalf("expected a failure without server paths, got %v", err)
	}
	if leftovers := namesIn(t, libraryDir); len(leftovers) != 0 {
		t.Fatalf("a failed download must leave the library empty, got %v", leftovers)
	}
}

func TestExecuteRemoteFetchStepNeverOverwritesAnExistingLibraryFile(t *testing.T) {
	libraryDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libraryDir, "clip.mp3"), []byte("already here"), 0o644); err != nil {
		t.Fatalf("seed library: %v", err)
	}
	fetchStep := remoteFetchStepInto(t, libraryDir, fakeYtDlpReadsOutputDir+"printf 'new download' > \"$output_dir/clip.mp3\"\n")

	if err := executeRemoteFetchStep(nil, fetchStep); err != nil {
		t.Fatalf("download failed: %v", err)
	}

	keptContent, _ := os.ReadFile(filepath.Join(libraryDir, "clip.mp3"))
	numberedContent, _ := os.ReadFile(filepath.Join(libraryDir, "clip (2).mp3"))
	if string(keptContent) != "already here" || string(numberedContent) != "new download" {
		t.Fatalf("expected the original kept and the new file numbered, got %q and %q", keptContent, numberedContent)
	}
}
