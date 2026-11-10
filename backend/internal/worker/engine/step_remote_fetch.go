package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	ingestapi "nas-go/api/internal/api/v1/ingest"
	jobs "nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
)

// downloadProgressRe extracts the percentage from a yt-dlp `[download]` line
// (e.g. "[download]  45.6% of 12.3MiB at ...").
var downloadProgressRe = regexp.MustCompile(`(\d{1,3}(?:\.\d+)?)%`)

// remoteFetchTimeout tells how long one download may run before it is killed.
var remoteFetchTimeout = config.IngestFetchTimeout

// killedDownloadPipeGrace bounds how long a canceled download may keep its
// output pipes open through a child process before they are closed by force.
const killedDownloadPipeGrace = 5 * time.Second

// tempArtifactExts are the in-flight suffixes yt-dlp/ffmpeg leave behind; they
// must never be moved into the library as if they were finished media.
var tempArtifactExts = []string{".part", ".ytdl", ".temp", ".tmp"}

// executeRemoteFetchStep runs yt-dlp to pull a URL into a private temp dir, then
// atomically moves the finished file(s) into the target storage root. Dropping
// them into a watched root is enough: the folder watcher picks them up and the
// normal metadata/checksum/thumbnail/persist pipeline indexes them — so this
// step owns only the download + placement, nothing about indexing.
func executeRemoteFetchStep(workerContext *WorkerContext, step jobs.StepModel) error {
	fetchErr := runRemoteFetch(workerContext, step)
	notifyRemoteFetchOutcome(workerContext, step, fetchErr)
	return fetchErr
}

func runRemoteFetch(workerContext *WorkerContext, step jobs.StepModel) error {
	payload, err := decodeRemoteFetchPayload(step.Payload)
	if err != nil {
		return err
	}

	presetArgs, ok := ingestapi.ResolvePreset(payload.Preset)
	if !ok {
		return fmt.Errorf("remote fetch: unknown preset %q", payload.Preset)
	}
	if strings.TrimSpace(payload.OutputDir) == "" {
		return fmt.Errorf("remote fetch: output dir is required")
	}
	if err := os.MkdirAll(payload.OutputDir, 0o755); err != nil {
		return fmt.Errorf("remote fetch: create output dir: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "kuranas-fetch-")
	if err != nil {
		return fmt.Errorf("remote fetch: temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	binary := strings.TrimSpace(payload.Binary)
	if binary == "" {
		binary = "yt-dlp"
	}
	args := buildYtDlpArgs(presetArgs, tmpDir, payload.URL)

	jobContext, stopWatchingJob := contextCanceledWithJob(workerContext, step.JobID)
	defer stopWatchingJob()
	downloadTimeout := remoteFetchTimeout()
	downloadContext, releaseDownloadTimer := context.WithTimeout(jobContext, downloadTimeout)
	defer releaseDownloadTimer()

	cmd := exec.CommandContext(downloadContext, binary, args...)
	cmd.WaitDelay = killedDownloadPipeGrace
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("remote fetch: stdout pipe: %w", err)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("remote fetch: start yt-dlp: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lastReported := -1
	lastReportAt := time.Time{}
	for scanner.Scan() {
		if pct, ok := parseDownloadProgress(scanner.Text()); ok {
			if shouldReportProgress(pct, lastReported, lastReportAt) {
				reportPullProgress(workerContext, step, pct)
				lastReported = pct
				lastReportAt = time.Now()
			}
		}
	}

	discardOutputLeftUnread(stdout)

	if err := cmd.Wait(); err != nil {
		if jobContext.Err() != nil {
			return ErrStepCanceled
		}
		if downloadContext.Err() != nil {
			return fmt.Errorf("remote fetch: download exceeded the %s limit", downloadTimeout)
		}
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("remote fetch: yt-dlp failed: %s", withoutServerPaths(lastLine(msg), tmpDir, payload.OutputDir))
		}
		return fmt.Errorf("remote fetch: yt-dlp failed: %w", err)
	}

	moved, err := moveFetchedFiles(tmpDir, payload.OutputDir)
	if err != nil {
		return fmt.Errorf("remote fetch: move output: %s", withoutServerPaths(err.Error(), tmpDir, payload.OutputDir))
	}
	if moved == 0 {
		return fmt.Errorf("remote fetch: yt-dlp produced no output file")
	}
	return nil
}

// discardOutputLeftUnread keeps the download from stalling on a full pipe when
// the progress reader gave up early, as it does on a line beyond its buffer.
func discardOutputLeftUnread(downloadOutput io.Reader) {
	_, _ = io.Copy(io.Discard, downloadOutput)
}

// notifyRemoteFetchOutcome tells the user a download ended, well or badly. A
// canceled download stays silent: whoever canceled it already knows.
func notifyRemoteFetchOutcome(workerContext *WorkerContext, step jobs.StepModel, fetchErr error) {
	if workerContext == nil || workerContext.NotificationService == nil || errors.Is(fetchErr, ErrStepCanceled) {
		return
	}
	payload, err := decodeRemoteFetchPayload(step.Payload)
	if err != nil {
		return
	}
	outcomeNotification := notifications.CreateNotificationDto{
		Type:    string(notifications.NotificationTypeSuccess),
		Title:   i18n.GetMessage("NOTIFICATION_REMOTE_FETCH_COMPLETED_TITLE"),
		Message: i18n.Translate("NOTIFICATION_REMOTE_FETCH_COMPLETED_MESSAGE", payload.URL),
	}
	if fetchErr != nil {
		outcomeNotification = notifications.CreateNotificationDto{
			Type:    string(notifications.NotificationTypeError),
			Title:   i18n.GetMessage("NOTIFICATION_REMOTE_FETCH_FAILED_TITLE"),
			Message: i18n.Translate("NOTIFICATION_REMOTE_FETCH_FAILED_MESSAGE", payload.URL),
		}
	}
	_, _ = workerContext.NotificationService.GroupOrCreate(outcomeNotification)
}

func decodeRemoteFetchPayload(raw []byte) (ingestapi.RemoteFetchStepPayload, error) {
	var payload ingestapi.RemoteFetchStepPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return payload, fmt.Errorf("decode remote fetch payload: %w", err)
	}
	if strings.TrimSpace(payload.URL) == "" {
		return payload, fmt.Errorf("remote fetch payload url is required")
	}
	return payload, nil
}

// buildYtDlpArgs composes the preset's format args with the fixed safety flags
// and an output template rooted at outputDir.
func buildYtDlpArgs(presetArgs []string, outputDir, url string) []string {
	args := make([]string, 0, len(presetArgs)+8)
	args = append(args, presetArgs...)
	args = append(args,
		"--no-playlist",
		"--restrict-filenames",
		"--no-overwrites",
		"--newline",
		"-o", filepath.Join(outputDir, "%(title)s.%(ext)s"),
		url,
	)
	return args
}

// parseDownloadProgress reads the integer percent out of a yt-dlp download line.
func parseDownloadProgress(line string) (int, bool) {
	if !strings.Contains(line, "[download]") {
		return 0, false
	}
	match := downloadProgressRe.FindStringSubmatch(line)
	if match == nil {
		return 0, false
	}
	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, false
	}
	pct := int(value)
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return pct, true
}

// moveFetchedFiles moves every finished file from srcDir into dstDir, skipping
// hidden files and in-flight temp artifacts. It returns how many files moved.
func moveFetchedFiles(srcDir, dstDir string) (int, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return 0, err
	}
	moved := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || isTempArtifact(name) {
			continue
		}
		if err := moveFile(filepath.Join(srcDir, name), unoccupiedPath(dstDir, name)); err != nil {
			return moved, err
		}
		moved++
	}
	return moved, nil
}

// unoccupiedPath returns where fileName can land inside dir without replacing
// what is already there: the name itself, or the name numbered " (2)", " (3)"…
func unoccupiedPath(dir string, fileName string) string {
	candidatePath := filepath.Join(dir, fileName)
	extension := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, extension)
	for copyNumber := 2; isPathOccupied(candidatePath); copyNumber++ {
		candidatePath = filepath.Join(dir, fmt.Sprintf("%s (%d)%s", baseName, copyNumber, extension))
	}
	return candidatePath
}

func isPathOccupied(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func isTempArtifact(name string) bool {
	lower := strings.ToLower(name)
	for _, ext := range tempArtifactExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// stagedCopySuffix marks a file still being copied into the library; it is one
// of tempArtifactExts, so a leftover is never taken for finished media.
const stagedCopySuffix = ".part"

// moveFile renames sourcePath to destinationPath. Across filesystems (rename
// returns EXDEV) it copies under a staging name and renames at the end, so the
// final name never holds a partial file.
func moveFile(sourcePath, destinationPath string) error {
	if err := os.Rename(sourcePath, destinationPath); err == nil {
		return nil
	}
	stagingPath := destinationPath + stagedCopySuffix
	if err := copyFileContent(sourcePath, stagingPath); err != nil {
		os.Remove(stagingPath)
		return err
	}
	if err := os.Rename(stagingPath, destinationPath); err != nil {
		os.Remove(stagingPath)
		return err
	}
	return os.Remove(sourcePath)
}

func copyFileContent(sourcePath, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		destinationFile.Close()
		return err
	}
	return destinationFile.Close()
}

// withoutServerPaths strips the given server directories from a message that
// is stored on the job and shown to clients.
func withoutServerPaths(message string, serverDirs ...string) string {
	for _, serverDir := range serverDirs {
		if serverDir == "" {
			continue
		}
		message = strings.ReplaceAll(message, serverDir+string(filepath.Separator), "")
		message = strings.ReplaceAll(message, serverDir, "")
	}
	return message
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}
